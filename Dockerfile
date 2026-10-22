FROM node:22-alpine AS web-build

WORKDIR /src/web/admin

COPY web/admin/package.json web/admin/package-lock.json ./
RUN npm ci

COPY web/admin/ ./
RUN npm run build

FROM golang:1.27.2-alpine AS go-build

WORKDIR /src

RUN apk add --no-cache ca-certificates git

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN rm -rf internal/adminui/dist

COPY --from=web-build \
    /src/internal/adminui/dist \
    /src/internal/adminui/dist

RUN CGO_ENABLED=0 GOOS=linux go build \
    -trimpath \
    -ldflags="-s -w" \
    -o /out/iam-server \
    ./cmd/iam-server

RUN CGO_ENABLED=0 GOOS=linux go build \
    -trimpath \
    -ldflags="-s -w" \
    -o /out/iam-migrate \
    ./cmd/iam-migrate

RUN CGO_ENABLED=0 GOOS=linux go build \
    -trimpath \
    -ldflags="-s -w" \
    -o /out/iamctl \
    ./cmd/iamctl

RUN CGO_ENABLED=0 GOOS=linux go build \
    -trimpath \
    -ldflags="-s -w" \
    -o /out/iam-keygen \
    ./cmd/iam-keygen

FROM alpine:3.21

RUN apk add --no-cache ca-certificates wget \
    && addgroup -S -g 10001 iam \
    && adduser -S -D -H -u 10001 -G iam iam

WORKDIR /app

COPY --from=go-build /out/iam-server /app/iam-server
COPY --from=go-build /out/iam-migrate /app/iam-migrate
COPY --from=go-build /out/iamctl /app/iamctl
COPY --from=go-build /out/iam-keygen /app/iam-keygen

USER 10001:10001

EXPOSE 8080

CMD ["/app/iam-server"]