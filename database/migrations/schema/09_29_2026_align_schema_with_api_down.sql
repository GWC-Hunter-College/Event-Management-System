-- 09_29_2026_align_schema_with_api_down.sql
-- Reverses 09_29_2026_align_schema_with_api_up.sql, in reverse order.
-- Lossy where the old schema can't hold the new data: club topics, image owner,
-- uploader and alt text are dropped, and cancelled events become archived.
-- Restoring the single-use UNIQUE indexes fails if an image has been given a second use.
-- Written for the legacy runner, which splits on semicolons: no comment contains one.

-- C1. Club topics.
DROP TABLE IF EXISTS `club_tags`;

DROP TABLE IF EXISTS `topics`;

-- I3, I2. Image columns and constraints.
ALTER TABLE `images`
  DROP FOREIGN KEY `fk_images_uploaded_by`,
  DROP FOREIGN KEY `fk_images_club`,
  DROP INDEX `uq_images_object_key`;

ALTER TABLE `images`
  DROP INDEX `fk_images_uploaded_by`,
  DROP INDEX `fk_images_club`,
  DROP COLUMN `alt_text`,
  DROP COLUMN `fk_uploaded_by`,
  DROP COLUMN `fk_club_id`,
  MODIFY `purpose` VARCHAR(100),
  MODIFY `object_key` VARCHAR(255),
  MODIFY `mimetype` VARCHAR(255),
  MODIFY `created_at` TIMESTAMP NULL;

-- I1. Single-use UNIQUE indexes on image references.
ALTER TABLE `event_images` ADD UNIQUE INDEX `fk_image_id` (`fk_image_id`);

ALTER TABLE `event_images` DROP INDEX `idx_event_images_image`;

ALTER TABLE `events` ADD UNIQUE INDEX `fk_thumbnail_id` (`fk_thumbnail_id`);

ALTER TABLE `events` DROP INDEX `idx_events_thumbnail`;

ALTER TABLE `clubs` ADD UNIQUE INDEX `fk_logo_id` (`fk_logo_id`);

ALTER TABLE `clubs` DROP INDEX `idx_clubs_logo`;

-- T2, K1. Timestamp defaults, the date check and the list index.
ALTER TABLE `events`
  DROP INDEX `idx_events_status_start`,
  DROP CHECK `chk_events_end_after_start`,
  MODIFY `created_at` TIMESTAMP NULL,
  MODIFY `updated_at` TIMESTAMP NULL;

-- T1. Back to wall-clock time in each row's timezone.
UPDATE `events`
SET `start_date` = CONVERT_TZ(`start_date`, '+00:00', `timezone`),
    `end_date` = CONVERT_TZ(`end_date`, '+00:00', `timezone`);

ALTER TABLE `events`
  MODIFY `title` VARCHAR(255),
  MODIFY `start_date` DATETIME,
  MODIFY `end_date` DATETIME,
  MODIFY `timezone` VARCHAR(60);

-- S1. Status vocabulary. The old ENUM has no cancelled, and archived is the closest
-- value that keeps a cancelled event out of public reads.
ALTER TABLE `events`
  MODIFY `status` ENUM ('drafted', 'draft', 'posted', 'cancelled', 'archived');

UPDATE `events` SET `status` = 'drafted' WHERE `status` = 'draft';

UPDATE `events` SET `status` = 'archived' WHERE `status` = 'cancelled';

ALTER TABLE `events`
  MODIFY `status` ENUM ('drafted', 'posted', 'archived');

-- M4. Club name.
ALTER TABLE `clubs` MODIFY `name` VARCHAR(255);

-- M3. Owner-club flag and one-owner index.
ALTER TABLE `events_to_clubs`
  DROP INDEX `uq_events_to_clubs_one_owner`,
  MODIFY `club_is_event_owner` BOOL;

-- M2. Role flags.
ALTER TABLE `club_members`
  MODIFY `member_is_eboard` BOOL,
  MODIFY `member_is_owner` BOOL;

-- M1. Primary keys back to nullable UNIQUE columns. The UNIQUE index is added before
-- the primary key is dropped, because the foreign key needs one of them.
ALTER TABLE `event_descriptions` ADD UNIQUE INDEX `fk_event_id` (`fk_event_id`);

ALTER TABLE `event_descriptions`
  DROP PRIMARY KEY,
  MODIFY `fk_event_id` INT;

ALTER TABLE `admins` ADD UNIQUE INDEX `fk_student_id` (`fk_student_id`);

ALTER TABLE `admins`
  DROP PRIMARY KEY,
  MODIFY `fk_student_id` CHAR(36);

ALTER TABLE `verified_clubs` ADD UNIQUE INDEX `fk_club_id` (`fk_club_id`);

ALTER TABLE `verified_clubs`
  DROP PRIMARY KEY,
  MODIFY `fk_club_id` INT;
