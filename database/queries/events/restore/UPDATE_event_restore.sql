-- UPDATE_event_restore.sql
-- Does: restores a soft-deleted event by clearing deleted_at, only when it was
--   deleted less than 30 days ago. The purge uses the same 30 days, so an event
--   that can still be restored is never purged.
-- Used by: POST /auth/events/{eventId}/restore, after
--   authorization/events/manages_owner_club/IS_student_owner_club_manager.sql or IS_admin.sql.
-- Params, in order:
--   1. event id
-- Returns: affected rows. 0 means the event doesn't exist, isn't deleted, or was
--   deleted 30 or more days ago (SELECT_event_deletion.sql tells them apart).
UPDATE events
SET deleted_at = NULL
WHERE id = ?
  AND deleted_at IS NOT NULL
  AND deleted_at > CURRENT_TIMESTAMP - INTERVAL 30 DAY;
