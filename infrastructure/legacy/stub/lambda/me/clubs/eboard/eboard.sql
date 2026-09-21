/* put the query of:

GET /me/clubs/eboard

in this folder */ 

-- the ? denotes the current user's id

SELECT cm.fk_student_id FROM club_members cm
WHERE ? = cm.fk_student_id
  AND cm.member_is_eboard = TRUE OR cm.member_is_owner = TRUE;

