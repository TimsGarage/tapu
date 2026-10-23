package main

import (
	"context"
	"crypto/ed25519"
	"errors"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/timsgarage/tapu/internal/admin"
	"github.com/timsgarage/tapu/internal/adminui"
	"github.com/timsgarage/tapu/internal/audit"
	"github.com/timsgarage/tapu/internal/config"
	"github.com/timsgarage/tapu/internal/database"
	"github.com/timsgarage/tapu/internal/health"
	"github.com/timsgarage/tapu/internal/identity"
	"github.com/timsgarage/tapu/internal/oauth"
	"github.com/timsgarage/tapu/internal/oidc"
	"github.com/timsgarage/tapu/internal/ratelimit"
	"github.com/timsgarage/tapu/internal/session"
	"github.com/timsgarage/tapu/internal/signing"
	"github.com/timsgarage/tapu/internal/token"
	"github.com/timsgarage/tapu/internal/web"
)

func main() {
	logger := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
		Level: slog.LevelInfo,
	}))

	if err := run(logger); err != nil {
		logger.Error("server stopped with an error", "error", err)
		os.Exit(1)
	}
}

func run(logger *slog.Logger) error {
	cfg, err := config.Load()
	if err != nil {
		return err
	}

	privateKey, err := signing.LoadPrivateKey(
		cfg.SigningPrivateKeyPath,
	)
	if err != nil {
		return err
	}

	publicKey, ok := privateKey.Public().(ed25519.PublicKey)
	if !ok {
		return errors.New("derive Ed25519 public key")
	}

	signingKeyID := signing.KeyID(publicKey)

	logger.Info(
		"signing key loaded",
		"key_id",
		signingKeyID,
	)

	startupCtx, startupCancel := context.WithTimeout(
		context.Background(),
		10*time.Second,
	)
	defer startupCancel()

	db, err := database.Open(startupCtx, cfg.DatabaseURL)
	if err != nil {
		return err
	}
	defer db.Close()

	passwords := identity.NewPasswordHasher()
	users := identity.NewRepository(db.GORM)

	authenticator, err := identity.NewAuthenticator(
		users,
		passwords,
	)
	if err != nil {
		return err
	}

	sessionManager := session.NewManager(
		db.GORM,
		cfg.SessionIdleTTL,
		cfg.SessionAbsoluteTTL,
	)

	auditRepository := audit.NewRepository(db.GORM)
	loginLimiter := ratelimit.New(5, 5*time.Minute)

	webHandler, err := web.NewHandler(
		logger,
		authenticator,
		sessionManager,
		auditRepository,
		loginLimiter,
		cfg.SecureCookies,
		cfg.UIOrigin,
	)
	if err != nil {
		return err
	}

	healthHandler := health.NewHandler(db.SQL)

	adminHandler := admin.NewHandler(
		logger,
		db.GORM,
		passwords,
		sessionManager,
		auditRepository,
	)

	applicationUIHandler, err := adminui.NewHandler()
	if err != nil {
		return err
	}

	oauthRepository := oauth.NewRepository(db.GORM)

	tokenSigner := token.NewSigner(
		cfg.IssuerURL,
		signingKeyID,
		privateKey,
		cfg.AccessTokenTTL,
		cfg.IDTokenTTL,
	)

	oauthHandler := oauth.NewHandler(
		logger,
		oauthRepository,
		tokenSigner,
		auditRepository,
		publicKey,
		cfg.IssuerURL,
		cfg.AuthorizationCodeTTL,
	)

	publicOIDCHandler := oidc.NewPublicHandler(
		cfg.IssuerURL,
		token.NewJSONWebKeySet(
			signingKeyID,
			publicKey,
		),
	)

	router := chi.NewRouter()
	router.Use(middleware.RequestID)
	router.Use(middleware.Recoverer)
	router.Use(middleware.CleanPath)
	router.Use(middleware.StripSlashes)
	router.Use(web.SecurityHeaders)

	router.Get("/health/live", healthHandler.Live)
	router.Get("/health/ready", healthHandler.Ready)

	router.Get(
		"/.well-known/openid-configuration",
		publicOIDCHandler.Discovery,
	)

	router.Get(
		"/.well-known/jwks.json",
		publicOIDCHandler.JWKS,
	)

	router.Handle("/_app/*", http.HandlerFunc(applicationUIHandler.Asset))
	router.Get("/login", applicationUIHandler.Shell)
	router.Post("/login", webHandler.Login)
	router.Get("/api/web/login", webHandler.LoginBootstrap)
	router.Post("/api/web/login", webHandler.Login)

	router.Route("/oauth", func(oauthRoutes chi.Router) {
		oauthRoutes.Use(oauthHandler.PublicClientCORS)

		oauthRoutes.Options("/token", func(
			writer http.ResponseWriter,
			_ *http.Request,
		) {
			writer.WriteHeader(http.StatusNoContent)
		})
		oauthRoutes.Post("/token", oauthHandler.Token)

		oauthRoutes.Options("/userinfo", func(
			writer http.ResponseWriter,
			_ *http.Request,
		) {
			writer.WriteHeader(http.StatusNoContent)
		})
		oauthRoutes.Get("/userinfo", oauthHandler.UserInfo)
		oauthRoutes.Post("/userinfo", oauthHandler.UserInfo)
	})

	router.Group(func(authorization chi.Router) {
		authorization.Use(webHandler.RequireAuthentication)
		authorization.Get(
			"/oauth/authorize",
			oauthHandler.Authorize,
		)
	})

	router.Group(func(protected chi.Router) {
		protected.Use(webHandler.RequireAuthentication)
		protected.Get("/account", applicationUIHandler.Shell)
		protected.Get("/api/web/account", webHandler.AccountBootstrap)
		protected.Post("/api/web/logout", webHandler.Logout)
	})

	router.Route("/api/admin", func(adminAPI chi.Router) {
		adminAPI.Use(webHandler.RequireAuthentication)
		adminAPI.Use(webHandler.RequireAdmin)
		adminAPI.Use(webHandler.RequireAPICSRF)

		adminAPI.Get("/session", adminHandler.CurrentSession)

		adminAPI.Get("/users", adminHandler.ListUsers)
		adminAPI.Post("/users", adminHandler.CreateUser)
		adminAPI.Patch("/users/{userID}", adminHandler.UpdateUser)
		adminAPI.Post(
			"/users/{userID}/password",
			adminHandler.ResetPassword,
		)

		adminAPI.Get(
			"/users/{userID}/services",
			adminHandler.ListUserServices,
		)
		adminAPI.Put(
			"/users/{userID}/services/{serviceID}",
			adminHandler.GrantUserService,
		)
		adminAPI.Delete(
			"/users/{userID}/services/{serviceID}",
			adminHandler.RevokeUserService,
		)

		adminAPI.Get("/services", adminHandler.ListServices)
		adminAPI.Post("/services", adminHandler.CreateService)
		adminAPI.Patch(
			"/services/{serviceID}",
			adminHandler.UpdateService,
		)

		adminAPI.Get("/clients", adminHandler.ListClients)
		adminAPI.Post("/clients", adminHandler.CreateClient)
		adminAPI.Patch(
			"/clients/{clientID}",
			adminHandler.UpdateClient,
		)
		adminAPI.Post(
			"/clients/{clientID}/rotate-secret",
			adminHandler.RotateClientSecret,
		)
	})

	router.Group(func(adminUI chi.Router) {
		adminUI.Use(webHandler.RequireAuthentication)
		adminUI.Use(webHandler.RequireAdmin)

		adminUI.Get("/admin", applicationUIHandler.Shell)
		adminUI.Get("/admin/*", applicationUIHandler.Shell)
	})

	server := &http.Server{
		Addr:              cfg.HTTPAddress,
		Handler:           router,
		ReadHeaderTimeout: 5 * time.Second,
		ReadTimeout:       15 * time.Second,
		WriteTimeout:      15 * time.Second,
		IdleTimeout:       60 * time.Second,
		MaxHeaderBytes:    1 << 20,
	}

	serverErrors := make(chan error, 1)

	go func() {
		logger.Info(
			"IAM server starting",
			"address",
			cfg.HTTPAddress,
			"environment",
			cfg.Environment,
			"secure_cookies",
			cfg.SecureCookies,
		)

		serverErrors <- server.ListenAndServe()
	}()

	shutdownSignals := make(chan os.Signal, 1)
	signal.Notify(
		shutdownSignals,
		syscall.SIGINT,
		syscall.SIGTERM,
	)
	defer signal.Stop(shutdownSignals)

	select {
	case receivedSignal := <-shutdownSignals:
		logger.Info(
			"shutdown signal received",
			"signal",
			receivedSignal.String(),
		)

	case serverErr := <-serverErrors:
		if !errors.Is(serverErr, http.ErrServerClosed) {
			return serverErr
		}

		return nil
	}

	shutdownCtx, shutdownCancel := context.WithTimeout(
		context.Background(),
		10*time.Second,
	)
	defer shutdownCancel()

	if err := server.Shutdown(shutdownCtx); err != nil {
		_ = server.Close()
		return err
	}

	logger.Info("IAM server stopped")
	return nil
}
