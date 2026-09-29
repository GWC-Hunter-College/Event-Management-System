-- INSERT_announcement.sql
-- Does: creates an announcement. posted_at is set to now when it's created as posted,
--   and left NULL for a draft. created_at and updated_at take their defaults.
-- Used by: POST /clubs/{clubId}/announcements, first statement of the create transaction.
-- Params, in order:
--   1. author id (the caller's sub)
--   2. title
--   3. body, or NULL for a draft
--   4. status: 'draft' or 'posted' (the API defaults to 'draft', and requires a body for 'posted')
-- Returns: the new announcement id as the last insert id.
INSERT INTO announcements
  (fk_author_id, title, body, status, posted_at)
SELECT
  p.author_id,
  p.title,
  p.body,
  p.status,
  IF(p.status = 'posted', CURRENT_TIMESTAMP, NULL)
FROM (
  SELECT
    CAST(? AS CHAR) AS author_id,
    CAST(? AS CHAR) AS title,
    CAST(? AS CHAR) AS body,
    CAST(? AS CHAR) AS status
) p;
