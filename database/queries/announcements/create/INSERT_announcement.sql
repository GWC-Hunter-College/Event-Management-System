-- INSERT_announcement.sql
-- Does: creates an announcement. created_at and updated_at take their defaults.
-- Used by: POST /clubs/{clubId}/announcements, first statement of the create transaction.
-- Params, in order:
--   1. author id (the caller's sub)
--   2. title
--   3. body, or NULL for a draft
--   4. status: 'draft' or 'posted' (the API defaults to 'draft', and requires a body for 'posted')
-- Returns: the new announcement id as the last insert id.
INSERT INTO announcements
  (fk_author_id, title, body, status)
VALUES
  (?, ?, ?, ?);
