## Production deployment

The server does not require the source repository, Go, or Node.js.

Required files:

- `compose.yml`
- `.env`

The IAM image is pulled from GitHub Container Registry.

Create the deployment directory:

    mkdir -p ~/tapu
    cd ~/tapu

Place the standalone production `compose.yml` in this directory.

Create `.env`:

    IAM_IMAGE=ghcr.io/timsgarage/tapu:v0.1.0
    IAM_PUBLIC_URL=https://auth.example.com
    POSTGRES_DB=iam
    POSTGRES_USER=iam
    POSTGRES_PASSWORD=<random hexadecimal password>

Start the deployment:

    docker-compose pull
    docker-compose up -d

Create an administrator:

    docker-compose run \
      --rm \
      --no-deps \
      iam \
      /app/iamctl \
      admin create \
      --username your-admin

Update by changing `IAM_IMAGE` to a newer version and running:

    docker-compose pull
    docker-compose up -d

Back up both PostgreSQL and the signing-key volume.
