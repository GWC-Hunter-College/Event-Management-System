-- SELECT_admin.sql
-- Does: reads one admin.
-- Used by: GET /admins/{studentId}, after the admin check.
-- Params, in order:
--   1. student id
-- Returns: student_id, email, first_name, last_name, or no row (404).
SELECT
  a.fk_student_id AS student_id,
  s.email,
  si.first_name,
  si.last_name
FROM admins a
JOIN students s ON s.id = a.fk_student_id
LEFT JOIN student_info si ON si.fk_student_id = a.fk_student_id
WHERE a.fk_student_id = ?;
