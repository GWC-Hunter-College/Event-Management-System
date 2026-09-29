-- UPDATE_announcement_soft_delete.sql
-- Does: soft-deletes an announcement in either status by setting deleted_at. Every
--   read stops returning it at once. It can be restored for 30 days
--   (announcements/restore/), and after that POST /admins/purge removes it for good.
-- Used by: DELETE /auth/announcements/{announcementId}
-- Params, in order:
--   1. announcement id
-- Returns: affected rows. 0 means it doesn't exist or was already deleted (404).
UPDATE announcements
SET deleted_at = CURRENT_TIMESTAMP
WHERE id = ?
  AND deleted_at IS NULL;
