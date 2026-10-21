-- +goose Up

CREATE TABLE users (
    id                  UUID PRIMARY KEY,
    username            TEXT NOT NULL,
    password_hash       TEXT NOT NULL,
    is_admin            BOOLEAN NOT NULL DEFAULT FALSE,
    status              TEXT NOT NULL DEFAULT 'active',
    password_changed_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    created_at          TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at          TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,

    CONSTRAINT users_username_not_empty
        CHECK (length(username) > 0),

    CONSTRAINT users_status_valid
        CHECK (status IN ('active', 'disabled'))
);

CREATE UNIQUE INDEX users_username_unique
    ON users (username);

CREATE TABLE iam_sessions (
    id                  UUID PRIMARY KEY,
    user_id             UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    token_hash          BYTEA NOT NULL,
    csrf_token_hash     BYTEA NOT NULL,
    created_at          TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    last_seen_at        TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    expires_at          TIMESTAMPTZ NOT NULL,
    absolute_expires_at TIMESTAMPTZ NOT NULL
);

CREATE UNIQUE INDEX iam_sessions_token_hash_unique
    ON iam_sessions (token_hash);

CREATE INDEX iam_sessions_user_id_index
    ON iam_sessions (user_id);

CREATE INDEX iam_sessions_expiration_index
    ON iam_sessions (expires_at, absolute_expires_at);

CREATE TABLE audit_events (
    id          UUID PRIMARY KEY,
    occurred_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    actor_type  TEXT NOT NULL,
    actor_id    UUID,
    action      TEXT NOT NULL,
    target_type TEXT,
    target_id   UUID,
    request_id  TEXT,
    source_ip   TEXT,
    metadata    JSONB NOT NULL DEFAULT '{}'::jsonb
);

CREATE INDEX audit_events_occurred_at_index
    ON audit_events (occurred_at DESC);

CREATE INDEX audit_events_actor_index
    ON audit_events (actor_type, actor_id);

CREATE INDEX audit_events_action_index
    ON audit_events (action);

-- +goose Down

DROP TABLE audit_events;
DROP TABLE iam_sessions;
DROP TABLE users;