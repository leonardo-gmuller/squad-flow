//go:build integration

package redisqueue

import (
	"context"
	"sync/atomic"
	"testing"
	"time"

	"github.com/redis/go-redis/v9"

	"github.com/leonardo-gmuller/squad-flow/backend/internal/app/domain/usecase/report"
	"github.com/leonardo-gmuller/squad-flow/backend/internal/app/pkg/identifier"
	"github.com/leonardo-gmuller/squad-flow/backend/internal/testinfra"
)

func TestRedisReclaimsWorkFromCrashedConsumer(t *testing.T) {
	t.Parallel()

	connectionURL := testinfra.Redis(t)

	id, _ := identifier.NewUUID()
	stream := "squadflow:test:reclaim:" + id

	queue, err := New(connectionURL, stream, "test-group", time.Millisecond*20)
	if err != nil {
		t.Fatal(err)
	}

	t.Cleanup(func() { _ = queue.client.Del(context.WithoutCancel(t.Context()), stream).Err(); _ = queue.Close() })

	if err = queue.client.XGroupCreateMkStream(t.Context(), stream, "test-group", "0").Err(); err != nil {
		t.Fatal(err)
	}

	event := report.QueueEvent{ID: id, JobID: id, Type: report.EventRequested}
	if err = queue.Publish(t.Context(), event); err != nil {
		t.Fatal(err)
	}

	if _, err = queue.client.XReadGroup(t.Context(), &redis.XReadGroupArgs{Group: "test-group", Consumer: "crashed-worker", Streams: []string{stream, ">"}, Count: 1}).Result(); err != nil {
		t.Fatal(err)
	}

	time.Sleep(30 * time.Millisecond)

	ctx, cancel := context.WithTimeout(t.Context(), 5*time.Second)
	defer cancel()

	var deliveries atomic.Int32

	done := make(chan error, 1)

	go func() {
		done <- queue.Consume(ctx, func(_ context.Context, received report.QueueEvent) error {
			if received != event {
				t.Errorf("corrupt queue event: %+v", received)
			}
			deliveries.Add(1)

			return nil
		})
	}()

	ticker := time.NewTicker(20 * time.Millisecond)
	defer ticker.Stop()

	for {
		length, lengthErr := queue.client.XLen(ctx, stream).Result()
		if lengthErr != nil {
			t.Fatal(lengthErr)
		}

		if deliveries.Load() == 1 && length == 0 {
			break
		}
		select {
		case <-ctx.Done():
			t.Fatal("pending message was not recovered and acknowledged")
		case <-ticker.C:
		}
	}

	cancel()
	select {
	case <-done:
	case <-time.After(2 * time.Second):
		t.Fatal("consumer failed to stop")
	}
}
