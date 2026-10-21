package authsdk

import (
	"crypto/subtle"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"time"

	"github.com/coreos/go-oidc/v3/oidc"
	"golang.org/x/oauth2"
)

const maximumCallbackValueLength = 4096

func (c *Client) LoginHandler(
	writer http.ResponseWriter,
	request *http.Request,
) {
	transaction, err := c.createTransaction()
	if err != nil {
		http.Error(
			writer,
			"Unable to start authentication",
			http.StatusInternalServerError,
		)
		return
	}

	c.setTransactionCookie(
		writer,
		transaction.ID,
		transaction.ExpiresAt,
	)

	authorizationURL := c.oauthConfig.AuthCodeURL(
		transaction.State,
		oauth2.SetAuthURLParam(
			"audience",
			c.config.Audience,
		),
		oauth2.SetAuthURLParam(
			"nonce",
			transaction.Nonce,
		),
		oauth2.SetAuthURLParam(
			"code_challenge",
			pkceChallenge(transaction.PKCEVerifier),
		),
		oauth2.SetAuthURLParam(
			"code_challenge_method",
			"S256",
		),
	)

	http.Redirect(
		writer,
		request,
		authorizationURL,
		http.StatusFound,
	)
}

func (c *Client) CallbackHandler(
	writer http.ResponseWriter,
	request *http.Request,
) {
	if request.URL.Query().Get("error") != "" {
		c.clearTransactionCookie(writer)

		http.Error(
			writer,
			"Authentication was denied",
			http.StatusUnauthorized,
		)
		return
	}

	code := request.URL.Query().Get("code")
	state := request.URL.Query().Get("state")

	if code == "" ||
		state == "" ||
		len(code) > maximumCallbackValueLength ||
		len(state) > maximumCallbackValueLength {
		http.Error(
			writer,
			"Invalid authentication callback",
			http.StatusBadRequest,
		)
		return
	}

	transactionCookie, err := request.Cookie(
		c.config.TransactionCookieName,
	)
	if err != nil {
		http.Error(
			writer,
			"Authentication transaction is missing",
			http.StatusBadRequest,
		)
		return
	}

	transaction, err := c.consumeTransaction(
		transactionCookie.Value,
	)
	c.clearTransactionCookie(writer)

	if err != nil {
		http.Error(
			writer,
			"Authentication transaction is invalid or expired",
			http.StatusBadRequest,
		)
		return
	}

	if !constantTimeEqual(state, transaction.State) {
		http.Error(
			writer,
			"Authentication state does not match",
			http.StatusBadRequest,
		)
		return
	}

	exchangeContext := oidc.ClientContext(
		request.Context(),
		c.config.HTTPClient,
	)

	oauthToken, err := c.oauthConfig.Exchange(
		exchangeContext,
		code,
		oauth2.VerifierOption(
			transaction.PKCEVerifier,
		),
	)
	if err != nil {
		http.Error(
			writer,
			"Authentication code exchange failed",
			http.StatusUnauthorized,
		)
		return
	}

	rawIDToken, ok := oauthToken.Extra(
		"id_token",
	).(string)
	if !ok || rawIDToken == "" {
		http.Error(
			writer,
			"Identity token is missing",
			http.StatusUnauthorized,
		)
		return
	}

	idToken, err := c.idTokenVerifier.Verify(
		exchangeContext,
		rawIDToken,
	)
	if err != nil {
		http.Error(
			writer,
			"Identity token is invalid",
			http.StatusUnauthorized,
		)
		return
	}

	var identityClaims idTokenClaims
	if err := idToken.Claims(&identityClaims); err != nil {
		http.Error(
			writer,
			"Identity claims are invalid",
			http.StatusUnauthorized,
		)
		return
	}

	if identityClaims.Subject == "" ||
		!constantTimeEqual(
			identityClaims.Nonce,
			transaction.Nonce,
		) {
		http.Error(
			writer,
			"Identity token nonce does not match",
			http.StatusUnauthorized,
		)
		return
	}

	if oauthToken.AccessToken == "" {
		http.Error(
			writer,
			"Access token is missing",
			http.StatusUnauthorized,
		)
		return
	}

	accessToken, err := c.accessTokenVerifier.Verify(
		exchangeContext,
		oauthToken.AccessToken,
	)
	if err != nil {
		http.Error(
			writer,
			"Access token is invalid",
			http.StatusUnauthorized,
		)
		return
	}

	var accessClaims accessTokenClaims
	if err := accessToken.Claims(&accessClaims); err != nil {
		http.Error(
			writer,
			"Access-token claims are invalid",
			http.StatusUnauthorized,
		)
		return
	}

	if accessClaims.Subject == "" ||
		accessClaims.Subject != identityClaims.Subject ||
		accessClaims.TokenType != "access_token" ||
		accessClaims.ClientID != c.config.ClientID {
		http.Error(
			writer,
			"Access token has invalid claims",
			http.StatusUnauthorized,
		)
		return
	}

	sessionID, err := randomValue(32)
	if err != nil {
		http.Error(
			writer,
			"Unable to create application session",
			http.StatusInternalServerError,
		)
		return
	}

	csrfToken, err := randomValue(32)
	if err != nil {
		http.Error(
			writer,
			"Unable to create application session",
			http.StatusInternalServerError,
		)
		return
	}

	now := time.Now().UTC()
	expiresAt := now.Add(c.config.SessionTTL)

	if !oauthToken.Expiry.IsZero() &&
		oauthToken.Expiry.Before(expiresAt) {
		expiresAt = oauthToken.Expiry
	}

	if !accessToken.Expiry.IsZero() &&
		accessToken.Expiry.Before(expiresAt) {
		expiresAt = accessToken.Expiry
	}

	if !expiresAt.After(now) {
		http.Error(
			writer,
			"Access token has already expired",
			http.StatusUnauthorized,
		)
		return
	}

	username := accessClaims.Username
	if username == "" {
		username = identityClaims.PreferredUsername
	}

	session := Session{
		ID: sessionID,
		Identity: Identity{
			Subject:          identityClaims.Subject,
			Username:         username,
			ClientID:         c.config.ClientID,
			Audience:         c.config.Audience,
			Scope:            accessClaims.Scope,
			AuthenticatedAt:  now,
			SessionExpiresAt: expiresAt,
		},
		AccessToken: oauthToken.AccessToken,
		CSRFToken:   csrfToken,
		ExpiresAt:   expiresAt,
		CreatedAt:   now,
	}

	if err := c.config.Store.Put(
		request.Context(),
		session,
	); err != nil {
		http.Error(
			writer,
			"Unable to save application session",
			http.StatusInternalServerError,
		)
		return
	}

	c.setSessionCookie(
		writer,
		session.ID,
		session.ExpiresAt,
	)

	http.Redirect(
		writer,
		request,
		c.config.AfterLoginURL,
		http.StatusSeeOther,
	)
}

func (c *Client) CurrentUserHandler(
	writer http.ResponseWriter,
	request *http.Request,
) {
	session, err := c.sessionFromRequest(request)
	if err != nil {
		c.clearSessionCookie(writer)

		writeSDKJSON(
			writer,
			http.StatusUnauthorized,
			map[string]string{
				"error": "unauthenticated",
			},
		)
		return
	}

	writeSDKJSON(
		writer,
		http.StatusOK,
		map[string]any{
			"authenticated": true,
			"user":          session.Identity,
			"csrfToken":     session.CSRFToken,
		},
	)
}

func (c *Client) LogoutHandler(
	writer http.ResponseWriter,
	request *http.Request,
) {
	if request.Method != http.MethodPost {
		writer.Header().Set("Allow", http.MethodPost)

		http.Error(
			writer,
			"Method not allowed",
			http.StatusMethodNotAllowed,
		)
		return
	}

	session, err := c.sessionFromRequest(request)
	if err != nil {
		c.clearSessionCookie(writer)

		writeSDKJSON(
			writer,
			http.StatusUnauthorized,
			map[string]string{
				"error": "unauthenticated",
			},
		)
		return
	}

	submittedCSRF := request.Header.Get("X-CSRF-Token")
	if submittedCSRF == "" {
		submittedCSRF = request.FormValue("csrf_token")
	}

	if !constantTimeEqual(
		submittedCSRF,
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

	if err := c.config.Store.Delete(
		request.Context(),
		session.ID,
	); err != nil {
		http.Error(
			writer,
			"Unable to delete application session",
			http.StatusInternalServerError,
		)
		return
	}

	c.clearSessionCookie(writer)
	writer.WriteHeader(http.StatusNoContent)
}

func (c *Client) sessionFromRequest(
	request *http.Request,
) (*Session, error) {
	cookie, err := request.Cookie(
		c.config.SessionCookieName,
	)
	if err != nil {
		return nil, ErrSessionNotFound
	}

	if len(cookie.Value) > 256 {
		return nil, ErrSessionNotFound
	}

	session, err := c.config.Store.Get(
		request.Context(),
		cookie.Value,
	)
	if err != nil {
		if errors.Is(err, ErrSessionNotFound) {
			return nil, ErrSessionNotFound
		}

		return nil, fmt.Errorf(
			"load application session: %w",
			err,
		)
	}

	return session, nil
}

func (c *Client) setSessionCookie(
	writer http.ResponseWriter,
	value string,
	expiresAt time.Time,
) {
	http.SetCookie(writer, &http.Cookie{
		Name:     c.config.SessionCookieName,
		Value:    value,
		Path:     "/",
		Expires:  expiresAt,
		Secure:   c.config.SecureCookies,
		HttpOnly: true,
		SameSite: http.SameSiteLaxMode,
	})
}

func (c *Client) clearSessionCookie(
	writer http.ResponseWriter,
) {
	http.SetCookie(writer, &http.Cookie{
		Name:     c.config.SessionCookieName,
		Value:    "",
		Path:     "/",
		Expires:  time.Unix(1, 0),
		MaxAge:   -1,
		Secure:   c.config.SecureCookies,
		HttpOnly: true,
		SameSite: http.SameSiteLaxMode,
	})
}

func (c *Client) setTransactionCookie(
	writer http.ResponseWriter,
	value string,
	expiresAt time.Time,
) {
	http.SetCookie(writer, &http.Cookie{
		Name:     c.config.TransactionCookieName,
		Value:    value,
		Path:     callbackCookiePath(c.config.RedirectURL),
		Expires:  expiresAt,
		Secure:   c.config.SecureCookies,
		HttpOnly: true,
		SameSite: http.SameSiteLaxMode,
	})
}

func (c *Client) clearTransactionCookie(
	writer http.ResponseWriter,
) {
	http.SetCookie(writer, &http.Cookie{
		Name:     c.config.TransactionCookieName,
		Value:    "",
		Path:     callbackCookiePath(c.config.RedirectURL),
		Expires:  time.Unix(1, 0),
		MaxAge:   -1,
		Secure:   c.config.SecureCookies,
		HttpOnly: true,
		SameSite: http.SameSiteLaxMode,
	})
}

func callbackCookiePath(
	redirectURL string,
) string {
	parsed, err := url.Parse(redirectURL)
	if err != nil || parsed.Path == "" {
		return "/"
	}

	return parsed.Path
}

func constantTimeEqual(
	first string,
	second string,
) bool {
	if len(first) == 0 ||
		len(first) != len(second) {
		return false
	}

	return subtle.ConstantTimeCompare(
		[]byte(first),
		[]byte(second),
	) == 1
}

func writeSDKJSON(
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
	writer.WriteHeader(status)

	_ = json.NewEncoder(writer).Encode(value)
}
