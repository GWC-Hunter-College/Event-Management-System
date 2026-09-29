-- SELECT_admins_for_update.sql
-- Does: locks every admin row until the transaction ends, so two admins can't remove
--   each other at the same time and leave no admin.
-- Used by: DELETE /admins/{studentId}, first statement of its transaction, before
--   DELETE_admin.sql.
-- Params: none.
-- Returns: fk_student_id of each admin.
SELECT fk_student_id
FROM admins
FOR UPDATE;
