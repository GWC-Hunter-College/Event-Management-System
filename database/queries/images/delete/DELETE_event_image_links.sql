-- DELETE_event_image_links.sql
-- Does: removes a file from every event gallery.
-- Used by: DELETE /images/{imageId} (takedown), step 3 of its transaction.
-- Params, in order:
--   1. image id
-- Returns: affected rows.
DELETE FROM event_images
WHERE fk_image_id = ?;
