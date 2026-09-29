-- DELETE_club_member.sql
-- Does: removes the caller's membership. An owner can leave only while the club
--   has another owner, so a club's last owner can't leave.
-- Used by: DELETE /clubs/{clubId}/members/me, after
--   clubs/members/update_role/SELECT_club_owners_for_update.sql in the same transaction.
-- Params, in order:
--   1. student id (the caller's sub)
--   2. club id
-- Returns: affected rows. 0 means not a member (404, check with IS_club_member.sql)
--   or the club's last owner (403).
DELETE cm
FROM club_members cm
LEFT JOIN (
  SELECT fk_club_id, COUNT(*) AS owner_count
  FROM club_members
  WHERE role = 'owner'
  GROUP BY fk_club_id
) o ON o.fk_club_id = cm.fk_club_id
WHERE cm.fk_student_id = ?
  AND cm.fk_club_id = ?
  AND (cm.role <> 'owner' OR COALESCE(o.owner_count, 0) > 1);
