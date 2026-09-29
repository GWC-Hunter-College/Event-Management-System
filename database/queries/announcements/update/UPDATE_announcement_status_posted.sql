-- UPDATE_announcement_status_posted.sql
-- Does: posts an announcement. draft -> posted is the only transition, and it needs
--   a body. Any other change affects 0 rows.
-- Used by: PATCH /auth/announcements/{announcementId} with "status": "posted", after
--   UPDATE_announcement.sql, when the current status is draft.
-- Params, in order:
--   1. announcement id
-- Returns: affected rows. 0 means the announcement is missing or soft-deleted (404),
--   or isn't a draft with a body (400). SELECT_announcement_status.sql tells them apart.
UPDATE announcements
SET status = 'posted'
WHERE id = ?
  AND deleted_at IS NULL
  AND status = 'draft'
  AND body IS NOT NULL
  AND body <> '';
