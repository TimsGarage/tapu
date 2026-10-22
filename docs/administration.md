# Administration

Tapu administrators manage global users, services, access assignments, and OAuth clients.

The administration UI is available at:

```text
https://auth.example.com/admin
```

Administrative access requires an active IAM session and administrator status.

## Initial administrator

Create an administrator from the deployed container:

```bash
docker-compose run \
  --rm \
  --no-deps \
  iam \
  /app/iamctl \
  admin create \
  --username your-admin
```

Create at least two administrators to reduce lockout risk.

## Users

Administrators can:

- Create users
- Disable or reactivate users
- Grant or remove administrator status
- Reset passwords
- Grant or revoke service access

Users cannot register themselves.

### User identifiers

Every user has an immutable UUID. Applications should store this UUID from the token `sub` claim.

Do not use usernames as permanent identifiers.

### Disabling users

Disabling a user prevents:

- New logins
- New authorization codes
- New token exchanges
- UserInfo access

Previously issued JWT access tokens may remain valid until they expire.

## Services

A service represents an application or API protected by Tapu.

Important fields:

```text
Key       Internal stable name
Name      Human-readable name
Audience  Access-token audience
Enabled   Whether new authorization is allowed
```

Example:

```text
Key: project-a
Name: Project A
Audience: project-a-api
Enabled: true
```

Audiences must be unique and stable.

Changing an audience requires updating every API and OAuth client that depends on it.

## User service access

A user must be explicitly assigned to a service before Tapu issues a token for that service.

Service access is currently binary:

```text
Allowed
Not allowed
```

Project-specific roles and permissions remain the responsibility of the application.

## OAuth clients

OAuth clients represent applications that request authentication.

### Public clients

Use for browser-only applications.

Properties:

- No client secret
- Authorization Code with PKCE
- Tokens are accessible to browser JavaScript
- Exact redirect URIs are required

### Confidential clients

Use for server-side backends.

Properties:

- Has a client secret
- Uses HTTP Basic authentication at the token endpoint
- Uses Authorization Code with PKCE
- Can keep tokens out of browser JavaScript

Prefer confidential clients when an application has a backend.

## Redirect URIs

Redirect URIs must match exactly.

These are different:

```text
https://app.example.com/auth/callback
https://app.example.com/auth/callback/
http://app.example.com/auth/callback
https://www.app.example.com/auth/callback
```

Do not register wildcard redirect URIs.

Use HTTPS in production.

## Client-to-service access

Each OAuth client must be assigned to the service audiences it may request.

Authorization requires both:

- Client access to the service
- User access to the service

This prevents a client from requesting tokens for unrelated APIs.

## Client secrets

Confidential-client secrets are:

- Generated randomly
- Displayed once
- Stored only as hashes
- Never recoverable from Tapu

Store the displayed secret immediately in the application’s secret-management system.

Never place it in:

- Git
- Svelte code
- `PUBLIC_` or `VITE_` environment variables
- URLs
- Logs
- Support messages

## Secret rotation

To rotate a client secret:

1. Generate a new secret in the admin UI.
2. Save the new secret securely.
3. Update the application configuration.
4. Restart or redeploy the application.
5. Verify authentication.
6. Review audit logs.

Rotation replaces the previous secret immediately. Coordinate deployment to avoid unnecessary downtime.

## Recommended application registration

### Go backend with SvelteKit

```text
Client type: confidential
Redirect URI: https://api.example.com/auth/callback
Audience: project-api
```

The frontend origin is not the OAuth callback. The Go backend receives and exchanges the authorization code.

### Browser-only SvelteKit

```text
Client type: public
Redirect URI: https://app.example.com/auth/callback
Audience: project-api
```

No client secret should exist.

## Administrative safety

Before disabling or demoting an administrator:

- Confirm another active administrator exists.
- Confirm that administrator can log in.
- Confirm that administrator can access `/admin`.

Do not remove the final working administrator.

## Audit review

Review audit events after:

- User creation
- Password resets
- Administrator changes
- Service-access changes
- OAuth client creation
- Client-secret rotation
- Unexpected token issuance

Audit logs should never contain raw credentials or tokens.
