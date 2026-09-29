-- SELECT_club_members.sql
-- Does: lists a club's members with their role, owners first, then e-board, then
--   members, each by name.
-- Used by: GET /clubs/{clubId}/members and GET /clubs/{clubId}/eboard (as role 'eboard'),
--   after the club role check. Callers are e-board or owners, so emails are included.
-- Params, in order:
--   1. club id
--   2. role filter: NULL for everyone, 'owner', 'eboard' (the e-board list: role
--      'eboard' or 'owner') or 'member'
--   3. limit
--   4. offset
-- Returns: student_id, email, first_name, last_name, role.
SELECT
  cm.fk_student_id AS student_id,
  s.email,
  si.first_name,
  si.last_name,
  cm.role
FROM club_members cm
CROSS JOIN (
  SELECT
    CAST(? AS SIGNED) AS club_id,
    CAST(? AS CHAR(6)) AS role_filter
) p
JOIN students s ON s.id = cm.fk_student_id
LEFT JOIN student_info si ON si.fk_student_id = cm.fk_student_id
WHERE cm.fk_club_id = p.club_id
  AND (
    p.role_filter IS NULL
    OR (p.role_filter = 'eboard' AND cm.role IN ('eboard', 'owner'))
    OR (p.role_filter <> 'eboard' AND cm.role = p.role_filter)
  )
ORDER BY
  FIELD(cm.role, 'owner', 'eboard', 'member'),
  si.last_name,
  si.first_name,
  s.email,
  cm.fk_student_id
LIMIT ? OFFSET ?;
