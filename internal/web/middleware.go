package web

import (
	"errors"
	"net/http"
	"net/url"

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
			"default-src 'none'; script-src 'self'; style-src 'self' 'unsafe-inline'; connect-src 'self'; img-src 'self' data:; form-action 'self'; frame-ancestors 'none'; base-uri 'none'",
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
			redirectToLogin(writer, request)
			return
		}

		found, err := h.sessions.Authenticate(
			request.Context(),
			cookie.Value,
		)
		if errors.Is(err, session.ErrUnauthenticated) {
			h.clearSessionCookies(writer)

			redirectToLogin(writer, request)
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

func (h *Handler) RequireAdmin(
	next http.Handler,
) http.Handler {
	return http.HandlerFunc(func(
		writer http.ResponseWriter,
		request *http.Request,
	) {
		user, ok := session.UserFromContext(request.Context())
		if !ok || !user.IsAdmin {
			http.Error(
				writer,
				"Forbidden",
				http.StatusForbidden,
			)
			return
		}

		next.ServeHTTP(writer, request)
	})
}

func (h *Handler) RequireAPICSRF(
	next http.Handler,
) http.Handler {
	return http.HandlerFunc(func(
		writer http.ResponseWriter,
		request *http.Request,
	) {
		if request.Method == http.MethodGet ||
			request.Method == http.MethodHead ||
			request.Method == http.MethodOptions {
			next.ServeHTTP(writer, request)
			return
		}

		found, ok := session.FromContext(request.Context())
		if !ok {
			http.Error(
				writer,
				"Unauthorized",
				http.StatusUnauthorized,
			)
			return
		}

		cookie, err := request.Cookie(h.csrfCookieName)
		submitted := request.Header.Get("X-CSRF-Token")

		if err != nil ||
			!constantTimeTokenEqual(submitted, cookie.Value) ||
			!h.sessions.ValidateCSRF(found, submitted) {
			http.Error(
				writer,
				"Invalid CSRF token",
				http.StatusForbidden,
			)
			return
		}

		next.ServeHTTP(writer, request)
	})
}

func redirectToLogin(
	writer http.ResponseWriter,
	request *http.Request,
) {
	destination := "/login"

	if request.URL.Path == "/oauth/authorize" {
		returnTo := request.URL.RequestURI()

		destination += "?return_to=" +
			url.QueryEscape(returnTo)
	}

	http.Redirect(
		writer,
		request,
		destination,
		http.StatusSeeOther,
	)
}
