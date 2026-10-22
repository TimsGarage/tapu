package authsdk

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestCORSAllowsConfiguredOrigin(t *testing.T) {
	client := &Client{
		config: Config{
			AllowedOrigins: []string{
				"https://app.example.com",
			},
		},
	}

	handler := client.CORS(
		http.HandlerFunc(func(
			writer http.ResponseWriter,
			_ *http.Request,
		) {
			writer.WriteHeader(http.StatusOK)
		}),
	)

	request := httptest.NewRequest(
		http.MethodGet,
		"https://api.example.com/api/auth/me",
		nil,
	)
	request.Header.Set(
		"Origin",
		"https://app.example.com",
	)

	response := httptest.NewRecorder()
	handler.ServeHTTP(response, request)

	if response.Code != http.StatusOK {
		t.Fatalf(
			"expected status 200, got %d",
			response.Code,
		)
	}

	if actual := response.Header().Get(
		"Access-Control-Allow-Origin",
	); actual != "https://app.example.com" {
		t.Fatalf(
			"unexpected allowed origin: %q",
			actual,
		)
	}

	if actual := response.Header().Get(
		"Access-Control-Allow-Credentials",
	); actual != "true" {
		t.Fatalf(
			"expected credentialed CORS, got %q",
			actual,
		)
	}
}

func TestCORSRejectsUnknownOrigin(t *testing.T) {
	client := &Client{
		config: Config{
			AllowedOrigins: []string{
				"https://app.example.com",
			},
		},
	}

	handler := client.CORS(
		http.HandlerFunc(func(
			writer http.ResponseWriter,
			_ *http.Request,
		) {
			writer.WriteHeader(http.StatusOK)
		}),
	)

	request := httptest.NewRequest(
		http.MethodGet,
		"https://api.example.com/api/auth/me",
		nil,
	)
	request.Header.Set(
		"Origin",
		"https://evil.example.net",
	)

	response := httptest.NewRecorder()
	handler.ServeHTTP(response, request)

	if response.Code != http.StatusForbidden {
		t.Fatalf(
			"expected status 403, got %d",
			response.Code,
		)
	}

	if response.Header().Get(
		"Access-Control-Allow-Origin",
	) != "" {
		t.Fatal(
			"rejected origin must not receive a CORS allow header",
		)
	}
}

func TestCORSHandlesPreflight(t *testing.T) {
	client := &Client{
		config: Config{
			AllowedOrigins: []string{
				"https://app.example.com",
			},
		},
	}

	handler := client.CORS(
		http.HandlerFunc(func(
			writer http.ResponseWriter,
			_ *http.Request,
		) {
			t.Fatal(
				"preflight must not reach the next handler",
			)
		}),
	)

	request := httptest.NewRequest(
		http.MethodOptions,
		"https://api.example.com/api/auth/logout",
		nil,
	)
	request.Header.Set(
		"Origin",
		"https://app.example.com",
	)
	request.Header.Set(
		"Access-Control-Request-Method",
		http.MethodPost,
	)
	request.Header.Set(
		"Access-Control-Request-Headers",
		"X-CSRF-Token",
	)

	response := httptest.NewRecorder()
	handler.ServeHTTP(response, request)

	if response.Code != http.StatusNoContent {
		t.Fatalf(
			"expected status 204, got %d",
			response.Code,
		)
	}
}
