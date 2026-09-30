-- SELECT_club_announcement_drafts.sql
-- Does: lists a club's draft announcements, most recently updated first.
-- Used by: GET /clubs/{clubId}/announcements/drafts, after clubs/get/EXISTS_club.sql (404)
--   and authorization/clubs/can_manage/IS_student_authorized_club.sql (403).
-- Params, in order:
--   1. club id
--   2. limit: 1 to 100, default 50, enforced by the API
--   3. offset: page * limit
-- Returns: announcement_details rows, all with status 'draft'.
SELECT ad.*
FROM announcement_details ad
WHERE ad.status = 'draft'
  AND EXISTS (
    SELECT 1
    FROM announcements_to_clubs ac
    WHERE ac.fk_announcement_id = ad.id
      AND ac.fk_club_id = ?
  )
ORDER BY ad.updated_at DESC, ad.id DESC
LIMIT ? OFFSET ?;
