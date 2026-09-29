-- SELECT_student_events.sql
-- Does: lists public events (posted and cancelled) linked to any club the caller
--   has joined, one row per event, with every linked club.
-- Used by: GET /me/events
-- Params, in order:
--   1. student id (the caller's sub)
--   2. when: 'upcoming', 'past' or NULL, as in events/read/SELECT_events.sql
--   3. range start, UTC DATETIME or NULL (inclusive)
--   4. range end, UTC DATETIME or NULL (exclusive)
--   5. limit: 1 to 100, default 50, enforced by the API
--   6. offset: page * limit
-- Date range: an event matches when it overlaps [range start, range end). The API
--   turns startDate=YYYY-MM-DD into 00:00 that day in America/New_York, and
--   endDate=YYYY-MM-DD into 00:00 the next day there (so the whole day is included),
--   both converted to UTC. Full timestamps with an offset are converted to UTC as given.
-- Returns: event_details rows.
SELECT ed.*
FROM event_details ed
CROSS JOIN (
  SELECT
    CAST(? AS CHAR(36)) AS student_id,
    CAST(? AS CHAR(8)) AS when_filter,
    CAST(? AS DATETIME) AS start_bound,
    CAST(? AS DATETIME) AS end_bound
) p
WHERE ed.status IN ('posted', 'cancelled')
  AND EXISTS (
    SELECT 1
    FROM events_to_clubs ec
    JOIN club_members cm ON cm.fk_club_id = ec.fk_club_id
    WHERE ec.fk_event_id = ed.id
      AND cm.fk_student_id = p.student_id
  )
  AND (
    p.when_filter IS NULL
    OR (p.when_filter = 'upcoming' AND ed.end_date > UTC_TIMESTAMP())
    OR (p.when_filter = 'past' AND ed.end_date <= UTC_TIMESTAMP())
  )
  AND (p.end_bound IS NULL OR ed.start_date < p.end_bound)
  AND (p.start_bound IS NULL OR ed.end_date > p.start_bound OR ed.start_date >= p.start_bound)
ORDER BY
  CASE WHEN p.when_filter = 'past' THEN ed.start_date END DESC,
  CASE WHEN p.when_filter = 'past' THEN ed.id END DESC,
  ed.start_date ASC,
  ed.id ASC
LIMIT ? OFFSET ?;
