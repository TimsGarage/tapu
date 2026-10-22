# End-to-End Tests

The E2E suite uses Playwright and an isolated Docker Compose environment to test:

- Login and logout
- Admin access control
- OAuth Authorization Code with PKCE
- Browser SSO
- Invalid PKCE rejection
- Credentialed CORS

## Architecture

```text
Playwright Chromium
├── IAM                 http://127.0.0.1:8080
├── Example Go API      http://127.0.0.1:3000
└── OAuth receiver      http://127.0.0.1:4100

PostgreSQL              127.0.0.1:55432
```

The suite uses the Docker Compose project `personal-iam-e2e`, keeping its database volume separate from normal development.

## Files

```text
e2e/
├── receiver/server.mjs       # Public OAuth client
├── scripts/seed.mjs          # Creates deterministic test data
├── tests/
│   ├── admin.spec.js
│   ├── auth.spec.js
│   ├── cors.spec.js
│   └── oauth.spec.js
├── .env.example
├── package.json
├── playwright.config.js
└── run.sh

deploy/compose/
├── compose.yml
└── compose.e2e.yml
```

## Prerequisites

Install:

- Docker and `docker-compose`
- Go
- Node.js and npm
- Playwright Chromium

```bash
cd e2e
npm install
npx playwright install chromium
cd ..
```

On Linux or in CI:

```bash
cd e2e
npx playwright install --with-deps chromium
cd ..
```

## Configuration

Create the local environment file:

```bash
cp e2e/.env.example e2e/.env
```

Required values:

```dotenv
POSTGRES_DB=iam
POSTGRES_USER=iam
POSTGRES_PASSWORD=e2e-local-postgres-password
E2E_DATABASE_URL=postgres://iam:e2e-local-postgres-password@127.0.0.1:55432/iam

E2E_IAM_URL=http://127.0.0.1:8080
E2E_RECEIVER_URL=http://127.0.0.1:4100

E2E_USERNAME=e2e-user
E2E_PASSWORD=e2e-user-password-123
E2E_ADMIN_USERNAME=e2e-admin
E2E_ADMIN_PASSWORD=e2e-admin-password-123

E2E_SERVICE_KEY=e2e-project
E2E_SERVICE_NAME=E2E Project
E2E_AUDIENCE=e2e-project-api
E2E_CLIENT_NAME=E2E Browser Client
E2E_REDIRECT_URI=http://127.0.0.1:4100/callback
```

`POSTGRES_PASSWORD` and the password in `E2E_DATABASE_URL` must match.

Never commit:

```text
e2e/.env
e2e/.generated.json
e2e/node_modules/
e2e/playwright-report/
e2e/test-results/
```

## Running the suite

For a clean run:

```bash
make e2e-reset
make e2e-up
make e2e-seed
make e2e-test
```

Other useful commands:

```bash
make e2e-down       # Stop containers, preserve test data
make e2e-headed     # Run tests with a visible browser
make e2e-report     # Open the Playwright HTML report
```

Verify IAM readiness:

```bash
curl --fail http://127.0.0.1:8080/health/ready
```

## Seed data

`make e2e-seed` creates:

- Normal user `e2e-user`
- Administrator `e2e-admin`
- Service `e2e-project-api`
- User-to-service access
- Public OAuth client for the test receiver
- Confidential OAuth client for the example Go API

Generated client credentials are written to:

```text
e2e/.generated.json
```

This file is test-only and must remain ignored by Git.

## Running selected tests

Run one file:

```bash
cd e2e
npm test -- tests/oauth.spec.js
```

Run one matching test:

```bash
cd e2e
npm test -- --grep "uses IAM SSO"
```

Debug interactively:

```bash
cd e2e
npm run test:debug
```

## Test coverage

### Authentication

- Invalid credentials return a generic error
- Valid credentials create an IAM session
- Logout removes the session
- Unauthenticated account access redirects to login

### Administration

- Administrators can access `/admin`
- Normal users receive HTTP 403

### OAuth and OIDC

- Authorization Code with PKCE completes successfully
- Existing IAM sessions provide SSO
- Incorrect PKCE verifiers return `invalid_grant`

### CORS

- Unknown origins are rejected
- Allowed credentialed preflight requests succeed

## Failure artifacts

Playwright stores failure information under:

```text
e2e/test-results/
```

Artifacts may include:

- Screenshot
- Video
- Network and DOM trace
- Error context

Open a trace:

```bash
cd e2e
npx playwright show-trace test-results/<test>/trace.zip
```

## Troubleshooting

### PostgreSQL password missing

Ensure E2E Compose commands use:

```text
--env-file e2e/.env
```

Then reset the failed volume:

```bash
make e2e-reset
make e2e-up
```

### Login remains on `/login`

Usually the seeded password does not match the configured password. Reset and reseed:

```bash
make e2e-reset
make e2e-up
make e2e-seed
```

### Port already in use

Check local listeners:

```bash
lsof -nP -iTCP:3000 -sTCP:LISTEN
lsof -nP -iTCP:4100 -sTCP:LISTEN
lsof -nP -iTCP:8080 -sTCP:LISTEN
lsof -nP -iTCP:55432 -sTCP:LISTEN
```

### CORS tests report connection refused

The example API did not start on port `3000`. Run the failed test in debug mode and inspect Playwright’s `webServer` output.

## Security

The E2E environment is for local testing only:

- Use test-only passwords.
- Never reuse production credentials.
- Keep PostgreSQL bound to `127.0.0.1`.
- Never expose E2E services to the internet.
- Do not commit `.env` or `.generated.json`.
- Do not run destructive E2E tests against a development or production database.
