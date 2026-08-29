-- Migration 014 — Make listen history anonymisable instead of cascade deleted.
-- RGPD right to erasure removes the identity, but the aggregate listening
-- statistics have no reason to disappear with it. user_id becomes nullable and
-- is set to NULL when the account is deleted, so the row survives without ever
-- being attributable to a person again.
-- The composite primary key included user_id, which cannot be NULL, so the
-- table moves to a surrogate key first.

ALTER TABLE listen_history ADD COLUMN IF NOT EXISTS id UUID NOT NULL DEFAULT gen_random_uuid();

ALTER TABLE listen_history DROP CONSTRAINT IF EXISTS listen_history_pkey;
ALTER TABLE listen_history ADD PRIMARY KEY (id);

ALTER TABLE listen_history ALTER COLUMN user_id DROP NOT NULL;

ALTER TABLE listen_history DROP CONSTRAINT IF EXISTS listen_history_user_id_fkey;
ALTER TABLE listen_history
    ADD CONSTRAINT listen_history_user_id_fkey
    FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE SET NULL;

-- Anonymised rows belong to nobody, so the owner policies must not match them.
DROP POLICY IF EXISTS listen_history_insert_own ON listen_history;
CREATE POLICY listen_history_insert_own ON listen_history
    FOR INSERT WITH CHECK (user_id IS NOT NULL AND user_id = auth.uid());

DROP POLICY IF EXISTS listen_history_select_own ON listen_history;
CREATE POLICY listen_history_select_own ON listen_history
    FOR SELECT USING (user_id IS NOT NULL AND user_id = auth.uid());
