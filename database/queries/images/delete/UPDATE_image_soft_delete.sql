-- UPDATE_image_soft_delete.sql
-- Does: soft-deletes a file after every reference to it is gone. The purge job
--   removes the row and the S3 object later.
-- Used by: DELETE /images/{imageId} (takedown), step 4 of its transaction.
-- Params, in order:
--   1. image id
-- Returns: affected rows. 0 means it was already deleted (404).
UPDATE images
SET deleted_at = CURRENT_TIMESTAMP
WHERE id = ?
  AND deleted_at IS NULL;
