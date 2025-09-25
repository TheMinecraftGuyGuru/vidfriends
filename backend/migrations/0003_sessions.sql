-- 0003_sessions.sql
-- Persist issued refresh tokens so sessions survive process restarts.

BEGIN;

CREATE TABLE IF NOT EXISTS sessions (
    refresh_token TEXT PRIMARY KEY,
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    expires_at TIMESTAMPTZ NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS sessions_user_id_idx ON sessions(user_id);

COMMIT;
