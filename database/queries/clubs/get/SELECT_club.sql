-- SELECT_club_by_id.sql
SELECT
  c.id,
  c.name,
  i.object_key   AS thumbnail_url,
  ci.description AS description,
  ci.website_url AS website_url
FROM clubs c
LEFT JOIN images     i  ON i.id = c.fk_logo_id
LEFT JOIN club_info  ci ON ci.fk_club_id = c.id
WHERE c.id = ?
LIMIT 1;
