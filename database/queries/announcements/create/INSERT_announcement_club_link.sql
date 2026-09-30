-- INSERT_announcement_club_link.sql
-- Does: ties an announcement to a club, as its owner club or as a co-owning club.
-- Used by: POST /clubs/{clubId}/announcements (the owner link, then one per associate),
--   and PATCH /auth/announcements/{announcementId} when the body includes associates.
-- Params, in order:
--   1. announcement id
--   2. club id
--   3. is owner: TRUE for the club in the path, FALSE for associates
-- Returns: nothing. A second owner link fails the one-owner index.
INSERT INTO announcements_to_clubs
  (fk_announcement_id, fk_club_id, club_is_announcement_owner)
VALUES
  (?, ?, ?);
