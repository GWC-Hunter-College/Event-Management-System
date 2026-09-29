-- INSERT_club_member.sql
-- Does: joins a student to a club as a regular member, when the club and the
--   student exist and the student isn't a member yet.
-- Used by: POST /clubs/{clubId}/members/me (the student is the caller).
-- Params, in order (club first, as in the legacy query):
--   1. club id
--   2. student id
-- Returns: affected rows. 0 means already a member (409) or no such club
--   (clubs/get/EXISTS_club.sql tells them apart, 404).
INSERT INTO club_members (fk_student_id, fk_club_id, role)
SELECT
  s.id,
  c.id,
  'member'
FROM students s
JOIN clubs c ON c.id = ?
WHERE s.id = ?
  AND NOT EXISTS (
    SELECT 1
    FROM club_members cm
    WHERE cm.fk_student_id = s.id
      AND cm.fk_club_id = c.id
  );
