-- INSERT_club_owner.sql
-- Does: makes the caller the owner of the club they just created.
-- Used by: POST /clubs, inside the create transaction, after INSERT_club_info.sql.
-- Params, in order:
--   1. student id (the caller's sub)
--   2. club id (the last insert id from INSERT_club.sql)
-- Returns: nothing.
INSERT INTO club_members
  (fk_student_id, fk_club_id, role)
VALUES
  (?, ?, 'owner');
