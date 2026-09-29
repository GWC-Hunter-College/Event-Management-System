-- UPDATE_club.sql
-- Does: renames a club. The API merges the request with SELECT_club.sql first, so
--   an omitted name is written back unchanged. A duplicate name fails (409).
-- Used by: PATCH /clubs/{clubId}, inside its transaction.
-- Params, in order:
--   1. name
--   2. club id
-- Returns: affected rows.
UPDATE clubs
SET name = ?
WHERE id = ?;
