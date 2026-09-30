-- IS_club_member.sql
-- Does: checks whether a student has any membership in a club (member, e-board or owner).
-- Used by: DELETE /clubs/{clubId}/members/me after a delete that affected 0 rows
--   (not a member is 404, a member is an owner who can't leave, 403), and
--   PUT /clubs/{clubId}/members/roles for the target member (404).
-- Params, in order:
--   1. student id
--   2. club id
-- Returns: is_member, 1 or 0.
SELECT EXISTS (
  SELECT 1
  FROM club_members
  WHERE fk_student_id = ?
    AND fk_club_id = ?
) AS is_member;
