# Club API

Club routes cover public discovery, authenticated creation, caller membership, club-scoped event listing, and thumbnail upload signing. Their registrations are in [`gateway/routes/club_routes.go`](../../gateway/routes/club_routes.go) and [`gateway/routes/club_image_routes.go`](../../gateway/routes/club_image_routes.go).

Authentication on a route is not the same as a database role check. The active create-club and create-event handlers do not query `admins`, `member_is_eboard`, or `member_is_owner`. See [Authorization](../architecture/authorization.md).

## Club objects

List responses use [`models.Club`](../../database/models/club.go):

```json
{
  "id": 7,
  "name": "Example Club",
  "thumbnailUrl": "clubs/7/thumbnails/example.png"
}
```

`thumbnailUrl` is populated directly from `images.object_key`; it is not signed by the club read handlers. Detail responses add `website_url` and `description`, using those exact snake-case JSON keys.

## `GET /clubs`

Returns public club summaries.

- JWT authorizer: not attached.
- Handler: [`lambda/api/clubs/get.go`](../../lambda/api/clubs/get.go).
- Query: [`clubs/SELECT_clubs.sql`](../../utils/query_client/queries/clubs/SELECT_clubs.sql).
- Entity: [`models.Club`](../../database/models/club.go).
- Body and path parameters: none.

Query parameter:

| Name | Behavior |
| --- | --- |
| `verified` | Only the exact strings `true` and `TRUE` select clubs present in `verified_clubs`. Missing or any other value returns all clubs. |

`200 OK`:

```json
{
  "message": "Succesfully fetched 1 clubs",
  "clubs": [
    {
      "id": 7,
      "name": "Example Club",
      "thumbnailUrl": "clubs/7/thumbnails/example.png"
    }
  ]
}
```

An empty result is `200` with `"clubs": []`. Query failures return `500` and include the database error text in the helper error message.

## `GET /clubs/{clubId}`

Returns one public club detail record.

- JWT authorizer: not attached.
- Handler: [`lambda/api/clubs/clubId/get.go`](../../lambda/api/clubs/clubId/get.go).
- Query: [`clubs/SELECT_club.sql`](../../utils/query_client/queries/clubs/SELECT_club.sql).
- Entity: [`models.ClubDetailed`](../../database/models/club.go).
- `clubId`: passed to SQL as received; the handler does not perform integer validation.
- Query parameters and body: none.

`200 OK`:

```json
{
  "message": "Succesfully fetched club 7",
  "club": {
    "id": 7,
    "name": "Example Club",
    "thumbnailUrl": "clubs/7/thumbnails/example.png",
    "website_url": "https://club.example.edu",
    "description": "A sanitized example club."
  }
}
```

The two detail fields can be `null`. A missing club follows the handler's client-error helper and returns `400`, not `404`. Other query failures return `500`.

## `POST /clubs`

Creates a club and its `club_info` row in one database transaction.

- JWT authorizer: attached.
- Handler: [`lambda/api/clubs/post/post.go`](../../lambda/api/clubs/post/post.go).
- Queries: [`clubs/INSERT_club.sql`](../../utils/query_client/queries/clubs/INSERT_club.sql) and [`clubs/INSERT_club_info.sql`](../../utils/query_client/queries/clubs/INSERT_club_info.sql).
- Claims used: `sub` and optional `email`; [`RequireStudent`](../../utils/auth/ensure_student.go) ensures the author exists in `students`.
- Path and query parameters: none.

Request body:

```json
{
  "club": {
    "name": "Example Club",
    "website_url": "https://club.example.edu",
    "description": "A sanitized example club."
  }
}
```

`club.name` is required and non-empty. `website_url` and `description` are optional and have no URL or length validation in the handler. Club `id` and `thumbnailUrl` fields can be deserialized through the embedded model but are not used by either insert.

`200 OK`:

```json
{
  "message": "Successfully inserted club into database",
  "clubId": 7
}
```

The handler does not insert the creator into `club_members`, set an owner, or add the club to `verified_clubs`. Any Cognito-authenticated caller can reach the operation; no admin record or role flag is checked.

Handler response paths are `400` for missing handler-visible claims, student-sync errors, invalid JSON, or validation failures, and `500` for transaction failures. A duplicate club name violates the database unique constraint and follows that `500` path. API Gateway rejects a missing or invalid token before the handler.

## `GET /clubs/{clubId}/events`

Returns posted events linked to one club. The full event contract is shared with [the event API](events.md#get-events).

- JWT authorizer: not attached.
- Handler: [`lambda/api/clubs/clubId/events/get.go`](../../lambda/api/clubs/clubId/events/get.go).
- Query: [`clubs/SELECT_club_events.sql`](../../utils/query_client/queries/clubs/SELECT_club_events.sql).
- `clubId`: required by the route and must parse as an integer.
- Body: none.

Query parameters:

| Name | Default | Validation and use |
| --- | --- | --- |
| `startDate` | `1970-01-01` | `YYYY-MM-DD`; SQL uses strict `start_date > value`. |
| `endDate` | `2100-01-01` | `YYYY-MM-DD`; SQL uses strict `end_date < value`. |
| `limit` | `10` | Positive integer. |
| `page` | `0` | Non-negative integer; offset is `page * limit`. |

The query always filters `status LIKE 'posted'`. `200` returns the common `message` plus an `events` array. Invalid dates, pagination, or `clubId` return `400`; query failures return `500`.

Because the SQL filters the joined association to the requested club, each event contains only that club in `owners.owner` or `owners.associates`, even when the event has other associated clubs. The handler uses a fixed placeholder thumbnail URL and does not populate club names. It also builds its final slice from a Go map, so output order is not deterministic after SQL execution.

## `POST /clubs/{clubId}/events`

Creates a drafted event for a club. The request schema, response, authorization boundary, and current SQL failure are documented under [Event creation](events.md#post-clubsclubidevents).

The route attaches the JWT authorizer, but the handler does not verify that the caller is a member, e-board member, or owner of `{clubId}` or any associate club.

## `POST /clubs/{clubId}/members/me`

Joins the authenticated student to a club as a regular member.

- JWT authorizer: attached.
- Handler: [`lambda/api/clubs/clubId/members/me/post/post.go`](../../lambda/api/clubs/clubId/members/me/post/post.go).
- Query: [`clubs/INSERT_club_member.sql`](../../utils/query_client/queries/clubs/INSERT_club_member.sql).
- `clubId`: required and must parse as an integer.
- Query parameters and body: none.

The insert selects the caller's `students` row and the requested `clubs` row, rejects an existing composite membership through `NOT EXISTS`, and stores both `member_is_eboard` and `member_is_owner` as false.

`200 OK`:

```json
{
  "message": "Successfully joined club",
  "clubId": 7,
  "joined": true
}
```

Invalid or missing `clubId`, missing handler-visible claims, a duplicate membership, or a nonexistent club returns `400`. Student-sync and database failures return `500`.

The insert does not consult `verified_clubs`; a valid club row can be joined regardless of verification state.

## `DELETE /clubs/{clubId}/members/me`

The active API route does **not** attach the JWT authorizer, while the first handler operation requires `requestContext.authorizer.jwt.claims`.

- JWT authorizer: not attached.
- Handler: [`lambda/api/clubs/clubId/members/me/delete/delete.go`](../../lambda/api/clubs/clubId/members/me/delete/delete.go).
- Query, if reached: [`clubs/DELETE_club_member.sql`](../../utils/query_client/queries/clubs/DELETE_club_member.sql).
- Request body and query parameters: none.

Through the registered API Gateway route, the request context has no JWT authorizer data, so the handler returns `400` with an extraction error before it parses `clubId` or runs the delete. Supplying an `Authorization` header alone does not change that route context.

When the handler is invoked with a populated JWT authorizer context, it accepts an integer `clubId` and deletes only a matching membership whose `member_is_owner` flag is false. Zero affected rows returns `400`; success would be:

```json
{
  "message": "Successfully left club",
  "clubId": 7,
  "left": true
}
```

## `POST /clubs/{clubId}/thumbnails`

This unauthenticated route returns a signed S3 upload URL and does not modify the club record. See [Image API](images.md#post-clubsclubidthumbnails) and [Image upload architecture](../architecture/image-uploads.md).
