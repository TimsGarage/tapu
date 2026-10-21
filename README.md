# Personal IAM

A small identity and access-management service for independently owned personal
applications.

## Current state

Phase 0 provides:

- Go HTTP server
- PostgreSQL
- Explicit SQL migrations
- Health checks
- Ed25519 signing-key generation
- Docker Compose deployment
- Tests and CI

Authentication is not implemented yet.

## Requirements

- Go matching the version in `go.mod`
- Docker with Docker Compose
- OpenSSL, or another secure random-password generator

## Initial setup

Create local environment configuration:

    cp .env.example .env

Replace the example PostgreSQL password in `.env`.

Generate the signing key:

    go run ./cmd/iam-keygen

Run tests:

    go test -race ./...

Start the stack:

    docker compose \
      --env-file .env \
      -f deploy/compose/compose.yml \
      up --build -d

Check container state:

    docker compose \
      --env-file .env \
      -f deploy/compose/compose.yml \
      ps

Test the endpoints:

    curl http://127.0.0.1:8080/health/live
    curl http://127.0.0.1:8080/health/ready

Both should return:

    {"status":"ok"}

## Security warning

The Phase 0 server is not ready for internet exposure.

Before deployment, it must be placed behind a correctly configured HTTPS reverse
proxy. User authentication and access control will be introduced in later
phases.

Never commit:

- `.env`
- Database passwords
- Signing keys
- Future client secrets
- Future session or refresh tokens

## Phase 1

Phase 1 adds:

- Administrator-created users
- Argon2id password hashing
- Database-backed sessions
- Hosted login
- Hosted account page
- Logout
- CSRF protection
- Login rate limiting
- Audit events

### Create an administrator

Start the stack and run:

    make admin-create

Then open:

    http://127.0.0.1:8080/login

The local Compose configuration uses `APP_ENV=development` so cookies work over
plain HTTP.

Never expose the development configuration to the internet. Production must use
`APP_ENV=production` behind HTTPS.

## OpenID Connect issuer

The IAM publishes OpenID Connect metadata at:

    /.well-known/openid-configuration

The signing public key is published at:

    /.well-known/jwks.json

The canonical issuer is configured through `ISSUER_URL`.

For local development:

    ISSUER_URL=http://127.0.0.1:8080

For production, it must use the externally visible HTTPS origin:

    ISSUER_URL=https://auth.example.com

The issuer value must remain stable. Applications validate it exactly.

Phase 3A does not yet implement the authorization or token endpoints.

## Authorization Code flow

The IAM supports OpenID Connect Authorization Code with mandatory PKCE `S256`.

Authorization endpoint:

    GET /oauth/authorize

Token endpoint:

    POST /oauth/token

UserInfo endpoint:

    GET /oauth/userinfo

Supported scopes:

- `openid`
- `profile`

Required authorization parameters:

- `response_type=code`
- `client_id`
- `redirect_uri`
- `scope`
- `audience`
- `code_challenge`
- `code_challenge_method=S256`

Confidential clients authenticate at the token endpoint using HTTP Basic
authentication.

Public clients send their `client_id` in the token request and rely on PKCE.

Authorization codes expire after one minute and can be used exactly once.
Access and ID tokens expire after five minutes. Refresh tokens are not currently
supported.

## Go backend integration

The reusable Go SDK is available at:

    sdk/go/authsdk

It provides:

- Authorization Code with PKCE
- State and nonce validation
- ID- and access-token validation
- Opaque application-local sessions
- Login, callback, current-user, and logout handlers
- Authentication middleware
- Request-context identity and access-token helpers

A minimal integration is available at:

    examples/go-svelte

The browser receives only an opaque application session cookie. OAuth tokens
remain in the Go backend.

The current SDK session store is in memory. Application sessions expire no later
than the five-minute IAM access token.
