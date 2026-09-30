-- SELECT_event_images.sql
-- Does: lists an event's gallery images other than its flyer, in the order they
--   were added. The API turns each object key into a signed URL.
-- Used by: GET /events/{eventId} (images), GET /auth/events/{eventId} (images),
--   GET /events/{eventId}/images, GET /auth/events/{eventId}/images and
--   GET /clubs/{clubId}/events/{eventId}/images. Routes that return a gallery on
--   its own check visibility first with events/read/SELECT_event_status.sql.
-- Params, in order:
--   1. event id
-- Returns: image_id, mimetype, object_key, alt_text. No rows for a soft-deleted event.
SELECT
  i.id AS image_id,
  i.mimetype,
  i.object_key,
  i.alt_text
FROM event_images ei
JOIN events e ON e.id = ei.fk_event_id
JOIN images i ON i.id = ei.fk_image_id
WHERE ei.fk_event_id = ?
  AND e.deleted_at IS NULL
  AND ei.fk_image_id <> COALESCE(e.fk_thumbnail_id, '')
ORDER BY ei.created_at, i.id;
