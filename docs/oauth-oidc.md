# OAuth 2.0 and OpenID Connect

Tapu implements a focused OAuth 2.0 and OpenID Connect subset for browser and backend applications.

## Discovery

OpenID Connect metadata:

```text
GET /.well-known/openid-configuration
```

JSON Web Key Set:

```text
GET /.well-known/jwks.json
```

Applications should use discovery rather than hard-coding endpoint URLs.

## Endpoints

```text
GET  /oauth/authorize
POST /oauth/token
GET  /oauth/userinfo
POST /oauth/userinfo
```

## Supported flow

Tapu currently supports:

```text
Authorization Code with PKCE
```

PKCE is mandatory for public and confidential clients.

Only this PKCE method is accepted:

```text
S256
```

The following grants are not currently supported:

- Refresh Token
- Client Credentials
- Device Authorization
- Implicit

## Client types

### Public client

Suitable for browser-only applications.

- Has no client secret
- Sends `client_id` in the token request
- Must use PKCE
- Cannot securely hold credentials

### Confidential client

Suitable for a server-side backend.

- Has a client secret
- Authenticates with HTTP Basic at the token endpoint
- Must also use PKCE
- Keeps tokens away from browser JavaScript when used as a backend-for-frontend

## Authorization request

Example:

```text
GET /oauth/authorize
    ?response_type=code
    &client_id=CLIENT_ID
    &redirect_uri=https://app.example.com/auth/callback
    &scope=openid%20profile
    &audience=project-api
    &state=RANDOM_STATE
    &nonce=RANDOM_NONCE
    &code_challenge=PKCE_CHALLENGE
    &code_challenge_method=S256
```

Required parameters:

- `response_type=code`
- `client_id`
- `redirect_uri`
- `scope`
- `audience`
- `code_challenge`
- `code_challenge_method=S256`

Applications should also send and validate:

- `state`
- `nonce`

Redirect URIs are matched exactly.

## Scopes

Supported scopes:

```text
openid
profile
```

`openid` is required.

`profile` enables username-related claims such as `preferred_username`.

Unknown scopes are rejected.

## Service access

The `audience` parameter selects the target service.

Authorization succeeds only when:

- The OAuth client is enabled
- The service is enabled
- The redirect URI is registered
- The client is assigned to the service
- The user is active
- The user has access to the service

These checks are repeated during authorization-code exchange.

## Authorization codes

Authorization codes are:

- Cryptographically random
- Stored only as SHA-256 hashes
- Short-lived
- Single-use
- Bound to the client
- Bound to the redirect URI
- Bound to the user and service
- Bound to the PKCE challenge
- Consumed atomically

## Token exchange

Public client example:

```bash
curl \
  --request POST \
  --header 'Content-Type: application/x-www-form-urlencoded' \
  --data-urlencode 'grant_type=authorization_code' \
  --data-urlencode 'client_id=CLIENT_ID' \
  --data-urlencode 'code=AUTHORIZATION_CODE' \
  --data-urlencode 'redirect_uri=https://app.example.com/auth/callback' \
  --data-urlencode 'code_verifier=PKCE_VERIFIER' \
  https://auth.example.com/oauth/token
```

Confidential client example:

```bash
curl \
  --request POST \
  --user 'CLIENT_ID:CLIENT_SECRET' \
  --header 'Content-Type: application/x-www-form-urlencoded' \
  --data-urlencode 'grant_type=authorization_code' \
  --data-urlencode 'code=AUTHORIZATION_CODE' \
  --data-urlencode 'redirect_uri=https://api.example.com/auth/callback' \
  --data-urlencode 'code_verifier=PKCE_VERIFIER' \
  https://auth.example.com/oauth/token
```

Successful response:

```json
{
	"access_token": "eyJ...",
	"token_type": "Bearer",
	"expires_in": 300,
	"id_token": "eyJ...",
	"scope": "openid profile"
}
```

Tapu does not currently issue refresh tokens.

## Access tokens

Access tokens are JWTs signed with EdDSA.

Important claims:

```text
iss          Tapu issuer
sub          Immutable user UUID
aud          Service audience
exp          Expiration
iat          Issue time
nbf          Not-before time
jti          Unique token identifier
token_type   access_token
client_id    Requesting OAuth client
scope        Granted scopes
username     Optional profile claim
```

APIs must verify:

- Signature
- Allowed signing algorithm
- Issuer
- Audience
- Expiration
- `token_type=access_token`

Decoding a JWT is not validation.

## ID tokens

ID tokens identify the authenticated user to the OAuth client.

Their audience is the OAuth client ID, not the service audience.

Applications must verify:

- Signature
- Issuer
- Client audience
- Expiration
- Nonce
- Subject

## UserInfo

Request:

```bash
curl \
  --header "Authorization: Bearer ACCESS_TOKEN" \
  https://auth.example.com/oauth/userinfo
```

Example response:

```json
{
	"sub": "immutable-user-uuid",
	"preferred_username": "alice"
}
```

The username is returned only when the `profile` scope was granted.

## Browser CORS

Browser-only public clients may call the token and UserInfo endpoints from an origin derived from one of their registered redirect URIs.

Origins are matched exactly.

Discovery and JWKS contain public information and may be read cross-origin without credentials.

## Error responses

OAuth errors follow this structure:

```json
{
	"error": "invalid_grant",
	"error_description": "Authorization grant is invalid."
}
```

Common errors include:

- `invalid_request`
- `invalid_client`
- `invalid_grant`
- `invalid_scope`
- `unsupported_grant_type`
- `unsupported_response_type`
- `access_denied`
- `server_error`
