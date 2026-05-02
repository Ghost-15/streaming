-- Migration 002 — Row Level Security policies
-- RLS garantit l'isolation des données au niveau DB, indépendamment du backend.

ALTER TABLE users          ENABLE ROW LEVEL SECURITY;
ALTER TABLE tracks         ENABLE ROW LEVEL SECURITY;
ALTER TABLE streams        ENABLE ROW LEVEL SECURITY;
ALTER TABLE playlists      ENABLE ROW LEVEL SECURITY;
ALTER TABLE playlist_tracks ENABLE ROW LEVEL SECURITY;

-- ─────────────────────────────────────────────────────────────
-- users
-- ─────────────────────────────────────────────────────────────

CREATE POLICY users_select_own ON users
    FOR SELECT USING (id = auth.uid());

CREATE POLICY users_update_own ON users
    FOR UPDATE USING (id = auth.uid());

CREATE POLICY users_admin_all ON users
    FOR ALL USING (
        EXISTS (SELECT 1 FROM users WHERE id = auth.uid() AND role = 'admin')
    );

-- ─────────────────────────────────────────────────────────────
-- tracks
-- ─────────────────────────────────────────────────────────────

-- Tout le monde peut lire les tracks (lecture publique)
CREATE POLICY tracks_select_all ON tracks
    FOR SELECT USING (TRUE);

-- Seul l'uploader peut modifier/supprimer sa track
CREATE POLICY tracks_insert_own ON tracks
    FOR INSERT WITH CHECK (uploaded_by = auth.uid());

CREATE POLICY tracks_update_own ON tracks
    FOR UPDATE USING (uploaded_by = auth.uid());

CREATE POLICY tracks_delete_own ON tracks
    FOR DELETE USING (uploaded_by = auth.uid());

-- Admin accès total
CREATE POLICY tracks_admin_all ON tracks
    FOR ALL USING (
        EXISTS (SELECT 1 FROM users WHERE id = auth.uid() AND role = 'admin')
    );

-- ─────────────────────────────────────────────────────────────
-- streams
-- ─────────────────────────────────────────────────────────────

CREATE POLICY streams_select_all ON streams
    FOR SELECT USING (TRUE);

CREATE POLICY streams_insert_own ON streams
    FOR INSERT WITH CHECK (broadcaster_id = auth.uid());

CREATE POLICY streams_update_own ON streams
    FOR UPDATE USING (broadcaster_id = auth.uid());

CREATE POLICY streams_admin_all ON streams
    FOR ALL USING (
        EXISTS (SELECT 1 FROM users WHERE id = auth.uid() AND role = 'admin')
    );

-- ─────────────────────────────────────────────────────────────
-- playlists
-- ─────────────────────────────────────────────────────────────

CREATE POLICY playlists_select_own ON playlists
    FOR SELECT USING (owner_id = auth.uid());

CREATE POLICY playlists_insert_own ON playlists
    FOR INSERT WITH CHECK (owner_id = auth.uid());

CREATE POLICY playlists_update_own ON playlists
    FOR UPDATE USING (owner_id = auth.uid());

CREATE POLICY playlists_delete_own ON playlists
    FOR DELETE USING (owner_id = auth.uid());

-- ─────────────────────────────────────────────────────────────
-- playlist_tracks (hérite des droits du propriétaire de la playlist)
-- ─────────────────────────────────────────────────────────────

CREATE POLICY playlist_tracks_owner ON playlist_tracks
    FOR ALL USING (
        EXISTS (
            SELECT 1 FROM playlists
            WHERE playlists.id = playlist_tracks.playlist_id
              AND playlists.owner_id = auth.uid()
        )
    );
