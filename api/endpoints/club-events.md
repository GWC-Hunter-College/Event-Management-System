# Club events

A club's event list, event creation, and the planned drafts list. The [event object](events.md#event-object) and [list parameters](events.md#list-parameters) are documented in events.md. Editing, publishing, and deleting are planned under [`/auth/events`](event-management.md).

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
- Has the shared [event list quirks](events.md#event-object). The Club page sends no parameters, so it gets 10 joined rows starting from 1970 and never reaches older semesters ([M9](../coverage.md#m9-list-defaults-dont-match-how-the-frontend-calls-lists)).

**Proposed changes** (Need: **Now**; screens: Club page Events tab and Manage tab, and later the GWC website):

- Return the [proposed event object](events.md#proposed-event-object), with `posted` and `cancelled` events ([Event status](events.md#event-status)), and every club linked to each event rather than only this one.
- Take the [proposed list parameters](events.md#proposed-list-parameters): `limit` 50 by default, 100 at most, counted in events. The Club page calls `when=upcoming` for its Upcoming section and `when=past` (start time descending) for past semesters.
- An unknown club returns `404` `{"error": "Club with id <clubId> not found"}`.
- Drafts stay out of this public list. The frontend reads them from [`GET /clubs/{clubId}/events/drafts`](#-get-clubsclubideventsdrafts) instead ([M1](../coverage.md#m1-drafts-in-public-lists)).

**Queries** (2): 1. [`clubs/get/EXISTS_club.sql`](../../database/queries/clubs/get/EXISTS_club.sql) (read, `404`) → 2. [`clubs/events/list/SELECT_club_events.sql`](../../database/queries/clubs/events/list/SELECT_club_events.sql) (read)

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

**Proposed contract** (Need: **Now**; screen: New event form, "SAVE DRAFT" and "POST EVENT"). Fixes [M2](../coverage.md#m2-create-event-body-timezone-associates-description) and [M11](../coverage.md#m11-create-event-extras-status-flyer-alttext-tags).

- **Auth:** 🔴 JWT + e-board or owner of `clubId`, using the [club role check](internal.md#-club-role-check). Otherwise `403`.
- **Request body:**

  ```json
  {
    "event": {
      "title": "Example Event",
      "location": "Campus",
      "rsvpLink": "https://events.example.edu/rsvp",
      "startDate": "2026-09-15T21:00:00Z",
      "endDate": "2026-09-15T23:00:00Z",
      "timezone": "America/New_York"
    },
    "description": "A sanitized example.",
    "status": "draft",
    "associates": []
  }
  ```

  | Field | Rule |
  | --- | --- |
  | `event.title` | Required and non-empty. |
  | `event.startDate`, `event.endDate` | Required ISO 8601 with an offset; `endDate` after `startDate`. The form always sends both, in UTC. |
  | `event.location` | Required when `status` is `posted`; may be empty for a draft (the form only requires it when posting). |
  | `event.timezone` | Optional IANA name; defaults to `America/New_York`, which the form assumes. |
  | `event.rsvpLink`, `description` | Optional. |
  | `status` | Optional, `draft` or `posted`; **omitted means `draft`** ([decision 1](../README.md#decisions)). The form always sends it. The design-mode mock defaults to `posted` and changes to match ([frontend changes](../coverage.md#frontend-changes)). |
  | `associates` | Optional array of club IDs, default `[]`; duplicates and `clubId` itself are dropped. No UI sends it yet (co-hosting is SOON). |
  | `flyer`, `altText`, `tags` | Not accepted. The flyer is uploaded after the event exists, through the [flyer flow](images.md#how-image-uploads-work); its alt text is stored on the flyer's `images` row ([decision 4](../README.md#decisions)), so it's sent with the [flyer confirm](images.md#-post-clubsclubideventseventidthumbnailsconfirm). Event tags wait until the UI has them. |

- **Response `200`:** unchanged, `{"message": "Successfully inserted event into database", "eventId": 42}` (the frontend reads `eventId`).
- **Errors:** `401` (API Gateway), `403` not a manager of the club, `404` unknown club, `400` validation with `"errors": [{"field", "message"}]`, `500` database failure.
- **Writes:** event, owner link, associate links, and description in one transaction.

**Queries** (5; steps 3–6 in one transaction): 1. [`clubs/get/EXISTS_club.sql`](../../database/queries/clubs/get/EXISTS_club.sql) (read, `404`) → 2. [`authorization/clubs/can_manage/IS_student_authorized_club.sql`](../../database/queries/authorization/clubs/can_manage/IS_student_authorized_club.sql) (auth, `403`) → 3. [`events/create/INSERT_event.sql`](../../database/queries/events/create/INSERT_event.sql) (write, its last insert id is the event id) → 4. [`events/create/INSERT_event_description.sql`](../../database/queries/events/create/INSERT_event_description.sql) (write, if there is a description) → 5. [`events/create/INSERT_event_club_link.sql`](../../database/queries/events/create/INSERT_event_club_link.sql) (write, the owner link, `TRUE`) → 6. [`events/create/INSERT_event_club_link.sql`](../../database/queries/events/create/INSERT_event_club_link.sql) (write, once per associate, `FALSE`, deduplicated). A bad associate id fails its foreign key and rolls everything back.

**Code:** route [`club_routes.go`](../../infrastructure/legacy/gateway/routes/club_routes.go) · handler [`clubs/clubId/events/post/post.go`](../../infrastructure/legacy/lambda/api/clubs/clubId/events/post/post.go) · SQL [`events/INSERT_event.sql`](../../infrastructure/legacy/utils/query_client/queries/events/INSERT_event.sql), [`events/INSERT_event_club_link.sql`](../../infrastructure/legacy/utils/query_client/queries/events/INSERT_event_club_link.sql), and [`events/INSERT_event_description.sql`](../../infrastructure/legacy/utils/query_client/queries/events/INSERT_event_description.sql) · [query group 12](../../database/README.md#12-create-event-draft)

## 🔴 GET `/clubs/{clubId}/events/drafts`

Lists a club's drafted events for its e-board and owners (PDF: check the caller's access, then return all drafts). Today the frontend tries to find drafts in `GET /clubs/{clubId}/events`, which never returns them.

**Auth:** 🔴 JWT + e-board or owner of the club (PDF).

**Path params:** `clubId`. Other parameters aren't defined.

**Response:** not defined.

**Proposed contract** (Need: **Now**; screens: New event step 1 "pick up a draft" panel, called once per managed club; Club page Manage tab). Fixes [M1](../coverage.md#m1-drafts-in-public-lists).

- **Auth:** 🔴 JWT + e-board or owner of `clubId`. Otherwise `403`; an unknown club is `404`.
- **Query params:** `limit` (default 50, maximum 100) and `page`, counted in events, as in the [proposed list parameters](events.md#proposed-list-parameters). `when` doesn't apply.
- **Response `200`:** the list envelope with [proposed event objects](events.md#proposed-event-object) (flyer and `imageCount`, no gallery), all with `"status": "draft"`, in `updatedAt` order, newest first:

  ```jsonc
  { "message": "Succesfully fetched 1 events", "events": [ /* event objects */ ] }
  ```

  Drafts may have no flyer (`thumbnailUrl` absent), which the step-1 panel shows as "no flyer yet".

**Status:** ⬜ Not built. The legacy `SELECT_club_events.sql` is status-parameterized, but its only handler always passes `posted`. The module's drafts query fixes the status to `draft` in the SQL, so no caller can choose it ([query group 21](../../database/README.md#21-club-draft-events)).

**Known issues:**

- The role check must run before the drafts query: drafts are for the club's e-board and owners only.

**Queries** (3): 1. [`clubs/get/EXISTS_club.sql`](../../database/queries/clubs/get/EXISTS_club.sql) (read, `404`) → 2. [`authorization/clubs/can_manage/IS_student_authorized_club.sql`](../../database/queries/authorization/clubs/can_manage/IS_student_authorized_club.sql) (auth, `403`) → 3. [`clubs/events/drafts/SELECT_club_drafts.sql`](../../database/queries/clubs/events/drafts/SELECT_club_drafts.sql) (read)

**Plan:** PDF "Endpoints Revamp", p. 17 · [query group 21](../../database/README.md#21-club-draft-events)
