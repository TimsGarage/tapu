# Personal IAM

A small identity and access-management service for independently owned personal
applications.

## Current state

Phase 0 provides:

- Go HTTP server
- PostgreSQL
- Explicit SQL migrations
- Health checks
- Ed25519 signing-key generation
- Docker Compose deployment
- Tests and CI

Authentication is not implemented yet.

## Requirements

- Go matching the version in `go.mod`
- Docker with Docker Compose
- OpenSSL, or another secure random-password generator

## Initial setup

Create local environment configuration:

    cp .env.example .env

Replace the example PostgreSQL password in `.env`.

Generate the signing key:

    go run ./cmd/iam-keygen

Run tests:

    go test -race ./...

Start the stack:

    docker compose \
      --env-file .env \
      -f deploy/compose/compose.yml \
      up --build -d

Check container state:

    docker compose \
      --env-file .env \
      -f deploy/compose/compose.yml \
      ps

Test the endpoints:

    curl http://127.0.0.1:8080/health/live
    curl http://127.0.0.1:8080/health/ready

Both should return:

    {"status":"ok"}

## Security warning

The Phase 0 server is not ready for internet exposure.

Before deployment, it must be placed behind a correctly configured HTTPS reverse
proxy. User authentication and access control will be introduced in later
phases.

Never commit:

- `.env`
- Database passwords
- Signing keys
- Future client secrets
- Future session or refresh tokens
