# Getting Started

This guide explains how to:

1. Run and develop the IAM locally
2. Integrate it into another Go and Svelte project

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

# 2. Integrating another project

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

Using gin needs a middleware adapter:

```golang
package httpmiddleware

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

type StandardMiddleware func(http.Handler) http.Handler

func FromStandard(
	middleware StandardMiddleware,
) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		reachedNextHandler := false

		next := http.HandlerFunc(func(
			_ http.ResponseWriter,
			request *http.Request,
		) {
			reachedNextHandler = true
			ctx.Request = request
		})

		middleware(next).ServeHTTP(
			ctx.Writer,
			ctx.Request,
		)

		if !reachedNextHandler {
			ctx.Abort()
			return
		}

		ctx.Next()
	}
}
```

The router would look like:

```golang
    router := gin.New()
    router.Use(gin.Recovery())

    router.Use(
        httpmiddleware.FromStandard(
            authClient.CORS,
        ),
    )

    router.GET(
        "/auth/login",
        gin.WrapF(authClient.LoginHandler),
    )
    router.GET(
        "/auth/callback",
        gin.WrapF(authClient.CallbackHandler),
    )
    router.GET(
        "/api/auth/me",
        gin.WrapF(authClient.CurrentUserHandler),
    )
    router.POST(
        "/api/auth/logout",
        gin.WrapF(authClient.LogoutHandler),
    )

    authenticated := router.Group("/api")
    authenticated.Use(
        httpmiddleware.FromStandard(
            authClient.RequireAuthentication,
        ),
    )

    mutations := authenticated.Group("/todos")
    mutations.Use(
        httpmiddleware.FromStandard(
            authClient.RequireCSRF,
        ),
    )

    mutations.GET("", listTodos)
    mutations.POST("", createTodo)
    mutations.PUT("/:todoID", updateTodo)
    mutations.DELETE("/:todoID", deleteTodo)
```

Retrieve the Identity:

```golang
func privateHandler(ctx *gin.Context) {
	identity, ok := authsdk.IdentityFromContext(
		ctx.Request.Context(),
	)

    // Similar for access token
    // accessToken, ok := authsdk.AccessTokenFromContext(
        // ctx.Request.Context(),
    // )
	if !ok {
		ctx.JSON(
			http.StatusUnauthorized,
			gin.H{
				"error": "unauthenticated",
			},
		)
		return
	}

	ctx.JSON(
		http.StatusOK,
		gin.H{
			"subject":  identity.Subject,
			"username": identity.Username,
			"audience": identity.Audience,
			"scope":    identity.Scope,
		},
	)
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
