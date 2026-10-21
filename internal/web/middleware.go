package web

import (
	"errors"
	"net/http"

	"github.com/go-chi/chi/v5/middleware"
	"github.com/timsgarage/tapu/internal/session"
)

func SecurityHeaders(next http.Handler) http.Handler {
	return http.HandlerFunc(func(
		writer http.ResponseWriter,
		request *http.Request,
	) {
		writer.Header().Set(
			"Content-Security-Policy",
			"default-src 'none'; style-src 'unsafe-inline'; form-action 'self'; frame-ancestors 'none'; base-uri 'none'",
		)
		writer.Header().Set("Referrer-Policy", "no-referrer")
		writer.Header().Set("X-Content-Type-Options", "nosniff")
		writer.Header().Set("X-Frame-Options", "DENY")
		writer.Header().Set(
			"Permissions-Policy",
			"camera=(), microphone=(), geolocation=()",
		)

		next.ServeHTTP(writer, request)
	})
}

func (h *Handler) RequireAuthentication(
	next http.Handler,
) http.Handler {
	return http.HandlerFunc(func(
		writer http.ResponseWriter,
		request *http.Request,
	) {
		cookie, err := request.Cookie(h.sessionCookieName)
		if err != nil {
			http.Redirect(
				writer,
				request,
				"/login",
				http.StatusSeeOther,
			)
			return
		}

		found, err := h.sessions.Authenticate(
			request.Context(),
			cookie.Value,
		)
		if errors.Is(err, session.ErrUnauthenticated) {
			h.clearSessionCookies(writer)

			http.Redirect(
				writer,
				request,
				"/login",
				http.StatusSeeOther,
			)
			return
		}

		if err != nil {
			h.logger.Error(
				"authenticate session",
				"error",
				err,
				"request_id",
				middleware.GetReqID(request.Context()),
			)

			http.Error(
				writer,
				"Internal server error",
				http.StatusInternalServerError,
			)
			return
		}

		ctx := session.WithAuthenticatedSession(
			request.Context(),
			found,
		)

		next.ServeHTTP(writer, request.WithContext(ctx))
	})
}
