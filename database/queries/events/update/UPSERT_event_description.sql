-- UPSERT_event_description.sql
-- Does: sets an event's description, creating the row when the event has none.
-- Used by: PATCH /auth/events/{eventId}, when the body includes description.
-- Params, in order:
--   1. event id
--   2. description, or NULL to clear it
-- Returns: nothing.
INSERT INTO event_descriptions (fk_event_id, description)
VALUES (?, ?) AS new
ON DUPLICATE KEY UPDATE description = new.description;
