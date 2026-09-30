-- EXISTS_club.sql
-- Does: checks that a club exists, for a 404 before other work.
-- Used by: GET /clubs/{clubId}/events, GET /clubs/{clubId}/events/drafts,
--   POST /clubs/{clubId}/events, POST and DELETE /clubs/{clubId}/members/me,
--   GET /clubs/{clubId}/members, PUT /clubs/{clubId}/members/roles,
--   POST /clubs/{clubId}/thumbnails, POST /clubs/{clubId}/thumbnails/confirm,
--   PATCH /clubs/{clubId}, and POST and DELETE /clubs/{clubId}/verification.
-- Params, in order:
--   1. club id
-- Returns: club_exists, 1 or 0.
SELECT EXISTS (
  SELECT 1
  FROM clubs
  WHERE id = ?
) AS club_exists;
