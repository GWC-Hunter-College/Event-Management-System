-- UPDATE_event_status_cancelled.sql
-- Does: cancels an event. Only posted events can move to cancelled. Any other
--   change affects 0 rows.
-- Used by: PATCH /auth/events/{eventId} with "status": "cancelled", when the current
--   status isn't already cancelled.
-- Params, in order:
--   1. event id
-- Returns: affected rows. 0 means the event is missing or soft-deleted (404) or the
--   transition isn't allowed (400). events/read/SELECT_event_status.sql tells them apart.
UPDATE events
SET status = 'cancelled'
WHERE id = ?
  AND deleted_at IS NULL
  AND status = 'posted';
