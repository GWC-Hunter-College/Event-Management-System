-- IS_admin.sql
-- Does: checks whether a student is an admin (has an admins row).
-- Used by: every /admins route, POST and DELETE /clubs/{clubId}/verification,
--   DELETE /images/{imageId} (an admin may take down any file), and isAdmin on GET /me.
-- Params, in order:
--   1. student id
-- Returns: is_admin, 1 or 0.
SELECT EXISTS (
  SELECT 1
  FROM admins
  WHERE fk_student_id = ?
) AS is_admin;
