-- IS_student_authorized_event.sql
-- Authorizes a student to modify an event if they are an e-board member OR owner
-- of ANY club linked to that event.
-- Returns 1 if authorized, 0 otherwise.
-- Params:
--   ? = fk_event_id  (event id)
--   ? = fk_student_id (user sub / student id)

SELECT EXISTS (
  SELECT 1
  FROM events_to_clubs etc
  JOIN club_members cm
    ON cm.fk_club_id = etc.fk_club_id
  WHERE etc.fk_event_id = ?
    AND cm.fk_student_id = ?
    AND (cm.member_is_eboard = 1 OR cm.member_is_owner = 1)
) AS is_authorized;
