-- IS_student_authorized_club.sql
-- Does: checks whether a student is on a club's e-board or owns it (can manage the club).
-- Used by: every club write that e-board members and owners may make: POST /clubs/{clubId}/events,
--   GET /clubs/{clubId}/events/drafts, GET /clubs/{clubId}/members, PATCH /clubs/{clubId},
--   the logo signer and confirm, and DELETE /images/{imageId} (for the owning club).
-- Params, in order:
--   1. student id
--   2. club id
-- Returns: is_authorized, 1 or 0.
SELECT EXISTS (
  SELECT 1
  FROM club_members
  WHERE fk_student_id = ?
    AND fk_club_id = ?
    AND role IN ('eboard', 'owner')
) AS is_authorized;
