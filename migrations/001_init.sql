-- Migration 001 — Initial schema
-- Tables: users, tracks, streams, playlists, playlist_tracks
-- Applied via: make migrate (depuis WSL) ou Supabase SQL Editor

CREATE EXTENSION IF NOT EXISTS "pgcrypto"; -- for gen_random_uuid()

-- ENUM types
CREATE TYPE user_role AS ENUM ('anon', 'user', 'diffuseur', 'admin');
CREATE TYPE stream_status AS ENUM ('live', 'ended');

-- ─────────────────────────────────────────────────────────────
-- Users
-- ─────────────────────────────────────────────────────────────
CREATE TABLE IF NOT EXISTS users (
    id            UUID        PRIMARY KEY DEFAULT gen_random_uuid(),
    email         TEXT        UNIQUE NOT NULL,
    password_hash TEXT        NOT NULL,
    first_name    TEXT        NOT NULL DEFAULT '',
    last_name     TEXT        NOT NULL DEFAULT '',
    role          user_role   NOT NULL DEFAULT 'user',
    created_at    TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- ─────────────────────────────────────────────────────────────
-- Tracks (fichiers audio stockés dans Supabase Storage)
-- ─────────────────────────────────────────────────────────────
CREATE TABLE IF NOT EXISTS tracks (
    id         UUID        PRIMARY KEY DEFAULT gen_random_uuid(),
    title      TEXT        NOT NULL,
    artist     TEXT        NOT NULL DEFAULT '',
    duration   INT         NOT NULL DEFAULT 0,  -- durée en secondes
    file_url   TEXT        NOT NULL,             -- URL Supabase Storage (bucket: audio)
    uploaded_by UUID       REFERENCES users(id) ON DELETE SET NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_tracks_uploaded_by ON tracks(uploaded_by);

-- ─────────────────────────────────────────────────────────────
-- Streams (diffusions live)
-- ─────────────────────────────────────────────────────────────
CREATE TABLE IF NOT EXISTS streams (
    id             UUID         PRIMARY KEY DEFAULT gen_random_uuid(),
    title          TEXT         NOT NULL,
    broadcaster_id UUID         NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    status         stream_status NOT NULL DEFAULT 'live',
    started_at     TIMESTAMPTZ  NOT NULL DEFAULT NOW(),
    ended_at       TIMESTAMPTZ,
    listener_count INT          NOT NULL DEFAULT 0
);

CREATE INDEX idx_streams_status      ON streams(status);
CREATE INDEX idx_streams_broadcaster ON streams(broadcaster_id);

-- ─────────────────────────────────────────────────────────────
-- Playlists
-- ─────────────────────────────────────────────────────────────
CREATE TABLE IF NOT EXISTS playlists (
    id         UUID        PRIMARY KEY DEFAULT gen_random_uuid(),
    owner_id   UUID        NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    title      TEXT        NOT NULL,
    is_queue   BOOLEAN     NOT NULL DEFAULT FALSE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_playlists_owner ON playlists(owner_id);

-- ─────────────────────────────────────────────────────────────
-- Playlist tracks (tracks ordonnées dans une playlist)
-- ─────────────────────────────────────────────────────────────
CREATE TABLE IF NOT EXISTS playlist_tracks (
    playlist_id UUID        NOT NULL REFERENCES playlists(id) ON DELETE CASCADE,
    track_id    UUID        NOT NULL REFERENCES tracks(id)    ON DELETE CASCADE,
    position    INT         NOT NULL,
    added_at    TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    PRIMARY KEY (playlist_id, position)
);
