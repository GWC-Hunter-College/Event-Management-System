INSERT INTO events 
    (
        fk_author_id, 
        fk_thumbnail_id, 
        title, 
        status, 
        start_date, 
        end_date, 
        location,
        timezone,
        rsvp_link, 
        created_at, 
        updated_at
    )
VALUES 
    (?, NULL, ?, 'drafted', ?, ?, ?, ?, NOW(), NOW());