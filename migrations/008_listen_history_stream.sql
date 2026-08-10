-- Migration 008 — Autoriser un historique d'écoute lié à un stream sans track.
-- On remplace la PK (user_id, track_id, listened_at) par une PK surrogate,
-- et on rend track_id nullable (un join de stream n'a pas forcément de track).

ALTER TABLE listen_history
    ADD COLUMN IF NOT EXISTS id UUID NOT NULL DEFAULT gen_random_uuid();

ALTER TABLE listen_history
    DROP CONSTRAINT IF EXISTS listen_history_pkey;

ALTER TABLE listen_history
    ALTER COLUMN track_id DROP NOT NULL;

ALTER TABLE listen_history
    ADD CONSTRAINT listen_history_pkey PRIMARY KEY (id);
