# Event API

The public event reads expose only rows whose database status is `posted`. Event creation is registered under a club and writes a `drafted` event. Route registration is split between [`gateway/routes/event_routes.go`](../../gateway/routes/event_routes.go) and [`gateway/routes/club_routes.go`](../../gateway/routes/club_routes.go).

## Event response object

Read handlers combine [`models.Event`](../../database/models/event.go), an optional description, and club ownership into [`event_schema.ResponseSchema`](../../lambda/api/events/schema/schema.go):

```json
{
  "id": 42,
  "authorId": "aaaaaaaa-bbbb-cccc-dddd-eeeeeeeeeeee",
  "thumbnailId": "11111111-2222-3333-4444-555555555555",
  "title": "Example Event",
  "location": "Campus",
  "rsvpLink": "https://events.example.edu/rsvp",
  "status": "posted",
  "startDate": "2026-09-15 17:00:00",
  "endDate": "2026-09-15 19:00:00",
  "timezone": "America/New_York",
  "createdAt": "2026-08-01 12:00:00",
  "updatedAt": "2026-08-01 12:00:00",
  "description": "A sanitized example.",
  "owners": {
    "owner": {
      "id": 7,
      "name": "",
      "thumbnailUrl": "<fixed-placeholder-url>"
    },
    "associates": [
      {
        "id": 8,
        "name": "",
        "thumbnailUrl": "<fixed-placeholder-url>"
      }
    ]
  }
}
```

`thumbnailId` and `deletedAt` are omitted when null. `rsvpLink` and `description` are present with `null` when null. The response-building code does not use the image object key or club name selected by SQL: it emits only each club ID, an empty `name`, and a fixed external placeholder thumbnail URL.

## `GET /events`

Returns public posted events within a date window.

- JWT authorizer: not attached.
- Handler: [`lambda/api/events/get.go`](../../lambda/api/events/get.go).
- Query: [`events/SELECT_events.sql`](../../utils/query_client/queries/events/SELECT_events.sql).
- Entities: [`event_schema.SQLSchema`](../../lambda/api/events/schema/schema.go), [`models.Event`](../../database/models/event.go), and [`models.EventOwners`](../../database/models/event.go).
- Body and path parameters: none.

Query parameters:

| Name | Default | Validation and use |
| --- | --- | --- |
| `startDate` | `1970-01-01` | Must parse as `YYYY-MM-DD`; SQL requires `start_date > value`. |
| `endDate` | `2100-01-01` | Must parse as `YYYY-MM-DD`; SQL requires `end_date < value`. |
| `limit` | `10` | Must be a positive integer; limits joined rows. |
| `page` | `0` | Must be a non-negative integer; offset is `page * limit`. |

`200 OK`:

```json
{
  "message": "Succesfully fetched 1 events",
  "events": [
    {
      "id": 42,
      "authorId": "aaaaaaaa-bbbb-cccc-dddd-eeeeeeeeeeee",
      "title": "Example Event",
      "location": "Campus",
      "rsvpLink": null,
      "status": "posted",
      "startDate": "2026-09-15 17:00:00",
      "endDate": "2026-09-15 19:00:00",
      "timezone": "America/New_York",
      "createdAt": "2026-08-01 12:00:00",
      "updatedAt": "2026-08-01 12:00:00",
      "description": null,
      "owners": {
        "owner": {"id": 7, "name": "", "thumbnailUrl": "<fixed-placeholder-url>"},
        "associates": []
      }
    }
  ]
}
```

Invalid dates, `limit`, or `page` return `400`. Query failures return `500`. No results return `200` with an empty array.

### Grouping and pagination behavior

The SQL result has one row per event-to-club link and applies `LIMIT` and `OFFSET` before Lambda groups by event ID. A page can split one event's associations or contain fewer distinct events than `limit`. The final slice is produced by iterating a Go map, so its order is not guaranteed even though SQL orders rows by `start_date`.

The date comparisons are strict: an event starting exactly at `startDate` or ending exactly at `endDate` is excluded.

The schema permits more than one event-to-club row with `club_is_event_owner = true`. When that occurs, the handler repeatedly assigns the single `owners.owner` field and only the last owner row processed remains there.

## `GET /events/{eventId}`

Returns one public posted event assembled from up to 100 event-to-club rows.

- JWT authorizer: not attached.
- Handler: [`lambda/api/events/eventId/get.go`](../../lambda/api/events/eventId/get.go).
- Query: [`events/SELECT_events.sql`](../../utils/query_client/queries/events/SELECT_events.sql).
- `eventId`: must be non-empty, but is not parsed as an integer. SQL compares it with `e.id LIKE ?`.
- Query parameters and body: none.

The handler fixes the date window to `1970-01-01` through `2100-01-01`, filters status to `posted`, and uses `LIMIT 100 OFFSET 0`.

`200 OK` wraps the event as `event`:

```json
{
  "message": "Succesfully fetched 2 events",
  "event": {
    "id": 42,
    "authorId": "aaaaaaaa-bbbb-cccc-dddd-eeeeeeeeeeee",
    "title": "Example Event",
    "location": "Campus",
    "rsvpLink": null,
    "status": "posted",
    "startDate": "2026-09-15 17:00:00",
    "endDate": "2026-09-15 19:00:00",
    "timezone": "America/New_York",
    "createdAt": "2026-08-01 12:00:00",
    "updatedAt": "2026-08-01 12:00:00",
    "description": null,
    "owners": {
      "owner": {"id": 7, "name": "", "thumbnailUrl": "<fixed-placeholder-url>"},
      "associates": [{"id": 8, "name": "", "thumbnailUrl": "<fixed-placeholder-url>"}]
    }
  }
}
```

The success message count is the number of joined rows, so it can say more than one “events” while returning one event object. Empty `eventId` returns `400`, no matching rows returns `404`, and query failures return `500`.

## `GET /clubs/{clubId}/events`

This public filtered list uses the same query parameters and response object. Its exact club-specific behavior is documented in [Club API](clubs.md#get-clubsclubidevents).

## `GET /me/events`

This JWT-protected list selects events connected to the caller's club memberships. See [Student API](students.md#get-meevents).

## `POST /clubs/{clubId}/events`

Accepts an event draft and associated club IDs.

- JWT authorizer: attached.
- Handler: [`lambda/api/clubs/clubId/events/post/post.go`](../../lambda/api/clubs/clubId/events/post/post.go).
- Queries: [`events/INSERT_event.sql`](../../utils/query_client/queries/events/INSERT_event.sql), [`events/INSERT_event_club_link.sql`](../../utils/query_client/queries/events/INSERT_event_club_link.sql), and [`events/INSERT_event_description.sql`](../../utils/query_client/queries/events/INSERT_event_description.sql).
- `clubId`: required and must parse as an integer.
- Claims used: `sub` becomes `events.fk_author_id`; optional `email` is used only by student synchronization.
- Query parameters: none.

Request body:

```json
{
  "event": {
    "title": "Example Event",
    "location": "Campus",
    "rsvpLink": "https://events.example.edu/rsvp",
    "startDate": "2026-09-15 17:00:00",
    "endDate": "2026-09-15 19:00:00",
    "timezone": "America/New_York"
  },
  "description": "A sanitized example.",
  "associates": [8, 9]
}
```

The validator requires non-empty `event.title`, `event.location`, `event.startDate`, `event.endDate`, `event.timezone`, and `description`. `associates` must be present as a non-null slice; there is no element validation. The handler does not parse dates, validate chronological order, verify an IANA timezone, validate the RSVP URL, or check that associate IDs exist before executing SQL. Input `authorId`, `status`, timestamps, ID, and thumbnail ID are not used by the insert.

If all inserts complete, the handler returns `200`:

```json
{
  "message": "Successfully inserted event into database",
  "eventId": 42
}
```

The owner link uses `{clubId}` with `club_is_event_owner = true`; each distinct associate ID other than `{clubId}` gets a false owner flag. The insert hard-codes status `drafted`, so public reads, which filter `posted`, do not expose a newly created row.

### Authorization actually enforced

The route proves only that the caller has a valid Cognito token. The handler ensures a `students` row exists but does not call either club- or event-authorization helper. It does not require membership, `member_is_eboard`, `member_is_owner`, or an admin record for the owner or associate clubs.

### Current SQL execution result

The current [`INSERT_event.sql`](../../utils/query_client/queries/events/INSERT_event.sql) lists `rsvp_link` among its columns but omits its value expression, while the handler supplies an RSVP argument. The statement therefore fails and the active handler returns `500` from the first insert rather than the success response above. In addition, the description insert contains a trailing comma in its column list; if execution reaches that transaction, it fails and rolls back the link-and-description transaction.

The initial event insert occurs before the link-and-description transaction. Consequently, a failure in that later transaction does not roll back the event row.

Other handler response paths are `400` for missing handler-visible claims, invalid `clubId`, invalid JSON, or validation failures, and `500` for student synchronization, insert, last-ID, or link transaction failures.

## Event images

Gallery and thumbnail paths are documented in [Image API](images.md). Their route registrations do not attach the Cognito authorizer.
