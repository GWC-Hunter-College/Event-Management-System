-- UPDATE_image_soft_delete_if_unused.sql
-- Does: soft-deletes a file once nothing references it, so removing one use never
--   removes a file another club logo, flyer or gallery still shows. POST /admins/purge
--   removes the row and the S3 object later.
-- Used by: the logo and flyer confirms, for the file they replaced, after the new
--   pointer is set, inside the same transaction.
-- Params, in order:
--   1. image id of the replaced file
-- Returns: affected rows. 0 means the file is still in use (or already deleted), which is fine.
UPDATE images i
SET i.deleted_at = CURRENT_TIMESTAMP
WHERE i.id = ?
  AND i.deleted_at IS NULL
  AND NOT EXISTS (SELECT 1 FROM clubs c WHERE c.fk_logo_id = i.id)
  AND NOT EXISTS (SELECT 1 FROM events e WHERE e.fk_thumbnail_id = i.id)
  AND NOT EXISTS (SELECT 1 FROM event_images ei WHERE ei.fk_image_id = i.id);
