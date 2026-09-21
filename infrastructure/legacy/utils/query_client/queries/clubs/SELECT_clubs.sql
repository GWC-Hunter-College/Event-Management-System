-- Queries for clubs
-- Filters based on verificaiton, if true only grab verified clubs
-- if false, grab all clubs

SELECT 
  c.id,
  c.name,
  i.object_key AS thumbnail_url
FROM clubs c
LEFT JOIN images i
  ON i.id = c.fk_logo_id
LEFT JOIN verified_clubs vc
  ON vc.fk_club_id = c.id
WHERE (? = FALSE OR vc.fk_club_id IS NOT NULL);
