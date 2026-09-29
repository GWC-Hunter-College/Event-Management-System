-- UPDATE_club_member_role.sql
-- Does: sets a member's role. Demoting the club's last owner affects 0 rows.
-- Used by: PUT /clubs/{clubId}/members/roles, after SELECT_club_owners_for_update.sql
--   and authorization/clubs/is_owner/IS_club_owner.sql (the caller must be an owner).
-- Params, in order:
--   1. club id
--   2. student id of the member to change
--   3. role: 'member', 'eboard' or 'owner'
-- Returns: affected rows. MySQL counts only changed rows, so 0 means one of: not a
--   member (404, check with IS_club_member.sql), already in that role (200), or the
--   last owner would be demoted (409).
UPDATE club_members cm
JOIN (
  SELECT
    CAST(? AS SIGNED) AS club_id,
    CAST(? AS CHAR(36)) AS student_id,
    CAST(? AS CHAR(6)) AS new_role
) p
  ON cm.fk_club_id = p.club_id
  AND cm.fk_student_id = p.student_id
LEFT JOIN (
  SELECT fk_club_id, COUNT(*) AS owner_count
  FROM club_members
  WHERE role = 'owner'
  GROUP BY fk_club_id
) o ON o.fk_club_id = cm.fk_club_id
SET cm.role = p.new_role
WHERE p.new_role IN ('member', 'eboard', 'owner')
  AND (
    p.new_role = 'owner'
    OR cm.role <> 'owner'
    OR COALESCE(o.owner_count, 0) > 1
  );
