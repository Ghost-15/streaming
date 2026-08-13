-- Migration 011 - Favorites and playlist items reference live streams.
-- In StreamPulse the favoritable / playlistable content is a live stream, not a
-- row in the (unused) tracks table. The original FKs pointed at tracks(id), so
-- every "add to favorites / playlist" of a live violated the foreign key. This
-- re-points both to streams(id).

-- Remove any orphaned rows from earlier failed attempts before re-pointing.
DELETE FROM favorites WHERE track_id NOT IN (SELECT id FROM streams);
DELETE FROM playlist_tracks WHERE track_id NOT IN (SELECT id FROM streams);

ALTER TABLE favorites DROP CONSTRAINT IF EXISTS favorites_track_id_fkey;
ALTER TABLE favorites
    ADD CONSTRAINT favorites_track_id_fkey
    FOREIGN KEY (track_id) REFERENCES streams(id) ON DELETE CASCADE;

ALTER TABLE playlist_tracks DROP CONSTRAINT IF EXISTS playlist_tracks_track_id_fkey;
ALTER TABLE playlist_tracks
    ADD CONSTRAINT playlist_tracks_track_id_fkey
    FOREIGN KEY (track_id) REFERENCES streams(id) ON DELETE CASCADE;
