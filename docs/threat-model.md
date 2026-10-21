# IAM threat model

## Purpose

The IAM provides centralized authentication for independently owned personal
applications.

The IAM owns global identities and service access. Each application owns its
application-specific user data.

## Protected assets

- Password hashes
- IAM sessions
- OAuth authorization codes
- Refresh tokens, when introduced
- Client secrets
- Signing private keys
- User-to-service access assignments
- Administrator privileges
- Audit records
- PostgreSQL data

## Trust boundaries

### Public internet to reverse proxy

All production traffic must enter through a TLS-enabled reverse proxy.

The IAM server must not be exposed directly to the internet.

### Reverse proxy to IAM

The reverse proxy forwards requests to the IAM over a private Docker network or
the local host.

Forwarded headers must only be trusted when requests can come exclusively from
the configured reverse proxy.

### IAM to PostgreSQL

PostgreSQL is available only on an internal Docker network. Its port is not
published to the host or internet.

### IAM to client applications

Client applications are not trusted merely because they share a parent domain.

Every client must have its own identity, exact redirect URI registration, and
permitted service audiences.

## Initial threat scenarios

### Credential guessing

Mitigations planned for Phase 1:

- Argon2id password hashing
- Login rate limiting
- Generic login errors
- Strong minimum password length
- Administrative user creation only

### Signing-key theft

Current mitigations:

- Private key excluded from Git
- Private key mounted as a Docker secret
- Container runs as a non-root user
- Key is validated at startup
- Key is not written to logs

Further mitigation:

- Document secure backup and rotation
- Keep previous public keys during rotation

### Database compromise

Password hashes and secret hashes must remain expensive or impossible to use
directly.

Raw passwords, session tokens, authorization codes, refresh tokens, and client
secrets must never be stored in the database.

### Subdomain compromise

Authentication cookies must be scoped to the IAM host rather than the parent
domain.

Applications use redirect-based SSO instead of sharing cookies.

### Malicious OAuth client

Planned mitigations:

- Exact redirect URI matching
- Authorization Code flow
- PKCE
- Client-specific service allowlists
- Audience-specific tokens
- No wildcard redirect URIs

### Cross-site request forgery

Cookie-authenticated state-changing requests will require CSRF protection.

Cookies will use Secure, HttpOnly, and appropriate SameSite attributes.

### Cross-site scripting

Planned mitigations:

- Content Security Policy
- No tokens exposed to JavaScript in backend-for-frontend applications
- No arbitrary HTML or CSS in client branding
- Context-aware output escaping

### Token theft and replay

Planned mitigations:

- Short-lived access tokens
- Opaque, hashed refresh tokens
- Refresh-token rotation
- Token-family revocation after reuse
- Audience verification
- Issuer verification
- Enforced signing algorithm

### Administrative lockout

At least two administrator accounts should exist once user management is
implemented.

A local command-line recovery procedure will be provided.

## Assumptions

- The home server operating system is maintained and patched.
- Docker access is equivalent to root-level administrative access.
- The reverse proxy provides valid HTTPS.
- PostgreSQL and the IAM are not directly exposed to the internet.
- Backups are protected separately from the running server.
- Applications validate issuer, audience, signature, and expiration.