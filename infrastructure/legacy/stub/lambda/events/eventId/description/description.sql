SELECT ed.fk_event_id, ed.description, e.status
FROM events e
INNER JOIN event_descriptions ed
  ON e.id  = ed.fk_event_id
WHERE e.id = ?
  AND e.status = "posted";

