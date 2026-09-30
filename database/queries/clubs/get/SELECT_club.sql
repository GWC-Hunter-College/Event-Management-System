-- SELECT_club.sql
-- Does: reads one club, verified or not.
-- Used by: GET /clubs/{clubId}, and PATCH /clubs/{clubId} to merge a partial update
--   and build its response.
-- Params, in order:
--   1. club id
-- Returns: id, name, thumbnail_url (the logo's object key, or NULL), website_url,
--   description, tags (JSON array), member_count, verified. No row means 404.
SELECT
  cd.id,
  cd.name,
  cd.thumbnail_url,
  cd.website_url,
  cd.description,
  cd.tags,
  cd.member_count,
  cd.verified
FROM club_details cd
WHERE cd.id = ?;
