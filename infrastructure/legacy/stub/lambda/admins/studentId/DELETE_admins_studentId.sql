-- Deletion or demotion of an admin

DELETE FROM admins
-- WHERE fk_student_id = :student_id;
WHERE fk_student_id = 2; -- Deleting Kyle's info