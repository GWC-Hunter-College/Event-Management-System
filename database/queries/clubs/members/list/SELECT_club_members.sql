-- SELECT_club_members.sql
-- Does: lists a club's members with one role each (owner over eboard over member),
--   owners first, then e-board, then members, each by name.
-- Used by: GET /clubs/{clubId}/members and GET /clubs/{clubId}/eboard (as role 'eboard'),
--   after the club role check. Callers are e-board or owners, so emails are included.
-- Params, in order:
--   1. club id
--   2. role filter: NULL for everyone, 'owner', 'eboard' (e-board members and owners,
--      as memberships.md defines the e-board list) or 'member' (neither flag)
--   3. limit
--   4. offset
-- Returns: student_id, email, first_name, last_name, role.
SELECT
  cm.fk_student_id AS student_id,
  s.email,
  si.first_name,
  si.last_name,
  CASE
    WHEN cm.member_is_owner THEN 'owner'
    WHEN cm.member_is_eboard THEN 'eboard'
    ELSE 'member'
  END AS role
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
    OR (p.role_filter = 'owner' AND cm.member_is_owner)
    OR (p.role_filter = 'eboard' AND (cm.member_is_eboard OR cm.member_is_owner))
    OR (p.role_filter = 'member' AND NOT cm.member_is_eboard AND NOT cm.member_is_owner)
  )
ORDER BY
  CASE
    WHEN cm.member_is_owner THEN 0
    WHEN cm.member_is_eboard THEN 1
    ELSE 2
  END,
  si.last_name,
  si.first_name,
  s.email,
  cm.fk_student_id
LIMIT ? OFFSET ?;
