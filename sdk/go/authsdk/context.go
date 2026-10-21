package authsdk

import "context"

type contextKey string

const sessionContextKey contextKey = "authsdk-session"

func withSession(
	ctx context.Context,
	session *Session,
) context.Context {
	return context.WithValue(
		ctx,
		sessionContextKey,
		session,
	)
}

func SessionFromContext(
	ctx context.Context,
) (*Session, bool) {
	session, ok := ctx.Value(
		sessionContextKey,
	).(*Session)

	return session, ok
}

func IdentityFromContext(
	ctx context.Context,
) (*Identity, bool) {
	session, ok := SessionFromContext(ctx)
	if !ok {
		return nil, false
	}

	return &session.Identity, true
}

func AccessTokenFromContext(
	ctx context.Context,
) (string, bool) {
	session, ok := SessionFromContext(ctx)
	if !ok || session.AccessToken == "" {
		return "", false
	}

	return session.AccessToken, true
}
