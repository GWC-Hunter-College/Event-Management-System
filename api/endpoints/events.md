# Events

Public reads of **posted** events. Drafted and archived events never appear here; managers will read those through the planned [`/auth/events`](event-management.md) routes. A club's event list and event creation are in [club-events.md](club-events.md), and event images are in [images.md](images.md).

Routes are registered in [`event_routes.go`](../../infrastructure/legacy/gateway/routes/event_routes.go). Response shapes come from the Go structs named below; example values are made up, following the sanitized examples in the [implementation's event docs](../../infrastructure/legacy/docs/api/events.md).

## Event object

Every event read (`GET /events`, `GET /events/{eventId}`, `GET /clubs/{clubId}/events`, `GET /me/events`) returns events in this shape, built from [`event_schema.ResponseSchema`](../../infrastructure/legacy/lambda/api/events/schema/schema.go), which embeds [`models.Event`](../../infrastructure/legacy/database/models/event.go):

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
    "owner": { "id": 7, "name": "", "thumbnailUrl": "<fixed-placeholder-url>" },
    "associates": [{ "id": 8, "name": "", "thumbnailUrl": "<fixed-placeholder-url>" }]
  }
}
```

| Field | Notes |
| --- | --- |
| `startDate`, `endDate`, `createdAt`, `updatedAt` | MySQL text in `YYYY-MM-DD HH:MM:SS` form, not ISO 8601. |
| `thumbnailId`, `deletedAt` | Omitted when null. |
| `rsvpLink`, `description` | Present as `null` when unset. |
| `owners.owner` | The linked club with `club_is_event_owner = true`. If that club isn't in the query result, it's the empty object `{"name": ""}`. |
| `owners.*.name` | Always `""`: the handlers don't copy the club name. |
| `owners.*.thumbnailUrl` | A fixed external placeholder image, not the club's logo. |

Quirks shared by every event list:

- SQL applies `LIMIT`/`OFFSET` to event-to-club rows, not to events, so a page can hold fewer than `limit` events or only some of an event's clubs.
- The handler groups rows in a Go map, so the JSON array isn't in the SQL's `start_date` order.
- Date bounds are strict: an event starting exactly at `startDate` or ending exactly at `endDate` is excluded.
- If an event has more than one link with `club_is_event_owner = true`, the last owner row processed wins.

## List parameters

`GET /events`, `GET /clubs/{clubId}/events`, and `GET /me/events` accept these optional query parameters. Every list is also filtered to `status = 'posted'`.

| Query param | Default | Rule |
| --- | --- | --- |
| `startDate` | `1970-01-01` | `YYYY-MM-DD`. Keeps events with `start_date > startDate`. |
| `endDate` | `2100-01-01` | `YYYY-MM-DD`. Keeps events with `end_date < endDate`. |
| `limit` | `10` | Positive integer. Counts joined rows, not events. |
| `page` | `0` | Non-negative integer. Offset is `page * limit`. |

An invalid value returns `400` with one of these `error` strings: `Invalid startDate format. Use YYYY-MM-DD.`, `Invalid endDate format. Use YYYY-MM-DD.`, `Invalid limit. Must be a positive integer.`, or `Invalid page number. Must be a non-negative integer.`

## 🟢 GET `/events`

Lists posted events in a date window. The frontend's Home (`/`) and Events (`/events`) pages call it. The GWC-Website is expected to read it too, but makes no API calls yet.

**Auth:** 🟢 Public. No role is needed; the handler hard-codes `posted`.

**Query params:** the [list parameters](#list-parameters).

**Response `200`** (each item is an [event object](#event-object)):

```jsonc
{
  "message": "Succesfully fetched 1 events",
  "events": [ /* event objects */ ]
}
```

No matches returns `200` with `"events": []`.

**Errors:**

| Status | When |
| --- | --- |
| `400` | Invalid list parameter (see [list parameters](#list-parameters)). |
| `500` | Query failed: `{"error": "Could not fetch events: <db error>"}`. |

**Status:** ✅ Implemented.

**Known issues:**

- The frontend's [`docs/api.md`](https://github.com/GWC-Hunter-College/Hunter-College-Clubs-Frontend/blob/staging/docs/api.md) says this list includes drafts and filters them out client-side. The backend returns posted events only.
- Has the shared [event list quirks](#event-object): joined-row paging, unordered output, and placeholder club fields.

**Code:** route [`event_routes.go`](../../infrastructure/legacy/gateway/routes/event_routes.go) · handler [`events/get.go`](../../infrastructure/legacy/lambda/api/events/get.go) · SQL [`events/SELECT_events.sql`](../../infrastructure/legacy/utils/query_client/queries/events/SELECT_events.sql) · [query group 11](../../database/README.md#11-public-and-composite-event-read)

## 🟢 GET `/events/{eventId}`

Returns one posted event with its description and linked clubs. The frontend's Event page (`/event/:eventId`) calls it.

**Auth:** 🟢 Public. The hard-coded `posted` filter hides drafted and archived events.

**Path params:** `eventId`, required and non-empty. It isn't parsed as an integer (see known issues).

**Response `200`** (an [event object](#event-object) under `event`):

```jsonc
{
  "message": "Succesfully fetched 2 events", // counts joined rows, not events
  "event": { /* event object */ }
}
```

**Errors:**

| Status | Body |
| --- | --- |
| `400` | `{"error": "Missing eventId path parameter"}` |
| `404` | `{"error": "No events found with eventId <eventId>"}` |
| `500` | `{"error": "Could not fetch events: <db error>"}` |

**Status:** ✅ Implemented.

**Known issues:**

- SQL matches `eventId` with `LIKE`, so `%` or `_` in the path act as wildcards and can merge several events into one response.
- Reads at most 100 joined rows, with a fixed window of `1970-01-01` to `2100-01-01`.
- Public reads don't filter on `deleted_at`.

**Covers two PDF subresources:** the planning PDF lists `GET /events/{eventId}/description` and `GET /events/{eventId}/clubs` as separate public routes. This response already includes `description` and `owners`, so the design folds both into this endpoint instead of giving them routes of their own. Hard-coded stubs for both paths exist only in the commented-out [`StubLambdaStack`](../../infrastructure/legacy/internal/stack/stubLambda.go): [`description/get.go`](../../infrastructure/legacy/stub/lambda/events/eventId/description/get.go) (with stale SQL in [`description.sql`](../../infrastructure/legacy/stub/lambda/events/eventId/description/description.sql)) and [`clubs/get.go`](../../infrastructure/legacy/stub/lambda/events/eventId/clubs/get.go). If a client ever needs those paths, add a thin adapter over this read rather than a second copy of the query.

**Code:** route [`event_routes.go`](../../infrastructure/legacy/gateway/routes/event_routes.go) · handler [`events/eventId/get.go`](../../infrastructure/legacy/lambda/api/events/eventId/get.go) · SQL [`events/SELECT_events.sql`](../../infrastructure/legacy/utils/query_client/queries/events/SELECT_events.sql) · [query group 11](../../database/README.md#11-public-and-composite-event-read)

## 🟢 GET `/events/{eventId}/images`

Public image gallery for a posted event. Per the PDF, it first checks that the event is public, then returns all of the event's images. No client calls it yet.

**Auth:** 🟢 Public (PDF). It must return images only when the event is `posted`.

**Path params:** `eventId`.

**Response:** not defined. The deployed club-scoped read declares `{"images": [{"imageId", "mimetype", "sourceUrl"}]}` ([details](images.md#-get-clubsclubideventseventidimages)); reusing that is an option, not a decision.

**Status:** ⬜ Not built. An inactive, hard-coded stub exists in [`stub/lambda/events/eventId/images/get.go`](../../infrastructure/legacy/stub/lambda/events/eventId/images/get.go); don't treat its response as a contract. The deployed [`GET /clubs/{clubId}/events/{eventId}/images`](images.md#-get-clubsclubideventseventidimages) is broken (missing SQL) and checks neither posted status nor roles.

**Needs:** the event-image list query, which is the missing `images`/`event_images` join ([query group 16](../../database/README.md#16-event-image-list)), plus a posted-visibility check and a storage adapter.

**Plan:** PDF "Endpoints Revamp", p. 17.
