# Operations

This document covers routine operation, monitoring, backups, cleanup, and incident handling.

## Service health

Liveness:

```text
GET /health/live
```

Readiness:

```text
GET /health/ready
```

Use readiness for deployment and reverse-proxy health checks.

Example:

```bash
curl --fail \
  https://auth.example.com/health/ready
```

## Container status

```bash
docker-compose ps
```

View all logs:

```bash
docker-compose logs -f
```

View selected services:

```bash
docker-compose logs -f iam postgres
```

After an update, always inspect:

```bash
docker-compose logs migrate iam
```

## Backups

Back up both PostgreSQL and the signing key.

### PostgreSQL

```bash
docker-compose exec -T postgres \
  pg_dump \
  -U iam \
  -d iam \
  > tapu-database.sql
```

### Signing key

```bash
docker run \
  --rm \
  -v tapu_signing-keys:/source:ro \
  -v "$PWD":/backup \
  alpine \
  tar -czf /backup/tapu-signing-keys.tar.gz \
  -C /source .
```

Store backups encrypted and outside the server.

Test restoration regularly.

## Restore procedure

1. Stop Tapu.
2. Restore the signing-key volume.
3. Restore PostgreSQL into an empty compatible database.
4. Set the original issuer URL.
5. Start migrations and Tapu.
6. Verify discovery, JWKS, login, and application authentication.

Do not generate a new signing key during recovery if the original key is available.

## Updates

Before updating:

1. Read the release notes.
2. Back up PostgreSQL.
3. Back up the signing key.
4. Record the current image version.

Update:

```bash
docker-compose pull
docker-compose up -d
```

Verify:

```bash
docker-compose ps
docker-compose logs migrate iam
curl --fail https://auth.example.com/health/ready
```

## Cleanup

Expired or consumed authorization codes may be removed:

```sql
DELETE FROM oauth_authorization_codes
WHERE expires_at <= CURRENT_TIMESTAMP
   OR used_at IS NOT NULL;
```

Run cleanup periodically until an internal scheduled cleanup worker exists.

Session cleanup should follow the same principle: expired sessions may be deleted without affecting active sessions.

Always back up the database before introducing new automated cleanup tasks.

## Monitoring

At minimum, monitor:

- IAM readiness
- Container restarts
- PostgreSQL health
- Disk usage
- Failed login volume
- HTTP 5xx responses
- Migration failures
- Certificate renewal
- Backup completion

Never place passwords, client secrets, cookies, authorization codes, or tokens in monitoring labels or log messages.

## Audit events

Tapu records security-relevant events such as:

- User creation and updates
- Password changes
- Service-access changes
- OAuth client changes
- Client-secret rotation
- Authorization-code issuance
- Token issuance

Audit events must not contain credentials or raw tokens.

Review audit events after unexpected account or access changes.

## Signing-key handling

The private signing key must:

- Remain outside Git
- Be readable only by the IAM
- Be backed up securely
- Never appear in logs
- Never be published through JWKS

The JWKS endpoint must contain only public key material.

Tapu currently uses one signing key. Replacing it immediately invalidates verification of tokens signed by the previous key.

## Database credentials

Use a unique, random PostgreSQL password.

PostgreSQL should be reachable only through the private Compose network. Do not publish port 5432 to the internet.

## Incident response

### Suspected client-secret exposure

1. Rotate the client secret in the admin UI.
2. Update the application secret.
3. Restart the application.
4. Review audit and access logs.
5. Disable the client if the impact is unclear.

### Suspected user-password exposure

1. Disable the user if necessary.
2. Reset the password.
3. Remove active IAM sessions if supported operationally.
4. Review service access and audit events.

### Suspected signing-key exposure

1. Treat all issued tokens as potentially forgeable.
2. Stop public token issuance.
3. Generate a new signing key.
4. Redeploy Tapu.
5. Restart relying applications if they cache JWKS.
6. Review access logs.
7. Notify affected users or operators.

Current Tapu releases do not provide graceful multi-key rotation, so emergency replacement invalidates previous tokens.

### Database compromise

Assume exposure of:

- Identity data
- Password hashes
- Client-secret hashes
- Session-token hashes
- Access assignments
- Audit history

Rotate database credentials, signing keys, client secrets, and user passwords as appropriate.

## Routine checklist

Weekly or automated:

- Verify readiness
- Check storage usage
- Check container restarts
- Confirm backups completed
- Review migration and error logs

Monthly:

- Test a backup restoration
- Review administrators
- Review enabled OAuth clients
- Review redirect URIs
- Review user-to-service access
- Update container images and dependencies
