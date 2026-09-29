-- 09_29_2026_align_schema_with_api_up.sql
-- The "change now" findings in database/docs/schema-review.md, applied on top of
-- 11_04_2025_create_core_tables_up.sql. Each section names its finding (S, T, I, M, K, C).
-- Run the pre-flight checks in the review first. A check that finds rows makes a
-- statement below fail on purpose instead of silently changing data.
-- Written for the legacy runner, which splits on semicolons: no comment contains one,
-- and no chunk is only a comment.
-- Not idempotent: T1 converts times once. Record it as applied and never re-run it.

-- M1. Rows with a NULL key mark nothing and no query can reach them.
DELETE FROM `verified_clubs` WHERE `fk_club_id` IS NULL;

DELETE FROM `admins` WHERE `fk_student_id` IS NULL;

DELETE FROM `event_descriptions` WHERE `fk_event_id` IS NULL;

-- M1. verified_clubs, admins and event_descriptions get primary keys. The primary key
-- is added before the UNIQUE index is dropped, because the foreign key needs one of them.
ALTER TABLE `verified_clubs`
  MODIFY `fk_club_id` INT NOT NULL,
  ADD PRIMARY KEY (`fk_club_id`);

ALTER TABLE `verified_clubs` DROP INDEX `fk_club_id`;

ALTER TABLE `admins`
  MODIFY `fk_student_id` CHAR(36) NOT NULL,
  ADD PRIMARY KEY (`fk_student_id`);

ALTER TABLE `admins` DROP INDEX `fk_student_id`;

ALTER TABLE `event_descriptions`
  MODIFY `fk_event_id` INT NOT NULL,
  ADD PRIMARY KEY (`fk_event_id`);

ALTER TABLE `event_descriptions` DROP INDEX `fk_event_id`;

-- M2. Role flags are never NULL. A NULL owner flag made the leave-club guard
-- (member_is_owner = FALSE) skip the row.
UPDATE `club_members`
SET `member_is_eboard` = COALESCE(`member_is_eboard`, FALSE),
    `member_is_owner` = COALESCE(`member_is_owner`, FALSE)
WHERE `member_is_eboard` IS NULL OR `member_is_owner` IS NULL;

ALTER TABLE `club_members`
  MODIFY `member_is_eboard` BOOL NOT NULL DEFAULT FALSE,
  MODIFY `member_is_owner` BOOL NOT NULL DEFAULT FALSE;

-- M3. At most one owner club per event. The functional index covers only owner rows.
-- It fails if an event already has two owner links, so fix those rows first.
UPDATE `events_to_clubs` SET `club_is_event_owner` = FALSE WHERE `club_is_event_owner` IS NULL;

ALTER TABLE `events_to_clubs`
  MODIFY `club_is_event_owner` BOOL NOT NULL DEFAULT FALSE,
  ADD UNIQUE INDEX `uq_events_to_clubs_one_owner` ((IF(`club_is_event_owner`, `fk_event_id`, NULL)));

-- M4. Every club has a name. Fails if a club has a NULL name.
ALTER TABLE `clubs` MODIFY `name` VARCHAR(255) NOT NULL;

-- S1. One status vocabulary for the API and the database, plus cancelled.
-- The transitional ENUM holds both spellings while rows are renamed.
-- NULL was never visible to any read, so it becomes draft.
ALTER TABLE `events`
  MODIFY `status` ENUM ('drafted', 'draft', 'posted', 'cancelled', 'archived');

UPDATE `events` SET `status` = 'draft' WHERE `status` = 'drafted' OR `status` IS NULL;

ALTER TABLE `events`
  MODIFY `status` ENUM ('draft', 'posted', 'cancelled', 'archived') NOT NULL DEFAULT 'draft';

-- T1. start_date and end_date hold UTC from now on. Existing rows hold wall-clock
-- time in the zone named by their timezone column, so they are converted once.
-- The columns become NOT NULL first. CONVERT_TZ returns NULL when the server has no
-- time zone tables or the name is unknown, and a NULL then fails the whole UPDATE
-- instead of converting only some rows.
UPDATE `events` SET `timezone` = 'America/New_York' WHERE `timezone` IS NULL OR `timezone` = '';

ALTER TABLE `events`
  MODIFY `title` VARCHAR(255) NOT NULL,
  MODIFY `start_date` DATETIME NOT NULL,
  MODIFY `end_date` DATETIME NOT NULL,
  MODIFY `timezone` VARCHAR(60) NOT NULL DEFAULT 'America/New_York';

UPDATE `events`
SET `start_date` = CONVERT_TZ(`start_date`, `timezone`, '+00:00'),
    `end_date` = CONVERT_TZ(`end_date`, `timezone`, '+00:00');

-- T2, K1. Timestamps default to the current time and updated_at follows row changes.
-- The event list index serves the status filter and start-time ordering.
UPDATE `events`
SET `created_at` = COALESCE(`created_at`, `updated_at`, CURRENT_TIMESTAMP),
    `updated_at` = COALESCE(`updated_at`, `created_at`, CURRENT_TIMESTAMP)
WHERE `created_at` IS NULL OR `updated_at` IS NULL;

ALTER TABLE `events`
  MODIFY `created_at` TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
  MODIFY `updated_at` TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  ADD CONSTRAINT `chk_events_end_after_start` CHECK (`end_date` >= `start_date`),
  ADD INDEX `idx_events_status_start` (`status`, `start_date`);

-- I1. An image row is a stored file. Uses reference it, and one file may have several
-- uses, so the single-use UNIQUE indexes become plain indexes. The plain index is added
-- first because each foreign key needs an index on its column.
ALTER TABLE `clubs` ADD INDEX `idx_clubs_logo` (`fk_logo_id`);

ALTER TABLE `clubs` DROP INDEX `fk_logo_id`;

ALTER TABLE `events` ADD INDEX `idx_events_thumbnail` (`fk_thumbnail_id`);

ALTER TABLE `events` DROP INDEX `fk_thumbnail_id`;

ALTER TABLE `event_images` ADD INDEX `idx_event_images_image` (`fk_image_id`);

ALTER TABLE `event_images` DROP INDEX `fk_image_id`;

-- I2. Owning club, uploader and alt text on the file.
ALTER TABLE `images`
  ADD COLUMN `fk_club_id` INT NULL AFTER `id`,
  ADD COLUMN `fk_uploaded_by` CHAR(36) NULL AFTER `fk_club_id`,
  ADD COLUMN `alt_text` VARCHAR(1000) NULL AFTER `mimetype`;

-- I2. Backfill the owning club from each existing use. M3 already guarantees one owner
-- club per event, so each join matches at most one club.
UPDATE `images` i
JOIN `clubs` c ON c.`fk_logo_id` = i.`id`
SET i.`fk_club_id` = c.`id`
WHERE i.`fk_club_id` IS NULL;

UPDATE `images` i
JOIN `events` e ON e.`fk_thumbnail_id` = i.`id`
JOIN `events_to_clubs` ec ON ec.`fk_event_id` = e.`id` AND ec.`club_is_event_owner` = TRUE
SET i.`fk_club_id` = ec.`fk_club_id`
WHERE i.`fk_club_id` IS NULL;

UPDATE `images` i
JOIN `event_images` ei ON ei.`fk_image_id` = i.`id`
JOIN `events_to_clubs` ec ON ec.`fk_event_id` = ei.`fk_event_id` AND ec.`club_is_event_owner` = TRUE
SET i.`fk_club_id` = ec.`fk_club_id`
WHERE i.`fk_club_id` IS NULL;

-- I3. The columns every read and confirm needs are required, and one S3 object has one row.
UPDATE `images` SET `created_at` = CURRENT_TIMESTAMP WHERE `created_at` IS NULL;

ALTER TABLE `images`
  MODIFY `purpose` VARCHAR(100) NOT NULL,
  MODIFY `object_key` VARCHAR(255) NOT NULL,
  MODIFY `mimetype` VARCHAR(255) NOT NULL,
  MODIFY `created_at` TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
  ADD CONSTRAINT `uq_images_object_key` UNIQUE (`object_key`),
  ADD CONSTRAINT `fk_images_club` FOREIGN KEY (`fk_club_id`) REFERENCES `clubs` (`id`),
  ADD CONSTRAINT `fk_images_uploaded_by` FOREIGN KEY (`fk_uploaded_by`) REFERENCES `students` (`id`) ON DELETE SET NULL;

-- C1. Club topics: a fixed list held as data, and at most three per club.
-- ON UPDATE CASCADE lets a topic key be renamed with one UPDATE. The binary collation
-- makes keys case-sensitive, so the foreign key only accepts the lowercase keys.
CREATE TABLE IF NOT EXISTS `topics` (
  `tag` VARCHAR(40) CHARACTER SET utf8mb4 COLLATE utf8mb4_bin PRIMARY KEY
);

INSERT IGNORE INTO `topics` (`tag`) VALUES
  ('technology'),
  ('arts'),
  ('community'),
  ('women in stem'),
  ('career'),
  ('sports');

CREATE TABLE IF NOT EXISTS `club_tags` (
  `fk_club_id` INT NOT NULL,
  `slot` TINYINT NOT NULL,
  `tag` VARCHAR(40) CHARACTER SET utf8mb4 COLLATE utf8mb4_bin NOT NULL,
  PRIMARY KEY (`fk_club_id`, `slot`),
  UNIQUE KEY `uq_club_tags_club_tag` (`fk_club_id`, `tag`),
  KEY `idx_club_tags_tag` (`tag`),
  CONSTRAINT `chk_club_tags_slot` CHECK (`slot` BETWEEN 1 AND 3),
  CONSTRAINT `fk_club_tags_club` FOREIGN KEY (`fk_club_id`) REFERENCES `clubs` (`id`) ON DELETE CASCADE,
  CONSTRAINT `fk_club_tags_topic` FOREIGN KEY (`tag`) REFERENCES `topics` (`tag`) ON UPDATE CASCADE
);
