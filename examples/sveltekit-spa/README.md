# SvelteKit client-side OAuth example

This is a SvelteKit application using:

- Svelte 5 runes
- Client-side rendering
- `@sveltejs/adapter-node`
- Authorization Code with PKCE
- A public Tapu OAuth client
- `oidc-client-ts`

## Security model

The project has no confidential backend and therefore has no client secret.

OAuth tokens are available to browser JavaScript and are stored in
`sessionStorage`. Preventing cross-site scripting is essential.

Client-side route protection is not an authorization boundary. APIs must
validate access tokens independently.

For applications with a Go backend, prefer the Go and SvelteKit example.

## IAM registration

Create:

    Service audience: sveltekit-spa-api
    Client type: public
    Redirect URI: http://127.0.0.1:5173/auth/callback

Grant the user and client access to the service.

## Development

    cp .env.example .env
    npm install
    npm run check
    npm run dev

Open:

    http://127.0.0.1:5173

## Production-style Node build

    npm run build

    HOST=127.0.0.1 \
    PORT=5173 \
    ORIGIN=http://127.0.0.1:5173 \
    node build
