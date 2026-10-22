# Project Handoff: Tapu IAM

This document gives a new developer or LLM enough context to continue developing the project. The repository is always the source of truth; inspect the current code before making changes.

## 1. Project purpose

Tapu is a small, self-hosted Identity and Access Management service for personal projects.

It provides:

- One central username/password login
- Browser single sign-on across multiple applications
- Admin-created users only
- Per-service access assignments
- OAuth 2.0 Authorization Code with PKCE
- OpenID Connect ID tokens
- Signed JWT access tokens
- Public and confidential OAuth clients
- A reusable Go integration SDK
- A Svelte administration UI
- Audit logging
- End-to-end browser tests

Each connected project keeps ownership of its own data. Projects store the immutable IAM user subject and do not store passwords or duplicate IAM identity data unnecessarily.

The current Go module path is expected to be:

```text
github.com/timsgarage/tapu
```

Confirm this in `go.mod` before adding imports.

---

## 2. Current implementation status

Completed:

- Phase 0: Go server, PostgreSQL, migrations, Docker Compose, health checks, signing keys
- Phase 1: users, Argon2id passwords, IAM sessions, login, logout, rate limiting, audit events
- Phase 2: services, user access, OAuth clients, admin API, Svelte admin UI
- Phase 3A: OIDC discovery, JWKS, token signing foundation
- Phase 3B: Authorization Code with PKCE, browser SSO, access tokens, ID tokens, UserInfo
- Phase 3C: reusable Go SDK, local application sessions, CORS support, example integration
- Playwright end-to-end test foundation

Not yet implemented:

- Client Credentials for machine-to-machine communication
- Refresh tokens
- Persistent sessions in the Go SDK
- Central OIDC logout
- GitHub login/linking
- Passkeys or MFA
- Per-service roles and permissions
- Signing-key rotation with multiple active public keys
- Production metrics stack
- Scheduled cleanup of expired rows

The next planned feature phase is machine clients and the Client Credentials grant.

---

## 3. High-level architecture

```text
                        ┌─────────────────────────┐
                        │       Tapu IAM          │
                        │ auth.example.com        │
                        │                         │
                        │ Users and passwords     │
                        │ IAM browser sessions    │
                        │ Service access          │
                        │ OAuth clients           │
                        │ Token issuance          │
                        │ Admin UI                │
                        └────────────┬────────────┘
                                     │
                                  PostgreSQL

┌─────────────────────────┐          │          ┌─────────────────────────┐
│ Project A               │          │          │ Project B               │
│ app.example.com         │◄─────────┴─────────►│ other.example.com       │
│ api.example.com         │     OIDC/OAuth       │ Go backend              │
│                         │                      │                         │
│ Own database            │                      │ Own database            │
│ Stores IAM subject      │                      │ Stores IAM subject      │
└─────────────────────────┘                      └─────────────────────────┘
```

The IAM owns:

- Global user identity
- Password credentials
- Administrator status
- IAM sessions
- Services and audiences
- User-to-service access
- OAuth clients and redirect URIs
- Client-to-service access
- Authorization codes
- Token signing
- Audit events

Each application owns:

- Application-specific user data
- Profiles and preferences
- Domain data
- Future project-specific roles
- A reference to the IAM user subject

Applications must use the token `sub` claim as the permanent IAM user identifier. Usernames may change and must not be used as foreign keys.

---

## 4. Main technology choices

### Backend

- Go
- Chi router
- GORM for data access
- Goose for explicit SQL migrations
- PostgreSQL
- `golang-jwt/jwt` for JWT signing
- Ed25519/EdDSA signing
- Argon2id password hashing

### Frontend

- Svelte
- Vite
- Embedded admin UI assets in the Go binary

### Deployment

- Docker
- `docker-compose`
- PostgreSQL container
- Migration container
- IAM container
- A TLS reverse proxy such as Caddy in production

### Testing

- Go unit tests
- Playwright browser tests
- Isolated E2E Docker Compose project
- Dedicated OAuth receiver and example Go API

Use `docker-compose`, not `docker compose`, in project documentation and commands.

---

## 5. Important repository structure

```text
.
├── cmd/
│   ├── iam-server/       # Main HTTP server
│   ├── iam-migrate/      # Goose migration runner
│   ├── iam-keygen/       # Ed25519 key generator
│   └── iamctl/           # Administrative and password-hash CLI
├── internal/
│   ├── admin/            # Admin models and API handlers
│   ├── adminui/          # Embedded Svelte build
│   ├── audit/            # Audit event persistence
│   ├── clientsecret/     # OAuth client-secret generation/hashing
│   ├── config/           # Environment configuration
│   ├── database/         # GORM connection and migrations
│   ├── health/           # Liveness and readiness
│   ├── identity/         # Users, usernames, passwords, authentication
│   ├── oauth/            # OAuth handlers, PKCE, codes, repository
│   ├── oidc/             # Discovery and public OIDC metadata
│   ├── ratelimit/        # In-memory login limiter
│   ├── session/          # IAM browser sessions
│   ├── signing/          # Ed25519 key handling
│   ├── token/            # JWT claims, signing, and JWKS
│   └── web/              # Login/account pages and web middleware
├── sdk/
│   └── go/authsdk/       # Reusable Go relying-party SDK
├── web/
│   └── admin/            # Svelte admin application
├── examples/
│   └── go-svelte/        # Go/Svelte integration example
├── e2e/
│   ├── receiver/         # Public OAuth test client
│   ├── scripts/          # Deterministic seed script
│   └── tests/            # Playwright specifications
├── deploy/
│   └── compose/
│       ├── compose.yml
│       └── compose.e2e.yml
├── docs/
├── secrets/
├── Dockerfile
├── Makefile
├── go.mod
└── go.sum
```

---

## 6. Database model

Migrations are embedded from:

```text
internal/database/migrations/
```

Use explicit SQL migrations. Do not use GORM `AutoMigrate` in production.

### Main tables

#### `users`

Contains:

- Immutable UUID
- Normalized username
- Argon2id password hash
- IAM administrator flag
- Active or disabled status
- Password-change timestamp
- Creation and update timestamps

#### `iam_sessions`

Contains hashed opaque IAM session and CSRF tokens.

Raw session and CSRF tokens are never stored.

#### `services`

Represents protected applications or APIs.

Important fields:

- Key
- Display name
- Unique audience
- Enabled status

#### `user_service_access`

Maps users to services they may access.

#### `oauth_clients`

Represents public or confidential clients.

Important fields:

- Public client ID
- Client type
- Hashed client secret for confidential clients
- Enabled status

Raw client secrets are displayed once and are not stored.

#### `oauth_client_redirect_uris`

Stores exact registered redirect URIs.

Wildcard redirect URIs are not supported.

#### `oauth_client_services`

Specifies which services and audiences a client may request.

#### `oauth_authorization_codes`

Stores hashed, single-use authorization codes bound to:

- User
- OAuth client
- Service
- Redirect URI
- Scope
- Nonce
- PKCE challenge
- Expiration

#### `audit_events`

Stores security-relevant actions without credentials or tokens.

### GORM acronym warning

GORM can incorrectly convert fields such as `OAuthClientID` to `o_auth_client_id`.

Always map acronym-heavy fields explicitly:

```go
OAuthClientID uuid.UUID `gorm:"column:oauth_client_id;type:uuid"`
```

Existing examples include:

- `OAuthClientID`
- `PKCEChallenge`
- `PKCEMethod`

---

## 7. Authentication and session flow

### IAM login

1. The user opens `/login`.
2. A login CSRF token is issued.
3. The user submits username and password.
4. The password is checked using Argon2id.
5. A random opaque IAM session token is generated.
6. Only its SHA-256 hash is stored.
7. The browser receives an `HttpOnly` session cookie.
8. The user is redirected to `/account` or back to an OAuth request.

IAM cookies are host-only. Do not set `Domain=.example.com`.

### Browser SSO

Applications redirect to:

```text
GET /oauth/authorize
```

If the user already has an IAM session, the IAM completes authorization without showing the login form again.

SSO is redirect-based. Applications do not share IAM cookies.

### IAM session properties

Production cookies use:

- `HttpOnly`
- `Secure`
- `SameSite=Lax`
- Host-only scope
- Path `/`

IAM sessions have both idle and absolute expiration.

---

## 8. OAuth and OpenID Connect behavior

### Supported grant

Implemented:

```text
authorization_code
```

Not yet implemented:

```text
client_credentials
refresh_token
device_code
```

### PKCE

PKCE is mandatory for every Authorization Code request.

Only this method is supported:

```text
S256
```

Plain PKCE is rejected.

### Supported scopes

```text
openid
profile
```

`openid` is mandatory.

The `profile` scope enables username-related claims.

### Audience

Tapu currently uses an explicit `audience` authorization parameter:

```text
audience=project-a-api
```

Both must be true before authorization succeeds:

- The user has access to the service.
- The OAuth client has access to the service.

Access is checked again during authorization-code exchange.

### Authorization codes

Authorization codes are:

- Cryptographically random
- Stored only as SHA-256 hashes
- Valid for approximately one minute
- Single-use
- Consumed atomically in a database transaction
- Bound to the exact client and redirect URI
- Bound to the PKCE challenge
- Bound to the user and service

### Access tokens

Access tokens are signed JWTs using EdDSA.

Important claims include:

```text
iss          IAM issuer
sub          immutable user UUID
aud          service audience
exp          expiration
iat          issue time
nbf          not-before time
jti          unique token ID
token_type   access_token
client_id    requesting OAuth client
scope        granted scopes
username     optional profile data
```

Default lifetime is approximately five minutes.

### ID tokens

ID tokens use the OAuth client ID as their audience.

Important claims include:

```text
iss
sub
aud
exp
iat
jti
nonce
preferred_username
```

### Public OIDC endpoints

```text
GET /.well-known/openid-configuration
GET /.well-known/jwks.json
```

### OAuth endpoints

```text
GET  /oauth/authorize
POST /oauth/token
GET  /oauth/userinfo
POST /oauth/userinfo
```

### UserInfo

UserInfo validates:

- EdDSA signature
- Issuer
- Expiration
- Access-token type
- Active user status

It returns the immutable user subject and optional profile information.

---

## 9. Main web and admin endpoints

### General

```text
GET  /
GET  /login
POST /login
GET  /account
POST /logout
```

### Health

```text
GET /health/live
GET /health/ready
```

### Admin API

The exact registered routes should be confirmed in `cmd/iam-server/main.go`.

Main groups include:

```text
GET    /api/admin/session

GET    /api/admin/users
POST   /api/admin/users
PATCH  /api/admin/users/{userID}
POST   /api/admin/users/{userID}/password

GET    /api/admin/users/{userID}/services
PUT    /api/admin/users/{userID}/services/{serviceID}
DELETE /api/admin/users/{userID}/services/{serviceID}

GET    /api/admin/services
POST   /api/admin/services
PATCH  /api/admin/services/{serviceID}

GET    /api/admin/clients
POST   /api/admin/clients
PATCH  /api/admin/clients/{clientID}
POST   /api/admin/clients/{clientID}/rotate-secret
```

Admin APIs require:

- A valid IAM session
- IAM administrator status
- CSRF protection for state-changing requests

---

## 10. Go SDK

The reusable SDK is located at:

```text
sdk/go/authsdk
```

It supports backend-for-frontend applications.

### SDK responsibilities

- OIDC discovery
- Authorization Code with PKCE
- State and nonce validation
- ID-token validation
- Access-token validation
- Application-local opaque sessions
- Login and callback handlers
- Current-user endpoint
- CSRF-protected local logout
- Authentication middleware
- Credentialed CORS
- Identity and access-token request context helpers

### Typical setup

```go
authClient, err := authsdk.New(
    context.Background(),
    authsdk.Config{
        IssuerURL:     os.Getenv("IAM_ISSUER_URL"),
        ClientID:      os.Getenv("IAM_CLIENT_ID"),
        ClientSecret:  os.Getenv("IAM_CLIENT_SECRET"),
        RedirectURL:   os.Getenv("IAM_REDIRECT_URL"),
        Audience:      os.Getenv("IAM_AUDIENCE"),
        AfterLoginURL: os.Getenv("APP_FRONTEND_URL"),
        AllowedOrigins: []string{
            os.Getenv("APP_FRONTEND_URL"),
        },
        SecureCookies: true,
    },
)
```

### Typical routes

```go
router.Use(authClient.CORS)

router.Get("/auth/login", authClient.LoginHandler)
router.Get("/auth/callback", authClient.CallbackHandler)
router.Get("/api/auth/me", authClient.CurrentUserHandler)
router.Post("/api/auth/logout", authClient.LogoutHandler)

router.With(
    authClient.RequireAuthentication,
).Get("/api/private", privateHandler)
```

### SDK limitations

The current SDK uses in-memory storage for:

- Application sessions
- Authorization transactions

Consequences:

- Application restarts sign users out.
- Multiple replicas cannot share sessions.
- OAuth callbacks must reach the same instance.
- Sessions expire no later than the access token.
- There are no refresh tokens.

The `SessionStore` interface exists so PostgreSQL or Redis-backed storage can be added later.

---

## 11. Separate frontend and API origins

The supported production pattern is:

```text
Frontend: https://app.example.com
API:      https://api.example.com
IAM:      https://auth.example.com
Callback: https://api.example.com/auth/callback
```

The Svelte frontend must use:

```javascript
fetch(`${apiOrigin}/api/auth/me`, {
	credentials: "include",
});
```

`credentials: 'same-origin'` is insufficient because the frontend and API are different origins.

The API must return exact credentialed CORS headers for known origins.

Never combine credentials with:

```text
Access-Control-Allow-Origin: *
```

The application session cookie remains host-only for `api.example.com`.

---

## 12. Security properties and constraints

The project intentionally follows these rules:

- Never implement cryptographic algorithms manually.
- Never store raw passwords.
- Never store raw session tokens.
- Never store raw authorization codes.
- Never store raw client secrets.
- Never log tokens or credentials.
- Match redirect URIs exactly.
- Require PKCE `S256`.
- Use generic login errors.
- Rate-limit login attempts.
- Keep IAM cookies host-only.
- Require CSRF protection for cookie-authenticated mutations.
- Enforce configured JWT algorithms.
- Verify issuer, audience, signature, expiration, and token type.
- Keep OAuth tokens out of browser JavaScript in backend-for-frontend applications.
- Use HTTPS and `Secure` cookies in production.
- Keep PostgreSQL off the public network.
- Keep signing keys and `.env` files out of Git.

Audit metadata must not contain:

- Passwords
- Password hashes
- Client secrets
- Session tokens
- CSRF tokens
- Authorization codes
- PKCE verifiers
- Access tokens
- ID tokens

---

## 13. Configuration

Important IAM environment variables include:

```dotenv
APP_ENV=development
HTTP_ADDRESS=:8080
DATABASE_URL=postgres://...
SIGNING_PRIVATE_KEY_PATH=/run/secrets/signing_private_key
ISSUER_URL=http://127.0.0.1:8080
```

Production requires:

```dotenv
APP_ENV=production
ISSUER_URL=https://auth.example.com
```

The issuer must remain stable and exactly match the externally visible IAM URL.

Application SDK configuration commonly uses:

```dotenv
IAM_ISSUER_URL=https://auth.example.com
IAM_CLIENT_ID=...
IAM_CLIENT_SECRET=...
IAM_REDIRECT_URL=https://api.example.com/auth/callback
IAM_AUDIENCE=project-a-api
APP_FRONTEND_URL=https://app.example.com
APP_SECURE_COOKIES=true
```

Never expose the client secret through a `VITE_` environment variable.

---

## 14. Local development commands

Create local configuration:

```bash
cp .env.example .env
```

Generate signing keys:

```bash
go run ./cmd/iam-keygen
```

Run Go checks:

```bash
go mod tidy
go fmt ./...
go vet ./...
go test -race ./...
```

Build the admin UI:

```bash
cd web/admin
npm install
npm run build
cd ../..
```

Start the stack:

```bash
docker-compose \
  --env-file .env \
  -f deploy/compose/compose.yml \
  up --build -d
```

Create an administrator:

```bash
make admin-create
```

Open:

```text
http://127.0.0.1:8080/login
http://127.0.0.1:8080/admin
```

Stop without deleting data:

```bash
docker-compose \
  --env-file .env \
  -f deploy/compose/compose.yml \
  down
```

---

## 15. End-to-end tests

Playwright tests cover:

- Invalid login
- Login and logout
- Protected account access
- Administrator access
- Normal-user admin denial
- Authorization Code with PKCE
- Browser SSO
- Incorrect PKCE rejection
- Credentialed CORS

The E2E environment uses:

```text
IAM:            http://127.0.0.1:8080
Example API:    http://127.0.0.1:3000
OAuth receiver: http://127.0.0.1:4100
PostgreSQL:     127.0.0.1:55432
```

Clean test run:

```bash
make e2e-reset
make e2e-up
make e2e-seed
make e2e-test
```

Important files:

```text
e2e/.env
e2e/.generated.json
deploy/compose/compose.e2e.yml
e2e/playwright.config.js
```

The E2E Compose project is:

```text
personal-iam-e2e
```

Always use `e2e/.env` for E2E Compose interpolation.

Never run E2E tests against production or a valued development database.

---

## 16. Production deployment expectations

Production requires:

- HTTPS reverse proxy
- Public issuer such as `https://auth.example.com`
- IAM bound only to loopback or a private container network
- PostgreSQL not published publicly
- Persistent PostgreSQL storage
- Backups of PostgreSQL and signing keys
- At least two administrator accounts
- `APP_ENV=production`
- Secure file permissions for secrets

Back up:

```text
PostgreSQL
secrets/signing-private.pem
.env or equivalent secret configuration
reverse-proxy configuration
```

Losing the signing private key prevents continuity of token verification after recovery.

---

## 17. Recommended next development phase

The next planned phase is Client Credentials for machine-to-machine communication.

A suitable design should include:

1. Machine clients represented separately or explicitly typed.
2. Hashed client secrets.
3. Service allowlists.
4. `grant_type=client_credentials`.
5. Audience-specific access tokens.
6. No user subject or username claims.
7. Machine identity in `sub`.
8. Short token lifetimes.
9. Secret rotation through the admin UI.
10. Audit events for machine-token issuance.
11. A Go SDK token source with safe caching.
12. E2E and unit tests.

Do not reuse ordinary user accounts for machine authentication.

---

## 18. Later expansion options

### Persistent SDK sessions

Implement `SessionStore` using PostgreSQL or Redis.

Requirements:

- Store only hashed opaque session identifiers where practical.
- Preserve expiration.
- Support deletion and cleanup.
- Support multiple application replicas.

### Refresh tokens

If introduced:

- Use opaque random values, not JWTs.
- Store only hashes.
- Rotate on every use.
- Track token families.
- Detect reuse.
- Revoke the family after reuse.
- Bind tokens to user, client, and service.
- Add explicit revocation.

### Key rotation

Support:

- Multiple public JWKS entries
- One active signing key
- Previously active verification keys
- Stable key IDs
- Safe retirement after all signed tokens expire

### Roles and permissions

The current central authorization model is binary service access.

A future model could add:

- Service-specific roles
- Role claims in audience-specific tokens
- Project-owned fine-grained authorization

Keep application-specific business permissions in the owning project unless there is a clear need to centralize them.

### External identity providers

GitHub login may later be linked to existing IAM users.

Account linking must require authenticated administrative or user confirmation. Do not automatically merge accounts based only on a username.

---

## 19. Development rules for future contributors or LLMs

Before changing code:

1. Inspect `go.mod` for the actual module path.
2. Inspect `cmd/iam-server/main.go` for current route wiring.
3. Inspect all existing migrations.
4. Check existing models for explicit GORM column names.
5. Run current tests before making changes.
6. Do not assume this document is newer than the repository.

When adding a feature:

1. Define the security boundary.
2. Add an explicit migration if persistence changes.
3. Add or update models.
4. Add repository logic.
5. Add handlers and middleware.
6. Register routes centrally.
7. Add unit tests.
8. Add E2E coverage for browser-visible behavior.
9. Update documentation.
10. Run formatting, vet, tests, and builds.

Required checks:

```bash
go mod tidy
go fmt ./...
go vet ./...
go test -race ./...
```

If the admin UI changes:

```bash
cd web/admin
npm ci
npm run build
cd ../..
```

If behavior spans browser boundaries:

```bash
make e2e-reset
make e2e-up
make e2e-seed
make e2e-test
```

Avoid large unrelated refactors while adding security-sensitive features.

---

## 20. Known areas requiring care

### Environment files

Normal development uses:

```text
.env
```

E2E Compose uses:

```text
e2e/.env
```

Using the wrong file can cause empty PostgreSQL variables or incorrect credentials.

### Generated E2E configuration

The seed writes:

```text
e2e/.generated.json
```

It contains test credentials and must remain ignored.

### Redirect URIs

They must match exactly. Differences in any of these matter:

- `http` versus `https`
- `localhost` versus `127.0.0.1`
- Port
- Path
- Trailing slash

### Issuer

Do not switch between:

```text
http://localhost:8080
http://127.0.0.1:8080
https://auth.example.com
```

A token issuer is compared exactly.

### Cookies

Local HTTP requires `SecureCookies: false`.

Production HTTPS requires:

```go
SecureCookies: true
```

### Admin lockout

The current administration API may allow an administrator to disable themselves or remove their own administrator status. Maintain at least two active administrators.

A future improvement should prevent removal or disabling of the final active administrator.

---

## 21. Guidance for the next assistant

When continuing development:

- Treat the repository as authoritative.
- Ask for current build or test failures before assuming a feature works.
- Provide changes file by file.
- Preserve the incremental phase structure.
- Do not proceed to a later phase until the current one is verified.
- Use `docker-compose` in all commands.
- Include migrations and tests with persistent changes.
- Explain security-sensitive decisions briefly.
- Do not introduce custom cryptography or non-standard authentication flows.
- Prefer standard OAuth 2.0 and OpenID Connect behavior.
- Never ask the user to commit secrets.
- Keep instructions executable and internally consistent.
