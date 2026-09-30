-- UPDATE_club_logo.sql
-- Does: points a club at its new logo.
-- Used by: POST /clubs/{clubId}/thumbnails/confirm, inside its transaction, after
--   images/create/INSERT_image.sql.
-- Params, in order:
--   1. image id of the new logo
--   2. club id
-- Returns: affected rows.
UPDATE clubs
SET fk_logo_id = ?
WHERE id = ?;
