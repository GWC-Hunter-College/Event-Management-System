-- SELECT_event_deletion.sql
-- Does: reads whether an event is deleted and whether it can still be restored,
--   to explain a restore that affected 0 rows.
-- Used by: POST /auth/events/{eventId}/restore
-- Params, in order:
--   1. event id
-- Returns: deleted_at, is_deleted, is_restorable, or no row (404: unknown or purged).
--   is_deleted 0 means 409 (not deleted). is_restorable 0 on a deleted event means
--   410 (the 30-day window has passed).
SELECT
  deleted_at,
  (deleted_at IS NOT NULL) AS is_deleted,
  (deleted_at IS NOT NULL AND deleted_at > CURRENT_TIMESTAMP - INTERVAL 30 DAY) AS is_restorable
FROM events
WHERE id = ?;
