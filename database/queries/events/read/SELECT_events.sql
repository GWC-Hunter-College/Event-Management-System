-- Grabs every single event that is 'posted'
-- Filters with startDate and endDate parameters if given (everything after start and before end)
-- Returns in ascending order by start_date (for better viewing)
-- Paginates with limit and offset parameters when given

SELECT 
	ec.fk_club_id AS club_id, 
	ec.club_is_event_owner AS is_owner,
    i.object_key AS object_key,
	e.*,
    ed.description AS description
FROM events e
INNER JOIN events_to_clubs ec ON e.id = ec.fk_event_id
INNER JOIN clubs c ON ec.fk_club_id = c.id
LEFT JOIN images i ON c.fk_logo_id = i.id
LEFT JOIN event_descriptions ed ON e.id = ed.fk_event_id
WHERE status LIKE ?
    AND e.id LIKE ?
    AND e.start_date > ?
    AND e.end_date < ?
ORDER BY e.start_date ASC
LIMIT ? OFFSET ?;