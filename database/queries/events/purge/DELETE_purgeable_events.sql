-- DELETE_purgeable_events.sql
-- Does: hard-deletes events deleted more than 30 days ago, which can no longer be
--   restored. Their description, tags, club links and gallery links go with them
--   (ON DELETE CASCADE).
-- Used by: POST /admins/purge, step 2, after UPDATE_images_of_purgeable_events.sql
--   in the same transaction.
-- Params: none. The 30 days match the restore window in events/restore/.
-- Returns: affected rows (the number of events purged).
DELETE FROM events
WHERE deleted_at < CURRENT_TIMESTAMP - INTERVAL 30 DAY;
