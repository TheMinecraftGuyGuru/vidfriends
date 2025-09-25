-- dev_seed.sql
-- Seed data to quickly populate a local VidFriends database with sample users,
-- friend connections, and video shares. Intended for development only.

BEGIN;

-- Users
INSERT INTO users (id, email, password_hash, created_at, updated_at) VALUES
  ('11111111-1111-1111-1111-111111111111', 'alice@example.com', '$2a$10$N9qo8uLOickgx2ZMRZoMyeIjZxYVSG3ZVSG7AV2XpFp7ZXDEh4XGa', NOW(), NOW()),
  ('22222222-2222-2222-2222-222222222222', 'bob@example.com', '$2a$10$N9qo8uLOickgx2ZMRZoMyeIjZxYVSG3ZVSG7AV2XpFp7ZXDEh4XGa', NOW(), NOW()),
  ('33333333-3333-3333-3333-333333333333', 'carol@example.com', '$2a$10$N9qo8uLOickgx2ZMRZoMyeIjZxYVSG3ZVSG7AV2XpFp7ZXDEh4XGa', NOW(), NOW())
ON CONFLICT (email) DO NOTHING;

-- Friend requests / relationships
INSERT INTO friend_requests (id, requester_id, receiver_id, status, created_at, responded_at) VALUES
  ('aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaaaa', '11111111-1111-1111-1111-111111111111', '22222222-2222-2222-2222-222222222222', 'accepted', NOW() - INTERVAL '3 days', NOW() - INTERVAL '3 days'),
  ('bbbbbbbb-bbbb-bbbb-bbbb-bbbbbbbbbbbb', '22222222-2222-2222-2222-222222222222', '33333333-3333-3333-3333-333333333333', 'pending', NOW() - INTERVAL '1 day', NULL)
ON CONFLICT (id) DO NOTHING;

-- Video shares
INSERT INTO video_shares (id, owner_id, url, title, description, thumbnail, created_at) VALUES
  ('cccccccc-cccc-cccc-cccc-cccccccccccc', '11111111-1111-1111-1111-111111111111',
   'https://www.youtube.com/watch?v=dQw4w9WgXcQ',
   'An 80s classic',
   'Sample share for local testing. Replace with your own links as needed.',
   'https://img.youtube.com/vi/dQw4w9WgXcQ/hqdefault.jpg',
   NOW() - INTERVAL '2 days'),
  ('dddddddd-dddd-dddd-dddd-dddddddddddd', '22222222-2222-2222-2222-222222222222',
   'https://vimeo.com/148751763',
   'Creative inspiration',
   'Another seeded video share to exercise metadata lookups.',
   NULL,
   NOW() - INTERVAL '12 hours')
ON CONFLICT (id) DO NOTHING;

COMMIT;
