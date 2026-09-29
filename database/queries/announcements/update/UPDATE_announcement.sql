-- UPDATE_announcement.sql
-- Does: writes an announcement's title and body after the API has merged the request
--   into the values read by SELECT_announcement_for_update.sql. Always bumps updated_at.
--   Posting uses UPDATE_announcement_status_posted.sql, never this one.
-- Used by: PATCH /auth/announcements/{announcementId}, inside its transaction.
-- Params, in order:
--   1. title
--   2. body, or NULL
--   3. announcement id
-- Returns: affected rows. It can be 0 when nothing changed within the same second,
--   so the locking read, not this count, is the existence check.
UPDATE announcements
SET
  title = ?,
  body = ?,
  updated_at = CURRENT_TIMESTAMP
WHERE id = ?
  AND deleted_at IS NULL;
