-- Grabs a student's id and email from the students table
-- Filters by the given student id (primary key)
-- Returns at most one row

SELECT 
    id,
    email
FROM students
WHERE id = ?;
