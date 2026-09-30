-- INSERT_club_tag.sql
-- Does: stores one of a club's topics. Run once per topic, in the user's order.
--   An unknown key, a wrong case, a duplicate or a fourth topic fails, and the API maps it to 400.
-- Used by: POST /clubs and PATCH /clubs/{clubId}, inside their transactions.
-- Params, in order:
--   1. club id
--   2. slot: the topic's position in the list plus 1, from 1 to 3
--   3. topic key, for example 'women in stem'
-- Returns: nothing.
INSERT INTO club_tags
  (fk_club_id, slot, tag)
VALUES
  (?, ?, ?);
