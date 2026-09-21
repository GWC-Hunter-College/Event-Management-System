-- Grabs all 'posted' events for clubs a user has joined (by their student sub/id)
-- Supports filtering by event id (LIKE), date window, and pagination
-- Returns ascending by start_date

SELECT
  ec.fk_club_id AS club_id,
  ec.club_is_event_owner AS is_owner,
  i.object_key AS object_key,
  e.*,
  ed.description AS description
FROM events e
INNER JOIN events_to_clubs ec ON e.id = ec.fk_event_id
INNER JOIN clubs c            ON ec.fk_club_id = c.id
LEFT  JOIN images i           ON c.fk_logo_id = i.id
LEFT  JOIN event_descriptions ed ON e.id = ed.fk_event_id
WHERE e.status LIKE ?
  AND e.id LIKE ?
  AND e.start_date > ?
  AND e.end_date   < ?
  AND ec.fk_club_id IN (
    SELECT cm.fk_club_id
    FROM club_members cm
    WHERE cm.fk_student_id = ?     -- user sub / students.id
  )
ORDER BY e.start_date ASC
LIMIT ? OFFSET ?;
