-- Inserts a student row with the given sub if one does not exist.
-- Does not modify email if the row already exists.
INSERT INTO students (id, email)
VALUES (?, NULL)
ON DUPLICATE KEY UPDATE
  id = students.id; -- no-op, preserves email
