-- INSERT_event_description.sql
-- Does: stores a new event's description. Skipped when the body has none.
-- Used by: POST /clubs/{clubId}/events, inside the create transaction.
-- Params, in order:
--   1. event id
--   2. description
-- Returns: nothing.
INSERT INTO event_descriptions
  (fk_event_id, description)
VALUES
  (?, ?);
