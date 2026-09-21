-- Upsert a student by Cognito sub (id) and email
INSERT INTO students (id, email)
VALUES (?, ?)
ON DUPLICATE KEY UPDATE
  email = VALUES(email);
