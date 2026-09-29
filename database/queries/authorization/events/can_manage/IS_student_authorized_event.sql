-- IS_student_authorized_event.sql
-- Does: checks whether a student is on the e-board of, or owns, ANY club linked to
--   an event (can manage the event).
-- Used by: GET, PATCH and DELETE /auth/events/{eventId}, the /auth/events image routes,
--   and the event thumbnail and gallery signers and confirms.
-- Params, in order (event first, as in the legacy query):
--   1. event id
--   2. student id
-- Returns: is_authorized, 1 or 0.
SELECT EXISTS (
  SELECT 1
  FROM events_to_clubs etc
  JOIN club_members cm
    ON cm.fk_club_id = etc.fk_club_id
  WHERE etc.fk_event_id = ?
    AND cm.fk_student_id = ?
    AND cm.role IN ('eboard', 'owner')
) AS is_authorized;
