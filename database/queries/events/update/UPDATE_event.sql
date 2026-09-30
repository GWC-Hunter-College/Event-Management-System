-- UPDATE_event.sql
-- Does: writes an event's fields after the API has merged the request into the
--   values read by SELECT_event_for_update.sql. Always bumps updated_at, so a
--   change to only the description or co-hosts still moves the draft up the list.
--   Status changes use the UPDATE_event_status_* queries, never this one.
-- Used by: PATCH /auth/events/{eventId}, inside its transaction.
-- Params, in order:
--   1. title
--   2. location, or NULL
--   3. rsvp link, or NULL
--   4. start date, UTC DATETIME
--   5. end date, UTC DATETIME, not before the start
--   6. timezone
--   7. event id
-- Returns: affected rows. It can be 0 when nothing changed within the same second,
--   so the locking read, not this count, is the existence check.
UPDATE events
SET
  title = ?,
  location = ?,
  rsvp_link = ?,
  start_date = ?,
  end_date = ?,
  timezone = ?,
  updated_at = CURRENT_TIMESTAMP
WHERE id = ?
  AND deleted_at IS NULL;
