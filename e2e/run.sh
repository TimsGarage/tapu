#!/usr/bin/env sh

set -eu

cd "$(dirname "$0")/.."

docker-compose \
  -p personal-iam-e2e \
  --env-file e2e/.env \
  -f deploy/compose/compose.yml \
  -f deploy/compose/compose.e2e.yml \
  up --build -d

trap '
  docker-compose \
    -p personal-iam-e2e \
    --env-file e2e/.env \
    -f deploy/compose/compose.yml \
    -f deploy/compose/compose.e2e.yml \
    down
' EXIT

cd e2e

npm run seed
npm test