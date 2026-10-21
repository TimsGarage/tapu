package authsdk

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"time"

	"github.com/coreos/go-oidc/v3/oidc"
	"golang.org/x/oauth2"
)

type Client struct {
	config Config

	oauthConfig         oauth2.Config
	idTokenVerifier     *oidc.IDTokenVerifier
	accessTokenVerifier *oidc.IDTokenVerifier

	transactionMutex sync.Mutex
	transactions     map[string]authorizationTransaction
}

type authorizationTransaction struct {
	ID           string
	State        string
	Nonce        string
	PKCEVerifier string
	ExpiresAt    time.Time
}

type idTokenClaims struct {
	Subject           string `json:"sub"`
	Nonce             string `json:"nonce"`
	PreferredUsername string `json:"preferred_username"`
}

type accessTokenClaims struct {
	Subject   string `json:"sub"`
	TokenType string `json:"token_type"`
	ClientID  string `json:"client_id"`
	Username  string `json:"username"`
	Scope     string `json:"scope"`
}

func New(
	ctx context.Context,
	config Config,
) (*Client, error) {
	config.applyDefaults()

	if err := config.validate(); err != nil {
		return nil, err
	}

	providerContext := oidc.ClientContext(
		ctx,
		config.HTTPClient,
	)

	provider, err := oidc.NewProvider(
		providerContext,
		config.IssuerURL,
	)
	if err != nil {
		return nil, fmt.Errorf(
			"discover OpenID Connect provider: %w",
			err,
		)
	}

	endpoint := provider.Endpoint()

	if config.PublicClient {
		endpoint.AuthStyle = oauth2.AuthStyleInParams
	} else {
		endpoint.AuthStyle = oauth2.AuthStyleInHeader
	}

	client := &Client{
		config: config,
		oauthConfig: oauth2.Config{
			ClientID:     config.ClientID,
			ClientSecret: config.ClientSecret,
			Endpoint:     endpoint,
			RedirectURL:  config.RedirectURL,
			Scopes: []string{
				oidc.ScopeOpenID,
				"profile",
			},
		},
		idTokenVerifier: provider.Verifier(
			&oidc.Config{
				ClientID: config.ClientID,
			},
		),
		accessTokenVerifier: provider.Verifier(
			&oidc.Config{
				ClientID: config.Audience,
			},
		),
		transactions: make(
			map[string]authorizationTransaction,
		),
	}

	return client, nil
}

func (c *Client) deleteExpiredTransactionsLocked(
	now time.Time,
) {
	for id, transaction := range c.transactions {
		if !transaction.ExpiresAt.After(now) {
			delete(c.transactions, id)
		}
	}
}

func (c *Client) createTransaction() (
	authorizationTransaction,
	error,
) {
	transactionID, err := randomValue(32)
	if err != nil {
		return authorizationTransaction{}, err
	}

	state, err := randomValue(32)
	if err != nil {
		return authorizationTransaction{}, err
	}

	nonce, err := randomValue(32)
	if err != nil {
		return authorizationTransaction{}, err
	}

	verifier, err := randomValue(32)
	if err != nil {
		return authorizationTransaction{}, err
	}

	transaction := authorizationTransaction{
		ID:           transactionID,
		State:        state,
		Nonce:        nonce,
		PKCEVerifier: verifier,
		ExpiresAt: time.Now().UTC().Add(
			c.config.TransactionTTL,
		),
	}

	c.transactionMutex.Lock()
	defer c.transactionMutex.Unlock()

	c.deleteExpiredTransactionsLocked(time.Now().UTC())
	c.transactions[transaction.ID] = transaction

	return transaction, nil
}

func (c *Client) consumeTransaction(
	id string,
) (authorizationTransaction, error) {
	c.transactionMutex.Lock()
	defer c.transactionMutex.Unlock()

	transaction, exists := c.transactions[id]
	delete(c.transactions, id)

	if !exists ||
		!transaction.ExpiresAt.After(time.Now().UTC()) {
		return authorizationTransaction{},
			errors.New("authorization transaction not found")
	}

	return transaction, nil
}
