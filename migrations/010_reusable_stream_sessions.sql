-- Migration 010 - Persistent reusable lives with isolated broadcast sessions.
-- A stream row is the broadcaster's reusable channel. active_session_id is
-- renewed every time that channel goes live, so delayed chunks from an older
-- session cannot revive or corrupt the current broadcast.

ALTER TABLE streams
    ADD COLUMN IF NOT EXISTS active_session_id UUID;

UPDATE streams
SET active_session_id = gen_random_uuid()
WHERE status = 'live' AND active_session_id IS NULL;

ALTER TABLE streams
    DROP CONSTRAINT IF EXISTS streams_session_matches_status;

ALTER TABLE streams
    ADD CONSTRAINT streams_session_matches_status CHECK (
        (status = 'live' AND active_session_id IS NOT NULL)
        OR (status = 'ended' AND active_session_id IS NULL)
    );

CREATE INDEX IF NOT EXISTS idx_streams_active_session
    ON streams(active_session_id)
    WHERE active_session_id IS NOT NULL;

DROP POLICY IF EXISTS streams_delete_own ON streams;
CREATE POLICY streams_delete_own ON streams
    FOR DELETE USING (broadcaster_id = auth.uid());
