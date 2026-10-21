FROM golang:1.27.2-alpine AS build

WORKDIR /src

RUN apk add --no-cache ca-certificates git

COPY go.mod go.sum ./
RUN go mod download

COPY . .

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

FROM alpine:3.21

RUN apk add --no-cache ca-certificates \
    && addgroup -S -g 10001 iam \
    && adduser -S -D -H -u 10001 -G iam iam

COPY --from=build /out/iam-server /app/iam-server
COPY --from=build /out/iam-migrate /app/iam-migrate
COPY --from=build /out/iamctl /app/iamctl

USER iam:iam

EXPOSE 8080

ENTRYPOINT ["/app/iam-server"]