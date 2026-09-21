-- make 4 users 
INSERT INTO students (id, email)
VALUES
  (1, 'kelly@example.com'),
  (2, 'kyle@example.com'), 
  (3, 'shohruz@example.com'),
  (4, 'anthony@example.com'),
  (5, 'maria@example.com'),
  (6, 'james@example.com'),
  (7, 'sophia@example.com'),
  (8, 'li@example.com'),
  (9, 'david@example.com'),
  (10, 'fatima@example.com');

-- random cooking club 
INSERT INTO clubs (name, fk_thumbnail_id)
VALUES ('cooking club', NULL);

-- make maria make cooking club 
INSERT INTO club_members (fk_student_id, fk_club_id, member_is_eboard, member_is_owner)
SELECT s.id, c.id, TRUE, TRUE
FROM students s
JOIN clubs c ON c.name = 'cooking club'
WHERE s.email = 'maria@example.com';

-- make anthony join 
INSERT INTO club_members (fk_student_id, fk_club_id, member_is_eboard, member_is_owner)
SELECT s.id, c.id, FALSE, FALSE
FROM students s
JOIN clubs c ON c.name = 'cooking club'
WHERE s.email = 'anthony@example.com';

-- make james join 
INSERT INTO club_members (fk_student_id, fk_club_id, member_is_eboard, member_is_owner)
SELECT s.id, c.id, FALSE, FALSE
FROM students s
JOIN clubs c ON c.name = 'cooking club'
WHERE s.email = 'james@example.com';

-- make james eboard member
UPDATE club_members cm
JOIN students s ON cm.fk_student_id = s.id
JOIN clubs c ON cm.fk_club_id = c.id
SET cm.member_is_eboard = TRUE
WHERE s.email = 'james@example.com'
  AND c.name = 'cooking club';

-- make kelly join 
INSERT INTO club_members (fk_student_id, fk_club_id, member_is_eboard, member_is_owner)
SELECT s.id, c.id, FALSE, FALSE
FROM students s
JOIN clubs c ON c.name = 'cooking club'
WHERE s.email = 'kelly@example.com';


-- make gwc a club 
INSERT INTO clubs (name, fk_thumbnail_id)
VALUES ('Girls Who Code @ Hunter', NULL);

-- make kelly make a gwc club 
INSERT INTO club_members (fk_student_id, fk_club_id, member_is_eboard, member_is_owner)
SELECT s.id, c.id, TRUE, TRUE
FROM students s
JOIN clubs c ON c.name = 'Girls Who Code @ Hunter'
WHERE s.email = 'kelly@example.com';

-- make sophia join 
INSERT INTO club_members (fk_student_id, fk_club_id, member_is_eboard, member_is_owner)
SELECT s.id, c.id, FALSE, FALSE
FROM students s
JOIN clubs c ON c.name = 'Girls Who Code @ Hunter'
WHERE s.email = 'sophia@example.com';

-- make anthony join 
INSERT INTO club_members (fk_student_id, fk_club_id, member_is_eboard, member_is_owner)
SELECT s.id, c.id, FALSE, FALSE
FROM students s
JOIN clubs c ON c.name = 'Girls Who Code @ Hunter'
WHERE s.email = 'anthony@example.com';

-- make anthony eboard member
UPDATE club_members cm
JOIN students s ON cm.fk_student_id = s.id
JOIN clubs c ON cm.fk_club_id = c.id
SET cm.member_is_eboard = TRUE
WHERE s.email = 'anthony@example.com'
  AND c.name = 'Girls Who Code @ Hunter';


-- make kyle join 
INSERT INTO club_members (fk_student_id, fk_club_id, member_is_eboard, member_is_owner)
SELECT s.id, c.id, TRUE, FALSE
FROM students s
JOIN clubs c ON c.name = 'Girls Who Code @ Hunter'
WHERE s.email = 'kyle@example.com';


-- make kyle eboard member
UPDATE club_members cm
JOIN students s ON cm.fk_student_id = s.id
JOIN clubs c ON cm.fk_club_id = c.id
SET cm.member_is_eboard = TRUE
WHERE s.email = 'kyle@example.com'
  AND c.name = 'Girls Who Code @ Hunter';

-- make li join 
INSERT INTO club_members (fk_student_id, fk_club_id, member_is_eboard, member_is_owner)
SELECT s.id, c.id, TRUE, FALSE
FROM students s
JOIN clubs c ON c.name = 'Girls Who Code @ Hunter'
WHERE s.email = 'li@example.com';


-- kelly makes an event 

INSERT INTO events (
  fk_author_id, fk_thumbnail_id, title, location, rsvp_link, status,
  start_date, end_date, timezone, created_at, updated_at, deleted_at
)
SELECT
  s.id, NULL,
  'First Kickoff Meeting',
  'Room 101, Hunter College',
  'some link',
  'posted',
  '2025-09-15 17:00:00',
  '2025-09-15 19:00:00',
  'America/New_York',
  NOW(), NOW(), NULL
FROM students s
WHERE s.email = 'kelly@example.com'; 

SET @event_id = LAST_INSERT_ID();

INSERT INTO event_descriptions (fk_event_id, description)
VALUES (@event_id, 'meow');

INSERT INTO events_to_clubs (fk_event_id, fk_club_id, club_is_event_owner)
SELECT @event_id, c.id, TRUE
FROM clubs c
WHERE c.name = 'Girls Who Code @ Hunter';

-- not all clubs will have thumbnails in the beginning 
INSERT INTO images (fk_event_id, fk_club_id, purpose, object_key, created_at)
SELECT e.id, c.id, 'thumbnail', 'https://bit.ly/fcc-relaxing-cat', NOW()
FROM events e
JOIN clubs c ON c.name = 'Girls Who Code @ Hunter'
WHERE e.title = 'First Kickoff Meeting';


-- make hunter cs a club 
INSERT INTO clubs (name, fk_thumbnail_id)
VALUES ('Hunter CS Club', NULL);

-- make shohruz make hunter cs club 
INSERT INTO club_members (fk_student_id, fk_club_id, member_is_eboard, member_is_owner)
SELECT s.id, c.id, TRUE, TRUE
FROM students s
JOIN clubs c ON c.name = 'Hunter CS Club'
WHERE s.email = 'shohruz@example.com';

-- shohruz makes two events
INSERT INTO events (
  fk_author_id, fk_thumbnail_id, title, location, rsvp_link, status,
  start_date, end_date, timezone, created_at, updated_at, deleted_at
)
SELECT
  s.id, NULL,
  'cs Meeting 2000',
  'Room 304, narnia',
  'some form!!!',
  'drafted',
  '2025-09-25 12:00:00',
  '2025-09-21 9:00:00',
  'America/New_York',
  NOW(), NOW(), NULL
FROM students s
WHERE s.email = 'shohruz@example.com'; 

SET @event_id = LAST_INSERT_ID();

INSERT INTO event_descriptions (fk_event_id, description)
VALUES (@event_id, 'cs club shennanigan');

INSERT INTO events_to_clubs (fk_event_id, fk_club_id, club_is_event_owner)
SELECT @event_id, c.id, TRUE
FROM clubs c
WHERE c.name = 'Hunter CS Club';

INSERT INTO images (fk_event_id, fk_club_id, purpose, object_key, created_at)
SELECT e.id, c.id, 'thumbnail', 'dummy img path 2', NOW()
FROM events e
JOIN clubs c ON c.name = 'Hunter CS Club'
WHERE e.title = 'cs Meeting 2000';

INSERT INTO events (
  fk_author_id, fk_thumbnail_id, title, location, rsvp_link, status,
  start_date, end_date, timezone, created_at, updated_at, deleted_at
)
SELECT
  s.id, NULL,
  'workshop',
  'hogwarts',
  'some form link',
  'drafted',
  '2025-10-28 11:00:00',
  '2025-10-28 19:00:00',
  'America/New_York',
  NOW(), NOW(), NULL
FROM students s
WHERE s.email = 'shohruz@example.com'; 

SET @event_id = LAST_INSERT_ID();

INSERT INTO event_descriptions (fk_event_id, description)
VALUES (@event_id, 'magic will happen');

INSERT INTO events_to_clubs (fk_event_id, fk_club_id, club_is_event_owner)
SELECT @event_id, c.id, TRUE
FROM clubs c
WHERE c.name = 'Hunter CS Club';

INSERT INTO images (fk_event_id, fk_club_id, purpose, object_key, created_at)
SELECT e.id, c.id, 'thumbnail', 'dummy img path', NOW()
FROM events e
JOIN clubs c ON c.name = 'Hunter CS Club'
WHERE e.title = 'workshop';

-- kyle join 
INSERT INTO club_members (fk_student_id, fk_club_id, member_is_eboard, member_is_owner)
SELECT s.id, c.id, TRUE, FALSE
FROM students s
JOIN clubs c ON c.name = 'Hunter CS Club'
WHERE s.email = 'kyle@example.com';

-- kyle eboard
UPDATE club_members cm
JOIN students s ON cm.fk_student_id = s.id
JOIN clubs c ON cm.fk_club_id = c.id
SET cm.member_is_eboard = TRUE
WHERE s.email = 'kyle@example.com'
  AND c.name = 'Hunter CS Club';

-- kyle make event
INSERT INTO events (
  fk_author_id, fk_thumbnail_id, title, location, rsvp_link, status,
  start_date, end_date, timezone, created_at, updated_at, deleted_at
)
SELECT
  s.id, NULL,
  'cs Meeting',
  'Room 234, Hunter College',
  'some form',
  'drafted',
  '2025-09-28 11:00:00',
  '2025-09-28 19:00:00',
  'America/New_York',
  NOW(), NOW(), NULL
FROM students s
WHERE s.email = 'kyle@example.com'; 

SET @event_id = LAST_INSERT_ID();

INSERT INTO event_descriptions (fk_event_id, description)
VALUES (@event_id, 'cs club things');

INSERT INTO events_to_clubs (fk_event_id, fk_club_id, club_is_event_owner)
SELECT @event_id, c.id, TRUE
FROM clubs c
WHERE c.name = 'Hunter CS Club';

INSERT INTO images (fk_event_id, fk_club_id, purpose, object_key, created_at)
SELECT e.id, c.id, 'thumbnail', 'dummy img path 3', NOW()
FROM events e
JOIN clubs c ON c.name = 'Hunter CS Club'
WHERE e.title = 'cs Meeting';


-- make kyle an admin
INSERT INTO admins (fk_student_id)
SELECT s.id
FROM students s
WHERE s.email= "kyle@example.com";

-- make anthony an admin 
INSERT INTO admins (fk_student_id)
SELECT s.id
FROM students s
WHERE s.email= "anthony@example.com";

-- verify Hunter CS Club
INSERT INTO verified_clubs (fk_club_id)
SELECT c.id
FROM clubs c 
WHERE c.name = 'Hunter CS Club';

-- verify Girls Who Code @ Hunter
INSERT INTO verified_clubs (fk_club_id)
SELECT c.id
FROM clubs c 
WHERE c.name = 'Girls Who Code @ Hunter';

-- note, when using last insert id, you need to make sure this is created right after the event is made 
