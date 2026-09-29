-- DELETE_admin.sql
-- Does: removes an admin, unless they're the last one. That applies to removing
--   yourself too.
-- Used by: DELETE /admins/{studentId}, after the admin check and
--   SELECT_admins_for_update.sql in the same transaction.
-- Params, in order:
--   1. student id
-- Returns: affected rows. 0 means the student isn't an admin (404, check with
--   authorization/admins/is_admin/IS_admin.sql) or is the last admin (409).
DELETE a
FROM admins a
JOIN (SELECT COUNT(*) AS admin_count FROM admins) c
WHERE a.fk_student_id = ?
  AND c.admin_count > 1;
