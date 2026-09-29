-- 2026_09_29_baseline_up.sql
-- The complete schema, created from scratch. It replaces the older files in
-- database/migrations/history, which are kept only as a record and are never applied.
-- The design decisions behind it are in database/docs/schema-review.md.
--
-- Conventions
--   Event start and end times are DATETIME in UTC. Compare them with UTC_TIMESTAMP().
--   created_at, updated_at and deleted_at are TIMESTAMP, which MySQL stores in UTC.
--   Compare them with CURRENT_TIMESTAMP.
--   deleted_at marks a soft-deleted row. Reads skip it. For 30 days the row can be
--   restored, and after that only an admin's purge can hard-delete it. The restore
--   and purge queries both hard-code the 30 days, so nothing restorable is purged.
--   Every foreign key is named, so a later migration can alter it.
--
-- Runs on an empty database. Written so a runner that splits on semicolons works:
-- no comment contains one, and no statement is only a comment.

-- Students: one row per Cognito user. id is the Cognito sub.
CREATE TABLE `students` (
  `id` CHAR(36) NOT NULL,
  `email` VARCHAR(255) NULL,
  PRIMARY KEY (`id`)
);

CREATE TABLE `student_info` (
  `fk_student_id` CHAR(36) NOT NULL,
  `username` VARCHAR(30) NULL,
  `first_name` VARCHAR(40) NULL,
  `last_name` VARCHAR(40) NULL,
  PRIMARY KEY (`fk_student_id`),
  UNIQUE KEY `uq_student_info_username` (`username`),
  CONSTRAINT `fk_student_info_student` FOREIGN KEY (`fk_student_id`) REFERENCES `students` (`id`) ON DELETE CASCADE
);

CREATE TABLE `admins` (
  `fk_student_id` CHAR(36) NOT NULL,
  PRIMARY KEY (`fk_student_id`),
  CONSTRAINT `fk_admins_student` FOREIGN KEY (`fk_student_id`) REFERENCES `students` (`id`)
);

-- Images: one row per stored file. Club logos, event flyers and gallery photos
-- reference a row, and one file may have several uses. fk_club_id is the owning
-- club, whose managers (and admins) may take the file down. A referenced image
-- never has deleted_at set, except through an event that is itself soft-deleted.
-- The foreign key to clubs is added after clubs exists, below.
CREATE TABLE `images` (
  `id` CHAR(36) NOT NULL,
  `fk_club_id` INT NULL,
  `fk_uploaded_by` CHAR(36) NULL,
  `purpose` VARCHAR(100) NOT NULL,
  `object_key` VARCHAR(255) NOT NULL,
  `filename` VARCHAR(255) NULL,
  `mimetype` VARCHAR(255) NOT NULL,
  `alt_text` VARCHAR(1000) NULL,
  `created_at` TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
  `deleted_at` TIMESTAMP NULL,
  PRIMARY KEY (`id`),
  UNIQUE KEY `uq_images_object_key` (`object_key`),
  KEY `idx_images_club` (`fk_club_id`),
  KEY `idx_images_deleted` (`deleted_at`),
  CONSTRAINT `fk_images_uploaded_by` FOREIGN KEY (`fk_uploaded_by`) REFERENCES `students` (`id`) ON DELETE SET NULL
);

-- Clubs. The name is unique under the default case- and accent-insensitive collation.
CREATE TABLE `clubs` (
  `id` INT NOT NULL AUTO_INCREMENT,
  `fk_logo_id` CHAR(36) NULL,
  `name` VARCHAR(255) NOT NULL,
  PRIMARY KEY (`id`),
  UNIQUE KEY `uq_clubs_name` (`name`),
  KEY `idx_clubs_logo` (`fk_logo_id`),
  CONSTRAINT `fk_clubs_logo` FOREIGN KEY (`fk_logo_id`) REFERENCES `images` (`id`)
);

ALTER TABLE `images`
  ADD CONSTRAINT `fk_images_club` FOREIGN KEY (`fk_club_id`) REFERENCES `clubs` (`id`);

CREATE TABLE `club_info` (
  `fk_club_id` INT NOT NULL,
  `website_url` VARCHAR(255) NULL,
  `description` TEXT NULL,
  PRIMARY KEY (`fk_club_id`),
  CONSTRAINT `fk_club_info_club` FOREIGN KEY (`fk_club_id`) REFERENCES `clubs` (`id`) ON DELETE CASCADE
);

-- Membership. Each member has exactly one role. The e-board list is
-- role IN ('eboard', 'owner'). A club's last owner can't be demoted and can't leave,
-- which the role-update and leave queries enforce.
CREATE TABLE `club_members` (
  `fk_student_id` CHAR(36) NOT NULL,
  `fk_club_id` INT NOT NULL,
  `role` ENUM ('member', 'eboard', 'owner') NOT NULL DEFAULT 'member',
  PRIMARY KEY (`fk_student_id`, `fk_club_id`),
  KEY `idx_club_members_club_role` (`fk_club_id`, `role`),
  CONSTRAINT `fk_club_members_student` FOREIGN KEY (`fk_student_id`) REFERENCES `students` (`id`),
  CONSTRAINT `fk_club_members_club` FOREIGN KEY (`fk_club_id`) REFERENCES `clubs` (`id`)
);

CREATE TABLE `verified_clubs` (
  `fk_club_id` INT NOT NULL,
  PRIMARY KEY (`fk_club_id`),
  CONSTRAINT `fk_verified_clubs_club` FOREIGN KEY (`fk_club_id`) REFERENCES `clubs` (`id`) ON DELETE CASCADE
) COMMENT = 'Table for making sure clubs are allowed to be shown to account for bad actors creating random new clubs';

-- Club topics: a fixed list held as data, at most three per club.
-- The binary collation makes keys case-sensitive, so only the exact lowercase keys pass.
-- ON UPDATE CASCADE lets a key be renamed with one UPDATE of topics.
CREATE TABLE `topics` (
  `tag` VARCHAR(40) CHARACTER SET utf8mb4 COLLATE utf8mb4_bin NOT NULL,
  PRIMARY KEY (`tag`)
);

INSERT INTO `topics` (`tag`) VALUES
  ('technology'),
  ('arts'),
  ('community'),
  ('women in stem'),
  ('career'),
  ('sports');

CREATE TABLE `club_tags` (
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

-- Events. status is the life of an event that still exists: draft, posted, cancelled.
-- DELETE sets deleted_at. start_date and end_date are UTC. timezone is the IANA zone
-- the event takes place in, for display and calendar export only.
CREATE TABLE `events` (
  `id` INT NOT NULL AUTO_INCREMENT,
  `fk_author_id` CHAR(36) NULL,
  `fk_thumbnail_id` CHAR(36) NULL,
  `title` VARCHAR(255) NOT NULL,
  `location` VARCHAR(255) NULL,
  `rsvp_link` VARCHAR(255) NULL,
  `status` ENUM ('draft', 'posted', 'cancelled') NOT NULL DEFAULT 'draft',
  `start_date` DATETIME NOT NULL,
  `end_date` DATETIME NOT NULL,
  `timezone` VARCHAR(60) NOT NULL DEFAULT 'America/New_York',
  `created_at` TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
  `updated_at` TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  `deleted_at` TIMESTAMP NULL,
  PRIMARY KEY (`id`),
  KEY `idx_events_status_start` (`status`, `start_date`),
  KEY `idx_events_thumbnail` (`fk_thumbnail_id`),
  KEY `idx_events_author` (`fk_author_id`),
  KEY `idx_events_deleted` (`deleted_at`),
  CONSTRAINT `chk_events_end_after_start` CHECK (`end_date` >= `start_date`),
  CONSTRAINT `fk_events_author` FOREIGN KEY (`fk_author_id`) REFERENCES `students` (`id`) ON DELETE SET NULL,
  CONSTRAINT `fk_events_thumbnail` FOREIGN KEY (`fk_thumbnail_id`) REFERENCES `images` (`id`)
);

CREATE TABLE `event_descriptions` (
  `fk_event_id` INT NOT NULL,
  `description` TEXT NULL,
  PRIMARY KEY (`fk_event_id`),
  CONSTRAINT `fk_event_descriptions_event` FOREIGN KEY (`fk_event_id`) REFERENCES `events` (`id`) ON DELETE CASCADE
);

-- Free-text event tags. Nothing reads or writes them yet.
CREATE TABLE `event_tags` (
  `fk_event_id` INT NOT NULL,
  `tag` VARCHAR(255) NOT NULL,
  PRIMARY KEY (`fk_event_id`, `tag`),
  CONSTRAINT `fk_event_tags_event` FOREIGN KEY (`fk_event_id`) REFERENCES `events` (`id`) ON DELETE CASCADE
);

-- Clubs linked to an event. At most one owner club per event: the functional index
-- only has a key for owner rows.
CREATE TABLE `events_to_clubs` (
  `fk_event_id` INT NOT NULL,
  `fk_club_id` INT NOT NULL,
  `club_is_event_owner` BOOL NOT NULL DEFAULT FALSE,
  PRIMARY KEY (`fk_event_id`, `fk_club_id`),
  KEY `idx_events_to_clubs_club` (`fk_club_id`),
  UNIQUE KEY `uq_events_to_clubs_one_owner` ((IF(`club_is_event_owner`, `fk_event_id`, NULL))),
  CONSTRAINT `fk_events_to_clubs_event` FOREIGN KEY (`fk_event_id`) REFERENCES `events` (`id`) ON DELETE CASCADE,
  CONSTRAINT `fk_events_to_clubs_club` FOREIGN KEY (`fk_club_id`) REFERENCES `clubs` (`id`)
);

-- An event's gallery. created_at orders the gallery, because one file may sit in
-- several galleries. Deleting an image that a gallery still uses is refused.
CREATE TABLE `event_images` (
  `fk_event_id` INT NOT NULL,
  `fk_image_id` CHAR(36) NOT NULL,
  `created_at` TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
  PRIMARY KEY (`fk_event_id`, `fk_image_id`),
  KEY `idx_event_images_image` (`fk_image_id`),
  CONSTRAINT `fk_event_images_event` FOREIGN KEY (`fk_event_id`) REFERENCES `events` (`id`) ON DELETE CASCADE,
  CONSTRAINT `fk_event_images_image` FOREIGN KEY (`fk_image_id`) REFERENCES `images` (`id`)
);

-- Announcements: updates a club's e-board posts, modelled on events. status is
-- draft or posted, and draft -> posted is the only transition. They follow the event
-- rules for deleted_at: soft delete, 30-day restore, then an admin's purge.
-- body may be empty on a draft, and posting requires one.
-- Room to grow the way events did: a flyer is a nullable fk_thumbnail_id to images,
-- a gallery is an announcement_images link table, and tags are announcement_tags,
-- each cascading from announcements like their event counterparts.
CREATE TABLE `announcements` (
  `id` INT NOT NULL AUTO_INCREMENT,
  `fk_author_id` CHAR(36) NULL,
  `title` VARCHAR(255) NOT NULL,
  `body` TEXT NULL,
  `status` ENUM ('draft', 'posted') NOT NULL DEFAULT 'draft',
  `created_at` TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
  `updated_at` TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  `deleted_at` TIMESTAMP NULL,
  PRIMARY KEY (`id`),
  KEY `idx_announcements_status_created` (`status`, `created_at`),
  KEY `idx_announcements_author` (`fk_author_id`),
  KEY `idx_announcements_deleted` (`deleted_at`),
  CONSTRAINT `fk_announcements_author` FOREIGN KEY (`fk_author_id`) REFERENCES `students` (`id`) ON DELETE SET NULL
);

-- Clubs an announcement belongs to, with at most one owner club, as events_to_clubs.
-- One announcement can go to several clubs (HunterHacks for the CS clubs, say).
CREATE TABLE `announcements_to_clubs` (
  `fk_announcement_id` INT NOT NULL,
  `fk_club_id` INT NOT NULL,
  `club_is_announcement_owner` BOOL NOT NULL DEFAULT FALSE,
  PRIMARY KEY (`fk_announcement_id`, `fk_club_id`),
  KEY `idx_announcements_to_clubs_club` (`fk_club_id`),
  UNIQUE KEY `uq_announcements_to_clubs_one_owner` ((IF(`club_is_announcement_owner`, `fk_announcement_id`, NULL))),
  CONSTRAINT `fk_announcements_to_clubs_announcement` FOREIGN KEY (`fk_announcement_id`) REFERENCES `announcements` (`id`) ON DELETE CASCADE,
  CONSTRAINT `fk_announcements_to_clubs_club` FOREIGN KEY (`fk_club_id`) REFERENCES `clubs` (`id`)
);

-- event_details: one row per event that isn't soft-deleted, with everything the
-- event object needs. Every event read selects from it, so the soft-delete filter
-- and the flyer, description, photo count and club columns are written once.
--   image_count counts gallery images other than the flyer.
--   owner_* is the owner club, or NULL when the event has no owner link.
--   associates is a JSON array of {id, name, logoObjectKey} in no guaranteed order
--   (the API sorts it by id), and an empty array when there are none.
CREATE VIEW `event_details` AS
SELECT
  e.`id`,
  e.`fk_author_id`,
  e.`fk_thumbnail_id`,
  flyer.`object_key` AS `thumbnail_object_key`,
  flyer.`alt_text` AS `alt_text`,
  e.`title`,
  e.`location`,
  e.`rsvp_link`,
  e.`status`,
  e.`start_date`,
  e.`end_date`,
  e.`timezone`,
  e.`created_at`,
  e.`updated_at`,
  ed.`description`,
  (
    SELECT COUNT(*)
    FROM `event_images` ei
    WHERE ei.`fk_event_id` = e.`id`
      AND ei.`fk_image_id` <> COALESCE(e.`fk_thumbnail_id`, '')
  ) AS `image_count`,
  oc.`id` AS `owner_club_id`,
  oc.`name` AS `owner_club_name`,
  ologo.`object_key` AS `owner_logo_object_key`,
  COALESCE(
    (
      SELECT JSON_ARRAYAGG(JSON_OBJECT('id', c.`id`, 'name', c.`name`, 'logoObjectKey', li.`object_key`))
      FROM `events_to_clubs` ac
      JOIN `clubs` c ON c.`id` = ac.`fk_club_id`
      LEFT JOIN `images` li ON li.`id` = c.`fk_logo_id`
      WHERE ac.`fk_event_id` = e.`id`
        AND ac.`club_is_event_owner` = FALSE
    ),
    JSON_ARRAY()
  ) AS `associates`
FROM `events` e
LEFT JOIN `images` flyer ON flyer.`id` = e.`fk_thumbnail_id`
LEFT JOIN `event_descriptions` ed ON ed.`fk_event_id` = e.`id`
LEFT JOIN `events_to_clubs` oec ON oec.`fk_event_id` = e.`id` AND oec.`club_is_event_owner` = TRUE
LEFT JOIN `clubs` oc ON oc.`id` = oec.`fk_club_id`
LEFT JOIN `images` ologo ON ologo.`id` = oc.`fk_logo_id`
WHERE e.`deleted_at` IS NULL;

-- club_details: one row per club, with everything the club object needs.
--   tags is a JSON array of topic keys in the order the club chose them.
--   member_count counts every membership: members, e-board and owners.
CREATE VIEW `club_details` AS
SELECT
  c.`id`,
  c.`name`,
  logo.`object_key` AS `thumbnail_url`,
  ci.`website_url`,
  ci.`description`,
  COALESCE(
    (
      SELECT CAST(CONCAT('[', GROUP_CONCAT(JSON_QUOTE(ct.`tag`) ORDER BY ct.`slot` SEPARATOR ','), ']') AS JSON)
      FROM `club_tags` ct
      WHERE ct.`fk_club_id` = c.`id`
    ),
    JSON_ARRAY()
  ) AS `tags`,
  (
    SELECT COUNT(*)
    FROM `club_members` cm
    WHERE cm.`fk_club_id` = c.`id`
  ) AS `member_count`,
  (vc.`fk_club_id` IS NOT NULL) AS `verified`
FROM `clubs` c
LEFT JOIN `images` logo ON logo.`id` = c.`fk_logo_id`
LEFT JOIN `club_info` ci ON ci.`fk_club_id` = c.`id`
LEFT JOIN `verified_clubs` vc ON vc.`fk_club_id` = c.`id`;

-- announcement_details: one row per announcement that isn't soft-deleted, with its
-- owner club and co-owning clubs, as event_details does for events.
--   associates is a JSON array of {id, name, logoObjectKey} in no guaranteed order.
CREATE VIEW `announcement_details` AS
SELECT
  a.`id`,
  a.`fk_author_id`,
  a.`title`,
  a.`body`,
  a.`status`,
  a.`created_at`,
  a.`updated_at`,
  oc.`id` AS `owner_club_id`,
  oc.`name` AS `owner_club_name`,
  ologo.`object_key` AS `owner_logo_object_key`,
  COALESCE(
    (
      SELECT JSON_ARRAYAGG(JSON_OBJECT('id', c.`id`, 'name', c.`name`, 'logoObjectKey', li.`object_key`))
      FROM `announcements_to_clubs` ac
      JOIN `clubs` c ON c.`id` = ac.`fk_club_id`
      LEFT JOIN `images` li ON li.`id` = c.`fk_logo_id`
      WHERE ac.`fk_announcement_id` = a.`id`
        AND ac.`club_is_announcement_owner` = FALSE
    ),
    JSON_ARRAY()
  ) AS `associates`
FROM `announcements` a
LEFT JOIN `announcements_to_clubs` oac ON oac.`fk_announcement_id` = a.`id` AND oac.`club_is_announcement_owner` = TRUE
LEFT JOIN `clubs` oc ON oc.`id` = oac.`fk_club_id`
LEFT JOIN `images` ologo ON ologo.`id` = oc.`fk_logo_id`
WHERE a.`deleted_at` IS NULL;
