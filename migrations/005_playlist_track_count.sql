-- Migration 005 — Ajout track_count sur playlists + trigger de maintenance
-- Justification scalabilité : évite un COUNT(*) JOIN à chaque lecture de playlist.
-- Le trigger garantit la cohérence sans intervention applicative.

-- ── 1. Ajout de la colonne ────────────────────────────────────────────────────
ALTER TABLE playlists
    ADD COLUMN IF NOT EXISTS track_count INT NOT NULL DEFAULT 0;

-- ── 2. Recalcul initial (playlists existantes) ────────────────────────────────
UPDATE playlists p
SET track_count = (
    SELECT COUNT(*) FROM playlist_tracks pt
    WHERE pt.playlist_id = p.id
);

-- ── 3. Fonction trigger ───────────────────────────────────────────────────────
CREATE OR REPLACE FUNCTION fn_update_playlist_track_count()
RETURNS TRIGGER AS $$
BEGIN
    IF TG_OP = 'INSERT' THEN
        UPDATE playlists
        SET track_count = track_count + 1
        WHERE id = NEW.playlist_id;

    ELSIF TG_OP = 'DELETE' THEN
        UPDATE playlists
        SET track_count = GREATEST(track_count - 1, 0) -- jamais négatif
        WHERE id = OLD.playlist_id;
    END IF;

    RETURN NULL; -- AFTER trigger, valeur de retour ignorée
END;
$$ LANGUAGE plpgsql;

-- ── 4. Attache le trigger sur playlist_tracks ─────────────────────────────────
DROP TRIGGER IF EXISTS trg_playlist_track_count ON playlist_tracks;

CREATE TRIGGER trg_playlist_track_count
    AFTER INSERT OR DELETE ON playlist_tracks
    FOR EACH ROW
    EXECUTE FUNCTION fn_update_playlist_track_count();
