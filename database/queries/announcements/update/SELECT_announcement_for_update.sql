-- SELECT_announcement_for_update.sql
-- Does: reads an announcement's editable fields and locks its row until the
--   transaction ends, so the API can merge a partial update.
-- Used by: PATCH /auth/announcements/{announcementId}, first statement of its transaction.
-- Params, in order:
--   1. announcement id
-- Returns: id, title, body, status, owner_club_id. No row means 404.
SELECT
  a.id,
  a.title,
  a.body,
  a.status,
  (
    SELECT ac.fk_club_id
    FROM announcements_to_clubs ac
    WHERE ac.fk_announcement_id = a.id
      AND ac.club_is_announcement_owner = TRUE
  ) AS owner_club_id
FROM announcements a
WHERE a.id = ?
  AND a.deleted_at IS NULL
FOR UPDATE;
