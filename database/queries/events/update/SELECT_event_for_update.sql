-- SELECT_event_for_update.sql
-- Does: reads an event's editable fields and locks its row until the transaction
--   ends, so the API can merge a partial update or replace the flyer safely.
-- Used by: PATCH /auth/events/{eventId}, POST /clubs/{clubId}/events/{eventId}/thumbnails/confirm
--   and POST /clubs/{clubId}/events/{eventId}/images/confirm, as the first statement
--   of their transactions.
-- Params, in order:
--   1. event id
-- Returns: id, fk_thumbnail_id, title, location, rsvp_link, status, start_date,
--   end_date, timezone, description, owner_club_id. No row means 404.
SELECT
  e.id,
  e.fk_thumbnail_id,
  e.title,
  e.location,
  e.rsvp_link,
  e.status,
  e.start_date,
  e.end_date,
  e.timezone,
  ed.description,
  (
    SELECT ec.fk_club_id
    FROM events_to_clubs ec
    WHERE ec.fk_event_id = e.id
      AND ec.club_is_event_owner = TRUE
  ) AS owner_club_id
FROM events e
LEFT JOIN event_descriptions ed ON ed.fk_event_id = e.id
WHERE e.id = ?
  AND e.deleted_at IS NULL
FOR UPDATE OF e;
