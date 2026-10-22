package authsdk

import (
	"net/http"
	"strconv"
	"strings"
)

func (c *Client) CORS(
	next http.Handler,
) http.Handler {
	allowedOrigins := make(map[string]struct{})

	for _, origin := range c.config.AllowedOrigins {
		allowedOrigins[origin] = struct{}{}
	}

	return http.HandlerFunc(func(
		writer http.ResponseWriter,
		request *http.Request,
	) {
		origin := strings.TrimSpace(
			request.Header.Get("Origin"),
		)

		// Requests such as top-level browser navigation generally have no
		// Origin header and do not require CORS handling.
		if origin == "" {
			next.ServeHTTP(writer, request)
			return
		}

		if _, allowed := allowedOrigins[origin]; !allowed {
			writeSDKJSON(
				writer,
				http.StatusForbidden,
				map[string]string{
					"error": "origin_not_allowed",
				},
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
			"Access-Control-Allow-Credentials",
			"true",
		)

		if request.Method == http.MethodOptions {
			writer.Header().Set(
				"Access-Control-Allow-Methods",
				"GET, POST, PUT, PATCH, DELETE, OPTIONS",
			)
			writer.Header().Set(
				"Access-Control-Allow-Headers",
				"Content-Type, X-CSRF-Token",
			)
			writer.Header().Set(
				"Access-Control-Max-Age",
				strconv.Itoa(10*60),
			)
			writer.WriteHeader(http.StatusNoContent)
			return
		}

		next.ServeHTTP(writer, request)
	})
}
