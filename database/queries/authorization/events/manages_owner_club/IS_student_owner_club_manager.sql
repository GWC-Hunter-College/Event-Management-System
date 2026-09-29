-- IS_student_owner_club_manager.sql
-- Does: checks whether a student is on the e-board of, or owns, an event's owning club.
--   Associate clubs don't count. Works for soft-deleted events, whose club links
--   stay until the purge.
-- Used by: POST /auth/events/{eventId}/restore (admins pass through IS_admin.sql instead).
-- Params, in order:
--   1. student id
--   2. event id
-- Returns: is_authorized, 1 or 0.
SELECT EXISTS (
  SELECT 1
  FROM events_to_clubs ec
  JOIN club_members cm
    ON cm.fk_club_id = ec.fk_club_id
  WHERE cm.fk_student_id = ?
    AND ec.fk_event_id = ?
    AND ec.club_is_event_owner = TRUE
    AND cm.role IN ('eboard', 'owner')
) AS is_authorized;
