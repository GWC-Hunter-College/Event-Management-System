-- DELETE_admin.sql
-- Does: removes an admin. Protecting the last admin is an open question, so this
--   query doesn't.
-- Used by: DELETE /admins/{studentId}, after the admin check.
-- Params, in order:
--   1. student id
-- Returns: affected rows. 0 means the student isn't an admin (404).
DELETE FROM admins
WHERE fk_student_id = ?;
