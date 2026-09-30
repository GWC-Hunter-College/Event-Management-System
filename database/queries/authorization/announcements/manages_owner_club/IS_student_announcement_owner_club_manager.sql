-- IS_student_announcement_owner_club_manager.sql
-- Does: checks whether a student is on the e-board of, or owns, an announcement's
--   owning club. Co-owning clubs don't count. Works for soft-deleted announcements,
--   whose club links stay until the purge.
-- Used by: POST /auth/announcements/{announcementId}/restore (admins pass through IS_admin.sql instead).
-- Params, in order:
--   1. student id
--   2. announcement id
-- Returns: is_authorized, 1 or 0.
SELECT EXISTS (
  SELECT 1
  FROM announcements_to_clubs ac
  JOIN club_members cm
    ON cm.fk_club_id = ac.fk_club_id
  WHERE cm.fk_student_id = ?
    AND ac.fk_announcement_id = ?
    AND ac.club_is_announcement_owner = TRUE
    AND cm.role IN ('eboard', 'owner')
) AS is_authorized;
