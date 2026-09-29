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

**Code:** route [`club_image_routes.go`](../../infrastructure/legacy/gateway/routes/club_image_routes.go) · handler [`clubs/thumbnails/post/post.go`](../../infrastructure/legacy/lambda/api/clubs/thumbnails/post/post.go)

## 🔴 POST `/clubs/{clubId}/thumbnails/confirm`

Saves an uploaded club logo: records the image metadata and points `clubs.fk_logo_id` at it (step 3 of the flow).

**Auth:** 🔴 JWT + e-board or owner of the club (intended).

**Request body (PDF):** `{ "filename", "mimetype", "imageId", "objectKey" }`.

**Response:** not defined.

**Status:** ⬜ Not built. No route, handler, query transaction, or S3 existence check.

**Notes:** the designed behavior is to create or reuse the image row, update `clubs.fk_logo_id`, and clean up a replaced logo, atomically where possible.

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

**Code:** route [`event_image_routes.go`](../../infrastructure/legacy/gateway/routes/event_image_routes.go) (`POST` and `OPTIONS`) · handler [`clubs/events/thumbnails/post/post.go`](../../infrastructure/legacy/lambda/api/clubs/events/thumbnails/post/post.go)

## 🔴 POST `/clubs/{clubId}/events/{eventId}/thumbnails/confirm`

Saves an uploaded event thumbnail by pointing `events.fk_thumbnail_id` at it (step 3 of the flow).

**Auth:** 🔴 JWT + e-board or owner of a linked club (intended).

**Request body (PDF):** `{ "filename", "mimetype", "imageId", "objectKey" }`.

**Response:** not defined.

**Status:** ⬜ Not built. No route, handler, query transaction, or S3 existence check.

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
- The two inserts commit separately, so a failed link leaves an orphan `images` row.

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
- The design splits this route into public reads of posted events at [`GET /events/{eventId}/images`](events.md#-get-eventseventidimages) and manager reads at [`GET /auth/events/{eventId}/images`](event-management.md#-get-autheventseventidimages).

**Code:** route [`event_image_routes.go`](../../infrastructure/legacy/gateway/routes/event_image_routes.go) · handler [`clubs/events/images/get/get.go`](../../infrastructure/legacy/lambda/api/clubs/events/images/get/get.go) · [query group 16](../../database/README.md#16-event-image-list)

## 🔴 DELETE `/images/{imageId}`

Deletes an image's metadata and, under an explicit lifecycle policy, its S3 object (PDF: "Deletes image").

**Auth:** 🔴 JWT + a policy that depends on what the image belongs to: its club, its event, or an admin.

**Path params:** `imageId`.

**Response:** not defined.

**Status:** ⬜ Not built. No route, handler, SQL cleanup, or S3 delete.

**Open:** link cleanup, clearing thumbnail foreign keys, deletion order, failure recovery, and how long orphans are kept.

**Plan:** PDF "Endpoints Revamp", p. 19 · [query group 28](../../database/README.md#28-image-deletion)
