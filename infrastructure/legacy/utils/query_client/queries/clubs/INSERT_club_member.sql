-- INSERT_club_member.sql
-- Join a club as a regular member if:
--   1) the club exists
--   2) the student exists (optional, see note)
--   3) the student is not already a member

INSERT INTO club_members (
    fk_student_id,
    fk_club_id,
    member_is_eboard,
    member_is_owner
)
SELECT
    s.id        AS fk_student_id,
    c.id        AS fk_club_id,
    FALSE       AS member_is_eboard,
    FALSE       AS member_is_owner
FROM students s
JOIN clubs c ON c.id = ?
WHERE s.id = ?
  AND NOT EXISTS (
      SELECT 1
      FROM club_members cm
      WHERE cm.fk_student_id = s.id
        AND cm.fk_club_id    = c.id
  );
