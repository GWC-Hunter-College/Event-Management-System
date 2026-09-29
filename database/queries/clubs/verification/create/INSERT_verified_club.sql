-- INSERT_verified_club.sql
-- Does: marks a club verified. Verifying a verified club changes nothing.
-- Used by: POST /clubs/{clubId}/verification, after the admin check.
-- Params, in order:
--   1. club id
-- Returns: affected rows. 0 means the club is already verified or doesn't exist
--   (clubs/get/EXISTS_club.sql tells them apart).
INSERT INTO verified_clubs (fk_club_id)
SELECT c.id
FROM clubs c
WHERE c.id = ?
  AND NOT EXISTS (
    SELECT 1
    FROM verified_clubs v
    WHERE v.fk_club_id = c.id
  );
