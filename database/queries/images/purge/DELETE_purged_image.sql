-- DELETE_purged_image.sql
-- Does: hard-deletes one file row soft-deleted more than 30 days ago. The foreign
--   keys refuse it if something references the file again. The handler deletes the
--   S3 object only after this succeeds, so a failed object delete leaves a harmless
--   orphan, never a row without its object.
-- Used by: POST /admins/purge, step 4, once per row from SELECT_purgeable_images.sql.
-- Params, in order:
--   1. image id
-- Returns: affected rows.
DELETE FROM images
WHERE id = ?
  AND deleted_at < CURRENT_TIMESTAMP - INTERVAL 30 DAY;
