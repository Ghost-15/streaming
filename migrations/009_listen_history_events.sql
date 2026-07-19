-- Migration 009 - Track stream listen history event type.
-- Stream joins and leaves are stored as explicit events for analytics.

ALTER TABLE listen_history
    ADD COLUMN IF NOT EXISTS event_type TEXT NOT NULL DEFAULT 'join';

ALTER TABLE listen_history
    DROP CONSTRAINT IF EXISTS listen_history_event_type_check;

ALTER TABLE listen_history
    ADD CONSTRAINT listen_history_event_type_check
    CHECK (event_type IN ('join', 'leave'));

CREATE INDEX IF NOT EXISTS idx_listen_history_stream_event
    ON listen_history(stream_id, event_type);
