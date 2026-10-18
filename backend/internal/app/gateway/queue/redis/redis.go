package redisqueue

import (
	"context"
	"errors"
	"log/slog"
	"strings"
	"time"

	"github.com/redis/go-redis/v9"

	"github.com/leonardo-gmuller/squad-flow/backend/internal/app/domain/erring"
	"github.com/leonardo-gmuller/squad-flow/backend/internal/app/domain/usecase/report"
	"github.com/leonardo-gmuller/squad-flow/backend/internal/app/pkg/identifier"
)

// Queue owns a dedicated Redis stream and one consumer group for report generation.
type Queue struct {
	client  *redis.Client
	stream  string
	group   string
	minIdle time.Duration
}

var _ report.EventQueue = (*Queue)(nil)

func New(connectionURL, stream, group string, minIdle time.Duration) (*Queue, error) {
	options, err := redis.ParseURL(connectionURL)
	if err != nil {
		return nil, erring.Wrap(err, "redisqueue.New")
	}

	options.ContextTimeoutEnabled = true

	return &Queue{client: redis.NewClient(options), stream: stream, group: group, minIdle: minIdle}, nil
}
func (q *Queue) Close() error { return erring.Wrap(q.client.Close(), "redisqueue.Close") }
func (q *Queue) Publish(ctx context.Context, event report.QueueEvent) error {
	_, err := q.client.XAdd(ctx, &redis.XAddArgs{Stream: q.stream, Values: map[string]any{"event_id": event.ID, "job_id": event.JobID, "type": event.Type}}).Result()

	return erring.Wrap(err, "redisqueue.Publish")
}

func (q *Queue) Consume(ctx context.Context, handle func(context.Context, report.QueueEvent) error) error {
	err := q.client.XGroupCreateMkStream(ctx, q.stream, q.group, "0").Err()
	if err != nil && !strings.Contains(err.Error(), "BUSYGROUP") {
		return erring.Wrap(err, "redisqueue.group")
	}

	consumer, err := identifier.NewUUID()
	if err != nil {
		return erring.Wrap(err, "redisqueue.Consume")
	}

	cursor := "0-0"
	for ctx.Err() == nil {
		pending, next, err := q.client.XAutoClaim(ctx, &redis.XAutoClaimArgs{Stream: q.stream, Group: q.group, Consumer: consumer, MinIdle: q.minIdle, Start: cursor, Count: 1}).Result()
		if err != nil && !errors.Is(err, redis.Nil) {
			return erring.Wrap(err, "redisqueue.reclaim")
		}

		cursor = next

		for _, message := range pending {
			if err = q.handle(ctx, message, handle); err != nil {
				return erring.Wrap(err, "redisqueue.Consume")
			}
		}

		streams, err := q.client.XReadGroup(ctx, &redis.XReadGroupArgs{Group: q.group, Consumer: consumer, Streams: []string{q.stream, ">"}, Count: 1, Block: time.Second}).Result()
		if errors.Is(err, redis.Nil) {
			continue
		}

		if err != nil {
			return erring.Wrap(err, "redisqueue.read")
		}

		for _, stream := range streams {
			for _, message := range stream.Messages {
				if err = q.handle(ctx, message, handle); err != nil {
					return erring.Wrap(err, "redisqueue.Consume")
				}
			}
		}
	}

	return erring.Wrap(ctx.Err(), "redisqueue.context")
}

func (q *Queue) handle(ctx context.Context, message redis.XMessage, handle func(context.Context, report.QueueEvent) error) error {
	eventID, _ := message.Values["event_id"].(string)
	jobID, _ := message.Values["job_id"].(string)
	kind, _ := message.Values["type"].(string)

	if identifier.IsUUID(eventID) && identifier.IsUUID(jobID) && kind == report.EventRequested {
		if err := handle(ctx, report.QueueEvent{ID: eventID, JobID: jobID, Type: kind}); err != nil {
			return erring.Wrap(err, "redisqueue.Consume")
		}
	} else {
		slog.Warn("discarding malformed report queue message", "stream_id", message.ID)
	}
	// The database transition must commit before acknowledging. A crash before ACK safely replays.
	if err := q.client.XAck(ctx, q.stream, q.group, message.ID).Err(); err != nil {
		return erring.Wrap(err, "redisqueue.ack")
	}

	return erring.Wrap(q.client.XDel(ctx, q.stream, message.ID).Err(), "redisqueue.delete")
}
