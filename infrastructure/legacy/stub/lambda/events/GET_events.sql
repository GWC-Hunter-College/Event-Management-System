-- Grabs every single event that is 'posted'
-- Filters with startDate and endDate parameters if given (everything after start and before end)
-- Returns in ascending order by start_date (for better viewing)
-- Paginates with limit and offset parameters when given

SELECT *
FROM events
WHERE status = 'posted'
    AND (start_date > '2025-09-01')
    AND (end_date < '2025-12-31')
ORDER BY start_date ASC
LIMIT 10 OFFSET 10;