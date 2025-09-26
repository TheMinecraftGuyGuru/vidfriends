-- Add asset ingestion tracking columns for video shares.
BEGIN;

ALTER TABLE video_shares
    ADD COLUMN IF NOT EXISTS asset_url TEXT NOT NULL DEFAULT '',
    ADD COLUMN IF NOT EXISTS asset_status TEXT NOT NULL DEFAULT 'pending',
    ADD COLUMN IF NOT EXISTS asset_size BIGINT NOT NULL DEFAULT 0;

COMMIT;
