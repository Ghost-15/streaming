-- Migration 004 — Ajout first_name/last_name sur users + FK playlist_tracks → tracks

-- ── Users : ajout prénom et nom ──────────────────────────────────────────────
ALTER TABLE users
    ADD COLUMN IF NOT EXISTS first_name TEXT NOT NULL DEFAULT '',
    ADD COLUMN IF NOT EXISTS last_name  TEXT NOT NULL DEFAULT '';

-- ── playlist_tracks : ajout FK vers tracks ───────────────────────────────────
-- On supprime l'ancienne contrainte si elle existe, puis on recrée proprement.
ALTER TABLE playlist_tracks
    DROP CONSTRAINT IF EXISTS playlist_tracks_track_id_fkey;

ALTER TABLE playlist_tracks
    ADD CONSTRAINT playlist_tracks_track_id_fkey
    FOREIGN KEY (track_id) REFERENCES tracks(id) ON DELETE CASCADE;
