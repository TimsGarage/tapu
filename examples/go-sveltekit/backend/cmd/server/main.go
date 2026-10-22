package main

import (
	"context"
	"errors"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"

	"github.com/timsgarage/tapu/examples/go-sveltekit/backend/internal/todos"
	"github.com/timsgarage/tapu/sdk/go/authsdk"
)

func main() {
	frontendURL := getEnv(
		"APP_FRONTEND_URL",
		"http://127.0.0.1:5174",
	)

	initializationContext, cancel :=
		context.WithTimeout(
			context.Background(),
			15*time.Second,
		)

	authClient, err := authsdk.New(
		initializationContext,
		authsdk.Config{
			IssuerURL: getEnv(
				"IAM_ISSUER_URL",
				"http://127.0.0.1:8080",
			),
			ClientID: strings.TrimSpace(
				os.Getenv("IAM_CLIENT_ID"),
			),
			ClientSecret: strings.TrimSpace(
				os.Getenv("IAM_CLIENT_SECRET"),
			),
			RedirectURL: getEnv(
				"IAM_REDIRECT_URL",
				"http://127.0.0.1:3000/auth/callback",
			),
			Audience: getEnv(
				"IAM_AUDIENCE",
				"go-sveltekit-api",
			),
			AfterLoginURL: frontendURL,
			AllowedOrigins: []string{
				frontendURL,
			},
			SecureCookies: getEnv(
				"APP_SECURE_COOKIES",
				"false",
			) == "true",
		},
	)
	cancel()

	if err != nil {
		log.Fatalf(
			"initialize authentication: %v",
			err,
		)
	}

	todoHandler := todos.NewHandler(
		todos.NewStore(),
	)

	router := chi.NewRouter()
	router.Use(middleware.RequestID)
	router.Use(middleware.RealIP)
	router.Use(middleware.Recoverer)
	router.Use(middleware.Timeout(30 * time.Second))
	router.Use(authClient.CORS)

	router.Get(
		"/health/live",
		func(
			writer http.ResponseWriter,
			_ *http.Request,
		) {
			writer.Header().Set(
				"Content-Type",
				"application/json",
			)
			_, _ = writer.Write(
				[]byte(`{"status":"ok"}`),
			)
		},
	)

	router.Get(
		"/auth/login",
		authClient.LoginHandler,
	)
	router.Get(
		"/auth/callback",
		authClient.CallbackHandler,
	)

	router.Get(
		"/api/auth/me",
		authClient.CurrentUserHandler,
	)
	router.Post(
		"/api/auth/logout",
		authClient.LogoutHandler,
	)

	router.Route(
		"/api/todos",
		func(protected chi.Router) {
			protected.Use(
				authClient.RequireAuthentication,
			)

			protected.Get(
				"/",
				todoHandler.List,
			)

			protected.Group(
				func(mutations chi.Router) {
					mutations.Use(
						authClient.RequireCSRF,
					)

					mutations.Post(
						"/",
						todoHandler.Create,
					)
					mutations.Put(
						"/{todoID}",
						todoHandler.Update,
					)
					mutations.Delete(
						"/{todoID}",
						todoHandler.Delete,
					)
				},
			)
		},
	)

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
			"Go API listening at http://127.0.0.1:3000",
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
	case signalValue := <-signals:
		log.Printf(
			"received shutdown signal: %s",
			signalValue,
		)

	case err := <-serverErrors:
		if !errors.Is(
			err,
			http.ErrServerClosed,
		) {
			log.Fatalf(
				"HTTP server failed: %v",
				err,
			)
		}

		return
	}

	shutdownContext, shutdownCancel :=
		context.WithTimeout(
			context.Background(),
			10*time.Second,
		)
	defer shutdownCancel()

	if err := server.Shutdown(
		shutdownContext,
	); err != nil {
		log.Printf(
			"graceful shutdown failed: %v",
			err,
		)
	}
}

func getEnv(
	name string,
	fallback string,
) string {
	value := strings.TrimSpace(os.Getenv(name))
	if value == "" {
		return fallback
	}

	return value
}
