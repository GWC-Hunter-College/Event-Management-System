-- SELECT_club_logo_for_update.sql
-- Does: reads a club's current logo and locks the club row until the transaction
--   ends, so two logo confirms can't both release the same old file.
-- Used by: POST /clubs/{clubId}/thumbnails/confirm, first statement of its transaction.
-- Params, in order:
--   1. club id
-- Returns: fk_logo_id (NULL when the club has no logo), or no row (404).
SELECT fk_logo_id
FROM clubs
WHERE id = ?
FOR UPDATE;
