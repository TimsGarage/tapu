package authsdk

import (
	"net/http"
	"strings"
)

func (c *Client) RequireAuthentication(
	next http.Handler,
) http.Handler {
	return http.HandlerFunc(func(
		writer http.ResponseWriter,
		request *http.Request,
	) {
		session, err := c.sessionFromRequest(request)
		if err != nil {
			c.clearSessionCookie(writer)

			writeSDKJSON(
				writer,
				http.StatusUnauthorized,
				map[string]string{
					"error": "unauthenticated",
					"login": "/auth/login",
				},
			)
			return
		}

		ctx := withSession(
			request.Context(),
			session,
		)

		next.ServeHTTP(
			writer,
			request.WithContext(ctx),
		)
	})
}

func (c *Client) RequireBrowserAuthentication(
	next http.Handler,
) http.Handler {
	return http.HandlerFunc(func(
		writer http.ResponseWriter,
		request *http.Request,
	) {
		session, err := c.sessionFromRequest(request)
		if err != nil {
			c.clearSessionCookie(writer)

			http.Redirect(
				writer,
				request,
				"/auth/login",
				http.StatusSeeOther,
			)
			return
		}

		ctx := withSession(
			request.Context(),
			session,
		)

		next.ServeHTTP(
			writer,
			request.WithContext(ctx),
		)
	})
}

func RequireScope(
	expectedScope string,
	next http.Handler,
) http.Handler {
	return http.HandlerFunc(func(
		writer http.ResponseWriter,
		request *http.Request,
	) {
		identity, ok := IdentityFromContext(
			request.Context(),
		)
		if !ok ||
			!scopeContains(identity.Scope, expectedScope) {
			writeSDKJSON(
				writer,
				http.StatusForbidden,
				map[string]string{
					"error": "insufficient_scope",
				},
			)
			return
		}

		next.ServeHTTP(writer, request)
	})
}

func scopeContains(
	scope string,
	expected string,
) bool {
	for _, value := range strings.Fields(scope) {
		if value == expected {
			return true
		}
	}

	return false
}
