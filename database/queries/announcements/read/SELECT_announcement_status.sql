-- SELECT_announcement_status.sql
-- Does: reads an announcement's status, to tell 404 from 400 after an UPDATE that
--   affected 0 rows.
-- Used by: PATCH and DELETE /auth/announcements/{announcementId}
-- Params, in order:
--   1. announcement id
-- Returns: one row (status), or no row when the announcement doesn't exist or is soft-deleted.
SELECT status
FROM announcements
WHERE id = ?
  AND deleted_at IS NULL;
