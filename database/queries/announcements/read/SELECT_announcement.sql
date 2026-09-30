-- SELECT_announcement.sql
-- Does: reads one announcement that isn't soft-deleted, for the public read or a manager's read.
-- Used by: GET /announcements/{announcementId} (public_only TRUE), and
--   GET /auth/announcements/{announcementId} and the PATCH and restore responses
--   (public_only FALSE, after the role check).
-- Params, in order:
--   1. announcement id
--   2. public_only: TRUE returns only posted announcements, FALSE drafts too
-- Returns: one announcement_details row, or no row (404).
SELECT ad.*
FROM announcement_details ad
WHERE ad.id = ?
  AND (? = FALSE OR ad.status = 'posted');
