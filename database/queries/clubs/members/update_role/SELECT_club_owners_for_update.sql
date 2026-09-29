-- SELECT_club_owners_for_update.sql
-- Does: locks a club's owner rows until the transaction ends, so two owners can't
--   demote each other at the same time and leave the club with none.
-- Used by: PUT /clubs/{clubId}/members/roles, first statement of its transaction,
--   before UPDATE_club_member_role.sql.
-- Params, in order:
--   1. club id
-- Returns: fk_student_id of each owner.
SELECT fk_student_id
FROM club_members
WHERE fk_club_id = ?
  AND member_is_owner = TRUE
FOR UPDATE;
