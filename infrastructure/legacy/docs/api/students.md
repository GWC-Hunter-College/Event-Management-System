# Student and current-identity API

The `/me` routes are the API's caller-relative reads. All three active routes attach the Cognito JWT authorizer and use the verified `sub` claim as `students.id`. The handlers also read the optional `email` claim.

See [Authentication](../architecture/authentication.md) for token validation and identity synchronization, [Authorization](../architecture/authorization.md) for the distinction between identity and database roles, and [Events](events.md) for the event object returned by `/me/events`.

## Common identity behavior

Before querying response data, each `/me` handler calls [`RequireStudent`](../../utils/auth/ensure_student.go):

1. Reject an empty `sub`.
2. Check for a `students` row with that ID.
3. If it is absent, upsert `sub` and `email`, or insert the `sub` with a null email when no email claim is present.

This is a request-time fallback in addition to the Cognito trigger sync. It establishes identity, not a club role or admin grant.

## `GET /me`

Returns the database student associated with the caller.

- JWT authorizer: attached.
- Handler: [`lambda/api/me/get.go`](../../lambda/api/me/get.go).
- Query: [`students/SELECT_student_by_sub.sql`](../../utils/query_client/queries/students/SELECT_student_by_sub.sql).
- Entity: [`models.Student`](../../database/models/student.go).
- Path, query, and body parameters: none.

`200 OK`:

```json
{
  "message": "Successfully fetched student 11111111-2222-3333-4444-555555555555",
  "student": {
    "id": "11111111-2222-3333-4444-555555555555",
    "email": "student@example.edu"
  }
}
```

Handler response paths:

| Status | Condition |
| --- | --- |
| `200` | The student row is read successfully. |
| `400` | The handler-visible claims contain no `sub`; the no-row branch also uses the `400` helper despite its source comment calling it not found. |
| `500` | Ensuring or querying the student fails. |

The model's `email` field is a Go `string`, while the database column is nullable. A row created without an email can therefore fail when SQL scans a null email into the response model.

## `GET /me/clubs`

Returns all club memberships for the caller and computes one role string per membership.

- JWT authorizer: attached.
- Handler: [`lambda/api/me/clubs/get.go`](../../lambda/api/me/clubs/get.go).
- Query: [`students/SELECT_student_clubs.sql`](../../utils/query_client/queries/students/SELECT_student_clubs.sql).
- Entity: [`models.ClubWithRole`](../../database/models/club.go).
- Path, query, and body parameters: none.

`200 OK`:

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

`role` is selected from the database flags in this order: `owner` when `member_is_owner = 1`, `eboard` when `member_is_eboard = 1`, and otherwise `member`. The value exposed as `thumbnailUrl` is the joined image `object_key`; this handler does not turn it into a signed or public URL. With no memberships, `clubs` is an empty array.

Handler response paths are `200` on success, `400` for a missing `sub`, and `500` for identity-sync or query errors.

## `GET /me/events`

Returns posted events associated with any club the caller has joined.

- JWT authorizer: attached.
- Handler: [`lambda/api/me/events/get.go`](../../lambda/api/me/events/get.go).
- Query: [`students/SELECT_student_events.sql`](../../utils/query_client/queries/students/SELECT_student_events.sql).
- Entities: [`event_schema.ResponseSchema`](../../lambda/api/events/schema/schema.go), [`models.Event`](../../database/models/event.go), and [`models.Club`](../../database/models/club.go).

### Query parameters

All query parameters are optional.

| Name | Default | Validation and use |
| --- | --- | --- |
| `startDate` | `1970-01-01` | Must parse as `YYYY-MM-DD`; SQL requires `event.start_date > startDate`. |
| `endDate` | `2100-01-01` | Must parse as `YYYY-MM-DD`; SQL requires `event.end_date < endDate`. |
| `limit` | `10` | Positive integer; applied to joined SQL rows. |
| `page` | `0` | Non-negative integer; offset is `page * limit`. |

The query always filters `status LIKE 'posted'`. It filters each event-to-club association through the caller's `club_members` rows, so the response represents only linked clubs that the caller joined. If the caller joined an associate club but not the owner club, the event can qualify while `owners.owner` remains the empty zero-value club object.

`200 OK` has the same event representation documented for [`GET /events`](events.md#get-events):

```json
{
  "message": "Succesfully fetched 1 events of student 11111111-2222-3333-4444-555555555555",
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
      "description": "A sanitized example.",
      "owners": {
        "owner": {
          "id": 7,
          "name": "",
          "thumbnailUrl": "<fixed-placeholder-url>"
        },
        "associates": []
      }
    }
  ]
}
```

Handler response paths:

| Status | Condition |
| --- | --- |
| `200` | The query and response grouping succeed; zero results are returned as an empty array. |
| `400` | Missing `sub`, invalid date, non-positive `limit`, or negative/non-integer `page`. |
| `500` | Identity synchronization or the SQL query fails. |

### Current grouping and pagination behavior

SQL paginates event-to-club rows before the handler groups rows by event ID. A page can therefore contain fewer distinct events than `limit`, or only part of an event's club associations. The handler builds the final slice from a Go map, so the SQL `ORDER BY start_date` is not preserved deterministically in the JSON array. Club names and database image keys are not copied into the owner objects; each returned club contains its ID, an empty `name`, and the handler's fixed placeholder thumbnail URL.
