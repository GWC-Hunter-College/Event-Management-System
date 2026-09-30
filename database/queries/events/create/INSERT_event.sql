-- INSERT_event.sql
-- Does: creates an event. created_at and updated_at take their defaults.
-- Used by: POST /clubs/{clubId}/events, first statement of the create transaction
--   (see "Transactions" in database/README.md).
-- Params, in order:
--   1. author id (the caller's sub)
--   2. title
--   3. location, or NULL for a draft
--   4. rsvp link, or NULL
--   5. status: 'draft' or 'posted' (the API defaults to 'draft' and rejects others)
--   6. start date, UTC DATETIME
--   7. end date, UTC DATETIME, not before the start
--   8. timezone: IANA name, default 'America/New_York' in the API
-- Returns: the new event id as the last insert id.
INSERT INTO events
  (fk_author_id, title, location, rsvp_link, status, start_date, end_date, timezone)
VALUES
  (?, ?, ?, ?, ?, ?, ?, ?);
