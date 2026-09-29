-- DELETE_announcement_associates.sql
-- Does: removes an announcement's co-owning clubs, keeping the owner club, before the
--   new list is inserted with announcements/create/INSERT_announcement_club_link.sql.
-- Used by: PATCH /auth/announcements/{announcementId}, when the body includes associates.
-- Params, in order:
--   1. announcement id
-- Returns: nothing.
DELETE FROM announcements_to_clubs
WHERE fk_announcement_id = ?
  AND club_is_announcement_owner = FALSE;
