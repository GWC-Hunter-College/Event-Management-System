-- DELETE_purgeable_announcements.sql
-- Does: hard-deletes announcements deleted more than 30 days ago, which can no longer
--   be restored. Their club links go with them (ON DELETE CASCADE). Announcements have
--   no images yet, so there are no files to release.
-- Used by: POST /admins/purge, in the same transaction as the event purge.
-- Params: none. The 30 days match the restore window in announcements/restore/.
-- Returns: affected rows (the number of announcements purged).
DELETE FROM announcements
WHERE deleted_at < CURRENT_TIMESTAMP - INTERVAL 30 DAY;
