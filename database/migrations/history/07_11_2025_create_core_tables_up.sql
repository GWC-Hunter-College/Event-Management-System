CREATE TABLE IF NOT EXISTS `STUDENTS` (
  `id` int PRIMARY KEY AUTO_INCREMENT,
  `first_name` VARCHAR(50),
  `last_name` VARCHAR(50),
  `email` VARCHAR(255),
  `password` VARCHAR(255) COMMENT 'hashed, to be added eventually'
);

CREATE TABLE IF NOT EXISTS `STUDENT_INFO` (
  `student_id` int UNIQUE,
  `major` VARCHAR(255),
  `emplid` VARCHAR(255),
  `grad_year` int,
  `interests_id` int COMMENT 'FK',
  `dietary_restrictions` VARCHAR(255),
  `comments` VARCHAR(255),
  CONSTRAINT `FK_StudentInfo_Students` FOREIGN KEY (`student_id`) REFERENCES `STUDENTS` (`id`)
);

CREATE TABLE IF NOT EXISTS `INTERESTS` (
  `id` int PRIMARY KEY AUTO_INCREMENT,
  `label` VARCHAR(255) COMMENT 'career, community, and volunteer when starting'
);

CREATE TABLE IF NOT EXISTS `STUDENTS_TO_INTERESTS` (
  `student_id` int COMMENT 'FK',
  `interest_id` int COMMENT 'FK',
  PRIMARY KEY (`student_id`, `interest_id`), -- Assuming this is a many-to-many relationship
  CONSTRAINT `FK_StudentsToInterests_Students` FOREIGN KEY (`student_id`) REFERENCES `STUDENTS` (`id`),
  CONSTRAINT `FK_StudentsToInterests_Interests` FOREIGN KEY (`interest_id`) REFERENCES `INTERESTS` (`id`)
);

CREATE TABLE IF NOT EXISTS `CLUBS` (
  `id` int PRIMARY KEY AUTO_INCREMENT,
  `club_name` VARCHAR(255),
  `club_icon` VARCHAR(255)
);

CREATE TABLE IF NOT EXISTS `CLUB_MEMBERS` (
  `student_id` int,
  `club_id` int,
  `role` ENUM ('member', 'eboard'),
  PRIMARY KEY (`student_id`, `club_id`),
  CONSTRAINT `FK_ClubMembers_Students` FOREIGN KEY (`student_id`) REFERENCES `STUDENTS` (`id`),
  CONSTRAINT `FK_ClubMembers_Clubs` FOREIGN KEY (`club_id`) REFERENCES `CLUBS` (`id`)
);

CREATE TABLE IF NOT EXISTS `EVENT_VERSIONS` (
  `id` int PRIMARY KEY AUTO_INCREMENT,
  `author_id` int COMMENT 'FK', -- Changed to reference students.id directly
  `event_name` VARCHAR(255),
  `event_img` VARCHAR(255),
  `event_status` ENUM ('drafted', 'posted'),
  `event_date` datetime,
  `type` ENUM ('create', 'edit', 'delete'),
  `timestamp` datetime,
  CONSTRAINT `FK_EventVersions_Students` FOREIGN KEY (`author_id`) REFERENCES `STUDENTS` (`id`)
);

CREATE TABLE IF NOT EXISTS `EVENTS` (
  `id` int PRIMARY KEY AUTO_INCREMENT,
  `current_version_id` int,
  `original_version_id` int,
  CONSTRAINT `FK_Events_CurrentVersion` FOREIGN KEY (`current_version_id`) REFERENCES `EVENT_VERSIONS` (`id`),
  CONSTRAINT `FK_Events_OriginalVersion` FOREIGN KEY (`original_version_id`) REFERENCES `EVENT_VERSIONS` (`id`)
);

CREATE TABLE IF NOT EXISTS `EVENTS_TO_CLUBS` (
  `event_id` int COMMENT 'FK',
  `club_id` int COMMENT 'FK',
  PRIMARY KEY (`event_id`, `club_id`),
  CONSTRAINT `FK_EventsToClubs_Events` FOREIGN KEY (`event_id`) REFERENCES `EVENTS` (`id`),
  CONSTRAINT `FK_EventsToClubs_Clubs` FOREIGN KEY (`club_id`) REFERENCES `CLUBS` (`id`)
);

CREATE TABLE IF NOT EXISTS `POINT_SOURCES` (
  `id` int PRIMARY KEY AUTO_INCREMENT,
  `title` VARCHAR(255),
  `points` int
);

CREATE TABLE IF NOT EXISTS `POINT_HISTORIES` (
  `member_id` int COMMENT 'FK', -- Changed to reference students.id directly
  `point_source_id` int COMMENT 'FK',
  `points_earned` int,
  `points_after_gain` int,
  `timestamp` datetime,
  PRIMARY KEY (`member_id`, `point_source_id`),
  CONSTRAINT `FK_PointHistories_Students` FOREIGN KEY (`member_id`) REFERENCES `STUDENTS` (`id`),
  CONSTRAINT `FK_PointHistories_PointSources` FOREIGN KEY (`point_source_id`) REFERENCES `POINT_SOURCES` (`id`)
);

CREATE TABLE IF NOT EXISTS `PURCHASES` (
  `id` int PRIMARY KEY AUTO_INCREMENT,
  `member_id` int COMMENT 'FK', -- Changed to reference students.id directly
  `item_title` VARCHAR(255),
  `points_cost` int,
  `points_after_purchase` int,
  `timestamp` datetime,
  CONSTRAINT `FK_Purchases_Students` FOREIGN KEY (`member_id`) REFERENCES `STUDENTS` (`id`)
);