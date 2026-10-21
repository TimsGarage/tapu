package oauth

import (
	"crypto/ed25519"
	"crypto/subtle"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/go-chi/chi/v5/middleware"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"

	"github.com/timsgarage/tapu/internal/admin"
	"github.com/timsgarage/tapu/internal/audit"
	"github.com/timsgarage/tapu/internal/clientsecret"
	"github.com/timsgarage/tapu/internal/identity"
	"github.com/timsgarage/tapu/internal/session"
	"github.com/timsgarage/tapu/internal/token"
)

const (
	maximumTokenRequestSize = 32 * 1024
	maximumOAuthValueLength = 4096
)

type Handler struct {
	logger               *slog.Logger
	repository           *Repository
	signer               *token.Signer
	audit                *audit.Repository
	publicKey            ed25519.PublicKey
	issuerURL            string
	authorizationCodeTTL time.Duration
}

func NewHandler(
	logger *slog.Logger,
	repository *Repository,
	signer *token.Signer,
	auditRepository *audit.Repository,
	publicKey ed25519.PublicKey,
	issuerURL string,
	authorizationCodeTTL time.Duration,
) *Handler {
	return &Handler{
		logger:               logger,
		repository:           repository,
		signer:               signer,
		audit:                auditRepository,
		publicKey:            publicKey,
		issuerURL:            issuerURL,
		authorizationCodeTTL: authorizationCodeTTL,
	}
}

func (h *Handler) Authorize(
	writer http.ResponseWriter,
	request *http.Request,
) {
	user, ok := session.UserFromContext(request.Context())
	if !ok {
		http.Error(
			writer,
			"Unauthorized",
			http.StatusUnauthorized,
		)
		return
	}

	query := request.URL.Query()

	clientID := query.Get("client_id")
	redirectURI := query.Get("redirect_uri")
	responseType := query.Get("response_type")
	rawScope := query.Get("scope")
	state := query.Get("state")
	nonce := query.Get("nonce")
	audience := query.Get("audience")
	codeChallenge := query.Get("code_challenge")
	codeChallengeMethod := query.Get("code_challenge_method")

	if anyValueTooLong(
		clientID,
		redirectURI,
		rawScope,
		state,
		nonce,
		audience,
		codeChallenge,
	) {
		writeOAuthError(
			writer,
			http.StatusBadRequest,
			"invalid_request",
			"One or more request values are too long.",
		)
		return
	}

	client, err := h.repository.FindEnabledClientByClientID(
		request.Context(),
		clientID,
	)
	if errors.Is(err, ErrClientNotFound) {
		writeOAuthError(
			writer,
			http.StatusBadRequest,
			"invalid_request",
			"Unknown client.",
		)
		return
	}

	if err != nil {
		h.internalError(writer, request, err)
		return
	}

	redirectAllowed, err := h.repository.RedirectURIAllowed(
		request.Context(),
		client.ID,
		redirectURI,
	)
	if err != nil {
		h.internalError(writer, request, err)
		return
	}

	if !redirectAllowed {
		writeOAuthError(
			writer,
			http.StatusBadRequest,
			"invalid_request",
			"Invalid redirect URI.",
		)
		return
	}

	if responseType != "code" {
		redirectOAuthError(
			writer,
			request,
			redirectURI,
			state,
			"unsupported_response_type",
			"Only the code response type is supported.",
		)
		return
	}

	scope, validScope := NormalizeScope(rawScope)
	if !validScope {
		redirectOAuthError(
			writer,
			request,
			redirectURI,
			state,
			"invalid_scope",
			"Unsupported or missing scope.",
		)
		return
	}

	if audience == "" {
		redirectOAuthError(
			writer,
			request,
			redirectURI,
			state,
			"invalid_request",
			"An audience is required.",
		)
		return
	}

	if codeChallengeMethod != PKCEMethodS256 ||
		!ValidPKCEChallenge(codeChallenge) {
		redirectOAuthError(
			writer,
			request,
			redirectURI,
			state,
			"invalid_request",
			"Valid S256 PKCE parameters are required.",
		)
		return
	}

	service, err := h.repository.FindEnabledServiceByAudience(
		request.Context(),
		audience,
	)
	if errors.Is(err, ErrServiceNotFound) {
		redirectOAuthError(
			writer,
			request,
			redirectURI,
			state,
			"invalid_target",
			"Unknown or disabled audience.",
		)
		return
	}

	if err != nil {
		h.internalError(writer, request, err)
		return
	}

	clientAllowed, err := h.repository.ClientCanAccessService(
		request.Context(),
		client.ID,
		service.ID,
	)
	if err != nil {
		h.internalError(writer, request, err)
		return
	}

	if !clientAllowed {
		redirectOAuthError(
			writer,
			request,
			redirectURI,
			state,
			"access_denied",
			"Client does not have access to the service.",
		)
		return
	}

	userAllowed, err := h.repository.UserCanAccessService(
		request.Context(),
		user.ID,
		service.ID,
	)
	if err != nil {
		h.internalError(writer, request, err)
		return
	}

	if !userAllowed {
		redirectOAuthError(
			writer,
			request,
			redirectURI,
			state,
			"access_denied",
			"User does not have access to the service.",
		)
		return
	}

	rawCode, codeHash, err := GenerateAuthorizationCode()
	if err != nil {
		h.internalError(writer, request, err)
		return
	}

	now := time.Now().UTC()

	var storedNonce *string
	if nonce != "" {
		storedNonce = &nonce
	}

	authorizationCode := AuthorizationCode{
		ID:            uuid.New(),
		CodeHash:      codeHash,
		UserID:        user.ID,
		OAuthClientID: client.ID,
		ServiceID:     service.ID,
		RedirectURI:   redirectURI,
		Scope:         scope,
		Nonce:         storedNonce,
		PKCEChallenge: codeChallenge,
		PKCEMethod:    PKCEMethodS256,
		CreatedAt:     now,
		ExpiresAt:     now.Add(h.authorizationCodeTTL),
	}

	if err := h.repository.CreateAuthorizationCode(
		request.Context(),
		&authorizationCode,
	); err != nil {
		h.internalError(writer, request, err)
		return
	}

	h.recordAudit(
		request,
		user.ID,
		"authorization_code.issued",
		"oauth_client",
		client.ID,
		map[string]any{
			"clientId": client.ClientID,
			"audience": service.Audience,
		},
	)

	destination, err := authorizationSuccessRedirect(
		redirectURI,
		rawCode,
		state,
	)
	if err != nil {
		h.internalError(writer, request, err)
		return
	}

	http.Redirect(
		writer,
		request,
		destination,
		http.StatusFound,
	)
}

func (h *Handler) Token(
	writer http.ResponseWriter,
	request *http.Request,
) {
	request.Body = http.MaxBytesReader(
		writer,
		request.Body,
		maximumTokenRequestSize,
	)

	if err := request.ParseForm(); err != nil {
		writeOAuthError(
			writer,
			http.StatusBadRequest,
			"invalid_request",
			"Invalid token request.",
		)
		return
	}

	if request.PostFormValue("grant_type") !=
		"authorization_code" {
		writeOAuthError(
			writer,
			http.StatusBadRequest,
			"unsupported_grant_type",
			"Only authorization_code is supported.",
		)
		return
	}

	client, authError := h.authenticateTokenClient(request)
	if authError != nil {
		writer.Header().Set(
			"WWW-Authenticate",
			`Basic realm="oauth-token"`,
		)

		writeOAuthError(
			writer,
			http.StatusUnauthorized,
			"invalid_client",
			"Client authentication failed.",
		)
		return
	}

	rawCode := request.PostFormValue("code")
	redirectURI := request.PostFormValue("redirect_uri")
	codeVerifier := request.PostFormValue("code_verifier")

	if anyValueTooLong(
		rawCode,
		redirectURI,
		codeVerifier,
	) {
		writeOAuthError(
			writer,
			http.StatusBadRequest,
			"invalid_grant",
			"Authorization grant is invalid.",
		)
		return
	}

	consumed, err := h.repository.ConsumeAuthorizationCode(
		request.Context(),
		ConsumeAuthorizationCodeInput{
			RawCode:       rawCode,
			OAuthClientID: client.ID,
			RedirectURI:   redirectURI,
			PKCEVerifier:  codeVerifier,
		},
	)
	if errors.Is(err, ErrInvalidGrant) {
		writeOAuthError(
			writer,
			http.StatusBadRequest,
			"invalid_grant",
			"Authorization grant is invalid.",
		)
		return
	}

	if err != nil {
		h.internalError(writer, request, err)
		return
	}

	nonce := ""
	if consumed.Code.Nonce != nil {
		nonce = *consumed.Code.Nonce
	}

	tokenInput := token.UserTokenInput{
		UserID:          consumed.User.ID,
		Username:        consumed.User.Username,
		ClientID:        consumed.Client.ClientID,
		ServiceAudience: consumed.Service.Audience,
		Scope:           consumed.Code.Scope,
		Nonce:           nonce,
	}

	accessToken, accessExpiration, err :=
		h.signer.SignAccessToken(tokenInput)
	if err != nil {
		h.internalError(writer, request, err)
		return
	}

	idToken, _, err := h.signer.SignIDToken(tokenInput)
	if err != nil {
		h.internalError(writer, request, err)
		return
	}

	h.recordAudit(
		request,
		consumed.User.ID,
		"token.issued",
		"oauth_client",
		consumed.Client.ID,
		map[string]any{
			"clientId": consumed.Client.ClientID,
			"audience": consumed.Service.Audience,
		},
	)

	expiresIn := int64(
		time.Until(accessExpiration).Round(time.Second).Seconds(),
	)
	if expiresIn < 0 {
		expiresIn = 0
	}

	writeTokenJSON(
		writer,
		http.StatusOK,
		map[string]any{
			"access_token": accessToken,
			"token_type":   "Bearer",
			"expires_in":   expiresIn,
			"id_token":     idToken,
			"scope":        consumed.Code.Scope,
		},
	)
}

func (h *Handler) UserInfo(
	writer http.ResponseWriter,
	request *http.Request,
) {
	rawToken, ok := bearerToken(request)
	if !ok {
		writeBearerError(
			writer,
			"invalid_token",
			"Bearer access token is required.",
		)
		return
	}

	claims := &token.AccessTokenClaims{}

	parsed, err := jwt.ParseWithClaims(
		rawToken,
		claims,
		func(parsedToken *jwt.Token) (any, error) {
			return h.publicKey, nil
		},
		jwt.WithValidMethods([]string{"EdDSA"}),
		jwt.WithIssuer(h.issuerURL),
		jwt.WithExpirationRequired(),
	)

	if err != nil || !parsed.Valid ||
		claims.TokenType != token.TypeAccessToken {
		writeBearerError(
			writer,
			"invalid_token",
			"Access token is invalid.",
		)
		return
	}

	userID, err := uuid.Parse(claims.Subject)
	if err != nil {
		writeBearerError(
			writer,
			"invalid_token",
			"Access token is invalid.",
		)
		return
	}

	user, err := h.repository.FindActiveUserByID(
		request.Context(),
		userID,
	)
	if errors.Is(err, identity.ErrUserNotFound) {
		writeBearerError(
			writer,
			"invalid_token",
			"Access token is invalid.",
		)
		return
	}

	if err != nil {
		h.internalError(writer, request, err)
		return
	}

	response := map[string]any{
		"sub": user.ID.String(),
	}

	if ScopeContains(claims.Scope, ScopeProfile) {
		response["preferred_username"] = user.Username
	}

	writeTokenJSON(writer, http.StatusOK, response)
}

func (h *Handler) authenticateTokenClient(
	request *http.Request,
) (*admin.OAuthClient, error) {
	basicClientID, basicSecret, hasBasic :=
		request.BasicAuth()

	formClientID := request.PostFormValue("client_id")

	clientID := formClientID
	if hasBasic {
		clientID = basicClientID
	}

	if clientID == "" {
		return nil, errors.New("missing client ID")
	}

	client, err := h.repository.FindEnabledClientByClientID(
		request.Context(),
		clientID,
	)
	if err != nil {
		return nil, err
	}

	switch client.ClientType {
	case admin.ClientTypeConfidential:
		if !hasBasic ||
			basicClientID != client.ClientID ||
			len(basicSecret) > 1024 ||
			len(client.SecretHash) != 32 {
			return nil, errors.New(
				"invalid confidential client authentication",
			)
		}

		providedHash := clientsecret.Hash(basicSecret)

		if subtle.ConstantTimeCompare(
			providedHash,
			client.SecretHash,
		) != 1 {
			return nil, errors.New(
				"invalid confidential client secret",
			)
		}

	case admin.ClientTypePublic:
		if hasBasic ||
			formClientID != client.ClientID ||
			len(client.SecretHash) != 0 {
			return nil, errors.New(
				"invalid public client authentication",
			)
		}

	default:
		return nil, errors.New("unknown client type")
	}

	return client, nil
}

func authorizationSuccessRedirect(
	redirectURI string,
	code string,
	state string,
) (string, error) {
	parsed, err := url.Parse(redirectURI)
	if err != nil {
		return "", err
	}

	query := parsed.Query()
	query.Set("code", code)

	if state != "" {
		query.Set("state", state)
	}

	parsed.RawQuery = query.Encode()
	return parsed.String(), nil
}

func redirectOAuthError(
	writer http.ResponseWriter,
	request *http.Request,
	redirectURI string,
	state string,
	code string,
	description string,
) {
	parsed, err := url.Parse(redirectURI)
	if err != nil {
		writeOAuthError(
			writer,
			http.StatusBadRequest,
			code,
			description,
		)
		return
	}

	query := parsed.Query()
	query.Set("error", code)
	query.Set("error_description", description)

	if state != "" {
		query.Set("state", state)
	}

	parsed.RawQuery = query.Encode()

	http.Redirect(
		writer,
		request,
		parsed.String(),
		http.StatusFound,
	)
}

func writeOAuthError(
	writer http.ResponseWriter,
	status int,
	code string,
	description string,
) {
	writeTokenJSON(
		writer,
		status,
		map[string]string{
			"error":             code,
			"error_description": description,
		},
	)
}

func writeBearerError(
	writer http.ResponseWriter,
	code string,
	description string,
) {
	writer.Header().Set(
		"WWW-Authenticate",
		fmt.Sprintf(
			`Bearer error=%q, error_description=%q`,
			code,
			description,
		),
	)

	writeOAuthError(
		writer,
		http.StatusUnauthorized,
		code,
		description,
	)
}

func writeTokenJSON(
	writer http.ResponseWriter,
	status int,
	value any,
) {
	writer.Header().Set(
		"Content-Type",
		"application/json",
	)
	writer.Header().Set(
		"Cache-Control",
		"no-store",
	)
	writer.Header().Set("Pragma", "no-cache")
	writer.WriteHeader(status)

	_ = json.NewEncoder(writer).Encode(value)
}

func bearerToken(
	request *http.Request,
) (string, bool) {
	authorization := request.Header.Get("Authorization")
	parts := strings.Fields(authorization)

	if len(parts) != 2 ||
		!strings.EqualFold(parts[0], "Bearer") ||
		parts[1] == "" {
		return "", false
	}

	return parts[1], true
}

func anyValueTooLong(values ...string) bool {
	for _, value := range values {
		if len(value) > maximumOAuthValueLength {
			return true
		}
	}

	return false
}

func (h *Handler) recordAudit(
	request *http.Request,
	actorID uuid.UUID,
	action string,
	targetType string,
	targetID uuid.UUID,
	metadata map[string]any,
) {
	event := audit.NewEvent(audit.ActorUser, action)
	event.ActorID = &actorID
	event.TargetType = &targetType
	event.TargetID = &targetID

	requestID := middleware.GetReqID(request.Context())
	event.RequestID = &requestID

	if err := h.audit.Record(
		request.Context(),
		event,
		metadata,
	); err != nil {
		h.logger.Error(
			"record OAuth audit event",
			"error",
			err,
			"action",
			action,
		)
	}
}

func (h *Handler) internalError(
	writer http.ResponseWriter,
	request *http.Request,
	err error,
) {
	h.logger.Error(
		"OAuth request failed",
		"error",
		err,
		"request_id",
		middleware.GetReqID(request.Context()),
	)

	writeOAuthError(
		writer,
		http.StatusInternalServerError,
		"server_error",
		"An internal error occurred.",
	)
}
