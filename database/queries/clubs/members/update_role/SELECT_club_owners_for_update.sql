-- SELECT_club_owners_for_update.sql
-- Does: locks a club's owner rows until the transaction ends, so two owners can't
--   demote each other, or both leave, at the same time and leave the club with none.
-- Used by: PUT /clubs/{clubId}/members/roles and DELETE /clubs/{clubId}/members/me,
--   first statement of their transactions.
-- Params, in order:
--   1. club id
-- Returns: fk_student_id of each owner.
SELECT fk_student_id
FROM club_members
WHERE fk_club_id = ?
  AND role = 'owner'
FOR UPDATE;
