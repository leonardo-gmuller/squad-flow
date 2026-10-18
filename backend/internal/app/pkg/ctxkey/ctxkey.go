// Package ctxkey centralizes typed values shared through request contexts.
package ctxkey

import (
	"context"

	"github.com/leonardo-gmuller/squad-flow/backend/internal/app/domain/entity"
)

type ctxKey int

const (
	keyAuthorizationHeader ctxKey = iota
	keyIdempotencyKey
	keyRequestID
	keyUser
	keyTransport
	keyAgent
)

func GetAuthorizationHeader(ctx context.Context) (string, bool) {
	value, ok := ctx.Value(keyAuthorizationHeader).(string)

	return value, ok
}

func PutAuthorizationHeader(ctx context.Context, value string) context.Context {
	return context.WithValue(ctx, keyAuthorizationHeader, value)
}

func GetIdempotencyKey(ctx context.Context) (string, bool) {
	value, ok := ctx.Value(keyIdempotencyKey).(string)

	return value, ok
}

func PutIdempotencyKey(ctx context.Context, value string) context.Context {
	return context.WithValue(ctx, keyIdempotencyKey, value)
}

func GetRequestID(ctx context.Context) (string, bool) {
	value, ok := ctx.Value(keyRequestID).(string)

	return value, ok
}

func PutRequestID(ctx context.Context, value string) context.Context {
	return context.WithValue(ctx, keyRequestID, value)
}

func GetUser(ctx context.Context) (entity.User, bool) {
	value, ok := ctx.Value(keyUser).(entity.User)

	return value, ok
}

func PutUser(ctx context.Context, value entity.User) context.Context {
	return context.WithValue(ctx, keyUser, value)
}

func GetTransport(ctx context.Context) (string, bool) {
	value, ok := ctx.Value(keyTransport).(string)

	return value, ok
}

func PutTransport(ctx context.Context, value string) context.Context {
	return context.WithValue(ctx, keyTransport, value)
}

func GetAgent(ctx context.Context) (string, bool) {
	value, ok := ctx.Value(keyAgent).(string)

	return value, ok
}

func PutAgent(ctx context.Context, value string) context.Context {
	return context.WithValue(ctx, keyAgent, value)
}
