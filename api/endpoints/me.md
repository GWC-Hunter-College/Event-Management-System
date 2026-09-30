# Me

Reads about the signed-in caller. Every route here needs a Cognito JWT, and the handler uses the token's `sub` as the student ID. Before reading anything, each built handler calls [`RequireStudent`](internal.md#-requirestudent-request-time-student-sync), which creates the caller's `students` row if it's missing.

Routes are registered in [`student_routes.go`](../../infrastructure/legacy/gateway/routes/student_routes.go). Response shapes come from the Go structs named below; example values are made up, following the [implementation's student docs](../../infrastructure/legacy/docs/api/students.md).

**Errors shared by every built `/me` route:**

| Status | Body | When |
| --- | --- | --- |
| `401` | From API Gateway | Token missing, expired, or invalid. The Lambda doesn't run. |
| `400` | `{"error": "missing sub in JWT claims", "code": "ERR_NO_SUB"}` | The claims have no `sub`. |
| `500` | `{"error": "failed to ensure student", "code": "ERR_REQUIRE_STUDENT", "detail": "<db error>"}` | `RequireStudent` failed. |

## 🔴 GET `/me`

Returns the caller's student record. No frontend page calls it; the frontend reads the email from the OIDC profile instead.

**Auth:** 🔴 JWT. Scoped to the caller's `sub`; no role is needed.

**Request:** no parameters or body.

**Response `200`** (shape: [`models.Student`](../../infrastructure/legacy/database/models/student.go)):

```json
{
  "message": "Successfully fetched student 11111111-2222-3333-4444-555555555555",
  "student": {
    "id": "11111111-2222-3333-4444-555555555555",
    "email": "student@example.edu"
  }
}
```

**Errors:** the shared errors above, plus:

| Status | Body | When |
| --- | --- | --- |
| `400` | `{"error": "Student with id <sub> not found", "code": "ERR_NO_SUB"}` | No student row. The source comment says 404, but the code uses the 400 helper. |
| `500` | `{"error": "failed to query student", "code": "ERR_DB_FAILURE", "detail": "<db error>"}` | The query failed. |

**Status:** ✅ Implemented.

**Known issues:**

- `students.email` is nullable, but `models.Student.Email` is a Go `string`. A row with a `NULL` email fails to scan, and this route returns `500`. `RequireStudent` creates exactly that kind of row when the token has no `email` claim (true of the access tokens the frontend sends; see [Auth](../README.md#auth)) and the Cognito trigger hadn't already written the row.

**Proposed changes** (Need: **Later**; no screen calls this route: the user menu reads the email from the ID token, and there's no admin page). Fixes [M5](../coverage.md#m5-the-access-token-has-no-email-claim) and gives a future admin page what it needs:

```json
{
  "message": "Successfully fetched student 11111111-2222-3333-4444-555555555555",
  "student": {
    "id": "11111111-2222-3333-4444-555555555555",
    "email": null,
    "isAdmin": false
  }
}
```

`email` may be `null`. `isAdmin` (new) is `true` when the caller has an `admins` row, so an admin page can decide whether to show itself without calling an admin-only route and handling `403`.

**Queries** (2): 1. [`me/get/SELECT_student_by_sub.sql`](../../database/queries/me/get/SELECT_student_by_sub.sql) (read, `404` when there is no row) → 2. [`authorization/admins/is_admin/IS_admin.sql`](../../database/queries/authorization/admins/is_admin/IS_admin.sql) (read, the proposed `isAdmin`)

**Code:** route [`student_routes.go`](../../infrastructure/legacy/gateway/routes/student_routes.go) · handler [`me/get.go`](../../infrastructure/legacy/lambda/api/me/get.go) · SQL [`students/SELECT_student_by_sub.sql`](../../infrastructure/legacy/utils/query_client/queries/students/SELECT_student_by_sub.sql) · query groups [1](../../database/README.md#1-student-existence-and-upsert) and [2](../../database/README.md#2-current-student)

## 🔴 GET `/me/clubs`

Lists every club the caller has joined, with the caller's role in each. The frontend calls it wherever it needs the viewer's role: the Club, My Clubs, Create, New event (step 1 and form), and Event pages.

**Auth:** 🔴 JWT. Scoped to the caller's memberships; no role is needed.

**Request:** no parameters or body.

**Response `200`** (shape: [`models.ClubWithRole`](../../infrastructure/legacy/database/models/club.go)):

```json
{
  "message": "Successfully fetched student 11111111-2222-3333-4444-555555555555",
  "clubs": [
    {
      "id": 7,
      "name": "Example Club",
      "thumbnailUrl": "clubs/7/thumbnails/example.png",
      "role": "eboard"
    }
  ]
}
```

`role` is `owner` when `member_is_owner = 1`, otherwise `eboard` when `member_is_eboard = 1`, otherwise `member`. With no memberships, `clubs` is `[]`.

**Errors:** the shared errors above, plus `500` `{"error": "failed to query student", "code": "ERR_DB_FAILURE", "detail": "<db error>"}` when the club query fails.

**Status:** ✅ Implemented.

**Known issues:**

- `thumbnailUrl` holds the logo's S3 object key, not a signed or public URL. It's omitted when the club has no logo.
- The schema doesn't make the two role flags mutually exclusive; `owner` wins.

**Proposed changes** (Need: **Now**; screens: My Clubs cards, and every page that derives the viewer's role): `thumbnailUrl` becomes a readable URL ([proposed club object](clubs.md#proposed-club-object)). `role` and the envelope are unchanged; in the baseline schema `role` is read straight from `club_members.role` ([decision 11](../README.md#decisions)). Clubs are ordered by name.

**Queries** (1): 1. [`me/clubs/list/SELECT_student_clubs.sql`](../../database/queries/me/clubs/list/SELECT_student_clubs.sql) (read)

**Code:** route [`student_routes.go`](../../infrastructure/legacy/gateway/routes/student_routes.go) · handler [`me/clubs/get.go`](../../infrastructure/legacy/lambda/api/me/clubs/get.go) · SQL [`students/SELECT_student_clubs.sql`](../../infrastructure/legacy/utils/query_client/queries/students/SELECT_student_clubs.sql) · [query group 3](../../database/README.md#3-current-students-clubs-and-roles)

## 🔴 GET `/me/events`

Lists posted events linked to any club the caller has joined. The frontend's My Clubs page (`/my-clubs`) calls it. The planning PDF calls it `GET /me/clubs/events`.

**Auth:** 🔴 JWT. Results are filtered through the caller's `club_members` rows; no role is needed.

**Query params:** the shared [list parameters](events.md#list-parameters) (`startDate`, `endDate`, `limit`, `page`).

**Response `200`** (each item is an [event object](events.md#event-object)):

```jsonc
{
  "message": "Succesfully fetched 1 events of student 11111111-2222-3333-4444-555555555555",
  "events": [ /* event objects */ ]
}
```

**Errors:** the shared errors above, plus `400` for an invalid list parameter and `500` `{"error": "Could not fetch events: <db error>"}` when the query fails.

**Status:** ✅ Implemented.

**Known issues:**

- Each event lists only the clubs the caller joined. If the caller joined an associate club but not the owner club, `owners.owner` is the empty object.
- Has the shared [event list quirks](events.md#event-object): joined-row paging and unordered output.

**Proposed changes** (Need: **Now**; screen: My Clubs stat cards and agenda): return the [proposed event object](events.md#proposed-event-object), including `cancelled` events, with every linked club (not only the ones the caller joined), and the [proposed list parameters](events.md#proposed-list-parameters): `limit` 50 by default and 100 at most, counted in events; My Clubs calls `when=upcoming` (start time ascending).

**PDF path:** the planning PDF writes this route as `GET /me/clubs/events?startDate=&endDate=`. A stub for that path exists only in the commented-out [`StubLambdaStack`](../../infrastructure/legacy/internal/stack/stubLambda.go), with a hard-coded handler ([`stub/lambda/me/clubs/events/get.go`](../../infrastructure/legacy/stub/lambda/me/clubs/events/get.go)) and stale SQL ([`GET_me_clubs_events.sql`](../../infrastructure/legacy/stub/lambda/me/clubs/events/GET_me_clubs_events.sql)). This route is the one implementation; add an alias for the PDF path only if a client needs it.

**Queries** (1): 1. [`me/events/list/SELECT_student_events.sql`](../../database/queries/me/events/list/SELECT_student_events.sql) (read)

**Code:** route [`student_routes.go`](../../infrastructure/legacy/gateway/routes/student_routes.go) · handler [`me/events/get.go`](../../infrastructure/legacy/lambda/api/me/events/get.go) · SQL [`students/SELECT_student_events.sql`](../../infrastructure/legacy/utils/query_client/queries/students/SELECT_student_events.sql) · [query group 4](../../database/README.md#4-current-students-events)

## 🔴 GET `/me/clubs/eboard`

Lists the clubs where the caller is e-board or owner (PDF). Until it exists, the frontend's Create and New event step 1 pages work this out from `GET /me/clubs`.

**Auth:** 🔴 JWT (PDF). Caller-scoped: memberships with role `eboard` or `owner`.

**Request:** none planned. A PDF note suggests the path might be `{me_id}/clubs/eboard` instead.

**Response:** not defined. It could reuse `GET /me/clubs` with a role filter ([query group 17](../../database/README.md#17-my-e-board-clubs)).

**Need:** none. The frontend filters `GET /me/clubs` by `role` itself.

**Status:** ⬜ Not built. Only an inactive, hard-coded stub ([`stub/lambda/me/clubs/eboard/get.go`](../../infrastructure/legacy/stub/lambda/me/clubs/eboard/get.go)) and stale SQL ([`eboard.sql`](../../infrastructure/legacy/stub/lambda/me/clubs/eboard/eboard.sql)) exist. That SQL has an `AND`/`OR` precedence bug and selects student IDs instead of clubs, so treat it as design notes, not working SQL.

**Queries** (1): 1. [`me/clubs/list/SELECT_student_clubs.sql`](../../database/queries/me/clubs/list/SELECT_student_clubs.sql) (read, the API keeps `role` `eboard` and `owner`). There's no SQL of its own ([query group 17](../../database/README.md#17-my-e-board-clubs)).

**Plan:** PDF "Endpoints Revamp", p. 15 · [query group 17](../../database/README.md#17-my-e-board-clubs)
