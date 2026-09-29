-- IS_student_authorized_announcement.sql
-- Does: checks whether a student is on the e-board of, or owns, ANY club an
--   announcement belongs to (can manage the announcement), as for events.
-- Used by: GET, PATCH and DELETE /auth/announcements/{announcementId}.
-- Params, in order:
--   1. student id
--   2. announcement id
-- Returns: is_authorized, 1 or 0.
SELECT EXISTS (
  SELECT 1
  FROM announcements_to_clubs ac
  JOIN club_members cm
    ON cm.fk_club_id = ac.fk_club_id
  WHERE cm.fk_student_id = ?
    AND ac.fk_announcement_id = ?
    AND cm.role IN ('eboard', 'owner')
) AS is_authorized;
