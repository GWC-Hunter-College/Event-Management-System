CREATE TABLE IF NOT EXISTS `students` (
  `id` CHAR(36) PRIMARY KEY,
  `email` VARCHAR(255)
);

CREATE TABLE IF NOT EXISTS `student_info` (
  `fk_student_id` CHAR(36) PRIMARY KEY,
  `username` VARCHAR(30) UNIQUE,
  `first_name` VARCHAR(40),
  `last_name` VARCHAR(40)
);

CREATE TABLE IF NOT EXISTS `clubs` (
  `id` INT PRIMARY KEY AUTO_INCREMENT,
  `fk_logo_id` CHAR(36) UNIQUE,
  `name` VARCHAR(255) UNIQUE
);

CREATE TABLE  IF NOT EXISTS `club_info` (
  `fk_club_id` INT PRIMARY KEY,
  `website_url` VARCHAR(255),
  `description` TEXT DEFAULT NULL
);

CREATE TABLE IF NOT EXISTS `club_members` (
  `fk_student_id` CHAR(36),
  `fk_club_id` INT,
  `member_is_eboard` BOOL,
  `member_is_owner` BOOL,
  PRIMARY KEY (`fk_student_id`, `fk_club_id`)
);

CREATE TABLE IF NOT EXISTS `verified_clubs` (
  `fk_club_id` INT UNIQUE
);

CREATE TABLE IF NOT EXISTS `admins` (
  `fk_student_id` CHAR(36) UNIQUE
);

CREATE TABLE IF NOT EXISTS `events` (
  `id` INT PRIMARY KEY AUTO_INCREMENT,
  `fk_author_id` CHAR(36),
  `fk_thumbnail_id` CHAR(36) UNIQUE,
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

CREATE TABLE IF NOT EXISTS `event_descriptions` (
  `fk_event_id` INT UNIQUE,
  `description` TEXT
);

CREATE TABLE IF NOT EXISTS `event_tags` (
  `fk_event_id` INT,
  `tag` VARCHAR(255),
  PRIMARY KEY (`fk_event_id`, `tag`)
);

CREATE TABLE IF NOT EXISTS `images` (
  `id` CHAR(36) PRIMARY KEY,
  `purpose` VARCHAR(100),
  `object_key` VARCHAR(255),
  `filename` VARCHAR(255),
  `mimetype` VARCHAR(255),
  `created_at` TIMESTAMP
);

CREATE TABLE IF NOT EXISTS `events_to_clubs` (
  `fk_event_id` INT,
  `fk_club_id` INT,
  `club_is_event_owner` BOOL,
  PRIMARY KEY (`fk_event_id`, `fk_club_id`)
);

CREATE TABLE IF NOT EXISTS `event_images` (
  `fk_event_id` INT,
  `fk_image_id` CHAR(36) UNIQUE,
  PRIMARY KEY (`fk_event_id`, `fk_image_id`)
);

ALTER TABLE `verified_clubs` COMMENT = 'Table for making sure clubs are allowed to be shown to account for bad actors creating random new clubs';

ALTER TABLE `student_info` ADD FOREIGN KEY (`fk_student_id`) REFERENCES `students` (`id`);

ALTER TABLE `clubs` ADD FOREIGN KEY (`fk_logo_id`) REFERENCES `images` (`id`);

ALTER TABLE `club_info` ADD FOREIGN KEY (`fk_club_id`) REFERENCES `clubs` (`id`);

ALTER TABLE `club_members` ADD FOREIGN KEY (`fk_student_id`) REFERENCES `students` (`id`);

ALTER TABLE `club_members` ADD FOREIGN KEY (`fk_club_id`) REFERENCES `clubs` (`id`);

ALTER TABLE `verified_clubs` ADD FOREIGN KEY (`fk_club_id`) REFERENCES `clubs` (`id`);

ALTER TABLE `admins` ADD FOREIGN KEY (`fk_student_id`) REFERENCES `students` (`id`);

ALTER TABLE `events` ADD FOREIGN KEY (`fk_author_id`) REFERENCES `students` (`id`);

ALTER TABLE `events` ADD FOREIGN KEY (`fk_thumbnail_id`) REFERENCES `images` (`id`);

ALTER TABLE `event_descriptions` ADD FOREIGN KEY (`fk_event_id`) REFERENCES `events` (`id`);

ALTER TABLE `event_tags` ADD FOREIGN KEY (`fk_event_id`) REFERENCES `events` (`id`);

ALTER TABLE `events_to_clubs` ADD FOREIGN KEY (`fk_event_id`) REFERENCES `events` (`id`);

ALTER TABLE `events_to_clubs` ADD FOREIGN KEY (`fk_club_id`) REFERENCES `clubs` (`id`);

ALTER TABLE `event_images` ADD FOREIGN KEY (`fk_event_id`) REFERENCES `events` (`id`);

ALTER TABLE `event_images` ADD FOREIGN KEY (`fk_image_id`) REFERENCES `images` (`id`);
