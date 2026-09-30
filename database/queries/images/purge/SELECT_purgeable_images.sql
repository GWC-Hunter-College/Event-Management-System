-- SELECT_purgeable_images.sql
-- Does: finds files soft-deleted more than 30 days ago that nothing references,
--   oldest first.
-- Used by: POST /admins/purge, step 3, after the event purge commits. For each row,
--   the handler deletes the row with DELETE_purged_image.sql, then the S3 object,
--   and repeats until a batch comes back empty.
-- Params, in order:
--   1. batch size
-- Returns: id, object_key.
SELECT
  i.id,
  i.object_key
FROM images i
WHERE i.deleted_at < CURRENT_TIMESTAMP - INTERVAL 30 DAY
  AND NOT EXISTS (SELECT 1 FROM clubs c WHERE c.fk_logo_id = i.id)
  AND NOT EXISTS (SELECT 1 FROM events e WHERE e.fk_thumbnail_id = i.id)
  AND NOT EXISTS (SELECT 1 FROM event_images ei WHERE ei.fk_image_id = i.id)
ORDER BY i.deleted_at, i.id
LIMIT ?;
