-- 2026_09_29_baseline_seed.sql
-- Development data for 2026_09_29_baseline_up.sql. Load it into an empty database
-- right after the baseline, never into production.
--
-- Ids are explicit, so the query tests in database/README.md can name rows.
-- Event times are UTC and relative to the day the seed is loaded, so there are
-- always past, ongoing and upcoming events. 21:00 UTC is 5 p.m. in New York during
-- daylight saving time and 4 p.m. outside it.
-- Every statement stands alone: no session variables and no LAST_INSERT_ID().
--
-- What it covers
--   Clubs: verified and unverified, with and without a logo, topics, club_info and members.
--     Hunter CS Club has two owners, every other club one. Hiking Club has no club_info row.
--   Events: posted past and upcoming, one ongoing, one cancelled, two drafts
--     (one with no location), two multi-club events, one with no description,
--     one soft-deleted more than 30 days ago (purgeable, no longer restorable) and one
--     soft-deleted 3 days ago (restorable).
--   Images: logos and flyers with alt text, a gallery whose flyer is also in it,
--     one photo in two galleries, and soft-deleted images inside and outside retention.
--   Students: one with a NULL email, as request-time sync creates from an access token.
--   Admins: two, so removing one is allowed and removing the second is refused.

-- Students
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
  ('aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaaaa', 'fatima@example.com'),
  ('bbbbbbbb-bbbb-bbbb-bbbb-bbbbbbbbbbbb', 'aaron@example.com'),
  ('cccccccc-cccc-cccc-cccc-cccccccccccc', 'nina@example.com'),
  ('dddddddd-dddd-dddd-dddd-dddddddddddd', 'ethan@example.com'),
  ('eeeeeeee-eeee-eeee-eeee-eeeeeeeeeeee', 'olivia@example.com'),
  ('ffffffff-ffff-ffff-ffff-ffffffffffff', 'ryan@example.com'),
  ('12121212-1212-1212-1212-121212121212', NULL);

INSERT INTO student_info (fk_student_id, username, first_name, last_name) VALUES
  ('11111111-1111-1111-1111-111111111111', 'kelly', 'Kelly', 'Smith'),
  ('22222222-2222-2222-2222-222222222222', 'kyle', 'Kyle', 'Jones'),
  ('33333333-3333-3333-3333-333333333333', 'shohruz', 'Shohruz', 'Aliyev'),
  ('44444444-4444-4444-4444-444444444444', 'anth', 'Anthony', 'Brown'),
  ('55555555-5555-5555-5555-555555555555', 'maria', 'Maria', 'Lopez'),
  ('66666666-6666-6666-6666-666666666666', 'james', 'James', 'Smith'),
  ('77777777-7777-7777-7777-777777777777', 'soph', 'Sophia', 'Wang'),
  ('88888888-8888-8888-8888-888888888888', 'li', 'Li', 'Zhang'),
  ('bbbbbbbb-bbbb-bbbb-bbbb-bbbbbbbbbbbb', 'aaron', 'Aaron', 'Taylor'),
  ('cccccccc-cccc-cccc-cccc-cccccccccccc', 'nina', 'Nina', 'Martinez');

-- Clubs. Logos are attached after the image rows exist.
INSERT INTO clubs (id, name, fk_logo_id) VALUES
  (1, 'Cooking Club', NULL),
  (2, 'Girls Who Code @ Hunter', NULL),
  (3, 'Hunter CS Club', NULL),
  (4, 'Math Club', NULL),
  (5, 'Drama Club', NULL),
  (6, 'Robotics Club', NULL),
  (7, 'Art Club', NULL),
  (8, 'Music Club', NULL),
  (9, 'Hiking Club', NULL);

INSERT INTO club_info (fk_club_id, website_url, description) VALUES
  (1, 'https://www.example.com/cooking', 'Cook, eat and share recipes every other Friday.'),
  (2, 'https://www.example.com/gwc', 'Girls Who Code at Hunter College: workshops, hack nights and mentoring.'),
  (3, 'https://www.example.com/huntercs', 'Talks, office hours and the fall hackathon.'),
  (4, 'https://www.example.com/math', NULL),
  (5, 'https://www.example.com/drama', 'Two productions a year. Everyone can audition.'),
  (6, 'https://www.example.com/robotics', 'Build robots for the spring competition.'),
  (7, NULL, 'Open studio on Wednesdays.'),
  (8, 'https://www.example.com/music', NULL);

INSERT INTO club_tags (fk_club_id, slot, tag) VALUES
  (1, 1, 'community'),
  (1, 2, 'arts'),
  (2, 1, 'technology'),
  (2, 2, 'women in stem'),
  (2, 3, 'community'),
  (3, 1, 'technology'),
  (3, 2, 'career'),
  (4, 1, 'technology'),
  (5, 1, 'arts'),
  (6, 1, 'technology'),
  (7, 1, 'arts'),
  (8, 1, 'arts'),
  (8, 2, 'community'),
  (9, 1, 'sports');

INSERT INTO verified_clubs (fk_club_id) VALUES (2), (3), (4), (6);

INSERT INTO club_members (fk_student_id, fk_club_id, role) VALUES
  ('55555555-5555-5555-5555-555555555555', 1, 'owner'),
  ('66666666-6666-6666-6666-666666666666', 1, 'eboard'),
  ('44444444-4444-4444-4444-444444444444', 1, 'member'),
  ('11111111-1111-1111-1111-111111111111', 1, 'member'),
  ('11111111-1111-1111-1111-111111111111', 2, 'owner'),
  ('44444444-4444-4444-4444-444444444444', 2, 'eboard'),
  ('22222222-2222-2222-2222-222222222222', 2, 'eboard'),
  ('77777777-7777-7777-7777-777777777777', 2, 'member'),
  ('88888888-8888-8888-8888-888888888888', 2, 'member'),
  ('12121212-1212-1212-1212-121212121212', 2, 'member'),
  ('33333333-3333-3333-3333-333333333333', 3, 'owner'),
  ('99999999-9999-9999-9999-999999999999', 3, 'owner'),
  ('88888888-8888-8888-8888-888888888888', 3, 'member'),
  ('bbbbbbbb-bbbb-bbbb-bbbb-bbbbbbbbbbbb', 4, 'owner'),
  ('aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaaaa', 4, 'member'),
  ('cccccccc-cccc-cccc-cccc-cccccccccccc', 5, 'owner'),
  ('dddddddd-dddd-dddd-dddd-dddddddddddd', 6, 'owner'),
  ('eeeeeeee-eeee-eeee-eeee-eeeeeeeeeeee', 7, 'owner'),
  ('ffffffff-ffff-ffff-ffff-ffffffffffff', 8, 'owner'),
  ('aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaaaa', 9, 'owner');

INSERT INTO admins (fk_student_id) VALUES
  ('22222222-2222-2222-2222-222222222222'),
  ('aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaaaa');

-- Images. Ids are fixed UUIDs, and object keys follow the signers' layout.
INSERT INTO images (id, fk_club_id, fk_uploaded_by, purpose, object_key, filename, mimetype, alt_text, created_at, deleted_at) VALUES
  ('a0000000-0000-0000-0000-000000000001', 2, '11111111-1111-1111-1111-111111111111', 'club-thumbnail', 'clubs/2/thumbnails/a0000000-0000-0000-0000-000000000001.png', 'gwc-logo.png', 'image/png', 'Girls Who Code at Hunter logo', CURRENT_TIMESTAMP - INTERVAL 90 DAY, NULL),
  ('a0000000-0000-0000-0000-000000000002', 3, '33333333-3333-3333-3333-333333333333', 'club-thumbnail', 'clubs/3/thumbnails/a0000000-0000-0000-0000-000000000002.png', 'cs-logo.png', 'image/png', 'Hunter CS Club logo', CURRENT_TIMESTAMP - INTERVAL 90 DAY, NULL),
  ('a0000000-0000-0000-0000-000000000003', 2, '11111111-1111-1111-1111-111111111111', 'event-thumbnail', 'events/1/thumbnails/a0000000-0000-0000-0000-000000000003.jpg', 'kickoff.jpg', 'image/jpeg', 'Poster for the fall kickoff meeting', CURRENT_TIMESTAMP - INTERVAL 30 DAY, NULL),
  ('a0000000-0000-0000-0000-000000000004', 2, '11111111-1111-1111-1111-111111111111', 'event-image', 'events/1/images/a0000000-0000-0000-0000-000000000004.jpg', 'room.jpg', 'image/jpeg', 'Students at laptops in Room 101', CURRENT_TIMESTAMP - INTERVAL 19 DAY, NULL),
  ('a0000000-0000-0000-0000-000000000005', 2, '44444444-4444-4444-4444-444444444444', 'event-image', 'events/1/images/a0000000-0000-0000-0000-000000000005.jpg', 'group.jpg', 'image/jpeg', NULL, CURRENT_TIMESTAMP - INTERVAL 19 DAY, NULL),
  ('a0000000-0000-0000-0000-000000000006', 3, '33333333-3333-3333-3333-333333333333', 'event-thumbnail', 'events/3/thumbnails/a0000000-0000-0000-0000-000000000006.png', 'hackathon.png', 'image/png', 'Fall Community Hackathon poster', CURRENT_TIMESTAMP - INTERVAL 10 DAY, NULL),
  ('a0000000-0000-0000-0000-000000000007', 5, 'cccccccc-cccc-cccc-cccc-cccccccccccc', 'event-thumbnail', 'events/8/thumbnails/a0000000-0000-0000-0000-000000000007.png', 'auditions.png', 'image/png', 'Audition flyer', CURRENT_TIMESTAMP - INTERVAL 60 DAY, NULL),
  ('a0000000-0000-0000-0000-000000000008', 5, 'cccccccc-cccc-cccc-cccc-cccccccccccc', 'event-image', 'events/8/images/a0000000-0000-0000-0000-000000000008.jpg', 'stage.jpg', 'image/jpeg', NULL, CURRENT_TIMESTAMP - INTERVAL 50 DAY, NULL),
  ('a0000000-0000-0000-0000-000000000009', 2, '11111111-1111-1111-1111-111111111111', 'club-thumbnail', 'clubs/2/thumbnails/a0000000-0000-0000-0000-000000000009.png', 'old-gwc-logo.png', 'image/png', NULL, CURRENT_TIMESTAMP - INTERVAL 200 DAY, CURRENT_TIMESTAMP - INTERVAL 40 DAY),
  ('a0000000-0000-0000-0000-00000000000a', 3, '33333333-3333-3333-3333-333333333333', 'event-thumbnail', 'events/3/thumbnails/a0000000-0000-0000-0000-00000000000a.png', 'hackathon-v1.png', 'image/png', NULL, CURRENT_TIMESTAMP - INTERVAL 12 DAY, CURRENT_TIMESTAMP - INTERVAL 2 DAY);

UPDATE clubs SET fk_logo_id = 'a0000000-0000-0000-0000-000000000001' WHERE id = 2;

UPDATE clubs SET fk_logo_id = 'a0000000-0000-0000-0000-000000000002' WHERE id = 3;

-- Events
INSERT INTO events (id, fk_author_id, fk_thumbnail_id, title, location, rsvp_link, status, start_date, end_date, timezone, created_at, updated_at, deleted_at) VALUES
  (1, '11111111-1111-1111-1111-111111111111', 'a0000000-0000-0000-0000-000000000003', 'Fall Kickoff Meeting', 'Room 101', 'https://events.example.edu/kickoff', 'posted',
     TIMESTAMP(UTC_DATE() - INTERVAL 20 DAY, '21:00:00'), TIMESTAMP(UTC_DATE() - INTERVAL 20 DAY, '23:00:00'), 'America/New_York',
     CURRENT_TIMESTAMP - INTERVAL 30 DAY, CURRENT_TIMESTAMP - INTERVAL 25 DAY, NULL),
  (2, '44444444-4444-4444-4444-444444444444', NULL, 'Resume Workshop', 'Room 715', NULL, 'posted',
     TIMESTAMP(UTC_DATE() + INTERVAL 5 DAY, '22:00:00'), TIMESTAMP(UTC_DATE() + INTERVAL 5 DAY, '23:30:00'), 'America/New_York',
     CURRENT_TIMESTAMP - INTERVAL 7 DAY, CURRENT_TIMESTAMP - INTERVAL 7 DAY, NULL),
  (3, '33333333-3333-3333-3333-333333333333', 'a0000000-0000-0000-0000-000000000006', 'Fall Community Hackathon', 'Library Hall', 'https://events.example.edu/hackathon', 'posted',
     TIMESTAMP(UTC_DATE() + INTERVAL 12 DAY, '13:00:00'), TIMESTAMP(UTC_DATE() + INTERVAL 13 DAY, '01:00:00'), 'America/New_York',
     CURRENT_TIMESTAMP - INTERVAL 10 DAY, CURRENT_TIMESTAMP - INTERVAL 2 DAY, NULL),
  (4, '55555555-5555-5555-5555-555555555555', NULL, 'International Food Festival', 'Student Center', 'https://events.example.edu/foodfest', 'cancelled',
     TIMESTAMP(UTC_DATE() + INTERVAL 9 DAY, '20:00:00'), TIMESTAMP(UTC_DATE() + INTERVAL 10 DAY, '00:00:00'), 'America/New_York',
     CURRENT_TIMESTAMP - INTERVAL 14 DAY, CURRENT_TIMESTAMP - INTERVAL 1 DAY, NULL),
  (5, '11111111-1111-1111-1111-111111111111', NULL, 'Game Night', NULL, NULL, 'draft',
     TIMESTAMP(UTC_DATE() + INTERVAL 20 DAY, '22:00:00'), TIMESTAMP(UTC_DATE() + INTERVAL 21 DAY, '01:00:00'), 'America/New_York',
     CURRENT_TIMESTAMP - INTERVAL 3 DAY, CURRENT_TIMESTAMP - INTERVAL 1 HOUR, NULL),
  (6, '22222222-2222-2222-2222-222222222222', NULL, 'Speaker Series: Women in Tech', 'Room 101', NULL, 'draft',
     TIMESTAMP(UTC_DATE() + INTERVAL 30 DAY, '22:00:00'), TIMESTAMP(UTC_DATE() + INTERVAL 30 DAY, '23:30:00'), 'America/New_York',
     CURRENT_TIMESTAMP - INTERVAL 5 DAY, CURRENT_TIMESTAMP - INTERVAL 4 DAY, NULL),
  (7, 'bbbbbbbb-bbbb-bbbb-bbbb-bbbbbbbbbbbb', NULL, 'Math Club Orientation', 'Room 201', NULL, 'posted',
     TIMESTAMP(UTC_DATE() - INTERVAL 40 DAY, '22:00:00'), TIMESTAMP(UTC_DATE() - INTERVAL 39 DAY, '00:00:00'), 'America/New_York',
     CURRENT_TIMESTAMP - INTERVAL 50 DAY, CURRENT_TIMESTAMP - INTERVAL 50 DAY, NULL),
  (8, 'cccccccc-cccc-cccc-cccc-cccccccccccc', 'a0000000-0000-0000-0000-000000000007', 'Spring Play Auditions', 'Auditorium', NULL, 'posted',
     TIMESTAMP(UTC_DATE() - INTERVAL 55 DAY, '21:00:00'), TIMESTAMP(UTC_DATE() - INTERVAL 54 DAY, '01:00:00'), 'America/New_York',
     CURRENT_TIMESTAMP - INTERVAL 60 DAY, CURRENT_TIMESTAMP - INTERVAL 45 DAY, CURRENT_TIMESTAMP - INTERVAL 45 DAY),
  (9, 'ffffffff-ffff-ffff-ffff-ffffffffffff', NULL, 'Charity Music Night', 'Auditorium', NULL, 'posted',
     TIMESTAMP(UTC_DATE() + INTERVAL 15 DAY, '22:00:00'), TIMESTAMP(UTC_DATE() + INTERVAL 16 DAY, '02:00:00'), 'America/New_York',
     CURRENT_TIMESTAMP - INTERVAL 20 DAY, CURRENT_TIMESTAMP - INTERVAL 3 DAY, CURRENT_TIMESTAMP - INTERVAL 3 DAY),
  (10, 'dddddddd-dddd-dddd-dddd-dddddddddddd', NULL, 'Robotics Demo Day', 'Gym', NULL, 'posted',
     TIMESTAMP(UTC_DATE() - INTERVAL 2 DAY, '19:00:00'), TIMESTAMP(UTC_DATE() - INTERVAL 2 DAY, '22:00:00'), 'America/New_York',
     CURRENT_TIMESTAMP - INTERVAL 9 DAY, CURRENT_TIMESTAMP - INTERVAL 9 DAY, NULL),
  (11, '99999999-9999-9999-9999-999999999999', NULL, 'CS Office Hours', 'Room 1001', NULL, 'posted',
     UTC_TIMESTAMP() - INTERVAL 1 HOUR, UTC_TIMESTAMP() + INTERVAL 2 HOUR, 'America/New_York',
     CURRENT_TIMESTAMP - INTERVAL 1 DAY, CURRENT_TIMESTAMP - INTERVAL 1 DAY, NULL);

-- Event 10 has no description.
INSERT INTO event_descriptions (fk_event_id, description) VALUES
  (1, 'Meet the board and hear the plan for the semester.'),
  (2, 'Bring a resume and get feedback from CS Club alumni.'),
  (3, 'A full-day hackathon co-hosted by the CS, Math and Robotics clubs.'),
  (4, 'Food and culture from around the world, hosted by the Cooking, Drama and Art clubs.'),
  (5, NULL),
  (6, 'A panel of alumni working in tech.'),
  (7, 'Welcome to the Math Club!'),
  (8, 'Auditions for the spring play.'),
  (9, 'A night of performances to raise money for charity.'),
  (11, 'Drop in with homework questions.');

INSERT INTO events_to_clubs (fk_event_id, fk_club_id, club_is_event_owner) VALUES
  (1, 2, TRUE),
  (2, 2, TRUE),
  (2, 3, FALSE),
  (3, 3, TRUE),
  (3, 4, FALSE),
  (3, 6, FALSE),
  (4, 1, TRUE),
  (4, 5, FALSE),
  (4, 7, FALSE),
  (5, 2, TRUE),
  (6, 2, TRUE),
  (7, 4, TRUE),
  (8, 5, TRUE),
  (9, 8, TRUE),
  (10, 6, TRUE),
  (11, 3, TRUE);

-- Gallery. Event 1's flyer is also in its gallery, so its image count is 2, not 3.
-- The group photo (image 5) is also in event 10's gallery.
-- Image 8 is used only by event 8, which is soft-deleted past retention.
INSERT INTO event_images (fk_event_id, fk_image_id, created_at) VALUES
  (1, 'a0000000-0000-0000-0000-000000000003', CURRENT_TIMESTAMP - INTERVAL 19 DAY),
  (1, 'a0000000-0000-0000-0000-000000000004', CURRENT_TIMESTAMP - INTERVAL 19 DAY + INTERVAL 1 MINUTE),
  (1, 'a0000000-0000-0000-0000-000000000005', CURRENT_TIMESTAMP - INTERVAL 19 DAY + INTERVAL 2 MINUTE),
  (10, 'a0000000-0000-0000-0000-000000000005', CURRENT_TIMESTAMP - INTERVAL 1 DAY),
  (8, 'a0000000-0000-0000-0000-000000000008', CURRENT_TIMESTAMP - INTERVAL 50 DAY);
