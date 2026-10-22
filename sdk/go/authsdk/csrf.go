package authsdk

import "net/http"

func (c *Client) RequireCSRF(
	next http.Handler,
) http.Handler {
	return http.HandlerFunc(func(
		writer http.ResponseWriter,
		request *http.Request,
	) {
		switch request.Method {
		case http.MethodGet,
			http.MethodHead,
			http.MethodOptions:
			next.ServeHTTP(writer, request)
			return
		}

		session, ok := SessionFromContext(
			request.Context(),
		)
		if !ok {
			writeSDKJSON(
				writer,
				http.StatusUnauthorized,
				map[string]string{
					"error": "unauthenticated",
				},
			)
			return
		}

		submitted := request.Header.Get(
			"X-CSRF-Token",
		)

		if !constantTimeEqual(
			submitted,
			session.CSRFToken,
		) {
			writeSDKJSON(
				writer,
				http.StatusForbidden,
				map[string]string{
					"error": "invalid_csrf_token",
				},
			)
			return
		}

		next.ServeHTTP(writer, request)
	})
}
