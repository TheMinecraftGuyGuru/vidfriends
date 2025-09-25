-- 0002_friend_request_constraints.sql
-- Enforce friend request status values and uniqueness constraints.

BEGIN;

ALTER TABLE friend_requests
    ADD CONSTRAINT friend_requests_status_check CHECK (status IN ('pending', 'accepted', 'blocked')),
    ADD CONSTRAINT friend_requests_unique_pair UNIQUE (requester_id, receiver_id);

CREATE INDEX IF NOT EXISTS friend_requests_receiver_idx ON friend_requests (receiver_id);
CREATE INDEX IF NOT EXISTS friend_requests_requester_idx ON friend_requests (requester_id);

COMMIT;
