# Tapu

Tapu is a small, self-hosted Identity and Access Management service for personal projects.

It provides one central login, browser single sign-on, service-level access control, OAuth 2.0/OpenID Connect, an administration UI, and reusable Go and SvelteKit integration libraries.

## Features

- Username and password authentication
- Invite-only user creation
- Browser single sign-on
- OAuth Authorization Code with PKCE
- OpenID Connect discovery and ID tokens
- EdDSA-signed JWT access tokens
- Public and confidential OAuth clients
- Exact redirect URI validation
- User-to-service access assignments
- Client-to-service access assignments
- Svelte administration UI
- Go backend integration SDK
- SvelteKit integration examples
- Audit events
- PostgreSQL persistence
- Docker and GHCR deployment

## Architecture

Tapu owns global identity and authentication.

Each connected project owns its application data and stores the immutable Tapu user subject as its user reference.

```text
SvelteKit frontend
        |
        v
Go backend, optional
        |
        v
Tapu IAM
        |
        v
PostgreSQL
```

For applications with a backend, the recommended integration keeps OAuth tokens in Go and gives the browser only an opaque application session cookie.

## Quick start for development

Requirements:

- Go
- Node.js and npm
- Docker
- `docker-compose`

Create local configuration:

```bash
cp .env.example .env
```

Generate signing keys:

```bash
go run ./cmd/iam-keygen
```

Build and start:

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
```

Administration UI:

```text
http://127.0.0.1:8080/admin
```

## Production deployment

Production deployments pull the Tapu image from GitHub Container Registry. The server does not need the source repository, Go, or Node.js.

Required server files:

```text
compose.yml
.env
```

Typical configuration:

```dotenv
IAM_IMAGE=ghcr.io/timsgarage/tapu:v0.1.0
IAM_PUBLIC_URL=https://auth.example.com

POSTGRES_DB=iam
POSTGRES_USER=iam
POSTGRES_PASSWORD=replace-with-a-random-password
```

Deploy:

```bash
docker-compose pull
docker-compose up -d
```

See `docs/deployment.md` for HTTPS, backups, updates, and rollback.

## Integration options

### Go backend with SvelteKit

Recommended for most applications.

- Confidential OAuth client
- OAuth tokens remain in Go
- Opaque browser session
- CSRF-protected API requests
- Credentialed CORS for separate frontend and API origins

Example:

```text
examples/go-sveltekit
```

### Browser-only SvelteKit

Suitable when no backend is available.

- Public OAuth client
- Authorization Code with PKCE
- Tokens stored in browser session storage
- Strong XSS protection required

Example:

```text
examples/sveltekit-spa
```

## Development checks

```bash
go mod tidy
go fmt ./...
go vet ./...
go test -race ./...
```

Build the administration UI:

```bash
cd web/admin
npm ci
npm run build
```

Run end-to-end tests:

```bash
make e2e-reset
make e2e-up
make e2e-seed
make e2e-test
```

## Documentation

- `docs/architecture.md` — system design and trust boundaries
- `docs/getting-started.md` — development, deployment, and integration
- `docs/deployment.md` — production container deployment
- `docs/oauth-oidc.md` — supported OAuth and OIDC behavior
- `docs/administration.md` — users, services, access, and clients
- `docs/operations.md` — backups, monitoring, updates, and incidents
- `docs/e2e-testing.md` — end-to-end test setup
- `docs/project-handoff.md` — complete project context

## Current limitations

- No refresh tokens
- No machine Client Credentials grant yet
- No MFA or passkeys
- No email recovery
- No signing-key rotation
- Default Go SDK sessions are in memory
- Single IAM instance is the intended deployment model

Tapu is designed for personal and non-critical projects. Sensitive or high-value deployments should receive additional hardening and an independent security review.

## Security

Never commit:

- `.env` files
- Signing private keys
- Database credentials
- OAuth client secrets
- Generated E2E credentials

Production deployments must use HTTPS and secure cookies.
