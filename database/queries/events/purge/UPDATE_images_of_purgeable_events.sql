-- UPDATE_images_of_purgeable_events.sql
-- Does: soft-deletes the flyers and gallery images of events that are about to be
--   purged, unless a club logo or an event outside the purge still uses them. Each
--   image takes its event's deleted_at, so the image purge in the same run removes it.
-- Used by: the purge batch job (internal), step 1, before DELETE_purgeable_events.sql,
--   in the same transaction.
-- Params, in order:
--   1. retention period in days: events deleted longer ago than this are purged
-- Returns: affected rows.
UPDATE images i
JOIN (SELECT CURRENT_TIMESTAMP - INTERVAL ? DAY AS cutoff) p
JOIN events e
  ON e.deleted_at < p.cutoff
  AND (
    e.fk_thumbnail_id = i.id
    OR EXISTS (
      SELECT 1
      FROM event_images ei
      WHERE ei.fk_event_id = e.id
        AND ei.fk_image_id = i.id
    )
  )
SET i.deleted_at = e.deleted_at
WHERE i.deleted_at IS NULL
  AND NOT EXISTS (SELECT 1 FROM clubs c WHERE c.fk_logo_id = i.id)
  AND NOT EXISTS (
    SELECT 1
    FROM events keep
    WHERE keep.fk_thumbnail_id = i.id
      AND (keep.deleted_at IS NULL OR keep.deleted_at >= p.cutoff)
  )
  AND NOT EXISTS (
    SELECT 1
    FROM event_images kei
    JOIN events keep ON keep.id = kei.fk_event_id
    WHERE kei.fk_image_id = i.id
      AND (keep.deleted_at IS NULL OR keep.deleted_at >= p.cutoff)
  );
