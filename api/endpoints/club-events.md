# Club events

A club's event list, event creation, and the planned drafts list. The [event object](events.md#event-object) and [list parameters](events.md#list-parameters) are documented in events.md. Editing, publishing, and archiving are planned under [`/auth/events`](event-management.md).

Routes are registered in [`club_routes.go`](../../infrastructure/legacy/gateway/routes/club_routes.go). Examples follow the [implementation's event docs](../../infrastructure/legacy/docs/api/events.md), with made-up values.

## 🟢 GET `/clubs/{clubId}/events`

Lists posted events linked to one club. The frontend's Club page (`/club/:clubId`) calls it. The New event pages also call it to find drafts, which it never returns (see known issues).

**Auth:** 🟢 Public. The handler hard-codes `posted`.

**Path params:** `clubId`, a required integer.

**Query params:** the [list parameters](events.md#list-parameters).

**Response `200`** (each item is an [event object](events.md#event-object)):

```jsonc
{
  "message": "Succesfully fetched 1 events",
  "events": [ /* event objects */ ]
}
```

An unknown club returns `200` with `"events": []`.

**Errors:**

| Status | When |
| --- | --- |
| `400` | Invalid list parameter, or `{"error": "invalid clubId: <value>"}`. |
| `500` | `{"error": "Could not fetch events: <db error>"}` |

**Status:** ✅ Implemented.

**Known issues:**

- Each event shows only this club, as `owners.owner` or in `owners.associates`, even when other clubs are linked.
- The frontend expects this list to include drafts (for its Manage tab, draft picker, and draft prefill) and a `404` for unknown clubs. Because the backend returns posted events only, those draft features never see anything.
- Has the shared [event list quirks](events.md#event-object).

**Code:** route [`club_routes.go`](../../infrastructure/legacy/gateway/routes/club_routes.go) · handler [`clubs/clubId/events/get.go`](../../infrastructure/legacy/lambda/api/clubs/clubId/events/get.go) · SQL [`clubs/SELECT_club_events.sql`](../../infrastructure/legacy/utils/query_client/queries/clubs/SELECT_club_events.sql) · [query group 10](../../database/README.md#10-club-event-list)

## 🔴 POST `/clubs/{clubId}/events`

Creates a draft event, links it to the club as owner and to any associate clubs, and stores its description. The frontend's New event form calls it for both Save Draft and Post Event. **It's broken:** the SQL is invalid, so every call ends in `500`.

**Auth:** 🔴 JWT. The PDF requires e-board or owner of `clubId`, but **no role check runs**: any signed-in user can create an event for any club and link any associates.

**Path params:** `clubId`, a required integer.

**Request body** (the handler's `RequestBodySchema`, with [`models.Event`](../../infrastructure/legacy/database/models/event.go) under `event`):

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

| Field | Rule |
| --- | --- |
| `event.title`, `event.location`, `event.startDate`, `event.endDate`, `event.timezone` | Required and non-empty. Date format, date order, and the timezone name aren't checked. |
| `event.rsvpLink` | Optional; not validated. |
| `description` | Required. |
| `associates` | Required array of club IDs (can be empty). IDs aren't checked; an entry equal to `clubId` is skipped. |
| Other `event` fields | `id`, `authorId`, `status`, `thumbnailId`, and timestamps are ignored. The author is the caller, and status is always `drafted`. |

**Response `200`** (what the handler returns once the SQL works):

```json
{ "message": "Successfully inserted event into database", "eventId": 42 }
```

**Errors:**

| Status | When |
| --- | --- |
| `401` | Token missing or invalid (from API Gateway). |
| `400` | No claims or `sub`; bad `clubId`; the body isn't JSON; or validation failed (adds `"errors": [{"field", "message"}]`). |
| `500` | Every valid request today, from the first insert (see Status). Also returned for `RequireStudent`, last-insert-ID, or link-transaction failures. |

**Status:** 🟨 Broken: SQL invalid. `INSERT_event.sql` lists 11 columns but supplies 10 values, with no value for `rsvp_link`, and the handler passes 7 arguments for its 6 placeholders. `INSERT_event_description.sql` has a trailing comma in its column list.

**Known issues:**

- No role check (see Auth).
- The event insert commits before the link-and-description transaction starts, so a later failure leaves an orphan event row.
- New events are always `drafted`, and nothing can publish them yet; [`PATCH /auth/events/{eventId}`](event-management.md#-patch-autheventseventid) is planned.
- The frontend sends `status`, `flyer`, and `altText`, leaves out `timezone` and `associates`, and may leave out `description`. Even with working SQL, that body fails validation with `400`.
- A repeated ID in `associates` would violate the `events_to_clubs` primary key.

**Code:** route [`club_routes.go`](../../infrastructure/legacy/gateway/routes/club_routes.go) · handler [`clubs/clubId/events/post/post.go`](../../infrastructure/legacy/lambda/api/clubs/clubId/events/post/post.go) · SQL [`events/INSERT_event.sql`](../../infrastructure/legacy/utils/query_client/queries/events/INSERT_event.sql), [`events/INSERT_event_club_link.sql`](../../infrastructure/legacy/utils/query_client/queries/events/INSERT_event_club_link.sql), and [`events/INSERT_event_description.sql`](../../infrastructure/legacy/utils/query_client/queries/events/INSERT_event_description.sql) · [query group 12](../../database/README.md#12-create-event-draft)

## 🔴 GET `/clubs/{clubId}/events/drafts`

Lists a club's drafted events for its e-board and owners (PDF: check the caller's access, then return all drafts). Today the frontend tries to find drafts in `GET /clubs/{clubId}/events`, which never returns them.

**Auth:** 🔴 JWT + e-board or owner of the club (PDF).

**Path params:** `clubId`. Other parameters aren't defined.

**Response:** not defined. **Proposed:** return the same [event objects](events.md#event-object), since the club event query already takes a status parameter.

**Status:** ⬜ Not built. `SELECT_club_events.sql` is status-parameterized, but its only handler always passes `posted`.

**Known issues:**

- Don't let callers choose the status filter until the role check is in place.

**Plan:** PDF "Endpoints Revamp", p. 17 · [query group 21](../../database/README.md#21-club-draft-events)
