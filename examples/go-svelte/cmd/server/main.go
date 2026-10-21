package main

import (
	"context"
	"encoding/json"
	"errors"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"

	"github.com/timsgarage/tapu/sdk/go/authsdk"
)

func main() {
	ctx, cancel := context.WithTimeout(
		context.Background(),
		15*time.Second,
	)

	authClient, err := authsdk.New(
		ctx,
		authsdk.Config{
			IssuerURL: getEnv(
				"IAM_ISSUER_URL",
				"http://127.0.0.1:8080",
			),
			ClientID: os.Getenv(
				"IAM_CLIENT_ID",
			),
			ClientSecret: os.Getenv(
				"IAM_CLIENT_SECRET",
			),
			RedirectURL: getEnv(
				"IAM_REDIRECT_URL",
				"http://127.0.0.1:3000/auth/callback",
			),
			Audience: getEnv(
				"IAM_AUDIENCE",
				"project-a-api",
			),
			AfterLoginURL: "/",
			SecureCookies: false,
		},
	)
	cancel()

	if err != nil {
		log.Fatalf("initialize authentication: %v", err)
	}

	router := chi.NewRouter()
	router.Use(middleware.RequestID)
	router.Use(middleware.Recoverer)

	router.Get("/auth/login", authClient.LoginHandler)
	router.Get(
		"/auth/callback",
		authClient.CallbackHandler,
	)
	router.Get(
		"/api/auth/me",
		authClient.CurrentUserHandler,
	)
	router.Post(
		"/auth/logout",
		authClient.LogoutHandler,
	)

	router.With(
		authClient.RequireAuthentication,
	).Get(
		"/api/private",
		privateHandler,
	)

	router.Get("/", indexHandler)

	server := &http.Server{
		Addr:              "127.0.0.1:3000",
		Handler:           router,
		ReadHeaderTimeout: 5 * time.Second,
		ReadTimeout:       15 * time.Second,
		WriteTimeout:      15 * time.Second,
		IdleTimeout:       60 * time.Second,
	}

	serverErrors := make(chan error, 1)

	go func() {
		log.Printf(
			"example application listening at http://127.0.0.1:3000",
		)

		serverErrors <- server.ListenAndServe()
	}()

	signals := make(chan os.Signal, 1)
	signal.Notify(
		signals,
		syscall.SIGINT,
		syscall.SIGTERM,
	)
	defer signal.Stop(signals)

	select {
	case received := <-signals:
		log.Printf(
			"shutdown signal received: %s",
			received,
		)

	case err := <-serverErrors:
		if !errors.Is(err, http.ErrServerClosed) {
			log.Fatalf("HTTP server failed: %v", err)
		}

		return
	}

	shutdownCtx, shutdownCancel := context.WithTimeout(
		context.Background(),
		10*time.Second,
	)
	defer shutdownCancel()

	if err := server.Shutdown(shutdownCtx); err != nil {
		log.Printf("graceful shutdown failed: %v", err)
	}
}

func indexHandler(
	writer http.ResponseWriter,
	_ *http.Request,
) {
	writer.Header().Set(
		"Content-Type",
		"text/html; charset=utf-8",
	)
	writer.Header().Set(
		"Cache-Control",
		"no-store",
	)

	_, _ = writer.Write([]byte(`<!doctype html>
<html lang="en">
<head>
  <meta charset="utf-8">
  <meta
    name="viewport"
    content="width=device-width, initial-scale=1"
  >
  <title>IAM SDK example</title>
</head>
<body>
  <main>
    <h1>IAM SDK example</h1>
    <p><a href="/auth/login">Sign in</a></p>
    <p><a href="/api/auth/me">Current user JSON</a></p>
    <p><a href="/api/private">Protected API JSON</a></p>
    <p><a href="/auth/logout">Sign out</a></p>
  </main>
</body>
</html>`))
}

func privateHandler(
	writer http.ResponseWriter,
	request *http.Request,
) {
	identity, ok := authsdk.IdentityFromContext(
		request.Context(),
	)
	if !ok {
		http.Error(
			writer,
			"Identity missing",
			http.StatusInternalServerError,
		)
		return
	}

	writer.Header().Set(
		"Content-Type",
		"application/json",
	)
	writer.Header().Set(
		"Cache-Control",
		"no-store",
	)

	_ = json.NewEncoder(writer).Encode(
		map[string]any{
			"message": "This endpoint is protected.",
			"user":    identity,
		},
	)
}

func getEnv(
	name string,
	fallback string,
) string {
	value := os.Getenv(name)
	if value == "" {
		return fallback
	}

	return value
}
