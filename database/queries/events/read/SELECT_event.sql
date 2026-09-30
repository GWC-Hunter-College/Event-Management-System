-- SELECT_event.sql
-- Does: reads one event that isn't soft-deleted, for the public read or a manager's read.
-- Used by: GET /events/{eventId} (public_only TRUE), GET /auth/events/{eventId} and
--   the PATCH /auth/events/{eventId} response (public_only FALSE, after the event role check).
-- Params, in order:
--   1. event id
--   2. public_only: TRUE returns only posted and cancelled events, FALSE any status
-- Returns: one event_details row, or no row (404).
--   The gallery comes from events/images/list/SELECT_event_images.sql.
SELECT ed.*
FROM event_details ed
WHERE ed.id = ?
  AND (? = FALSE OR ed.status IN ('posted', 'cancelled'));
