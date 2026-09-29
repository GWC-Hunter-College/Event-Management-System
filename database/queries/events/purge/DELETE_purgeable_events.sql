-- DELETE_purgeable_events.sql
-- Does: hard-deletes events that were soft-deleted longer ago than the retention
--   period. Their description, tags, club links and gallery links go with them
--   (ON DELETE CASCADE).
-- Used by: the purge batch job (internal), step 2, after UPDATE_images_of_purgeable_events.sql
--   with the same retention period, in the same transaction.
-- Params, in order:
--   1. retention period in days
-- Returns: affected rows.
DELETE FROM events
WHERE deleted_at < CURRENT_TIMESTAMP - INTERVAL ? DAY;
