-- SELECT_admins.sql
-- Does: lists admins with their email and name. What "their clubs" in the PDF means
--   is an open question, so clubs aren't included.
-- Used by: GET /admins, after the admin check.
-- Params: none.
-- Returns: student_id, email, first_name, last_name.
SELECT
  a.fk_student_id AS student_id,
  s.email,
  si.first_name,
  si.last_name
FROM admins a
JOIN students s ON s.id = a.fk_student_id
LEFT JOIN student_info si ON si.fk_student_id = a.fk_student_id
ORDER BY s.email, a.fk_student_id;
