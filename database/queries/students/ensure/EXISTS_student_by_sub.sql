-- Checks whether a student row already exists for the given sub.
-- Returns a single row with value `1` if and only if:
--   • The `id` column matches the provided sub
-- If no row with that sub exists, the query returns no rows.
SELECT 1
FROM students
WHERE id = ?
LIMIT 1;
