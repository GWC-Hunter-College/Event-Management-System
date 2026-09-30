-- SELECT_club_drafts.sql
-- Does: lists a club's draft events, most recently updated first, one row per event.
-- Used by: GET /clubs/{clubId}/events/drafts, after clubs/get/EXISTS_club.sql (404)
--   and authorization/clubs/can_manage/IS_student_authorized_club.sql (403).
-- Params, in order:
--   1. club id
--   2. limit: 1 to 100, default 50, enforced by the API
--   3. offset: page * limit
-- Returns: event_details rows, all with status 'draft'.
SELECT ed.*
FROM event_details ed
WHERE ed.status = 'draft'
  AND EXISTS (
    SELECT 1
    FROM events_to_clubs ec
    WHERE ec.fk_event_id = ed.id
      AND ec.fk_club_id = ?
  )
ORDER BY ed.updated_at DESC, ed.id DESC
LIMIT ? OFFSET ?;
