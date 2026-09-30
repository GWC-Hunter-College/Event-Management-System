-- SELECT_student_clubs.sql
-- Does: lists the clubs the caller belongs to, with the caller's role in each.
-- Used by: GET /me/clubs
-- Params, in order:
--   1. student id (the caller's sub)
-- Returns: id, name, thumbnail_url (the logo's object key, or NULL), role
--   ('member', 'eboard' or 'owner').
SELECT
  c.id,
  c.name,
  i.object_key AS thumbnail_url,
  cm.role
FROM club_members cm
JOIN clubs c ON c.id = cm.fk_club_id
LEFT JOIN images i ON i.id = c.fk_logo_id
WHERE cm.fk_student_id = ?
ORDER BY c.name, c.id;
