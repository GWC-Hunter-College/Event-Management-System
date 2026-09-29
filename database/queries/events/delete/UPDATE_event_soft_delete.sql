-- UPDATE_event_soft_delete.sql
-- Does: soft-deletes an event in any status by setting deleted_at. Every read
--   stops returning it at once. It can be restored for 30 days
--   (events/restore/), and after that POST /admins/purge removes it for good.
-- Used by: DELETE /auth/events/{eventId}
-- Params, in order:
--   1. event id
-- Returns: affected rows. 0 means the event doesn't exist or was already deleted (404).
UPDATE events
SET deleted_at = CURRENT_TIMESTAMP
WHERE id = ?
  AND deleted_at IS NULL;
