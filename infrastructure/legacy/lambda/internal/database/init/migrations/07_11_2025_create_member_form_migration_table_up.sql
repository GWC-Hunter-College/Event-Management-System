CREATE TABLE IF NOT EXISTS `MEMBER_FORM_DATA` (
  `email` VARCHAR(255),
  `full_name` VARCHAR(255),
  `major` VARCHAR(255),
  `emplid` VARCHAR(255),
  `grad_year` int,
  `dietary_restrictions` VARCHAR(255),
  `comments` VARCHAR(255),
  `join_date` timestamp
);

ALTER TABLE `MEMBER_FORM_DATA` COMMENT = 'Staging table to insert into students and student_info tables';