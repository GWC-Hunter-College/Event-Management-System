-- seed students
INSERT INTO students (id, email) VALUES
  ('11111111-1111-1111-1111-111111111111', 'kelly@example.com'),
  ('22222222-2222-2222-2222-222222222222', 'kyle@example.com'),
  ('33333333-3333-3333-3333-333333333333', 'shohruz@example.com'),
  ('44444444-4444-4444-4444-444444444444', 'anthony@example.com'),
  ('55555555-5555-5555-5555-555555555555', 'maria@example.com'),
  ('66666666-6666-6666-6666-666666666666', 'james@example.com'),
  ('77777777-7777-7777-7777-777777777777', 'sophia@example.com'),
  ('88888888-8888-8888-8888-888888888888', 'li@example.com'),
  ('99999999-9999-9999-9999-999999999999', 'david@example.com'),
  ('aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaaaa', 'fatima@example.com');

-- seed student_info (FK to students.id)
INSERT INTO student_info (fk_student_id, username, first_name, last_name) VALUES
  ('11111111-1111-1111-1111-111111111111', 'kelly', 'Kelly', 'Smith'),
  ('22222222-2222-2222-2222-222222222222', 'kyle', 'Kyle', 'Jones'),
  ('33333333-3333-3333-3333-333333333333', 'shohruz', 'Shohruz', 'Aliyev'),
  ('44444444-4444-4444-4444-444444444444', 'anth', 'Anthony', 'Brown'),
  ('55555555-5555-5555-5555-555555555555', 'maria', 'Maria', 'Lopez'),
  ('66666666-6666-6666-6666-666666666666', 'james', 'James', 'Smith'),
  ('77777777-7777-7777-7777-777777777777', 'soph', 'Sophia', 'Wang'),
  ('88888888-8888-8888-8888-888888888888', 'li', 'Li', 'Zhang'),
  ('99999999-9999-9999-9999-999999999999', 'david', 'David', 'Kim'),
  ('aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaaaa', 'fat', 'Fatima', 'Hussein');

-- create cooking club (no logo yet)
INSERT INTO clubs (name, fk_logo_id) VALUES ('Cooking Club', NULL);
INSERT INTO club_info (fk_club_id, website_url)
SELECT id, 'https://www.example.com/cooking' FROM clubs WHERE name='Cooking Club';

-- add maria as owner of cooking club
INSERT INTO club_members (fk_student_id, fk_club_id, member_is_eboard, member_is_owner)
SELECT '55555555-5555-5555-5555-555555555555', id, TRUE, TRUE FROM clubs WHERE name='Cooking Club';

-- add anthony, james, kelly as members
INSERT INTO club_members (fk_student_id, fk_club_id, member_is_eboard, member_is_owner)
SELECT '44444444-4444-4444-4444-444444444444', id, FALSE, FALSE FROM clubs WHERE name='Cooking Club';
INSERT INTO club_members (fk_student_id, fk_club_id, member_is_eboard, member_is_owner)
SELECT '66666666-6666-6666-6666-666666666666', id, TRUE, FALSE FROM clubs WHERE name='Cooking Club';
INSERT INTO club_members (fk_student_id, fk_club_id, member_is_eboard, member_is_owner)
SELECT '11111111-1111-1111-1111-111111111111', id, FALSE, FALSE FROM clubs WHERE name='Cooking Club';

-- make GWC club
INSERT INTO clubs (name, fk_logo_id) VALUES ('Girls Who Code @ Hunter', NULL);
INSERT INTO club_info (fk_club_id, website_url)
SELECT id, 'https://www.example.com/gwc' FROM clubs WHERE name='Girls Who Code @ Hunter';

-- assign kelly as owner
INSERT INTO club_members (fk_student_id, fk_club_id, member_is_eboard, member_is_owner)
SELECT '11111111-1111-1111-1111-111111111111', id, TRUE, TRUE FROM clubs WHERE name='Girls Who Code @ Hunter';

-- add sophia and anthony, anthony becomes eboard
INSERT INTO club_members (fk_student_id, fk_club_id, member_is_eboard, member_is_owner)
SELECT '77777777-7777-7777-7777-777777777777', id, FALSE, FALSE FROM clubs WHERE name='Girls Who Code @ Hunter';
INSERT INTO club_members (fk_student_id, fk_club_id, member_is_eboard, member_is_owner)
SELECT '44444444-4444-4444-4444-444444444444', id, TRUE, FALSE FROM clubs WHERE name='Girls Who Code @ Hunter';

-- add kyle and li
INSERT INTO club_members (fk_student_id, fk_club_id, member_is_eboard, member_is_owner)
SELECT '22222222-2222-2222-2222-222222222222', id, TRUE, FALSE FROM clubs WHERE name='Girls Who Code @ Hunter';
INSERT INTO club_members (fk_student_id, fk_club_id, member_is_eboard, member_is_owner)
SELECT '88888888-8888-8888-8888-888888888888', id, TRUE, FALSE FROM clubs WHERE name='Girls Who Code @ Hunter';

-- create hunter cs club
INSERT INTO clubs (name, fk_logo_id) VALUES ('Hunter CS Club', NULL);
INSERT INTO club_info (fk_club_id, website_url)
SELECT id, 'https://www.example.com/huntercs' FROM clubs WHERE name='Hunter CS Club';

-- shohruz is owner
INSERT INTO club_members (fk_student_id, fk_club_id, member_is_eboard, member_is_owner)
SELECT '33333333-3333-3333-3333-333333333333', id, TRUE, TRUE FROM clubs WHERE name='Hunter CS Club';


INSERT INTO events (
  fk_author_id, fk_thumbnail_id, title, location, rsvp_link, status,
  start_date, end_date, timezone, created_at, updated_at
)
SELECT '11111111-1111-1111-1111-111111111111',
       NULL,
       'First Kickoff Meeting','Room 101','some link','posted',
       '2025-09-15 17:00:00','2025-09-15 19:00:00','America/New_York',NOW(),NOW();

SET @event_id = LAST_INSERT_ID();

INSERT INTO event_descriptions (fk_event_id, description)
VALUES(@event_id,'meow');

INSERT INTO events_to_clubs (fk_event_id,fk_club_id,club_is_event_owner)
SELECT @event_id,id,TRUE FROM clubs WHERE name='Girls Who Code @ Hunter';

-- verify clubs
INSERT INTO verified_clubs (fk_club_id) SELECT id FROM clubs WHERE name='Hunter CS Club';
INSERT INTO verified_clubs (fk_club_id) SELECT id FROM clubs WHERE name='Girls Who Code @ Hunter';

-- admins
INSERT INTO admins (fk_student_id) VALUES
  ('22222222-2222-2222-2222-222222222222'), -- kyle
  ('44444444-4444-4444-4444-444444444444'); -- anthony

-- ============================================
--  EXTRA STUDENTS (adds 30 more)
-- ============================================
INSERT INTO students (id, email) VALUES
  ('bbbbbbbb-bbbb-bbbb-bbbb-bbbbbbbbbbbb', 'aaron@example.com'),
  ('cccccccc-cccc-cccc-cccc-cccccccccccc', 'nina@example.com'),
  ('dddddddd-dddd-dddd-dddd-dddddddddddd', 'ethan@example.com'),
  ('eeeeeeee-eeee-eeee-eeee-eeeeeeeeeeee', 'olivia@example.com'),
  ('ffffffff-ffff-ffff-ffff-ffffffffffff', 'ryan@example.com'),
  ('11111111-aaaa-aaaa-aaaa-aaaaaaaaaaaa', 'emma@example.com'),
  ('22222222-bbbb-bbbb-bbbb-bbbbbbbbbbbb', 'michael@example.com'),
  ('33333333-cccc-cccc-cccc-cccccccccccc', 'isabella@example.com'),
  ('44444444-dddd-dddd-dddd-dddddddddddd', 'daniel@example.com'),
  ('55555555-eeee-eeee-eeee-eeeeeeeeeeee', 'grace@example.com'),
  ('66666666-ffff-ffff-ffff-ffffffffffff', 'alex@example.com'),
  ('77777777-aaaa-aaaa-aaaa-aaaaaaaaaaaa', 'mia@example.com'),
  ('88888888-bbbb-bbbb-bbbb-bbbbbbbbbbbb', 'jackson@example.com'),
  ('99999999-cccc-cccc-cccc-cccccccccccc', 'hannah@example.com'),
  ('aaaaaaaa-bbbb-bbbb-bbbb-bbbbbbbbbbbb', 'noah@example.com'),
  ('bbbbbbbb-cccc-cccc-cccc-cccccccccccc', 'ava@example.com'),
  ('cccccccc-dddd-dddd-dddd-dddddddddddd', 'william@example.com'),
  ('dddddddd-eeee-eeee-eeee-eeeeeeeeeeee', 'lily@example.com'),
  ('eeeeeeee-ffff-ffff-ffff-ffffffffffff', 'benjamin@example.com'),
  ('ffffffff-aaaa-aaaa-aaaa-aaaaaaaaaaaa', 'zoe@example.com'),
  ('12121212-1212-1212-1212-121212121212', 'sebastian@example.com'),
  ('13131313-1313-1313-1313-131313131313', 'ella@example.com'),
  ('14141414-1414-1414-1414-141414141414', 'lucas@example.com'),
  ('15151515-1515-1515-1515-151515151515', 'scarlett@example.com'),
  ('16161616-1616-1616-1616-161616161616', 'leo@example.com'),
  ('17171717-1717-1717-1717-171717171717', 'claire@example.com'),
  ('18181818-1818-1818-1818-181818181818', 'josh@example.com'),
  ('19191919-1919-1919-1919-191919191919', 'audrey@example.com'),
  ('20202020-2020-2020-2020-202020202020', 'adam@example.com'),
  ('21212121-2121-2121-2121-212121212121', 'sarah@example.com');

-- matching student_info
INSERT INTO student_info (fk_student_id, username, first_name, last_name) VALUES
  ('bbbbbbbb-bbbb-bbbb-bbbb-bbbbbbbbbbbb','aaron','Aaron','Taylor'),
  ('cccccccc-cccc-cccc-cccc-cccccccccccc','nina','Nina','Martinez'),
  ('dddddddd-dddd-dddd-dddd-dddddddddddd','ethan','Ethan','Lee'),
  ('eeeeeeee-eeee-eeee-eeee-eeeeeeeeeeee','olivia','Olivia','Davis'),
  ('ffffffff-ffff-ffff-ffff-ffffffffffff','ryan','Ryan','Nguyen'),
  ('11111111-aaaa-aaaa-aaaa-aaaaaaaaaaaa','emma','Emma','Lewis'),
  ('22222222-bbbb-bbbb-bbbb-bbbbbbbbbbbb','michael','Michael','Patel'),
  ('33333333-cccc-cccc-cccc-cccccccccccc','isabella','Isabella','Chen'),
  ('44444444-dddd-dddd-dddd-dddddddddddd','daniel','Daniel','Singh'),
  ('55555555-eeee-eeee-eeee-eeeeeeeeeeee','grace','Grace','Johnson'),
  ('66666666-ffff-ffff-ffff-ffffffffffff','alex','Alex','Rodriguez'),
  ('77777777-aaaa-aaaa-aaaa-aaaaaaaaaaaa','mia','Mia','Rossi'),
  ('88888888-bbbb-bbbb-bbbb-bbbbbbbbbbbb','jackson','Jackson','White'),
  ('99999999-cccc-cccc-cccc-cccccccccccc','hannah','Hannah','Morales'),
  ('aaaaaaaa-bbbb-bbbb-bbbb-bbbbbbbbbbbb','noah','Noah','Sharma'),
  ('bbbbbbbb-cccc-cccc-cccc-cccccccccccc','ava','Ava','Khan'),
  ('cccccccc-dddd-dddd-dddd-dddddddddddd','william','William','Brown'),
  ('dddddddd-eeee-eeee-eeee-eeeeeeeeeeee','lily','Lily','Stein'),
  ('eeeeeeee-ffff-ffff-ffff-ffffffffffff','ben','Benjamin','Miller'),
  ('ffffffff-aaaa-aaaa-aaaa-aaaaaaaaaaaa','zoe','Zoe','Adams'),
  ('12121212-1212-1212-1212-121212121212','seb','Sebastian','Torres'),
  ('13131313-1313-1313-1313-131313131313','ella','Ella','Hoffman'),
  ('14141414-1414-1414-1414-141414141414','lucas','Lucas','Ali'),
  ('15151515-1515-1515-1515-151515151515','scarlett','Scarlett','Baker'),
  ('16161616-1616-1616-1616-161616161616','leo','Leo','Carter'),
  ('17171717-1717-1717-1717-171717171717','claire','Claire','Ng'),
  ('18181818-1818-1818-1818-181818181818','josh','Josh','Ward'),
  ('19191919-1919-1919-1919-191919191919','audrey','Audrey','Goldberg'),
  ('20202020-2020-2020-2020-202020202020','adam','Adam','Hernandez'),
  ('21212121-2121-2121-2121-212121212121','sarah','Sarah','Young');

-- ============================================
--  MORE CLUBS
-- ============================================
INSERT INTO clubs (name,fk_logo_id) VALUES
  ('Math Club',NULL),
  ('Drama Club',NULL),
  ('Robotics Club',NULL),
  ('Art Club',NULL),
  ('Debate Society',NULL),
  ('Photography Club',NULL),
  ('Hiking Club',NULL),
  ('Finance Society',NULL),
  ('Music Club',NULL);

-- give each club a URL
INSERT INTO club_info (fk_club_id, website_url)
SELECT id,'https://www.example.com/math' FROM clubs WHERE name='Math Club';
INSERT INTO club_info (fk_club_id, website_url)
SELECT id,'https://www.example.com/drama' FROM clubs WHERE name='Drama Club';
INSERT INTO club_info (fk_club_id, website_url)
SELECT id,'https://www.example.com/robotics' FROM clubs WHERE name='Robotics Club';
INSERT INTO club_info (fk_club_id, website_url)
SELECT id,'https://www.example.com/art' FROM clubs WHERE name='Art Club';
INSERT INTO club_info (fk_club_id, website_url)
SELECT id,'https://www.example.com/debate' FROM clubs WHERE name='Debate Society';
INSERT INTO club_info (fk_club_id, website_url)
SELECT id,'https://www.example.com/photo' FROM clubs WHERE name='Photography Club';
INSERT INTO club_info (fk_club_id, website_url)
SELECT id,'https://www.example.com/hiking' FROM clubs WHERE name='Hiking Club';
INSERT INTO club_info (fk_club_id, website_url)
SELECT id,'https://www.example.com/finance' FROM clubs WHERE name='Finance Society';
INSERT INTO club_info (fk_club_id, website_url)
SELECT id,'https://www.example.com/music' FROM clubs WHERE name='Music Club';

-- assign owners (first in list is owner)
INSERT INTO club_members (fk_student_id,fk_club_id,member_is_eboard,member_is_owner)
SELECT 'bbbbbbbb-bbbb-bbbb-bbbb-bbbbbbbbbbbb', id, TRUE, TRUE FROM clubs WHERE name='Math Club';
INSERT INTO club_members (fk_student_id,fk_club_id,member_is_eboard,member_is_owner)
SELECT 'cccccccc-cccc-cccc-cccc-cccccccccccc', id, TRUE, TRUE FROM clubs WHERE name='Drama Club';
INSERT INTO club_members (fk_student_id,fk_club_id,member_is_eboard,member_is_owner)
SELECT 'dddddddd-dddd-dddd-dddd-dddddddddddd', id, TRUE, TRUE FROM clubs WHERE name='Robotics Club';
INSERT INTO club_members (fk_student_id,fk_club_id,member_is_eboard,member_is_owner)
SELECT 'eeeeeeee-eeee-eeee-eeee-eeeeeeeeeeee', id, TRUE, TRUE FROM clubs WHERE name='Art Club';
INSERT INTO club_members (fk_student_id,fk_club_id,member_is_eboard,member_is_owner)
SELECT 'ffffffff-ffff-ffff-ffff-ffffffffffff', id, TRUE, TRUE FROM clubs WHERE name='Debate Society';
INSERT INTO club_members (fk_student_id,fk_club_id,member_is_eboard,member_is_owner)
SELECT '11111111-aaaa-aaaa-aaaa-aaaaaaaaaaaa', id, TRUE, TRUE FROM clubs WHERE name='Photography Club';
INSERT INTO club_members (fk_student_id,fk_club_id,member_is_eboard,member_is_owner)
SELECT '22222222-bbbb-bbbb-bbbb-bbbbbbbbbbbb', id, TRUE, TRUE FROM clubs WHERE name='Hiking Club';
INSERT INTO club_members (fk_student_id,fk_club_id,member_is_eboard,member_is_owner)
SELECT '33333333-cccc-cccc-cccc-cccccccccccc', id, TRUE, TRUE FROM clubs WHERE name='Finance Society';
INSERT INTO club_members (fk_student_id,fk_club_id,member_is_eboard,member_is_owner)
SELECT '44444444-dddd-dddd-dddd-dddddddddddd', id, TRUE, TRUE FROM clubs WHERE name='Music Club';

-- add a few members to Math Club as example
INSERT INTO club_members (fk_student_id,fk_club_id,member_is_eboard,member_is_owner)
SELECT '55555555-eeee-eeee-eeee-eeeeeeeeeeee',id,FALSE,FALSE FROM clubs WHERE name='Math Club';
INSERT INTO club_members (fk_student_id,fk_club_id,member_is_eboard,member_is_owner)
SELECT '66666666-ffff-ffff-ffff-ffffffffffff',id,FALSE,FALSE FROM clubs WHERE name='Math Club';

-- ============================================
--  MORE EVENTS
-- ============================================
INSERT INTO events (
  fk_author_id,fk_thumbnail_id,title,location,rsvp_link,status,
  start_date,end_date,timezone,created_at,updated_at
)
SELECT 'bbbbbbbb-bbbb-bbbb-bbbb-bbbbbbbbbbbb',
       NULL,
       'Math Club Orientation','Room 201','mathlink','posted',
       '2025-09-20 18:00:00','2025-09-20 20:00:00','America/New_York',NOW(),NOW();

SET @event2 = LAST_INSERT_ID();
INSERT INTO event_descriptions (fk_event_id,description)
VALUES(@event2,'Welcome to the Math Club!');

INSERT INTO events_to_clubs (fk_event_id,fk_club_id,club_is_event_owner)
SELECT @event2,id,TRUE FROM clubs WHERE name='Math Club';

-- second event for Drama Club
INSERT INTO events (
  fk_author_id,fk_thumbnail_id,title,location,rsvp_link,status,
  start_date,end_date,timezone,created_at,updated_at
)
SELECT 'cccccccc-cccc-cccc-cccc-cccccccccccc',
       NULL,
       'Drama Club Play Auditions','Auditorium','dramalink','posted',
       '2025-10-01 17:00:00','2025-10-01 21:00:00','America/New_York',NOW(),NOW();

SET @event3 = LAST_INSERT_ID();
INSERT INTO event_descriptions (fk_event_id,description)
VALUES(@event3,'Auditions for fall play!');

INSERT INTO events_to_clubs (fk_event_id,fk_club_id,club_is_event_owner)
SELECT @event3,id,TRUE FROM clubs WHERE name='Drama Club';

-- ============================================
--  VERIFIED CLUBS
-- ============================================
INSERT INTO verified_clubs (fk_club_id) SELECT id FROM clubs WHERE name='Math Club';
INSERT INTO verified_clubs (fk_club_id) SELECT id FROM clubs WHERE name='Robotics Club';

-- ============================================
--  ADMINS (add more admins)
-- ============================================
INSERT INTO admins (fk_student_id) VALUES
  ('77777777-aaaa-aaaa-aaaa-aaaaaaaaaaaa'), -- mia
  ('88888888-bbbb-bbbb-bbbb-bbbbbbbbbbbb'), -- jackson
  ('99999999-cccc-cccc-cccc-cccccccccccc'); -- hannah
  
  
  -- ============================================
--  COLLABORATIVE EVENTS
-- ============================================

-- Event 1: Fall Community Hackathon
INSERT INTO events (
  fk_author_id, fk_thumbnail_id, title, location, rsvp_link, status,
  start_date, end_date, timezone, created_at, updated_at
)
SELECT '33333333-3333-3333-3333-333333333333', -- shohruz (CS club lead)
       NULL,
       'Fall Community Hackathon','Library Hall','hackathon2025','posted',
       '2025-11-10 09:00:00','2025-11-10 20:00:00','America/New_York',NOW(),NOW();

SET @event_hackathon = LAST_INSERT_ID();
INSERT INTO event_descriptions (fk_event_id, description)
VALUES(@event_hackathon,'A full-day hackathon co-hosted by CS, Math, and Robotics clubs.');

INSERT INTO events_to_clubs (fk_event_id,fk_club_id,club_is_event_owner)
SELECT @event_hackathon,id,TRUE FROM clubs WHERE name='Hunter CS Club';
INSERT INTO events_to_clubs (fk_event_id,fk_club_id,club_is_event_owner)
SELECT @event_hackathon,id,FALSE FROM clubs WHERE name='Math Club';
INSERT INTO events_to_clubs (fk_event_id,fk_club_id,club_is_event_owner)
SELECT @event_hackathon,id,FALSE FROM clubs WHERE name='Robotics Club';


-- Event 2: International Food Festival
INSERT INTO events (
  fk_author_id, fk_thumbnail_id, title, location, rsvp_link, status,
  start_date, end_date, timezone, created_at, updated_at
)
SELECT '55555555-5555-5555-5555-555555555555', -- maria (Cooking Club owner)
       NULL,
       'International Food Festival','Student Center','foodfest2025','posted',
       '2025-10-15 16:00:00','2025-10-15 20:00:00','America/New_York',NOW(),NOW();

SET @event_food = LAST_INSERT_ID();
INSERT INTO event_descriptions (fk_event_id, description)
VALUES(@event_food,'Celebrating food and culture, hosted by Cooking, Drama, and Art Clubs.');

INSERT INTO events_to_clubs (fk_event_id,fk_club_id,club_is_event_owner)
SELECT @event_food,id,TRUE FROM clubs WHERE name='Cooking Club';
INSERT INTO events_to_clubs (fk_event_id,fk_club_id,club_is_event_owner)
SELECT @event_food,id,FALSE FROM clubs WHERE name='Drama Club';
INSERT INTO events_to_clubs (fk_event_id,fk_club_id,club_is_event_owner)
SELECT @event_food,id,FALSE FROM clubs WHERE name='Art Club';


-- Event 3: Charity Music Night
INSERT INTO events (
  fk_author_id, fk_thumbnail_id, title, location, rsvp_link, status,
  start_date, end_date, timezone, created_at, updated_at
)
SELECT '44444444-dddd-dddd-dddd-dddddddddddd', -- daniel (Music Club owner)
       NULL,
       'Charity Music Night','Auditorium','musicnight2025','posted',
       '2025-12-05 18:00:00','2025-12-05 22:00:00','America/New_York',NOW(),NOW();

SET @event_music = LAST_INSERT_ID();
INSERT INTO event_descriptions (fk_event_id, description)
VALUES(@event_music,'A night of performances to raise money for charity, hosted by Music, Debate, and Photography Clubs.');

INSERT INTO events_to_clubs (fk_event_id,fk_club_id,club_is_event_owner)
SELECT @event_music,id,TRUE FROM clubs WHERE name='Music Club';
INSERT INTO events_to_clubs (fk_event_id,fk_club_id,club_is_event_owner)
SELECT @event_music,id,FALSE FROM clubs WHERE name='Debate Society';
INSERT INTO events_to_clubs (fk_event_id,fk_club_id,club_is_event_owner)
SELECT @event_music,id,FALSE FROM clubs WHERE name='Photography Club';