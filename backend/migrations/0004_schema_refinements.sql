-- 0004_schema_refinements.sql
-- Tighten constraints and performance characteristics for core VidFriends tables.

BEGIN;

-- Automatically maintain updated_at timestamps on users.
CREATE OR REPLACE FUNCTION set_updated_at_timestamp()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = NOW();
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

DROP TRIGGER IF EXISTS users_set_updated_at ON users;
CREATE TRIGGER users_set_updated_at
BEFORE UPDATE ON users
FOR EACH ROW
EXECUTE FUNCTION set_updated_at_timestamp();

-- Prevent nonsensical friend requests and ensure uniqueness regardless of direction.
ALTER TABLE friend_requests
    ADD CONSTRAINT friend_requests_no_self CHECK (requester_id <> receiver_id);

CREATE UNIQUE INDEX IF NOT EXISTS friend_requests_any_pair_idx
    ON friend_requests (LEAST(requester_id, receiver_id), GREATEST(requester_id, receiver_id));

CREATE INDEX IF NOT EXISTS friend_requests_created_at_idx ON friend_requests (created_at DESC);

-- Ensure session expiry lookups remain fast for cleanup routines.
CREATE INDEX IF NOT EXISTS sessions_expires_at_idx ON sessions (expires_at);

-- Harden video share data quality and query performance.
ALTER TABLE video_shares
    ADD CONSTRAINT video_shares_url_check CHECK (url ~ '^https?://');

CREATE UNIQUE INDEX IF NOT EXISTS video_shares_owner_url_idx
    ON video_shares (owner_id, url);

CREATE INDEX IF NOT EXISTS video_shares_created_at_idx ON video_shares (created_at DESC);
CREATE INDEX IF NOT EXISTS video_shares_owner_created_at_idx ON video_shares (owner_id, created_at DESC);

COMMIT;
