-- SELECT_club_announcements.sql
-- Does: lists a club's posted announcements, most recently posted first, one row
--   per announcement.
--   Each row carries every club the announcement belongs to, not only this one.
-- Used by: GET /clubs/{clubId}/announcements (check the club with clubs/get/EXISTS_club.sql for 404)
-- Params, in order:
--   1. club id
--   2. limit: 1 to 100, default 50, enforced by the API
--   3. offset: page * limit
-- Returns: announcement_details rows (see the baseline for the columns).
SELECT ad.*
FROM announcement_details ad
WHERE ad.status = 'posted'
  AND EXISTS (
    SELECT 1
    FROM announcements_to_clubs ac
    WHERE ac.fk_announcement_id = ad.id
      AND ac.fk_club_id = ?
  )
ORDER BY ad.posted_at DESC, ad.id DESC
LIMIT ? OFFSET ?;
