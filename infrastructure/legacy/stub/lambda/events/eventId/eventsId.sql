SELECT e.*, ed.description
FROM events e
LEFT JOIN event_descriptions ed 
  ON e.id = ed.fk_event_id
WHERE e.status = 'posted';

