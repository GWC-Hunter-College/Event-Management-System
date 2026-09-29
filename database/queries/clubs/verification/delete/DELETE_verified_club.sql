-- DELETE_verified_club.sql
-- Does: removes a club's verification. Memberships and roles are unchanged.
-- Used by: DELETE /clubs/{clubId}/verification, after the admin check.
-- Params, in order:
--   1. club id
-- Returns: affected rows. 0 means the club wasn't verified.
DELETE FROM verified_clubs
WHERE fk_club_id = ?;
