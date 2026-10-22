# Getting Started

This guide explains how to:

1. Run and develop the IAM locally
2. Deploy it on a server
3. Integrate it into another Go and Svelte project

The IAM provides centralized login while each connected project keeps ownership of its own application data.

---

# 1. Local development

## Requirements

Install:

- Git
- Go matching the version in `go.mod`
- Node.js and npm
- Docker
- `docker-compose`

Verify the installations:

```bash
go version
node --version
npm --version
docker --version
docker-compose version
```

## Clone the repository

```bash
git clone https://github.com/your-user/your-iam.git
cd your-iam
```

## Create the environment file

```bash
cp .env.example .env
```

Open `.env` and set a long, random PostgreSQL password:

```dotenv
POSTGRES_DB=iam
POSTGRES_USER=iam
POSTGRES_PASSWORD=replace-with-a-long-random-password
```

A password can be generated with:

```bash
openssl rand -base64 32
```

Never commit `.env`.

## Generate the signing key

The IAM uses an Ed25519 key to sign access and ID tokens:

```bash
go run ./cmd/iam-keygen
```

This creates:

```text
secrets/signing-private.pem
secrets/signing-public.pem
```

Never commit these files. Losing the private key invalidates existing tokens, so back it up before using the IAM in production.

## Install frontend dependencies

```bash
cd web/admin
npm install
cd ../..
```

## Run checks

```bash
go mod tidy
go fmt ./...
go vet ./...
go test -race ./...
```

Build the admin UI:

```bash
cd web/admin
npm run build
cd ../..
```

## Start the local stack

```bash
docker-compose \
  --env-file .env \
  -f deploy/compose/compose.yml \
  up --build -d
```

Check that all services started:

```bash
docker-compose \
  --env-file .env \
  -f deploy/compose/compose.yml \
  ps
```

Test readiness:

```bash
curl --fail http://127.0.0.1:8080/health/ready
```

Expected response:

```json
{ "status": "ok" }
```

## Create the first administrator

```bash
make admin-create
```

Enter a username and a password of at least 12 characters.

Create at least two administrators before production use to reduce the risk of administrative lockout.

## Open the IAM

Open:

```text
http://127.0.0.1:8080/login
```

After signing in, open:

```text
http://127.0.0.1:8080/admin
```

From the admin UI you can:

- Create and disable users
- Reset passwords
- Create services
- Grant users access to services
- Register OAuth clients
- Manage redirect URIs
- Rotate client secrets

## View logs

```bash
docker-compose \
  --env-file .env \
  -f deploy/compose/compose.yml \
  logs -f
```

## Stop the stack

```bash
docker-compose \
  --env-file .env \
  -f deploy/compose/compose.yml \
  down
```

This preserves PostgreSQL data.

Do not add `--volumes` unless you intend to delete the local database.

---

# 2. Production deployment

The production setup requires:

- A server with Docker and `docker-compose`
- A domain such as `auth.example.com`
- DNS pointing that domain to the server
- HTTPS through a reverse proxy such as Caddy
- Persistent storage and backups

The IAM must not be exposed directly to the internet over plain HTTP.

## Prepare the server

Clone the repository:

```bash
git clone https://github.com/your-user/your-iam.git
cd your-iam
```

Create the production environment:

```bash
cp .env.example .env
```

Set a unique PostgreSQL password:

```dotenv
POSTGRES_DB=iam
POSTGRES_USER=iam
POSTGRES_PASSWORD=replace-with-a-long-random-password
```

Protect the file:

```bash
chmod 600 .env
```

## Generate or restore the signing key

For a new installation:

```bash
go run ./cmd/iam-keygen
```

For a restored installation, place the backed-up private key at:

```text
secrets/signing-private.pem
```

Protect it:

```bash
chmod 600 secrets/signing-private.pem
```

## Configure the production issuer

The IAM container must use its public HTTPS address:

```yaml
environment:
    APP_ENV: production
    HTTP_ADDRESS: ":8080"
    ISSUER_URL: "https://auth.example.com"
```

The issuer is part of every token and must exactly match the address used by applications. Do not change it casually after deployment.

Production mode enables `Secure` cookies and requires an HTTPS issuer.

## Configure the reverse proxy

The IAM should remain bound to the server’s loopback interface:

```yaml
ports:
    - "127.0.0.1:8080:8080"
```

A minimal Caddy configuration is:

```caddyfile
auth.example.com {
    reverse_proxy 127.0.0.1:8080
}
```

Caddy obtains and renews TLS certificates automatically when:

- DNS points to the server
- Ports 80 and 443 are reachable
- The domain is publicly valid

Do not publish PostgreSQL to the host or internet.

## Start the production stack

```bash
docker-compose \
  --env-file .env \
  -f deploy/compose/compose.yml \
  up --build -d
```

Check the containers:

```bash
docker-compose \
  --env-file .env \
  -f deploy/compose/compose.yml \
  ps
```

Verify the public endpoints:

```bash
curl --fail https://auth.example.com/health/ready
curl --fail https://auth.example.com/.well-known/openid-configuration
curl --fail https://auth.example.com/.well-known/jwks.json
```

## Create administrators

```bash
make admin-create
```

Create at least two administrators and use strong, unique passwords.

## Back up persistent data

Back up:

- PostgreSQL
- `secrets/signing-private.pem`
- Reverse-proxy configuration
- Any locally stored TLS or deployment files

Example PostgreSQL backup:

```bash
docker-compose \
  --env-file .env \
  -f deploy/compose/compose.yml \
  exec -T postgres \
  pg_dump -U iam -d iam > iam-backup.sql
```

Protect backups because they contain user data and password hashes.

Test restoration periodically. A backup that has never been restored is not a verified backup.

## Update the deployment

Before updating, create a backup. Then:

```bash
git pull
```

Rebuild and restart:

```bash
docker-compose \
  --env-file .env \
  -f deploy/compose/compose.yml \
  up --build -d
```

The migration container applies new database migrations before the IAM starts.

Check logs after every update:

```bash
docker-compose \
  --env-file .env \
  -f deploy/compose/compose.yml \
  logs migrate iam
```

---

# 3. Integrating another project

The recommended architecture is:

```text
Svelte frontend
    |
    | requests with an application session cookie
    v
Go API
    |
    | OAuth Authorization Code with PKCE
    v
IAM
```

The browser receives only an opaque application cookie. OAuth tokens stay in the Go backend.

For separate subdomains:

```text
Frontend: https://app.example.com
API:      https://api.example.com
IAM:      https://auth.example.com
Callback: https://api.example.com/auth/callback
```

## Create the service in the IAM

In the IAM admin UI, create a service:

```text
Key: project-a
Name: Project A
Audience: project-a-api
```

The audience identifies the API for which access tokens are issued.

## Grant user access

Open the user in the admin UI and grant access to `Project A`.

A user without this assignment cannot obtain a token for the service.

## Create an OAuth client

Create a confidential OAuth client:

```text
Name: Project A Web
Type: confidential
Redirect URI: https://api.example.com/auth/callback
Service: Project A
```

Copy the client ID and client secret.

The secret is displayed only once. Store it in the API’s secret-management system, not in source control or frontend code.

Redirect URIs must match exactly, including:

- Scheme
- Host
- Port
- Path

## Add the Go SDK

If the SDK is part of the same repository:

```go
import "github.com/your-user/your-iam/sdk/go/authsdk"
```

For a separate repository, add the IAM module:

```bash
go get github.com/your-user/your-iam/sdk/go/authsdk
```

Private repositories may require Git authentication and `GOPRIVATE` configuration.

## Configure the Go API

Set environment variables:

```dotenv
IAM_ISSUER_URL=https://auth.example.com
IAM_CLIENT_ID=replace-with-client-id
IAM_CLIENT_SECRET=replace-with-client-secret
IAM_REDIRECT_URL=https://api.example.com/auth/callback
IAM_AUDIENCE=project-a-api
APP_FRONTEND_URL=https://app.example.com
```

Never expose `IAM_CLIENT_SECRET` to Svelte or any browser-delivered code.

Create the SDK client:

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
if err != nil {
    log.Fatal(err)
}
```

`SecureCookies` must be `true` in production.

## Register authentication routes

Using Chi:

```go
router := chi.NewRouter()

router.Use(authClient.CORS)

router.Get("/auth/login", authClient.LoginHandler)
router.Get("/auth/callback", authClient.CallbackHandler)
router.Get("/api/auth/me", authClient.CurrentUserHandler)
router.Post("/api/auth/logout", authClient.LogoutHandler)
```

Protect an API route:

```go
router.With(
    authClient.RequireAuthentication,
).Get("/api/private", privateHandler)
```

Read the authenticated identity:

```go
func privateHandler(
    writer http.ResponseWriter,
    request *http.Request,
) {
    identity, ok := authsdk.IdentityFromContext(
        request.Context(),
    )
    if !ok {
        http.Error(
            writer,
            "Unauthorized",
            http.StatusUnauthorized,
        )
        return
    }

    // Use identity.Subject as the immutable IAM user ID.
    // Keep all project-specific data in this project's database.
}
```

Store `identity.Subject` as the IAM user reference. Do not use the username as a permanent database identifier because usernames may change.

## Configure the Svelte frontend

Set the API origin:

```dotenv
VITE_API_ORIGIN=https://api.example.com
```

Call the API with credentials included:

```javascript
const apiOrigin = import.meta.env.VITE_API_ORIGIN;

const response = await fetch(`${apiOrigin}/api/auth/me`, {
	credentials: "include",
	headers: {
		Accept: "application/json",
	},
});
```

Start login with a browser navigation:

```javascript
window.location.assign(`${apiOrigin}/auth/login`);
```

For logout, send the CSRF token returned by `/api/auth/me`:

```javascript
await fetch(`${apiOrigin}/api/auth/logout`, {
	method: "POST",
	credentials: "include",
	headers: {
		"X-CSRF-Token": csrfToken,
	},
});
```

Because `app.example.com` and `api.example.com` are different origins, `credentials: 'include'` is required.

## Cookie and CORS behavior

The API session cookie should be:

- Host-only for `api.example.com`
- `HttpOnly`
- `Secure`
- `SameSite=Lax`
- Path `/`

Do not configure `Domain=.example.com`. The frontend does not need direct access to the cookie.

The API must allow the exact frontend origin:

```text
https://app.example.com
```

Do not use wildcard CORS together with credentials.

## Local project testing

A local setup can use:

```text
Frontend: http://127.0.0.1:5173
API:      http://127.0.0.1:3000
IAM:      http://127.0.0.1:8080
Callback: http://127.0.0.1:3000/auth/callback
```

Register the local callback as a separate OAuth client redirect URI.

For local HTTP only:

```go
SecureCookies: false
```

Always use `true` behind production HTTPS.

## Integration checklist

Before considering the integration complete, verify:

- The service and audience exist
- The user has service access
- The OAuth client has service access
- The redirect URI matches exactly
- The client secret exists only in the Go API
- Login redirects through the IAM
- SSO works when an IAM session already exists
- `/api/auth/me` returns the user identity
- Protected APIs return HTTP 401 without a session
- Logout requires a valid CSRF token
- The frontend uses `credentials: 'include'`
- CORS permits only known frontend origins
- Project data uses the immutable IAM subject as its user reference
