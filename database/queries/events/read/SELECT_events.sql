-- SELECT_events.sql
-- Does: lists public events (posted and cancelled), one row per event, so paging
--   counts events rather than event-to-club rows.
-- Used by: GET /events
-- Params, in order:
--   1. when: 'upcoming' (not ended yet, start ascending), 'past' (ended, start
--      descending), or NULL (no time filter, start ascending). The API rejects other values.
--   2. range start, UTC DATETIME or NULL (inclusive)
--   3. range end, UTC DATETIME or NULL (exclusive)
--   4. limit: 1 to 100, default 50, enforced by the API
--   5. offset: page * limit
-- Date range: an event matches when it overlaps [range start, range end). The API
--   turns startDate=YYYY-MM-DD into 00:00 that day in America/New_York, and
--   endDate=YYYY-MM-DD into 00:00 the next day there (so the whole day is included),
--   both converted to UTC. Full timestamps with an offset are converted to UTC as given.
-- Returns: event_details rows (see the baseline for the columns).
SELECT ed.*
FROM event_details ed
CROSS JOIN (
  SELECT
    CAST(? AS CHAR(8)) AS when_filter,
    CAST(? AS DATETIME) AS start_bound,
    CAST(? AS DATETIME) AS end_bound
) p
WHERE ed.status IN ('posted', 'cancelled')
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
