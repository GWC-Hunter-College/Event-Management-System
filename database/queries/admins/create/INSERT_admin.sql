-- INSERT_admin.sql
-- Does: makes an existing student an admin. Promoting an admin changes nothing.
-- Used by: POST /admins, after the admin check.
-- Params, in order:
--   1. student id
-- Returns: affected rows. 0 means the student doesn't exist or is already an admin
--   (me/get/SELECT_student_by_sub.sql tells them apart).
INSERT INTO admins (fk_student_id)
SELECT s.id
FROM students s
WHERE s.id = ?
  AND NOT EXISTS (
    SELECT 1
    FROM admins a
    WHERE a.fk_student_id = s.id
  );
