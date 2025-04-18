-- +goose Up
-- +goose StatementBegin
CREATE TABLE refresh_sessions (
    id SERIAL PRIMARY KEY,
    user_id UUID NOT NULL,
    access_uuid UUID NOT NULL,
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
