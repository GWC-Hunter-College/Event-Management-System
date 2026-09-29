-- IS_club_owner.sql
-- Does: checks whether a student owns a club. Owner-only actions need this, not
--   the e-board-or-owner check in can_manage.
-- Used by: PUT /clubs/{clubId}/members/roles (403 otherwise).
-- Params, in order:
--   1. student id
--   2. club id
-- Returns: is_owner, 1 or 0.
SELECT EXISTS (
  SELECT 1
  FROM club_members
  WHERE fk_student_id = ?
    AND fk_club_id = ?
    AND member_is_owner = TRUE
) AS is_owner;
