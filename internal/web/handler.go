package web

import (
	"crypto/rand"
	"crypto/subtle"
	"encoding/base64"
	"errors"
	"fmt"
	"html/template"
	"log/slog"
	"net"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/go-chi/chi/v5/middleware"
	"github.com/timsgarage/tapu/internal/audit"
	"github.com/timsgarage/tapu/internal/identity"
	"github.com/timsgarage/tapu/internal/ratelimit"
	"github.com/timsgarage/tapu/internal/session"
)

const (
	loginCSRFCookieName = "iam_login_csrf"
	maximumFormSize     = 16 * 1024
)

type Handler struct {
	logger            *slog.Logger
	templates         *template.Template
	authenticator     *identity.Authenticator
	sessions          *session.Manager
	audit             *audit.Repository
	loginLimiter      *ratelimit.Limiter
	sessionCookieName string
	csrfCookieName    string
	secureCookies     bool
}

type loginPageData struct {
	CSRFToken string
	Error     string
	ReturnTo  string
}

type accountPageData struct {
	Username  string
	IsAdmin   bool
	CSRFToken string
}

func NewHandler(
	logger *slog.Logger,
	authenticator *identity.Authenticator,
	sessions *session.Manager,
	auditRepository *audit.Repository,
	loginLimiter *ratelimit.Limiter,
	secureCookies bool,
) (*Handler, error) {
	templates, err := parseTemplates()
	if err != nil {
		return nil, fmt.Errorf("parse templates: %w", err)
	}

	return &Handler{
		logger:            logger,
		templates:         templates,
		authenticator:     authenticator,
		sessions:          sessions,
		audit:             auditRepository,
		loginLimiter:      loginLimiter,
		sessionCookieName: "iam_session",
		csrfCookieName:    "iam_csrf",
		secureCookies:     secureCookies,
	}, nil
}

func (h *Handler) Home(
	writer http.ResponseWriter,
	request *http.Request,
) {
	http.Redirect(
		writer,
		request,
		"/account",
		http.StatusSeeOther,
	)
}
func (h *Handler) LoginPage(
	writer http.ResponseWriter,
	request *http.Request,
) {
	returnTo := validLoginReturnTo(
		request.URL.Query().Get("return_to"),
	)

	if h.hasAuthenticatedSession(request) {
		if returnTo != "" {
			http.Redirect(
				writer,
				request,
				returnTo,
				http.StatusSeeOther,
			)
			return
		}

		http.Redirect(
			writer,
			request,
			"/account",
			http.StatusSeeOther,
		)
		return
	}

	csrfToken, err := generateBrowserToken()
	if err != nil {
		h.internalServerError(writer, request, err)
		return
	}

	h.setLoginCSRFCookie(writer, csrfToken)

	h.renderLogin(writer, request, loginPageData{
		CSRFToken: csrfToken,
		ReturnTo:  returnTo,
	})
}

func (h *Handler) Login(
	writer http.ResponseWriter,
	request *http.Request,
) {
	request.Body = http.MaxBytesReader(
		writer,
		request.Body,
		maximumFormSize,
	)

	if err := request.ParseForm(); err != nil {
		http.Error(
			writer,
			"Invalid request",
			http.StatusBadRequest,
		)
		return
	}

	returnTo := validLoginReturnTo(
		request.PostFormValue("return_to"),
	)

	csrfToken := request.PostFormValue("csrf_token")
	csrfCookie, err := request.Cookie(loginCSRFCookieName)

	if err != nil ||
		!constantTimeTokenEqual(csrfToken, csrfCookie.Value) {
		http.Error(
			writer,
			"Invalid CSRF token",
			http.StatusForbidden,
		)
		return
	}

	username := identity.NormalizeUsername(
		request.PostFormValue("username"),
	)
	password := request.PostFormValue("password")
	sourceIP := requestSourceIP(request)

	usernameKey := "username:" + username
	ipKey := "ip:" + sourceIP

	if !h.loginLimiter.Allow(usernameKey, ipKey) {
		h.recordRequestAudit(
			request,
			audit.NewEvent(audit.ActorSystem, "login.rate_limited"),
			map[string]any{
				"username": username,
			},
		)

		h.renderLogin(writer, request, loginPageData{
			CSRFToken: csrfToken,
			Error:     "Invalid username or password.",
			ReturnTo:  returnTo,
		})
		return
	}

	user, err := h.authenticator.Authenticate(
		request.Context(),
		username,
		password,
	)

	if errors.Is(err, identity.ErrInvalidCredentials) {
		h.recordRequestAudit(
			request,
			audit.NewEvent(audit.ActorSystem, "login.failed"),
			map[string]any{
				"username": username,
			},
		)

		h.renderLogin(writer, request, loginPageData{
			CSRFToken: csrfToken,
			Error:     "Invalid username or password.",
			ReturnTo:  returnTo,
		})
		return
	}

	if err != nil {
		h.internalServerError(writer, request, err)
		return
	}

	created, err := h.sessions.Create(
		request.Context(),
		user.ID,
	)
	if err != nil {
		h.internalServerError(writer, request, err)
		return
	}

	h.setSessionCookies(
		writer,
		created.SessionToken,
		created.CSRFToken,
		created.Session.AbsoluteExpiresAt,
	)
	h.clearLoginCSRFCookie(writer)

	h.loginLimiter.Reset(usernameKey, ipKey)

	event := audit.NewEvent(audit.ActorUser, "login.succeeded")
	event.ActorID = &user.ID

	h.recordRequestAudit(request, event, nil)

	destination := "/account"
	if returnTo != "" {
		destination = returnTo
	}

	http.Redirect(
		writer,
		request,
		destination,
		http.StatusSeeOther,
	)
}

func (h *Handler) Account(
	writer http.ResponseWriter,
	request *http.Request,
) {
	user, ok := session.UserFromContext(request.Context())
	if !ok {
		http.Error(
			writer,
			"Internal server error",
			http.StatusInternalServerError,
		)
		return
	}

	csrfCookie, err := request.Cookie(h.csrfCookieName)
	if err != nil {
		h.clearSessionCookies(writer)

		http.Redirect(
			writer,
			request,
			"/login",
			http.StatusSeeOther,
		)
		return
	}

	found, ok := session.FromContext(request.Context())
	if !ok || !h.sessions.ValidateCSRF(found, csrfCookie.Value) {
		h.clearSessionCookies(writer)

		http.Redirect(
			writer,
			request,
			"/login",
			http.StatusSeeOther,
		)
		return
	}

	writer.Header().Set(
		"Cache-Control",
		"no-store, max-age=0",
	)

	err = h.templates.ExecuteTemplate(
		writer,
		"account",
		accountPageData{
			Username:  user.Username,
			IsAdmin:   user.IsAdmin,
			CSRFToken: csrfCookie.Value,
		},
	)
	if err != nil {
		h.logger.Error(
			"render account page",
			"error",
			err,
		)
	}
}

func (h *Handler) Logout(
	writer http.ResponseWriter,
	request *http.Request,
) {
	request.Body = http.MaxBytesReader(
		writer,
		request.Body,
		maximumFormSize,
	)

	if err := request.ParseForm(); err != nil {
		http.Error(
			writer,
			"Invalid request",
			http.StatusBadRequest,
		)
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

	csrfCookie, cookieErr := request.Cookie(h.csrfCookieName)
	submittedCSRF := request.PostFormValue("csrf_token")

	if cookieErr != nil ||
		!constantTimeTokenEqual(submittedCSRF, csrfCookie.Value) ||
		!h.sessions.ValidateCSRF(found, submittedCSRF) {
		http.Error(
			writer,
			"Invalid CSRF token",
			http.StatusForbidden,
		)
		return
	}

	sessionCookie, err := request.Cookie(h.sessionCookieName)
	if err == nil {
		if err := h.sessions.Delete(
			request.Context(),
			sessionCookie.Value,
		); err != nil {
			h.internalServerError(writer, request, err)
			return
		}
	}

	event := audit.NewEvent(audit.ActorUser, "logout.succeeded")
	event.ActorID = &found.UserID

	h.recordRequestAudit(request, event, nil)
	h.clearSessionCookies(writer)

	http.Redirect(
		writer,
		request,
		"/login",
		http.StatusSeeOther,
	)
}

func (h *Handler) renderLogin(
	writer http.ResponseWriter,
	request *http.Request,
	data loginPageData,
) {
	writer.Header().Set(
		"Cache-Control",
		"no-store, max-age=0",
	)

	if data.Error != "" {
		writer.WriteHeader(http.StatusUnauthorized)
	}

	if err := h.templates.ExecuteTemplate(
		writer,
		"login",
		data,
	); err != nil {
		h.logger.Error(
			"render login page",
			"error",
			err,
			"request_id",
			middleware.GetReqID(request.Context()),
		)
	}
}

func (h *Handler) hasAuthenticatedSession(
	request *http.Request,
) bool {
	cookie, err := request.Cookie(h.sessionCookieName)
	if err != nil {
		return false
	}

	_, err = h.sessions.Authenticate(
		request.Context(),
		cookie.Value,
	)

	return err == nil
}

func (h *Handler) setSessionCookies(
	writer http.ResponseWriter,
	sessionToken string,
	csrfToken string,
	expiresAt time.Time,
) {
	http.SetCookie(writer, &http.Cookie{
		Name:     h.sessionCookieName,
		Value:    sessionToken,
		Path:     "/",
		Expires:  expiresAt,
		Secure:   h.secureCookies,
		HttpOnly: true,
		SameSite: http.SameSiteLaxMode,
	})

	http.SetCookie(writer, &http.Cookie{
		Name:     h.csrfCookieName,
		Value:    csrfToken,
		Path:     "/",
		Expires:  expiresAt,
		Secure:   h.secureCookies,
		HttpOnly: true,
		SameSite: http.SameSiteLaxMode,
	})
}

func (h *Handler) clearSessionCookies(
	writer http.ResponseWriter,
) {
	h.clearCookie(writer, h.sessionCookieName)
	h.clearCookie(writer, h.csrfCookieName)
}

func (h *Handler) setLoginCSRFCookie(
	writer http.ResponseWriter,
	token string,
) {
	http.SetCookie(writer, &http.Cookie{
		Name:     loginCSRFCookieName,
		Value:    token,
		Path:     "/login",
		MaxAge:   600,
		Secure:   h.secureCookies,
		HttpOnly: true,
		SameSite: http.SameSiteStrictMode,
	})
}

func (h *Handler) clearLoginCSRFCookie(
	writer http.ResponseWriter,
) {
	http.SetCookie(writer, &http.Cookie{
		Name:     loginCSRFCookieName,
		Value:    "",
		Path:     "/login",
		MaxAge:   -1,
		Expires:  time.Unix(1, 0),
		Secure:   h.secureCookies,
		HttpOnly: true,
		SameSite: http.SameSiteStrictMode,
	})
}

func (h *Handler) clearCookie(
	writer http.ResponseWriter,
	name string,
) {
	http.SetCookie(writer, &http.Cookie{
		Name:     name,
		Value:    "",
		Path:     "/",
		MaxAge:   -1,
		Expires:  time.Unix(1, 0),
		Secure:   h.secureCookies,
		HttpOnly: true,
		SameSite: http.SameSiteLaxMode,
	})
}

func (h *Handler) recordRequestAudit(
	request *http.Request,
	event audit.Event,
	metadata map[string]any,
) {
	requestID := middleware.GetReqID(request.Context())
	sourceIP := requestSourceIP(request)

	event.RequestID = &requestID
	event.SourceIP = &sourceIP

	if err := h.audit.Record(
		request.Context(),
		event,
		metadata,
	); err != nil {
		h.logger.Error(
			"record audit event",
			"error",
			err,
			"action",
			event.Action,
			"request_id",
			requestID,
		)
	}
}

func (h *Handler) internalServerError(
	writer http.ResponseWriter,
	request *http.Request,
	err error,
) {
	h.logger.Error(
		"request failed",
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
}

func generateBrowserToken() (string, error) {
	value := make([]byte, 32)

	if _, err := rand.Read(value); err != nil {
		return "", fmt.Errorf("generate browser token: %w", err)
	}

	return base64.RawURLEncoding.EncodeToString(value), nil
}

func constantTimeTokenEqual(first string, second string) bool {
	firstBytes, firstErr := base64.RawURLEncoding.Strict().
		DecodeString(first)
	secondBytes, secondErr := base64.RawURLEncoding.Strict().
		DecodeString(second)

	if firstErr != nil ||
		secondErr != nil ||
		len(firstBytes) != 32 ||
		len(secondBytes) != 32 {
		return false
	}

	return subtle.ConstantTimeCompare(firstBytes, secondBytes) == 1
}

func requestSourceIP(request *http.Request) string {
	host, _, err := net.SplitHostPort(request.RemoteAddr)
	if err == nil {
		return host
	}

	return strings.TrimSpace(request.RemoteAddr)
}

func validLoginReturnTo(value string) string {
	if value == "" || len(value) > 4096 {
		return ""
	}

	parsed, err := url.Parse(value)
	if err != nil {
		return ""
	}

	if parsed.IsAbs() ||
		parsed.Host != "" ||
		parsed.Path != "/oauth/authorize" ||
		parsed.Fragment != "" {
		return ""
	}

	return parsed.RequestURI()
}
