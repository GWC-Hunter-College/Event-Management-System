-- DELETE_club_member.sql
-- Leave a club if the student is a member and not the owner.

DELETE FROM club_members
WHERE fk_student_id   = ?
  AND fk_club_id      = ?
  AND member_is_owner = FALSE;
