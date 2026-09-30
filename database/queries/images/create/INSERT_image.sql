-- INSERT_image.sql
-- Does: records a file after its upload, with its owning club, uploader and alt text.
--   This is the shared image metadata write (internal, the PDF's POST /images).
-- Used by: POST /clubs/{clubId}/thumbnails/confirm, POST /clubs/{clubId}/events/{eventId}/thumbnails/confirm
--   and POST /clubs/{clubId}/events/{eventId}/images/confirm, each inside its transaction.
-- Params, in order:
--   1. image id: the UUID the signer returned
--   2. owning club id: the logo's club, or the event's owner club
--   3. uploader id (the caller's sub), or NULL
--   4. purpose: 'club-thumbnail', 'event-thumbnail' or 'event-image'
--   5. object key: the key the signer returned, unique
--   6. filename, or NULL
--   7. mimetype
--   8. alt text, or NULL
-- Returns: nothing. created_at takes its default.
INSERT INTO images
  (id, fk_club_id, fk_uploaded_by, purpose, object_key, filename, mimetype, alt_text)
VALUES
  (?, ?, ?, ?, ?, ?, ?, ?);
