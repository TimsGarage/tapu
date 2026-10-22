# Architecture

Tapu is a self-hosted Identity and Access Management service for multiple independent applications.

It provides central authentication and service-level access while each application remains responsible for its own data and permissions.

## Components

```text
                         ┌──────────────────────┐
                         │      Tapu IAM        │
                         │ auth.example.com     │
                         │                      │
                         │ Users and passwords  │
                         │ Browser sessions     │
                         │ OAuth clients        │
                         │ Service access       │
                         │ Token issuance       │
                         │ Administration UI    │
                         └──────────┬───────────┘
                                    │
                              PostgreSQL

┌─────────────────────────┐         │
│ Application             │         │ OAuth 2.0 / OpenID Connect
│                         │◄────────┘
│ SvelteKit frontend      │
│ Go backend, optional    │
│ Application database    │
└─────────────────────────┘
```

## Ownership boundaries

Tapu owns:

- Global users and passwords
- IAM administrator status
- IAM browser sessions
- Services and token audiences
- User-to-service access
- OAuth clients and redirect URIs
- Client-to-service access
- Authorization codes
- Token signing
- Audit events

Applications own:

- Application data
- Project-specific profiles
- Business roles and permissions
- Application sessions
- A reference to the immutable Tapu user subject

Applications should use the token `sub` claim as the permanent user identifier. Usernames must not be used as database foreign keys because they may change.

## Main technologies

- Go
- Chi and standard `net/http`
- PostgreSQL
- GORM
- Goose SQL migrations
- Svelte 5
- Ed25519 token signing
- Argon2id password hashing
- Docker and Docker Compose

## Authentication flow

### IAM login

1. The browser opens `/login`.
2. Tapu issues a login CSRF token.
3. The user submits a username and password.
4. Tapu verifies the Argon2id password hash.
5. Tapu creates an opaque browser session.
6. Only the session-token hash is stored.
7. The browser receives an `HttpOnly` session cookie.

### Browser single sign-on

Applications redirect the browser to:

```text
GET /oauth/authorize
```

If an IAM session already exists, Tapu can complete authorization without showing the login form again.

Applications do not share IAM cookies. SSO is redirect-based.

## Application integration models

### Go backend with SvelteKit

This is the recommended model.

```text
SvelteKit → Go backend → Tapu
```

- The Go backend is a confidential OAuth client.
- OAuth tokens remain in Go.
- The browser receives an opaque application cookie.
- API mutations use CSRF protection.
- The frontend and backend may use separate origins.

### Browser-only SvelteKit

```text
SvelteKit → Tapu
```

- The application is a public OAuth client.
- Authorization Code with PKCE is mandatory.
- Tokens are accessible to browser JavaScript.
- APIs must independently validate access tokens.
- Strong XSS protection is essential.

## Database

Important tables include:

- `users`
- `iam_sessions`
- `services`
- `user_service_access`
- `oauth_clients`
- `oauth_client_redirect_uris`
- `oauth_client_services`
- `oauth_authorization_codes`
- `audit_events`

Schema changes use explicit SQL migrations. GORM `AutoMigrate` is not used in production.

## Token model

Access tokens:

- Signed JWTs using EdDSA
- Intended for one service audience
- Short-lived
- Contain the immutable user subject
- Contain the requesting client ID and granted scopes

ID tokens:

- Signed JWTs using EdDSA
- Intended for the OAuth client
- Contain identity and nonce claims

Tapu currently publishes one signing key through JWKS.

## Current limitations

- No refresh tokens
- No Client Credentials grant yet
- No MFA or passkeys
- No account recovery
- No signing-key rotation
- Go SDK sessions are in memory by default
- Existing JWTs remain valid until expiration
- Login rate limiting is local to one IAM instance

These limitations are acceptable for the intended small, single-instance, personal-project environment but should be considered before protecting sensitive systems.
