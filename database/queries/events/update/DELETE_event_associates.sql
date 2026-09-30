-- DELETE_event_associates.sql
-- Does: removes an event's co-host links, keeping the owner link, before the new
--   list is inserted with events/create/INSERT_event_club_link.sql.
-- Used by: PATCH /auth/events/{eventId}, when the body includes associates.
-- Params, in order:
--   1. event id
-- Returns: nothing.
DELETE FROM events_to_clubs
WHERE fk_event_id = ?
  AND club_is_event_owner = FALSE;
