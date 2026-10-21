package session

import (
	"context"

	"github.com/timsgarage/tapu/internal/identity"
)

type contextKey string

const (
	sessionContextKey contextKey = "iam-session"
	userContextKey    contextKey = "iam-user"
)

func WithAuthenticatedSession(
	ctx context.Context,
	found *Session,
) context.Context {
	ctx = context.WithValue(ctx, sessionContextKey, found)
	ctx = context.WithValue(ctx, userContextKey, &found.User)

	return ctx
}

func FromContext(ctx context.Context) (*Session, bool) {
	found, ok := ctx.Value(sessionContextKey).(*Session)
	return found, ok
}

func UserFromContext(ctx context.Context) (*identity.User, bool) {
	user, ok := ctx.Value(userContextKey).(*identity.User)
	return user, ok
}
