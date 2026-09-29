-- UPDATE_event_thumbnail.sql
-- Does: points an event at its new flyer.
-- Used by: POST /clubs/{clubId}/events/{eventId}/thumbnails/confirm, inside its
--   transaction, after images/create/INSERT_image.sql.
-- Params, in order:
--   1. image id of the new flyer
--   2. event id
-- Returns: affected rows.
UPDATE events
SET fk_thumbnail_id = ?
WHERE id = ?
  AND deleted_at IS NULL;
