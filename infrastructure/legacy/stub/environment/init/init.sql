DROP DATABASE IF EXISTS clubsys;
CREATE DATABASE clubsys;
USE clubsys;

CREATE TABLE `students` (
  `id` CHAR(36) PRIMARY KEY,
  `email` VARCHAR(255)
);

CREATE TABLE `clubs` (
  `id` INT PRIMARY KEY AUTO_INCREMENT,
  `fk_thumbnail_id` INT,
  `name` VARCHAR(255) UNIQUE
);

CREATE TABLE `club_members` (
  `fk_student_id` CHAR(36),
  `fk_club_id` INT,
  `member_is_eboard` BOOL,
  `member_is_owner` BOOL,
  PRIMARY KEY (`fk_student_id`, `fk_club_id`)
);

CREATE TABLE `verified_clubs` (
  `fk_club_id` INT UNIQUE
);

CREATE TABLE `admins` (
  `fk_student_id` CHAR(36) UNIQUE
);

CREATE TABLE `events` (
  `id` INT PRIMARY KEY AUTO_INCREMENT,
  `fk_author_id` CHAR(36),
  `fk_thumbnail_id` INT,
  `title` VARCHAR(255),
  `location` VARCHAR(255),
  `rsvp_link` VARCHAR(255),
  `status` ENUM ('drafted', 'posted', 'archived'),
  `start_date` DATETIME,
  `end_date` DATETIME,
  `timezone` VARCHAR(60),
  `created_at` TIMESTAMP,
  `updated_at` TIMESTAMP,
  `deleted_at` TIMESTAMP
);

CREATE TABLE `event_descriptions` (
  `fk_event_id` INT UNIQUE,
  `description` TEXT
);

CREATE TABLE `event_tags` (
  `fk_event_id` INT,
  `tag` VARCHAR(255),
  PRIMARY KEY (`fk_event_id`, `tag`)
);

CREATE TABLE `images` (
  `id` INT PRIMARY KEY AUTO_INCREMENT,
  `fk_event_id` INT,
  `fk_club_id` INT,
  `purpose` VARCHAR(100),
  `object_key` VARCHAR(255),
  `created_at` TIMESTAMP
);

CREATE TABLE `events_to_clubs` (
  `fk_event_id` INT,
  `fk_club_id` INT,
  `club_is_event_owner` BOOL,
  PRIMARY KEY (`fk_event_id`, `fk_club_id`)
);

ALTER TABLE `verified_clubs` COMMENT = 'Table for making sure clubs are allowed to be shown to account for bad actors creating random new clubs';

ALTER TABLE `clubs` ADD FOREIGN KEY (`fk_thumbnail_id`) REFERENCES `images` (`id`);

ALTER TABLE `club_members` ADD FOREIGN KEY (`fk_student_id`) REFERENCES `students` (`id`);

ALTER TABLE `club_members` ADD FOREIGN KEY (`fk_club_id`) REFERENCES `clubs` (`id`);

ALTER TABLE `verified_clubs` ADD FOREIGN KEY (`fk_club_id`) REFERENCES `clubs` (`id`);

ALTER TABLE `admins` ADD FOREIGN KEY (`fk_student_id`) REFERENCES `students` (`id`);

ALTER TABLE `events` ADD FOREIGN KEY (`fk_author_id`) REFERENCES `students` (`id`);

ALTER TABLE `events` ADD FOREIGN KEY (`fk_thumbnail_id`) REFERENCES `images` (`id`);

ALTER TABLE `event_descriptions` ADD FOREIGN KEY (`fk_event_id`) REFERENCES `events` (`id`);

ALTER TABLE `event_tags` ADD FOREIGN KEY (`fk_event_id`) REFERENCES `events` (`id`);

ALTER TABLE `images` ADD FOREIGN KEY (`fk_event_id`) REFERENCES `events` (`id`);

ALTER TABLE `images` ADD FOREIGN KEY (`fk_club_id`) REFERENCES `clubs` (`id`);

ALTER TABLE `events_to_clubs` ADD FOREIGN KEY (`fk_event_id`) REFERENCES `events` (`id`);

ALTER TABLE `events_to_clubs` ADD FOREIGN KEY (`fk_club_id`) REFERENCES `clubs` (`id`);
