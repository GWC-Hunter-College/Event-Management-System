-- SELECT_image.sql
-- Does: reads a stored file that isn't soft-deleted, for the takedown's 404 and
--   its authorization (manager of the owning club, or admin).
-- Used by: DELETE /images/{imageId}
-- Params, in order:
--   1. image id
-- Returns: id, fk_club_id, purpose, object_key, mimetype, alt_text, or no row (404).
SELECT
  id,
  fk_club_id,
  purpose,
  object_key,
  mimetype,
  alt_text
FROM images
WHERE id = ?
  AND deleted_at IS NULL;
