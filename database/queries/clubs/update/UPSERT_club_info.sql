-- UPSERT_club_info.sql
-- Does: writes a club's website and description, creating the club_info row when
--   the club has none. The API passes merged values, so omitted fields are written back unchanged.
-- Used by: PATCH /clubs/{clubId}, inside its transaction.
-- Params, in order:
--   1. club id
--   2. website url, or NULL
--   3. description, or NULL
-- Returns: nothing.
INSERT INTO club_info (fk_club_id, website_url, description)
VALUES (?, ?, ?) AS new
ON DUPLICATE KEY UPDATE
  website_url = new.website_url,
  description = new.description;
