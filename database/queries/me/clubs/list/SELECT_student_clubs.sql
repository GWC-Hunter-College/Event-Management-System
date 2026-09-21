-- Get clubs for a specific user with computed role
SELECT
  c.id,
  c.name,
  i.object_key AS thumbnail_url,
  CASE
    WHEN cm.member_is_owner = 1 THEN 'owner'
    WHEN cm.member_is_eboard = 1 THEN 'eboard'
    ELSE 'member'
  END AS role
FROM club_members cm
JOIN clubs c            ON c.id = cm.fk_club_id
LEFT JOIN images i      ON i.id = c.fk_logo_id
WHERE cm.fk_student_id = ?;
