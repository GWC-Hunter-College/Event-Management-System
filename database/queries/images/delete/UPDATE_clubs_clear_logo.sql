-- UPDATE_clubs_clear_logo.sql
-- Does: removes a file from every club that uses it as a logo.
-- Used by: DELETE /images/{imageId} (takedown), step 1 of its transaction.
-- Params, in order:
--   1. image id
-- Returns: affected rows.
UPDATE clubs
SET fk_logo_id = NULL
WHERE fk_logo_id = ?;
