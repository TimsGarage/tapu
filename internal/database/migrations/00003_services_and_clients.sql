-- +goose Up

CREATE TABLE services (
    id         UUID PRIMARY KEY,
    key        TEXT NOT NULL,
    name       TEXT NOT NULL,
    audience   TEXT NOT NULL,
    enabled    BOOLEAN NOT NULL DEFAULT TRUE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,

    CONSTRAINT services_key_not_empty
        CHECK (length(key) > 0),

    CONSTRAINT services_name_not_empty
        CHECK (length(name) > 0),

    CONSTRAINT services_audience_not_empty
        CHECK (length(audience) > 0)
);

CREATE UNIQUE INDEX services_key_unique
    ON services (key);

CREATE UNIQUE INDEX services_audience_unique
    ON services (audience);

CREATE TABLE user_service_access (
    user_id    UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    service_id UUID NOT NULL REFERENCES services(id) ON DELETE CASCADE,
    created_by UUID REFERENCES users(id) ON DELETE SET NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,

    PRIMARY KEY (user_id, service_id)
);

CREATE INDEX user_service_access_service_id_index
    ON user_service_access (service_id);

CREATE TABLE oauth_clients (
    id          UUID PRIMARY KEY,
    client_id   TEXT NOT NULL,
    name        TEXT NOT NULL,
    client_type TEXT NOT NULL,
    secret_hash BYTEA,
    enabled     BOOLEAN NOT NULL DEFAULT TRUE,
    created_at  TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at  TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,

    CONSTRAINT oauth_clients_name_not_empty
        CHECK (length(name) > 0),

    CONSTRAINT oauth_clients_type_valid
        CHECK (client_type IN ('public', 'confidential')),

    CONSTRAINT oauth_clients_secret_valid
        CHECK (
            (client_type = 'public' AND secret_hash IS NULL)
            OR
            (client_type = 'confidential' AND secret_hash IS NOT NULL)
        )
);

CREATE UNIQUE INDEX oauth_clients_client_id_unique
    ON oauth_clients (client_id);

CREATE TABLE oauth_client_redirect_uris (
    id              UUID PRIMARY KEY,
    oauth_client_id UUID NOT NULL
        REFERENCES oauth_clients(id) ON DELETE CASCADE,
    redirect_uri    TEXT NOT NULL,

    CONSTRAINT oauth_redirect_uri_not_empty
        CHECK (length(redirect_uri) > 0),

    UNIQUE (oauth_client_id, redirect_uri)
);

CREATE INDEX oauth_client_redirect_uris_client_index
    ON oauth_client_redirect_uris (oauth_client_id);

CREATE TABLE oauth_client_services (
    oauth_client_id UUID NOT NULL
        REFERENCES oauth_clients(id) ON DELETE CASCADE,
    service_id      UUID NOT NULL
        REFERENCES services(id) ON DELETE CASCADE,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,

    PRIMARY KEY (oauth_client_id, service_id)
);

CREATE INDEX oauth_client_services_service_index
    ON oauth_client_services (service_id);

-- +goose Down

DROP TABLE oauth_client_services;
DROP TABLE oauth_client_redirect_uris;
DROP TABLE oauth_clients;
DROP TABLE user_service_access;
DROP TABLE services;