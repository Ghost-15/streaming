-- Migration 006 — stream_url sur streams + suspended_at sur users
-- US-003 (source audio diffusée) + US-013 (suspension de compte).

-- ── Streams : URL de la source audio diffusée ────────────────────────────────
-- Le diffuseur fournit une URL audio jouable (HLS/MP3) au démarrage du stream.
-- Les auditeurs la consomment côté client (just_audio). Le transport audio brut
-- (ingestion micro temps réel) reste hors périmètre — voir ADR-007.
ALTER TABLE streams
    ADD COLUMN IF NOT EXISTS stream_url TEXT NOT NULL DEFAULT '';

-- ── Users : suspension de compte (US-013) ────────────────────────────────────
-- NULL = compte actif. Date non-nulle = compte suspendu depuis cette date.
ALTER TABLE users
    ADD COLUMN IF NOT EXISTS suspended_at TIMESTAMPTZ;
