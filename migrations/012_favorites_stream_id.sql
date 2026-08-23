-- Migration 012 - Rename favorites.track_id to favorites.stream_id.
-- Migration 011 re-pointed the foreign key at streams(id) but kept the original
-- column name, so the schema still claimed to store track identifiers while it
-- actually stored stream identifiers. This aligns the name with the content.

ALTER TABLE favorites RENAME COLUMN track_id TO stream_id;

ALTER TABLE favorites RENAME CONSTRAINT favorites_track_id_fkey TO favorites_stream_id_fkey;
