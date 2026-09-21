-- Grabs all events of the clubs the student is a part of
-- Filters with startDate and endDate parameters if given
-- Returns in ascending order by start_date (for better viewing)

SELECT e.*
FROM club_members cm
JOIN events_to_clubs etc ON cm.fk_club_id = etc.fk_club_id
JOIN events e ON etc.fk_event_id = e.id
-- WHERE cm.fk_student_id = :student_id
WHERE cm.fk_student_id = 1 -- Testing Kelly's info
    AND e.status = 'posted'
    -- AND (:startDate IS NULL OR ev.event_date >= :startDate)
    -- AND (:endDate IS NULL OR ev.event_date <= :endDate)
    AND (e.start_date >= '2025-09-01')
    AND (e.end_date <= '2025-12-31')
ORDER BY e.start_date ASC;