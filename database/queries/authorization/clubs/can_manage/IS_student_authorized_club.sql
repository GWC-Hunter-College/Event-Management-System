-- IS_student_authorized_club.sql
-- Checks if a given student (by sub/id) is an e-board member or owner of the specified club.
-- Returns 1 if authorized, 0 otherwise.
-- Params:
--   ? = fk_student_id (user sub)
--   ? = fk_club_id (club id)

SELECT EXISTS (
  SELECT 1
  FROM club_members
  WHERE fk_student_id = ?
    AND fk_club_id    = ?
    AND (member_is_eboard = 1 OR member_is_owner = 1)
) AS is_authorized;
