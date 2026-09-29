-- UPDATE_announcement_restore.sql
-- Does: restores a soft-deleted announcement by clearing deleted_at, only when it was
--   deleted less than 30 days ago, the same window as events. The purge uses the same
--   30 days, so nothing restorable is purged.
-- Used by: POST /auth/announcements/{announcementId}/restore, after
--   authorization/announcements/manages_owner_club/IS_student_announcement_owner_club_manager.sql
--   or IS_admin.sql.
-- Params, in order:
--   1. announcement id
-- Returns: affected rows. 0 means it doesn't exist, isn't deleted, or was deleted
--   30 or more days ago (SELECT_announcement_deletion.sql tells them apart).
UPDATE announcements
SET deleted_at = NULL
WHERE id = ?
  AND deleted_at IS NOT NULL
  AND deleted_at > CURRENT_TIMESTAMP - INTERVAL 30 DAY;
