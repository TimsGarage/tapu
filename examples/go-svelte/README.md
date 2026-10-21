# Go and Svelte integration example

This example demonstrates the recommended backend-for-frontend integration.

The browser receives only an opaque application session cookie. OAuth access and
ID tokens remain in the Go backend.

## IAM configuration

Create or use:

- Service audience: `project-a-api`
- Confidential OAuth client
- Redirect URI: `http://127.0.0.1:3000/auth/callback`
- Client access to `project-a-api`
- User access to `project-a-api`

## Environment

Set the generated client credentials:

    export IAM_ISSUER_URL='http://127.0.0.1:8080'
    export IAM_CLIENT_ID='replace-with-client-id'

Read the secret without placing it directly in shell history:

    read -rs IAM_CLIENT_SECRET
    export IAM_CLIENT_SECRET
    printf '

'

Optional values:

    export IAM_REDIRECT_URL='http://127.0.0.1:3000/auth/callback'
    export IAM_AUDIENCE='project-a-api'

## Start the application

From the repository root:

    go run ./examples/go-svelte/cmd/server

Open:

    http://127.0.0.1:3000

The application redirects to the IAM for authentication. If an IAM session
already exists, the redirect completes without another password prompt.

## Production configuration

Use HTTPS and set `SecureCookies` to `true`.

The configured redirect URL must exactly match the redirect URI registered in
the IAM.

The default in-memory application session store is suitable only for a
single-instance application. Restarting the application removes all local
sessions.

Applications requiring persistent sessions should implement the SDK's
`SessionStore` interface using PostgreSQL or another server-side store.

## Integrate the SDK into another Project

Go needs setup:

```go
authClient, err := authsdk.New(
	context.Background(),
	authsdk.Config{
		IssuerURL:    os.Getenv("IAM_ISSUER_URL"),
		ClientID:     os.Getenv("IAM_CLIENT_ID"),
		ClientSecret: os.Getenv("IAM_CLIENT_SECRET"),
		RedirectURL:  "https://project-a.example.com/auth/callback",
		Audience:     "project-a-api",
		AfterLoginURL: "/",
		SecureCookies: true,
	},
)
```

And four basic routes:

```go
router.Get("/auth/login", authClient.LoginHandler)
router.Get("/auth/callback", authClient.CallbackHandler)
router.Get("/api/auth/me", authClient.CurrentUserHandler)
router.Post("/api/auth/logout", authClient.LogoutHandler)
```

Protected routes then look like this:

```go
router.With(
	authClient.RequireAuthentication,
).Get(
	"/api/private",
	privateHandler,
)
```

Read identity:

```go
identity, ok := authsdk.IdentityFromContext(
	request.Context(),
)
```
