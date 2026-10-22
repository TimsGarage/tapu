# Deployment

Tapu is distributed as a container image through GitHub Container Registry.

A production server needs only:

- Docker
- `docker-compose`
- A production `compose.yml`
- A `.env` file
- A domain pointing to the server

The source repository, Go, and Node.js are not required on the server.

## Requirements

Prepare:

- A Linux server
- Docker and `docker-compose`
- A domain such as `auth.example.com`
- Ports 80 and 443 open
- Persistent storage and backups

Tapu must be exposed through HTTPS in production.

## Deployment files

Create a directory on the server:

```bash
mkdir -p ~/tapu
cd ~/tapu
```

Copy the standalone production Compose file from:

```text
deploy/production/compose.yml
```

to the server as:

```text
~/tapu/compose.yml
```

Create `.env`:

```dotenv
IAM_IMAGE=ghcr.io/timsgarage/tapu:v0.1.0
IAM_PUBLIC_URL=https://auth.example.com

POSTGRES_DB=iam
POSTGRES_USER=iam
POSTGRES_PASSWORD=replace-with-a-random-hexadecimal-password
```

Generate a database password:

```bash
openssl rand -hex 32
```

Protect the environment file:

```bash
chmod 600 .env
```

Use a versioned image tag in production instead of `latest`.

## Private registry access

If the GHCR package is private, create a GitHub token with `read:packages` and log in:

```bash
printf '%s' "$GITHUB_CONTAINER_TOKEN" |
docker login ghcr.io \
  --username YOUR_GITHUB_USERNAME \
  --password-stdin
```

Registry credentials should not be placed in `compose.yml`.

## Start Tapu

Pull the images:

```bash
docker-compose pull
```

Start the services:

```bash
docker-compose up -d
```

The production stack starts:

- PostgreSQL
- Signing-key initialization
- Database migrations
- Tapu IAM
- HTTPS reverse proxy

Check the status:

```bash
docker-compose ps
```

Inspect logs:

```bash
docker-compose logs migrate iam caddy
```

## Verify the deployment

```bash
curl --fail \
  https://auth.example.com/health/ready
```

Check OpenID Connect discovery:

```bash
curl --fail \
  https://auth.example.com/.well-known/openid-configuration
```

Check the public signing key:

```bash
curl --fail \
  https://auth.example.com/.well-known/jwks.json
```

The discovery document’s issuer must exactly equal:

```text
https://auth.example.com
```

## Create administrators

Create the first administrator:

```bash
docker-compose run \
  --rm \
  --no-deps \
  iam \
  /app/iamctl \
  admin create \
  --username your-admin
```

Create at least two administrators to reduce the risk of administrative lockout.

## Update Tapu

Back up the database before updating.

Change the image version in `.env`:

```dotenv
IAM_IMAGE=ghcr.io/timsgarage/tapu:v0.2.0
```

Then run:

```bash
docker-compose pull
docker-compose up -d
```

Check migration and server logs:

```bash
docker-compose logs migrate iam
```

## Rollback

Restore the previous image tag:

```dotenv
IAM_IMAGE=ghcr.io/timsgarage/tapu:v0.1.0
```

Then run:

```bash
docker-compose pull
docker-compose up -d
```

Database migrations may not be backward-compatible. A full rollback may require restoring a matching database backup.

## Backups

At minimum, back up:

- PostgreSQL
- The signing-key volume
- `.env`
- Reverse-proxy configuration

Create a database backup:

```bash
docker-compose exec -T postgres \
  pg_dump \
  -U iam \
  -d iam \
  > tapu-database.sql
```

Back up the signing-key volume:

```bash
docker run \
  --rm \
  -v tapu_signing-keys:/source:ro \
  -v "$PWD":/backup \
  alpine \
  tar -czf /backup/tapu-signing-keys.tar.gz \
  -C /source .
```

Protect backup files. The database contains password hashes and identity data, while the signing key controls Tapu’s token identity.

## Production checklist

- HTTPS is active
- `APP_ENV=production`
- The issuer uses the public HTTPS origin
- PostgreSQL is not exposed publicly
- Image versions are pinned
- Signing keys and database data are persistent
- Backups are automated and tested
- At least two administrators exist
- Redirect URIs use exact HTTPS URLs
- Client secrets are stored outside source control
- Health checks and logs are monitored
