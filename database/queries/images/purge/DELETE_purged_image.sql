-- DELETE_purged_image.sql
-- Does: hard-deletes one soft-deleted file row. The foreign keys refuse it if
--   something references the file again. The job deletes the S3 object only after
--   this succeeds, so a failed object delete leaves a harmless orphan, never a row
--   without its object.
-- Used by: the purge batch job (internal), step 4, once per row from SELECT_purgeable_images.sql.
-- Params, in order:
--   1. image id
-- Returns: affected rows.
DELETE FROM images
WHERE id = ?
  AND deleted_at IS NOT NULL;
