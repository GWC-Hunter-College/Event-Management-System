# Events

Public reads of **posted** events. Drafts and deleted events never appear here; the manager reads for them are the planned (⬜) [`/auth/events`](event-management.md) routes. A club's event list and event creation are in [club-events.md](club-events.md), and event images are in [images.md](images.md).

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

## Proposed event object

**Proposed.** Need: **Now**. Updated for [decisions](../README.md#decisions) 3 and 4 (2026-09-29). The shape every event read would return so the frontend's existing screens work (Home, Events, Club, Event, My Clubs). It keeps today's field names wherever the frontend's normalizer (`fromJsonEvent` in the frontend's `src/types/events.ts`) already reads them, and only adds or fills fields. See [coverage.md](../coverage.md#contract-mismatches) for the mismatch behind each change.

```json
{
  "id": 42,
  "authorId": "aaaaaaaa-bbbb-cccc-dddd-eeeeeeeeeeee",
  "thumbnailId": "11111111-2222-3333-4444-555555555555",
  "thumbnailUrl": "<readable-flyer-url>",
  "altText": "Poster for Example Event",
  "title": "Example Event",
  "location": "Campus",
  "rsvpLink": "https://events.example.edu/rsvp",
  "status": "posted",
  "startDate": "2026-09-15T21:00:00Z",
  "endDate": "2026-09-15T23:00:00Z",
  "timezone": "America/New_York",
  "createdAt": "2026-08-01T16:00:00Z",
  "updatedAt": "2026-08-01T16:00:00Z",
  "description": "A sanitized example.",
  "imageCount": 1,
  "images": ["<readable-image-url>"],
  "owners": {
    "owner": { "id": 7, "name": "Example Club", "thumbnailUrl": "<readable-logo-url>" },
    "associates": [{ "id": 8, "name": "Other Club", "thumbnailUrl": "<readable-logo-url>" }]
  }
}
```

| Change from today | Why | Mismatch |
| --- | --- | --- |
| `thumbnailUrl` (new, top level): a readable URL for the flyer, or absent when there's none. | The frontend maps it to `flyer`; today there's no flyer field at all. | M6 |
| `owners.*.thumbnailUrl`: the club's real logo URL instead of the fixed placeholder. | Host logos on every event list. | M6 |
| `owners.*.name`: the club's name instead of `""`. | Removes the frontend's lookup through `GET /clubs?verified=true`, which misses unverified clubs. | M8 |
| `imageCount` (new, every read): the number of gallery images, flyer excluded. | The Club page's past-event rows show "N PHOTOS"; the Event page's pill and "VIEW ALL N PHOTOS" can use it too. | — |
| `images` (new, **`GET /events/{eventId}` only**): readable URLs of gallery images, flyer excluded, in upload order; `[]` when none. Lists leave it out. | The Event page gallery reads `event.images`. | — |
| `altText` (new): the flyer's alt text, read from the flyer's `images` row; absent when there's no flyer or no alt text. | Collected by the New event form. Needs a new `images` column ([schema needs](../coverage.md#schema-needs)). | M11 |
| `status`: one of the API values in [Event status](#event-status). | The frontend's vocabulary. | M10 |
| `startDate`, `endDate`, `createdAt`, `updatedAt`: ISO 8601 in UTC with a `Z` offset. | The frontend parses offset-less text as browser-local time. | M14 |

Unchanged: `id`, `authorId`, `thumbnailId`, `title`, `location`, `rsvpLink`, `timezone`, `description`, and the `owners` structure. Lists also change how they page; see [proposed list parameters](#proposed-list-parameters).

Lists carry the flyer (`thumbnailUrl`) and `imageCount`, not the gallery ([decision 3](../README.md#decisions)). The frontend's `PastEventRow` photo count reads `images.length` today, so it has to switch to `imageCount` ([frontend changes](../coverage.md#frontend-changes)).

## Proposed list parameters

**Proposed.** Need: **Now**. For `GET /events`, `GET /clubs/{clubId}/events`, `GET /me/events`, and the [drafts list](club-events.md#-get-clubsclubideventsdrafts). Follows [decision 3](../README.md#decisions).

| Query param | Default | Rule |
| --- | --- | --- |
| `when` | none | New. `upcoming`: events that haven't ended, ordered by start time ascending. `past`: events that have ended, ordered by start time descending. Anything else is `400`. |
| `startDate`, `endDate` | none | Optional. A date (`YYYY-MM-DD`) covers the whole day in `America/New_York`: `startDate=2026-10-01` starts at 00:00 on October 1, and `endDate=2026-10-31` ends just before 00:00 on November 1. A full ISO 8601 timestamp with an offset is taken as given; `endDate` is always exclusive. An event matches when it overlaps the range, so a multi-day event that started before `startDate` is included ([decision 10](../README.md#decisions)). Without `when`, results are ordered by start time ascending. Anything else is `400`. |
| `limit` | `50` | Counts events, not joined rows. `1`–`100`; anything else is `400` `Invalid limit. Must be between 1 and 100.` |
| `page` | `0` | As today, counted in events. |

The Club page needs two calls: `when=upcoming` for its Upcoming section and `when=past` (paging for older semesters) for past events. Home, Events, and My Clubs use `when=upcoming`. `when` is the proposed parameter name; the decision fixes the ordering, not the name.

## Event status

The deployed (legacy) database stores `events.status` as `drafted`, `posted`, or `archived`. The frontend uses `draft`, `posted`, and `cancelled`. The [baseline schema](../../database/migrations/schema/2026_09_29_baseline_up.sql) stores the API's words directly. **Proposed** API vocabulary (Need: **Now**):

| API value | Stored as | Who sees it | Meaning |
| --- | --- | --- | --- |
| `draft` | `draft` | The club's e-board and owners, through the [drafts list](club-events.md#-get-clubsclubideventsdrafts) and [`GET /auth/events/{eventId}`](event-management.md#-get-autheventseventid). | Not published. The default when an event is created. |
| `posted` | `posted` | Everyone. | Published. |
| `cancelled` | `cancelled` | Everyone. Stays in public lists with its details, so people who planned to go find out. | Called off. Set with [`PATCH`](event-management.md#-patch-autheventseventid). |

There's no `archived` status ([decision 2](../README.md#decisions)). Deleting an event with [`DELETE`](event-management.md#-delete-autheventseventid) is a soft delete: it sets `deleted_at`, and from then on **every** read, public or manager, treats the event as not found. For 30 days a deleted event can be [restored](event-management.md#-post-autheventseventidrestore); after that, an admin's [purge](admins.md#-post-adminspurge) removes it for good ([decisions 8 and 9](../README.md#decisions)).

Public reads (`GET /events`, `GET /events/{eventId}`, `GET /clubs/{clubId}/events`, `GET /me/events`) return `posted` and `cancelled`, and never deleted events.

**Allowed transitions** ([decision 2](../README.md#decisions), 2026-09-29):

| From | To | How |
| --- | --- | --- |
| `draft` | `posted` | [`PATCH`](event-management.md#-patch-autheventseventid) with `"status": "posted"` |
| `posted` | `cancelled` | `PATCH` with `"status": "cancelled"` |
| `cancelled` | `posted` | `PATCH` with `"status": "posted"` |

Nothing goes back to `draft`. Saving a draft's fields without a `status` isn't a transition. Any other change is `400`. The SQL enforces the table itself: each status `UPDATE` lists its allowed starting statuses, so a disallowed change affects no rows ([queries](../../database/README.md#transactions)). A deleted event can't change status at all.

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
- The frontend calls it with no parameters and expects every upcoming event; the defaults return 10 joined rows starting from 1970 ([M9](../coverage.md#m9-list-defaults-dont-match-how-the-frontend-calls-lists)).

**Proposed changes** (Need: **Now**; screens: Home, Events; later the GWC website):

- Return the [proposed event object](#proposed-event-object), with `posted` and `cancelled` events ([Event status](#event-status)).
- Take the [proposed list parameters](#proposed-list-parameters): `limit` defaults to 50 (maximum 100) and counts events; `when=upcoming` orders by start ascending, `when=past` descending.
- Unchanged: access 🟢, no role, and the `{"message", "events"}` envelope. The frontend sends `when=upcoming`.

**Queries** (1): 1. [`events/read/SELECT_events.sql`](../../database/queries/events/read/SELECT_events.sql) (read)

**Code:** route [`event_routes.go`](../../infrastructure/legacy/gateway/routes/event_routes.go) · handler [`events/get.go`](../../infrastructure/legacy/lambda/api/events/get.go) · SQL [`events/SELECT_events.sql`](../../infrastructure/legacy/utils/query_client/queries/events/SELECT_events.sql) · [query group 11](../../database/README.md#11-public-and-composite-event-read)

## 🟢 GET `/events/{eventId}`

Returns one posted event with its description and linked clubs. The frontend's Event page (`/event/:eventId`) calls it.

**Auth:** 🟢 Public. The hard-coded `posted` filter hides drafted and archived events. (That describes the deployed handler; the proposed read also returns `cancelled` and never returns deleted events.)

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

**Proposed changes** (Need: **Now**; screen: Event page): return the [proposed event object](#proposed-event-object), including `images`, the full gallery (this is the only read that returns it), and return `cancelled` events as well as `posted` ones; anything else, including drafts, stays `404`. Parse `eventId` as an integer (`400` otherwise) instead of matching with `LIKE`. Access stays 🟢 with no role; the envelope stays `{"message", "event"}`.

**Covers two PDF subresources:** the planning PDF lists `GET /events/{eventId}/description` and `GET /events/{eventId}/clubs` as separate public routes. This response already includes `description` and `owners`, so the design folds both into this endpoint instead of giving them routes of their own. Hard-coded stubs for both paths exist only in the commented-out [`StubLambdaStack`](../../infrastructure/legacy/internal/stack/stubLambda.go): [`description/get.go`](../../infrastructure/legacy/stub/lambda/events/eventId/description/get.go) (with stale SQL in [`description.sql`](../../infrastructure/legacy/stub/lambda/events/eventId/description/description.sql)) and [`clubs/get.go`](../../infrastructure/legacy/stub/lambda/events/eventId/clubs/get.go). If a client ever needs those paths, add a thin adapter over this read rather than a second copy of the query.

**Queries** (2): 1. [`events/read/SELECT_event.sql`](../../database/queries/events/read/SELECT_event.sql) (read, `public_only` `TRUE`; no row is `404`) → 2. [`events/images/list/SELECT_event_images.sql`](../../database/queries/events/images/list/SELECT_event_images.sql) (read, `images`)

**Code:** route [`event_routes.go`](../../infrastructure/legacy/gateway/routes/event_routes.go) · handler [`events/eventId/get.go`](../../infrastructure/legacy/lambda/api/events/eventId/get.go) · SQL [`events/SELECT_events.sql`](../../infrastructure/legacy/utils/query_client/queries/events/SELECT_events.sql) · [query group 11](../../database/README.md#11-public-and-composite-event-read)

## 🟢 GET `/events/{eventId}/images`

Public image gallery for a posted event. Per the PDF, it first checks that the event is public, then returns all of the event's images. No client calls it yet.

**Auth:** 🟢 Public (PDF). It must return images only when the event is `posted`.

**Path params:** `eventId`.

**Response:** not defined. The deployed club-scoped read declares `{"images": [{"imageId", "mimetype", "sourceUrl"}]}` ([details](images.md#-get-clubsclubideventseventidimages)).

**Proposed** (Need: **Later**): reuse that declared shape, with `sourceUrl` a readable URL, for `posted` and `cancelled` events; `404` otherwise. The frontend's Event page doesn't need this route while the [proposed event object](#proposed-event-object) carries `images`; it becomes useful if lists drop `images` or a gallery pages through many photos.

**Status:** ⬜ Not built. An inactive, hard-coded stub exists in [`stub/lambda/events/eventId/images/get.go`](../../infrastructure/legacy/stub/lambda/events/eventId/images/get.go); don't treat its response as a contract. The deployed [`GET /clubs/{clubId}/events/{eventId}/images`](images.md#-get-clubsclubideventseventidimages) is broken (missing SQL) and checks neither posted status nor roles.

**Needs:** a storage adapter to sign the URLs. The gallery query and the visibility check exist ([query group 16](../../database/README.md#16-event-image-list)).

**Queries** (2): 1. [`events/read/SELECT_event_status.sql`](../../database/queries/events/read/SELECT_event_status.sql) (read, `404` unless `posted` or `cancelled`) → 2. [`events/images/list/SELECT_event_images.sql`](../../database/queries/events/images/list/SELECT_event_images.sql) (read)

**Plan:** PDF "Endpoints Revamp", p. 17.
