package realtime

import (
	"context"
	"encoding/json"
	"log/slog"
	"net/http"
	"sync"
	"time"

	"github.com/gorilla/websocket"

	"github.com/leonardo-gmuller/squad-flow/backend/internal/app/domain/entity"
	"github.com/leonardo-gmuller/squad-flow/backend/internal/app/domain/erring"
	"github.com/leonardo-gmuller/squad-flow/backend/internal/app/domain/usecase/report"
)

type Event struct {
	Type      string `json:"type"`
	RequestID string `json:"request_id,omitempty"`
	Data      any    `json:"data,omitempty"`
}
type peer struct {
	conn   *websocket.Conn
	userID string
	send   chan Event
}
type Hub struct {
	mu     sync.Mutex
	peers  map[*peer]bool
	closed bool
}

func New() *Hub { return &Hub{peers: map[*peer]bool{}} }
func (h *Hub) Publish(userID string, event Event) {
	h.mu.Lock()
	defer h.mu.Unlock()

	for p := range h.peers {
		if userID == "" || p.userID == userID {
			select {
			case p.send <- event:
			default:
				_ = p.conn.Close()
			}
		}
	}
}

func (h *Hub) Close() {
	h.mu.Lock()
	defer h.mu.Unlock()

	h.closed = true
	for p := range h.peers {
		_ = p.conn.Close()
	}
}

type Submit func(context.Context, entity.User, report.Request, string) (report.Job, error)

func (h *Hub) Serve(w http.ResponseWriter, r *http.Request, user entity.User, originAllowed func(string) bool, valid func(context.Context) bool, submit Submit) {
	upgrader := websocket.Upgrader{CheckOrigin: func(req *http.Request) bool { return originAllowed(req.Header.Get("Origin")) }, ReadBufferSize: 4096, WriteBufferSize: 4096}

	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		return
	}

	client := &peer{conn: conn, userID: user.ID, send: make(chan Event, 32)}

	h.mu.Lock()
	if h.closed {
		h.mu.Unlock()

		_ = conn.Close()

		return
	}

	h.peers[client] = true
	h.mu.Unlock()
	slog.Info("websocket connected")

	ctx, cancel := context.WithCancel(r.Context())

	defer func() {
		cancel()
		h.mu.Lock()
		delete(h.peers, client)
		h.mu.Unlock()

		_ = conn.Close()

		slog.Info("websocket disconnected")
	}()

	go client.write(ctx, valid)
	conn.SetReadLimit(8192)

	if err = conn.SetReadDeadline(time.Now().Add(60 * time.Second)); err != nil {
		return
	}

	conn.SetPongHandler(func(string) error {
		return erring.Wrap(conn.SetReadDeadline(time.Now().Add(60*time.Second)), "realtime.pong")
	})
	client.emit(ctx, Event{Type: "connected"})
	client.read(ctx, user, valid, submit)
}

func (p *peer) emit(ctx context.Context, event Event) {
	select {
	case p.send <- event:
	case <-ctx.Done():
	}
}

func (p *peer) write(ctx context.Context, valid func(context.Context) bool) {
	ticker := time.NewTicker(20 * time.Second)
	defer ticker.Stop()
	defer p.conn.Close()

	for {
		select {
		case <-ctx.Done():
			return
		case event := <-p.send:
			if !valid(ctx) {
				return
			}

			if err := p.conn.SetWriteDeadline(time.Now().Add(10 * time.Second)); err != nil {
				return
			}

			if err := p.conn.WriteJSON(event); err != nil {
				return
			}
		case <-ticker.C:
			if !valid(ctx) {
				return
			}

			if err := p.conn.WriteControl(websocket.PingMessage, nil, time.Now().Add(10*time.Second)); err != nil {
				return
			}
		}
	}
}

type incomingMessage struct {
	IdempotencyKey string         `json:"idempotency_key"`
	Type           string         `json:"type"`
	RequestID      string         `json:"request_id"`
	Request        report.Request `json:"request"`
}

func (p *peer) read(ctx context.Context, user entity.User, valid func(context.Context) bool, submit Submit) {
	for {
		_, raw, err := p.conn.ReadMessage()
		if err != nil || !valid(ctx) {
			return
		}

		var input incomingMessage
		if err = json.Unmarshal(raw, &input); err != nil || input.Type != "report.generate" || len(input.RequestID) > 80 || input.RequestID == "" {
			p.emit(ctx, Event{Type: "error", Data: "mensagem WebSocket inválida"})

			continue
		}

		job, err := submit(ctx, user, input.Request, input.IdempotencyKey)
		if err != nil {
			p.emit(ctx, Event{Type: "report.error", RequestID: input.RequestID, Data: err.Error()})

			continue
		}

		p.emit(ctx, Event{Type: "report.queued", RequestID: input.RequestID, Data: job})
	}
}
