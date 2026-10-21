-- +goose Up

CREATE TABLE oauth_authorization_codes (
    id                  UUID PRIMARY KEY,
    code_hash           BYTEA NOT NULL,
    user_id             UUID NOT NULL
        REFERENCES users(id) ON DELETE CASCADE,
    oauth_client_id     UUID NOT NULL
        REFERENCES oauth_clients(id) ON DELETE CASCADE,
    service_id          UUID NOT NULL
        REFERENCES services(id) ON DELETE CASCADE,
    redirect_uri        TEXT NOT NULL,
    scope               TEXT NOT NULL,
    nonce               TEXT,
    pkce_challenge      TEXT NOT NULL,
    pkce_method         TEXT NOT NULL,
    created_at          TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    expires_at          TIMESTAMPTZ NOT NULL,
    used_at             TIMESTAMPTZ,

    CONSTRAINT oauth_authorization_codes_pkce_method_valid
        CHECK (pkce_method = 'S256'),

    CONSTRAINT oauth_authorization_codes_redirect_uri_not_empty
        CHECK (length(redirect_uri) > 0),

    CONSTRAINT oauth_authorization_codes_scope_not_empty
        CHECK (length(scope) > 0),

    CONSTRAINT oauth_authorization_codes_challenge_not_empty
        CHECK (length(pkce_challenge) > 0)
);

CREATE UNIQUE INDEX oauth_authorization_codes_code_hash_unique
    ON oauth_authorization_codes (code_hash);

CREATE INDEX oauth_authorization_codes_expiration_index
    ON oauth_authorization_codes (expires_at);

CREATE INDEX oauth_authorization_codes_client_index
    ON oauth_authorization_codes (oauth_client_id);

CREATE INDEX oauth_authorization_codes_user_index
    ON oauth_authorization_codes (user_id);

-- +goose Down

DROP TABLE oauth_authorization_codes;