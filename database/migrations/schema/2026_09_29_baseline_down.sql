-- 2026_09_29_baseline_down.sql
-- Drops everything 2026_09_29_baseline_up.sql creates, children before parents.
-- Destroys all data. Only for rebuilding a local or test database.
-- The images-to-clubs foreign key is dropped first, because it closes a cycle with clubs.

DROP VIEW IF EXISTS `announcement_details`;

DROP VIEW IF EXISTS `club_details`;

DROP VIEW IF EXISTS `event_details`;

DROP TABLE IF EXISTS `announcements_to_clubs`;

DROP TABLE IF EXISTS `announcements`;

DROP TABLE IF EXISTS `event_images`;

DROP TABLE IF EXISTS `events_to_clubs`;

DROP TABLE IF EXISTS `event_tags`;

DROP TABLE IF EXISTS `event_descriptions`;

DROP TABLE IF EXISTS `events`;

DROP TABLE IF EXISTS `club_tags`;

DROP TABLE IF EXISTS `topics`;

DROP TABLE IF EXISTS `verified_clubs`;

DROP TABLE IF EXISTS `club_members`;

DROP TABLE IF EXISTS `club_info`;

ALTER TABLE `images` DROP FOREIGN KEY `fk_images_club`;

DROP TABLE IF EXISTS `clubs`;

DROP TABLE IF EXISTS `images`;

DROP TABLE IF EXISTS `admins`;

DROP TABLE IF EXISTS `student_info`;

DROP TABLE IF EXISTS `students`;
