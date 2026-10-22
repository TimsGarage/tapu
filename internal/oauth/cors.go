package oauth

import (
	"log/slog"
	"net/http"
	"strings"

	"github.com/go-chi/chi/v5/middleware"
)

func (h *Handler) PublicClientCORS(
	next http.Handler,
) http.Handler {
	return http.HandlerFunc(func(
		writer http.ResponseWriter,
		request *http.Request,
	) {
		origin := strings.TrimSpace(
			request.Header.Get("Origin"),
		)

		// Server-to-server requests generally do not contain Origin.
		if origin == "" {
			next.ServeHTTP(writer, request)
			return
		}

		allowed, err := h.repository.PublicClientOriginAllowed(
			request.Context(),
			origin,
		)
		if err != nil {
			h.logger.Error(
				"check OAuth CORS origin",
				slog.String("error", err.Error()),
				slog.String(
					"request_id",
					middleware.GetReqID(
						request.Context(),
					),
				),
			)

			writeOAuthError(
				writer,
				http.StatusInternalServerError,
				"server_error",
				"An internal error occurred.",
			)
			return
		}

		if !allowed {
			writeOAuthError(
				writer,
				http.StatusForbidden,
				"access_denied",
				"Origin is not allowed.",
			)
			return
		}

		writer.Header().Add("Vary", "Origin")
		writer.Header().Add(
			"Vary",
			"Access-Control-Request-Method",
		)
		writer.Header().Add(
			"Vary",
			"Access-Control-Request-Headers",
		)
		writer.Header().Set(
			"Access-Control-Allow-Origin",
			origin,
		)
		writer.Header().Set(
			"Access-Control-Allow-Methods",
			"GET, POST, OPTIONS",
		)
		writer.Header().Set(
			"Access-Control-Allow-Headers",
			"Authorization, Content-Type",
		)
		writer.Header().Set(
			"Access-Control-Max-Age",
			"600",
		)

		// OAuth bearer tokens are sent through Authorization headers.
		// IAM cookies are not needed by token or UserInfo requests.
		if request.Method == http.MethodOptions {
			writer.WriteHeader(http.StatusNoContent)
			return
		}

		next.ServeHTTP(writer, request)
	})
}
