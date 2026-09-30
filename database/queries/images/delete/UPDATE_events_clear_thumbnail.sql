-- UPDATE_events_clear_thumbnail.sql
-- Does: removes a file from every event that uses it as a flyer, soft-deleted events included.
-- Used by: DELETE /images/{imageId} (takedown), step 2 of its transaction.
-- Params, in order:
--   1. image id
-- Returns: affected rows.
UPDATE events
SET fk_thumbnail_id = NULL
WHERE fk_thumbnail_id = ?;
