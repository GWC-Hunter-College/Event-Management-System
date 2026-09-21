INSERT INTO events_to_clubs
    (
        fk_event_id,
        fk_club_id,
        club_is_event_owner
    )
VALUES
    (?, ?, ?);