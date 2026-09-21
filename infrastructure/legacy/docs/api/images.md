# Image API

Image routes issue short-lived S3 signed URLs or record event-gallery metadata. All active image routes are registered on both HTTP APIs without the Cognito authorizer, and none of their handlers performs a database role check.

The shared bucket blocks public access and enforces TLS. A successful upload therefore uses the returned signed `PUT` URL rather than a public object URL. See [Image upload architecture](../architecture/image-uploads.md) for the S3 and IAM flow and [Authorization](../architecture/authorization.md) for the access boundary.

## Route summary

| Method and path | JWT attached | Purpose |
| --- | --- | --- |
| `POST /clubs/{clubId}/thumbnails` | No | Sign a club-thumbnail `PUT`. |
| `GET /clubs/{clubId}/events/{eventId}/images` | No | Read event-gallery metadata and sign each `GET`. |
| `POST /clubs/{clubId}/events/{eventId}/images` | No | Sign an event-gallery `PUT`. |
| `POST /clubs/{clubId}/events/{eventId}/images/confirm` | No | Insert gallery image metadata and its event link. |
| `POST /clubs/{clubId}/events/{eventId}/thumbnails` | No | Sign an event-thumbnail `PUT`. |

The three event `POST` paths are also explicitly registered for `OPTIONS` using the same Lambda integrations. The handlers contain no method branch and apply their normal body validation whenever that integration receives an `OPTIONS` request. API Gateway also has API-level CORS configuration.

On `200`, the three signing handlers and the gallery-read handler set `Access-Control-Allow-Origin: *`, `Access-Control-Allow-Methods: OPTIONS,POST,GET`, and `Access-Control-Allow-Headers: *`. Their explicit error responses do not set those headers. The confirmation handler sets no response headers on either success or failure, and none of these handlers explicitly sets `Content-Type`.

## Signed-upload request

All three upload-signing handlers accept the same JSON fields:

```json
{
  "filename": "poster.png",
  "mimetype": "image/png"
}
```

Both strings must be non-empty. The handlers do not restrict MIME types, filename extensions, or file size. The filename is used only to obtain its final extension; generated object keys use a new UUID, not the supplied base filename. The requested MIME type is included as the S3 `Content-Type` for the signed operation.

Signed URLs expire after one minute. Examples below replace the URL and UUID with sanitized placeholders.

## `POST /clubs/{clubId}/thumbnails`

Creates a signed `PUT` for this object-key form:

```text
clubs/{clubId}/thumbnails/{generated-uuid}{filename-extension}
```

- JWT authorizer: not attached.
- Handler: [`lambda/api/clubs/thumbnails/post/post.go`](../../lambda/api/clubs/thumbnails/post/post.go).
- `clubId`: used as an object-key segment; not validated against the database.
- Query parameters: none.

`200 OK`:

```json
{
  "uploadUrl": "<one-minute-signed-put-url>",
  "key": "clubs/7/thumbnails/11111111-2222-3333-4444-555555555555.png"
}
```

Missing or invalid JSON, `filename`, or `mimetype` returns `400` with a plain-text, JSON-like body: `message: "Missing filename or mimetype"`. Signing failures return `500` with a JSON `message`.

There is no active club-thumbnail confirmation route. This operation does not insert an `images` row or assign `clubs.fk_logo_id`, so the signed upload alone does not make the object appear in club read responses. The `GET /clubs/{clubId}/thumbnails` registration is commented out and is not active.

## `POST /clubs/{clubId}/events/{eventId}/images`

Creates a signed event-gallery `PUT` for:

```text
events/{eventId}/images/{generated-uuid}{filename-extension}
```

- JWT authorizer: not attached.
- Handler: [`lambda/api/clubs/events/images/post/post.go`](../../lambda/api/clubs/events/images/post/post.go).
- `eventId`: used as an object-key segment; not checked against the database.
- `clubId`: ignored by the handler.
- Query parameters: none.

`200 OK`:

```json
{
  "uploadUrl": "<one-minute-signed-put-url>",
  "imageId": "11111111-2222-3333-4444-555555555555",
  "objectKey": "events/42/images/11111111-2222-3333-4444-555555555555.png"
}
```

The `imageId` and `objectKey` are inputs to the confirmation call after the client uploads the object. Invalid input returns the same `400` body as the club signer; signing failures return `500` with a JSON `message`.

## `POST /clubs/{clubId}/events/{eventId}/images/confirm`

Records an already uploaded gallery object.

- JWT authorizer: not attached.
- Handler: [`lambda/api/clubs/events/images/confirm/post.go`](../../lambda/api/clubs/events/images/confirm/post.go).
- Queries: [`images/INSERT_image.sql`](../../utils/query_client/queries/images/INSERT_image.sql) and [`images/INSERT_event_image.sql`](../../utils/query_client/queries/images/INSERT_event_image.sql).
- Entity/table fields: [`models.Image`](../../database/models/image.go), `images`, and `event_images`.
- `eventId`: passed to the event-image insert without integer validation.
- `clubId`: ignored by the handler.
- Query parameters: none.

Request body:

```json
{
  "imageId": "11111111-2222-3333-4444-555555555555",
  "objectKey": "events/42/images/11111111-2222-3333-4444-555555555555.png",
  "filename": "poster.png",
  "mimetype": "image/png"
}
```

All four strings must be non-empty. The handler trusts their values: it does not check S3 object existence, key shape, UUID format, MIME type, `{clubId}`, event ownership, or caller identity. It stores purpose as the fixed string `event-image`.

`200 OK`:

```json
{
  "message": "Image confirmed successfully"
}
```

The two inserts are separate operations rather than one transaction. If the `images` insert succeeds and the `event_images` insert fails, the image row remains. Input validation returns `400` with a plain-text, JSON-like body. Either insert failure returns `500` and includes the database error text in that body.

The Lambda keeps its query client in package state but closes it at the end of every invocation. A subsequent warm invocation can therefore encounter a closed database connection.

## `GET /clubs/{clubId}/events/{eventId}/images`

The handler is written to query event-gallery images and issue one-minute signed `GET` URLs.

- JWT authorizer: not attached.
- Handler: [`lambda/api/clubs/events/images/get/get.go`](../../lambda/api/clubs/events/images/get/get.go).
- `eventId`: used as the query argument.
- `clubId`: ignored by the handler.
- Query parameters and body: none.

Its declared `200` response shape is:

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

In the current repository, the handler requests `images/SELECT_event_images.sql`, but that embedded SQL file does not exist under [`utils/query_client/queries/images`](../../utils/query_client/queries/images). The active operation therefore takes its `500` query-error path before generating any signed URLs. The error body is JSON with a `message` beginning `Could not retrieve images:`.

Like the confirmation handler, this Lambda closes its package-level database connection after each request.

## `POST /clubs/{clubId}/events/{eventId}/thumbnails`

Creates a signed event-thumbnail `PUT` for:

```text
events/{eventId}/thumbnails/{generated-uuid}{filename-extension}
```

- JWT authorizer: not attached.
- Handler: [`lambda/api/clubs/events/thumbnails/post/post.go`](../../lambda/api/clubs/events/thumbnails/post/post.go).
- `eventId`: used as an object-key segment; not checked against the database.
- `clubId`: ignored by the handler.
- Query parameters: none.

`200 OK`:

```json
{
  "uploadUrl": "<one-minute-signed-put-url>",
  "imageId": "11111111-2222-3333-4444-555555555555",
  "objectKey": "events/42/thumbnails/11111111-2222-3333-4444-555555555555.png"
}
```

Input and signing errors follow the same `400` and `500` behavior as the event-gallery signer. No active endpoint confirms this metadata or assigns `events.fk_thumbnail_id`, so the signed upload alone does not connect the thumbnail to an event row.
