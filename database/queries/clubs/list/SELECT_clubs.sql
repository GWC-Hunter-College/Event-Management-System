-- SELECT_clubs.sql
-- Does: lists clubs, optionally only verified ones, ordered by name.
-- Used by: GET /clubs and GET /clubs?verified=true
-- Params, in order:
--   1. verified only: TRUE for verified clubs only, FALSE for every club
-- Returns: id, name, thumbnail_url (the logo's object key, or NULL), description,
--   tags (JSON array of topic keys in the club's order), member_count, verified.
SELECT
  cd.id,
  cd.name,
  cd.thumbnail_url,
  cd.description,
  cd.tags,
  cd.member_count,
  cd.verified
FROM club_details cd
WHERE (? = FALSE OR cd.verified)
ORDER BY cd.name, cd.id;
