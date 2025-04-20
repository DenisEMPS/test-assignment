-- +goose Up
-- +goose StatementBegin
CREATE TABLE refresh_sessions (
    user_id UUID NOT NULL,
    access_uuid UUID NOT NULL,
    PRIMARY KEY (access_uuid, user_id),
    refresh_hash TEXT NOT NULL,
    ip VARCHAR(45) NOT NULL,
    expires_at TIMESTAMPTZ NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS refresh_sessions;
-- +goose StatementEnd
