# Images

Images go straight from the browser to S3 through short-lived presigned URLs. The API only signs URLs and records metadata. S3 is the current storage adapter, not part of the API contract; the [image-upload architecture doc](../../infrastructure/legacy/docs/architecture/image-uploads.md) covers the S3 and IAM side.

**No image route checks auth today.** Every deployed image route is 🟢 public, and the event-image handlers ignore `{clubId}`. The PDF intended these routes to be protected (🔴); its upload guide says auth "will be added before we release". The frontend doesn't call any of them yet; logos and flyers are placeholder `blob:` URLs in its design mode.

Routes are registered in [`club_image_routes.go`](../../infrastructure/legacy/gateway/routes/club_image_routes.go) and [`event_image_routes.go`](../../infrastructure/legacy/gateway/routes/event_image_routes.go). The three event `POST` routes also register `OPTIONS` against the same Lambda for CORS preflight; the handlers don't branch on method. Response shapes come from each handler's response struct; example values are made up, following the [implementation's image docs](../../infrastructure/legacy/docs/api/images.md).

## How image uploads work

This is the PDF's flow (pp. 25–26), with what the code does at each step.

1. **Request a presigned URL.** Send `POST` with `{ "filename": "poster.png", "mimetype": "image/png" }` to the route for the image's purpose:
   - club logo: [`POST /clubs/{clubId}/thumbnails`](#-post-clubsclubidthumbnails)
   - event thumbnail: [`POST /clubs/{clubId}/events/{eventId}/thumbnails`](#-post-clubsclubideventseventidthumbnails)
   - event gallery image: [`POST /clubs/{clubId}/events/{eventId}/images`](#-post-clubsclubideventseventidimages)

   The response is `{ "uploadUrl", "imageId", "objectKey" }`, and `uploadUrl` expires after one minute. **Exception:** the club-logo route returns `{ "uploadUrl", "key" }`, with no `imageId`.

2. **PUT the file to S3** at `uploadUrl`. The `Content-Type` header must match the `mimetype` you sent, because it's part of the signature. The PDF's example:

   ```js
   const uploadResponse = await fetch(uploadUrl, {
     method: "PUT",
     headers: { "Content-Type": file.type },
     body: file,
   });
   ```

3. **Confirm the upload** once the PUT succeeds. Send `POST` with `{ "filename", "mimetype", "imageId", "objectKey" }` to the matching confirm route:
   - club logo: [`POST /clubs/{clubId}/thumbnails/confirm`](#-post-clubsclubidthumbnailsconfirm) (⬜ not built)
   - event thumbnail: [`POST /clubs/{clubId}/events/{eventId}/thumbnails/confirm`](#-post-clubsclubideventseventidthumbnailsconfirm) (⬜ not built)
   - event gallery image: [`POST /clubs/{clubId}/events/{eventId}/images/confirm`](#-post-clubsclubideventseventidimagesconfirm) (🟨 built, fails on warm Lambda invocations)

**Reading images back:** gallery images are listed as one-minute signed GET URLs by [`GET /clubs/{clubId}/events/{eventId}/images`](#-get-clubsclubideventseventidimages), which is broken today (missing SQL). Logos and thumbnails aren't served yet: club reads return the logo's object key as `thumbnailUrl`, and event reads return a fixed placeholder.

**What works end to end today:** only event gallery images can be uploaded and confirmed, and confirmation only works on a cold Lambda. Club logos and event thumbnails can be uploaded to S3, but nothing attaches them to the club or event, because their confirm routes don't exist.

**Other rules:** the bucket blocks public access and enforces TLS. Object keys use a fresh UUID plus the file extension; the rest of the filename is dropped. MIME types, extensions, and file sizes aren't restricted. The PDF's generic internal `POST /images` metadata write isn't built; see [internal.md](internal.md#-image-metadata-write-pdf-post-images).

**Errors shared by the three signing routes:**

| Status | Body |
| --- | --- |
| `400` | The plain-text (not JSON) body `message: "Missing filename or mimetype"`, for bad JSON or a missing field. |
| `500` | `{"message": "Could not generate presigned URL: <reason>"}` |

## 🟢 POST `/clubs/{clubId}/thumbnails`

Returns a presigned S3 PUT URL for a club logo. The planning PDF writes it as protected `POST /clubs/thumbnails?filetype=<FILETYPE>&filename=<FILENAME>`; the built route takes `clubId` in the path and a JSON body instead.

**Auth:** 🟢 None. PDF intent: 🔴, for the club's e-board or owner. The handler doesn't check that the club exists.

**Path params:** `clubId`, used only as part of the object key.

**Request body:**

```json
{ "filename": "logo.png", "mimetype": "image/png" }
```

Both fields are required and non-empty.

**Response `200`** (handler `S3Response`):

```json
{
  "uploadUrl": "<one-minute-signed-put-url>",
  "key": "clubs/7/thumbnails/11111111-2222-3333-4444-555555555555.png"
}
```

**Errors:** the [shared signing errors](#how-image-uploads-work).

**Status:** ✅ Implemented.

**Known issues:**

- Returns `key` instead of the `imageId` and `objectKey` that the other signers and the PDF's flow use.
- There's no confirm route, so an upload never sets `clubs.fk_logo_id` and never appears in club reads.
- The `GET /clubs/{clubId}/thumbnails` registration is commented out.

**Proposed changes** (Need: **Now**; screen: New club form logo dropzone, called after `POST /clubs` returns the `clubId`). Fixes [M7](../coverage.md#m7-the-club-logo-signer-returns-key-not-imageid-and-objectkey).

- **Auth:** 🔴 JWT + owner or e-board of `clubId`; `404` for an unknown club.
- **Request body:** unchanged, `{ "filename", "mimetype" }`.
- **Response `200`:** the same shape as the event signers:

  ```json
  {
    "uploadUrl": "<one-minute-signed-put-url>",
    "imageId": "11111111-2222-3333-4444-555555555555",
    "objectKey": "clubs/7/thumbnails/11111111-2222-3333-4444-555555555555.png"
  }
  ```

**Queries** (2): 1. [`clubs/get/EXISTS_club.sql`](../../database/queries/clubs/get/EXISTS_club.sql) (read, `404`) → 2. [`authorization/clubs/can_manage/IS_student_authorized_club.sql`](../../database/queries/authorization/clubs/can_manage/IS_student_authorized_club.sql) (auth, `403`) → then the S3 presign (no query). The deployed route runs none (S3 only); the proposed contract adds these checks.

**Code:** route [`club_image_routes.go`](../../infrastructure/legacy/gateway/routes/club_image_routes.go) · handler [`clubs/thumbnails/post/post.go`](../../infrastructure/legacy/lambda/api/clubs/thumbnails/post/post.go)

## 🔴 POST `/clubs/{clubId}/thumbnails/confirm`

Saves an uploaded club logo: records the image metadata and points `clubs.fk_logo_id` at it (step 3 of the flow).

**Auth:** 🔴 JWT + e-board or owner of the club (intended).

**Request body (PDF):** `{ "filename", "mimetype", "imageId", "objectKey" }`.

**Response:** not defined.

**Status:** ⬜ Not built. No route, handler, or S3 existence check. The SQL and its transaction order exist ([query group 27](../../database/README.md#27-club-and-event-thumbnail-metadata-assignment)).

**Notes:** the designed behavior is to create or reuse the image row, update `clubs.fk_logo_id`, and clean up a replaced logo, atomically where possible.

**Proposed contract** (Need: **Now**; screen: New club form logo dropzone; later an edit-club screen).

- **Auth:** 🔴 JWT + owner or e-board of `clubId`.
- **Request body:** `{ "filename", "mimetype", "imageId", "objectKey" }`, the values the signer returned. `objectKey` must start with `clubs/{clubId}/thumbnails/`, and the object must exist in storage.
- **Response `200`:** `{ "message": "Logo saved", "clubId": 7, "thumbnailUrl": "<readable-logo-url>" }`, so the page can show the logo without re-reading the club.
- **Errors:** `400` bad body, key outside the club's prefix, or missing object; `401`; `403`; `404` unknown club; `500`.

**Queries** (5; steps 2–5 in one transaction): 1. [`authorization/clubs/can_manage/IS_student_authorized_club.sql`](../../database/queries/authorization/clubs/can_manage/IS_student_authorized_club.sql) (auth, `403`) → the S3 existence check (no query) → 2. [`clubs/thumbnails/confirm/SELECT_club_logo_for_update.sql`](../../database/queries/clubs/thumbnails/confirm/SELECT_club_logo_for_update.sql) (read, locks the club; no row is `404`) → 3. [`images/create/INSERT_image.sql`](../../database/queries/images/create/INSERT_image.sql) (write, `club-thumbnail`) → 4. [`clubs/thumbnails/confirm/UPDATE_club_logo.sql`](../../database/queries/clubs/thumbnails/confirm/UPDATE_club_logo.sql) (write) → 5. [`images/release/UPDATE_image_soft_delete_if_unused.sql`](../../database/queries/images/release/UPDATE_image_soft_delete_if_unused.sql) (write, the old logo, if there was one)

**Plan:** PDF upload guide, p. 25 · [query group 27](../../database/README.md#27-club-and-event-thumbnail-metadata-assignment)

## 🟢 POST `/clubs/{clubId}/events/{eventId}/thumbnails`

Returns a presigned S3 PUT URL for an event thumbnail. Its protected version in the design is [`POST /auth/events/{eventId}/thumbnails`](event-management.md#-post-autheventseventidthumbnails).

**Auth:** 🟢 None. PDF intent: 🔴, for the event's e-board or owner. The handler doesn't check that the event exists.

**Path params:** `eventId`, used only as part of the object key. `clubId` is ignored.

**Request body:** `{ "filename": "poster.png", "mimetype": "image/png" }`, both required.

**Response `200`** (handler `S3Response`):

```json
{
  "uploadUrl": "<one-minute-signed-put-url>",
  "imageId": "11111111-2222-3333-4444-555555555555",
  "objectKey": "events/42/thumbnails/11111111-2222-3333-4444-555555555555.png"
}
```

**Errors:** the [shared signing errors](#how-image-uploads-work).

**Status:** ✅ Implemented.

**Known issues:**

- There's no confirm route, so nothing sets `events.fk_thumbnail_id` and the returned `imageId` and `objectKey` go unused.

**Proposed changes** (Need: **Now**; screen: New event form flyer dropzone): add the Cognito authorizer and the event role check, or replace this route with [`POST /auth/events/{eventId}/thumbnails`](event-management.md#-post-autheventseventidthumbnails). Request and response stay as they are.

**Queries** (2): 1. [`authorization/events/can_manage/IS_student_authorized_event.sql`](../../database/queries/authorization/events/can_manage/IS_student_authorized_event.sql) (auth, `403`) → 2. [`events/read/SELECT_event_status.sql`](../../database/queries/events/read/SELECT_event_status.sql) (read, no row is `404`) → then the S3 presign (no query). The deployed route runs none (S3 only); the proposed contract adds these checks.

**Code:** route [`event_image_routes.go`](../../infrastructure/legacy/gateway/routes/event_image_routes.go) (`POST` and `OPTIONS`) · handler [`clubs/events/thumbnails/post/post.go`](../../infrastructure/legacy/lambda/api/clubs/events/thumbnails/post/post.go)

## 🔴 POST `/clubs/{clubId}/events/{eventId}/thumbnails/confirm`

Saves an uploaded event thumbnail by pointing `events.fk_thumbnail_id` at it (step 3 of the flow).

**Auth:** 🔴 JWT + e-board or owner of a linked club (intended).

**Request body (PDF):** `{ "filename", "mimetype", "imageId", "objectKey" }`.

**Response:** not defined.

**Status:** ⬜ Not built. No route, handler, or S3 existence check. The SQL and its transaction order exist ([query group 27](../../database/README.md#27-club-and-event-thumbnail-metadata-assignment)).

**Proposed contract** (Need: **Now**; screen: New event form flyer dropzone and alt text). Its `/auth/events/{eventId}/thumbnails/confirm` equivalent has the same contract if the protected routes move there ([open question](../README.md#open-questions)).

- **Auth:** 🔴 JWT + e-board or owner of any club linked to the event.
- **Request body:** `{ "filename", "mimetype", "imageId", "objectKey", "altText" }`. `altText` is optional; the rest are the signer's values. `objectKey` must start with `events/{eventId}/thumbnails/`, and the object must exist in storage.
- **Response `200`:** `{ "message": "Flyer saved", "eventId": 42, "thumbnailUrl": "<readable-flyer-url>" }`.
- **Errors:** `400`, `401`, `403`, `404` unknown event, `500`.
- **Writes:** the `images` row (with `altText` in its new alt-text column, [decision 4](../README.md#decisions); see [schema needs](../coverage.md#schema-needs)) and `events.fk_thumbnail_id`, in one transaction. Confirming a new flyer is also how its alt text changes.

**Queries** (5; steps 2–5 in one transaction): 1. [`authorization/events/can_manage/IS_student_authorized_event.sql`](../../database/queries/authorization/events/can_manage/IS_student_authorized_event.sql) (auth, `403`) → the S3 existence check (no query) → 2. [`events/update/SELECT_event_for_update.sql`](../../database/queries/events/update/SELECT_event_for_update.sql) (read, locks the event; no row is `404`) → 3. [`images/create/INSERT_image.sql`](../../database/queries/images/create/INSERT_image.sql) (write, `event-thumbnail`, the owner club, the caller, `altText`) → 4. [`events/thumbnails/confirm/UPDATE_event_thumbnail.sql`](../../database/queries/events/thumbnails/confirm/UPDATE_event_thumbnail.sql) (write) → 5. [`images/release/UPDATE_image_soft_delete_if_unused.sql`](../../database/queries/images/release/UPDATE_image_soft_delete_if_unused.sql) (write, the old flyer, if there was one)

**Plan:** PDF upload guide, p. 26 · [query group 27](../../database/README.md#27-club-and-event-thumbnail-metadata-assignment)

## 🟢 POST `/clubs/{clubId}/events/{eventId}/images`

Returns a presigned S3 PUT URL for an event gallery image. Its protected version in the design is [`POST /auth/events/{eventId}/images`](event-management.md#-post-autheventseventidimages).

**Auth:** 🟢 None. PDF intent: 🔴, for the event's e-board or owner. The handler doesn't check that the event exists or that it belongs to the club.

**Path params:** `eventId`, used only as part of the object key. `clubId` is ignored.

**Request body:** `{ "filename": "poster.png", "mimetype": "image/png" }`, both required.

**Response `200`** (handler `S3Response`):

```json
{
  "uploadUrl": "<one-minute-signed-put-url>",
  "imageId": "11111111-2222-3333-4444-555555555555",
  "objectKey": "events/42/images/11111111-2222-3333-4444-555555555555.png"
}
```

**Errors:** the [shared signing errors](#how-image-uploads-work).

**Status:** ✅ Implemented.

**Queries** (2): 1. [`authorization/events/can_manage/IS_student_authorized_event.sql`](../../database/queries/authorization/events/can_manage/IS_student_authorized_event.sql) (auth, `403`) → 2. [`events/read/SELECT_event_status.sql`](../../database/queries/events/read/SELECT_event_status.sql) (read, no row is `404`) → then the S3 presign (no query). The deployed route runs none (S3 only); the protected version adds these checks.

**Code:** route [`event_image_routes.go`](../../infrastructure/legacy/gateway/routes/event_image_routes.go) (`POST` and `OPTIONS`) · handler [`clubs/events/images/post/post.go`](../../infrastructure/legacy/lambda/api/clubs/events/images/post/post.go)

## 🟢 POST `/clubs/{clubId}/events/{eventId}/images/confirm`

Records an uploaded gallery image: inserts an `images` row with purpose `event-image`, then links it to the event in `event_images`.

**Auth:** 🟢 None. PDF intent: 🔴, for the event's e-board or owner.

**Path params:** `eventId`, passed to the insert without validation. `clubId` is ignored.

**Request body:**

```json
{
  "imageId": "11111111-2222-3333-4444-555555555555",
  "objectKey": "events/42/images/11111111-2222-3333-4444-555555555555.png",
  "filename": "poster.png",
  "mimetype": "image/png"
}
```

All four fields are required, non-empty, and stored as sent.

**Response `200`:**

```json
{ "message": "Image confirmed successfully" }
```

**Errors:**

| Status | Body |
| --- | --- |
| `400` | Plain text: `message: "Missing imageId, objectKey, filename, or mimetype in request body"` |
| `500` | Plain text: `message: "Database error: <reason>"` |

**Status:** 🟨 Broken on warm invocations. The handler closes its shared database client at the end of every request (`defer queryClient.Close()`), so later requests on the same Lambda instance fail until a cold start.

**Known issues:**

- Trusts client-supplied IDs, keys, and MIME types, and doesn't check that the S3 object exists or belongs to the event.
- **Need: Later.** No screen uploads gallery images yet ("More photos" is SOON).
- The two inserts commit separately, so a failed link leaves an orphan `images` row.

**Queries** (4; steps 2–4 in one transaction): 1. [`authorization/events/can_manage/IS_student_authorized_event.sql`](../../database/queries/authorization/events/can_manage/IS_student_authorized_event.sql) (auth, `403`) → 2. [`events/update/SELECT_event_for_update.sql`](../../database/queries/events/update/SELECT_event_for_update.sql) (read, the owner club; no row is `404`) → 3. [`images/create/INSERT_image.sql`](../../database/queries/images/create/INSERT_image.sql) (write, `event-image`) → 4. [`events/images/confirm/INSERT_event_image.sql`](../../database/queries/events/images/confirm/INSERT_event_image.sql) (write). One transaction, replacing the legacy handler's two separate commits.

**Code:** route [`event_image_routes.go`](../../infrastructure/legacy/gateway/routes/event_image_routes.go) (`POST` and `OPTIONS`) · handler [`clubs/events/images/confirm/post.go`](../../infrastructure/legacy/lambda/api/clubs/events/images/confirm/post.go) · SQL [`images/INSERT_image.sql`](../../infrastructure/legacy/utils/query_client/queries/images/INSERT_image.sql) and [`images/INSERT_event_image.sql`](../../infrastructure/legacy/utils/query_client/queries/images/INSERT_event_image.sql) · [query group 15](../../database/README.md#15-event-image-metadata-confirmation)

## 🟢 GET `/clubs/{clubId}/events/{eventId}/images`

Lists an event's gallery images with one-minute signed GET URLs. **It's broken:** the SQL file it loads doesn't exist, so every call returns `500`.

**Auth:** 🟢 None. It checks neither the event's posted status nor any role, and it ignores `clubId`.

**Path params:** `eventId`, used as the query argument.

**Response `200`** (declared by the handler's `ResponseJSONSchema`; never returned today):

```json
{
  "images": [
    {
      "imageId": "11111111-2222-3333-4444-555555555555",
      "mimetype": "image/png",
      "sourceUrl": "<one-minute-signed-get-url>"
    }
  ]
}
```

**Errors:** `500` `{"message": "Could not retrieve images: <reason>"}` on every call today.

**Status:** 🟨 Broken: the handler loads `images/SELECT_event_images.sql`, which doesn't exist, so it fails before signing any URLs.

**Known issues:**

- Like the confirm route, it closes its shared database connection after each request (`defer queryClient.Conn.Close()`).
- **Need: Later.** The Event page reads gallery URLs from the event object's `images` ([proposed event object](events.md#proposed-event-object)), not from this route.
- The design splits this route into public reads of posted events at [`GET /events/{eventId}/images`](events.md#-get-eventseventidimages) and manager reads at [`GET /auth/events/{eventId}/images`](event-management.md#-get-autheventseventidimages).

**Queries** (2): 1. [`events/read/SELECT_event_status.sql`](../../database/queries/events/read/SELECT_event_status.sql) (read, `404` unless `posted` or `cancelled`) → 2. [`events/images/list/SELECT_event_images.sql`](../../database/queries/events/images/list/SELECT_event_images.sql) (read). The design replaces this route with the public and manager gallery reads.

**Code:** route [`event_image_routes.go`](../../infrastructure/legacy/gateway/routes/event_image_routes.go) · handler [`clubs/events/images/get/get.go`](../../infrastructure/legacy/lambda/api/clubs/events/images/get/get.go) · [query group 16](../../database/README.md#16-event-image-list)

## 🔴 DELETE `/images/{imageId}`

Deletes an image's metadata and, under an explicit lifecycle policy, its S3 object (PDF: "Deletes image").

**Auth:** 🔴 JWT + e-board or owner of the image's owning club (`images.fk_club_id`), or an admin ([schema decision 5](../../database/docs/schema-review.md#decisions)).

**Path params:** `imageId`.

**Response:** not defined.

**Status:** ⬜ Not built. No route or handler. The SQL exists ([query group 28](../../database/README.md#28-image-deletion)).

**Behavior** ([schema decision 5](../../database/docs/schema-review.md#decisions) and [I4](../../database/docs/schema-review.md#i4-deleting-an-image)): a takedown clears every logo and flyer pointer to the file and every gallery link, then soft-deletes it, in one transaction. The file stops being served at once. An admin's [purge](admins.md#-post-adminspurge) removes the row and then the S3 object once it has been deleted for more than 30 days; a failed S3 delete leaves a harmless orphan file.

**Queries** (7; steps 3–6 in one transaction): 1. [`images/get/SELECT_image.sql`](../../database/queries/images/get/SELECT_image.sql) (read, no row is `404`; `fk_club_id` for the check) → 2. [`authorization/clubs/can_manage/IS_student_authorized_club.sql`](../../database/queries/authorization/clubs/can_manage/IS_student_authorized_club.sql) or [`authorization/admins/is_admin/IS_admin.sql`](../../database/queries/authorization/admins/is_admin/IS_admin.sql) (auth, manager of the owning club, or an admin; `403`) → 3. [`images/delete/UPDATE_clubs_clear_logo.sql`](../../database/queries/images/delete/UPDATE_clubs_clear_logo.sql) (write) → 4. [`images/delete/UPDATE_events_clear_thumbnail.sql`](../../database/queries/images/delete/UPDATE_events_clear_thumbnail.sql) (write) → 5. [`images/delete/DELETE_event_image_links.sql`](../../database/queries/images/delete/DELETE_event_image_links.sql) (write) → 6. [`images/delete/UPDATE_image_soft_delete.sql`](../../database/queries/images/delete/UPDATE_image_soft_delete.sql) (write). The S3 object is deleted later, by [`POST /admins/purge`](admins.md#-post-adminspurge).

**Plan:** PDF "Endpoints Revamp", p. 19 · [query group 28](../../database/README.md#28-image-deletion)
