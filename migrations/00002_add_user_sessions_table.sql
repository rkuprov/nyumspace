-- +goose Up
CREATE TABLE user_sessions (
    session_token TEXT PRIMARY KEY,
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    expires_at TIMESTAMPTZ NOT NULL
);

-- +goose Down
DROP TABLE IF EXISTS user_sessions;

