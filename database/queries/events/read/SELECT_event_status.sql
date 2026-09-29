-- SELECT_event_status.sql
-- Does: reads an event's status, as an existence and visibility check.
-- Used by: GET /events/{eventId}/images (posted or cancelled, else 404),
--   GET /clubs/{clubId}/events/{eventId}/images, the event image signers and confirms (404),
--   and PATCH and DELETE /auth/events/{eventId} after an UPDATE that affected 0 rows
--   (no row means 404, a row means the transition wasn't allowed, 400).
-- Params, in order:
--   1. event id
-- Returns: one row (status), or no row when the event doesn't exist or is soft-deleted.
SELECT status
FROM events
WHERE id = ?
  AND deleted_at IS NULL;
