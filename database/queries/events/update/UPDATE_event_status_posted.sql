-- UPDATE_event_status_posted.sql
-- Does: posts an event. Only draft and cancelled events can move to posted, and
--   only when they have a location. Any other change affects 0 rows.
-- Used by: PATCH /auth/events/{eventId} with "status": "posted", after UPDATE_event.sql,
--   when the current status isn't already posted.
-- Params, in order:
--   1. event id
-- Returns: affected rows. 0 means the event is missing or soft-deleted (404) or the
--   transition isn't allowed (400). events/read/SELECT_event_status.sql tells them apart.
UPDATE events
SET status = 'posted'
WHERE id = ?
  AND deleted_at IS NULL
  AND status IN ('draft', 'cancelled')
  AND location IS NOT NULL
  AND location <> '';
