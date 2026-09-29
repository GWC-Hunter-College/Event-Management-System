-- DELETE_club_tags.sql
-- Does: removes all of a club's topics before the new list is inserted.
-- Used by: PATCH /clubs/{clubId} when the body includes tags, inside its transaction.
-- Params, in order:
--   1. club id
-- Returns: nothing.
DELETE FROM club_tags
WHERE fk_club_id = ?;
