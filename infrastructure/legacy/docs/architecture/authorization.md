# Authorization

The current backend has two distinct access-control layers:

1. A Cognito JWT authorizer on selected API Gateway routes authenticates the caller and supplies a verified `sub`.
2. MySQL tables represent application roles and relationships, but a handler must query those tables to authorize an operation.

An attached Cognito authorizer is therefore evidence of identity, not evidence of club or administrative privilege. [Authentication](authentication.md) documents token creation, validation, and student synchronization.

## Database authorization state

The fresh-install schema in [`11_04_2025_create_core_tables_up.sql`](../../lambda/internal/database/init/migrations/11_04_2025_create_core_tables_up.sql) stores these relevant values:

| Table or field | Meaning in active code |
| --- | --- |
| `students.id` | Cognito `sub`, used as the application identity key. |
| `club_members(fk_student_id, fk_club_id)` | Membership relation. |
| `club_members.member_is_eboard` | Club-level e-board flag. |
| `club_members.member_is_owner` | Club-level owner flag. |
| `admins.fk_student_id` | Admin record; no active API handler queries it. |
| `verified_clubs.fk_club_id` | Controls the optional `GET /clubs?verified=true` filter; it is not an authorization grant. |
| `events_to_clubs.club_is_event_owner` | Identifies the owner club for an event; it is not a user-role flag. |

[`GET /me/clubs`](../api/students.md#get-meclubs) presents one role string per membership. It gives `owner` precedence over `eboard`, then falls back to `member`. The schema does not make the two boolean flags mutually exclusive.

## Available role-check helpers

The repository includes role-aware queries and Go helpers:

- [`IS_student_authorized_club.sql`](../../utils/query_client/queries/authorization/IS_student_authorized_club.sql) returns true when a student is an e-board member or owner of the specified club.
- [`IS_student_authorized_event.sql`](../../utils/query_client/queries/authorization/IS_student_authorized_event.sql) returns true when a student is an e-board member or owner of any club linked to the event.
- [`utils/auth/club_authorization.go`](../../utils/auth/club_authorization.go) and [`utils/auth/event_authorization.go`](../../utils/auth/event_authorization.go) wrap those queries. Equivalent internal packages also exist under `lambda/internal/auth`.

No active API handler calls these helpers or loads either authorization query. Their presence does not add enforcement to a route.

## Active enforcement matrix

“Cognito” means API Gateway validates a token before Lambda. The final column states the application authorization that the active handler actually performs.

| Method and path | Cognito | Active handler enforcement |
| --- | --- | --- |
| `GET /health` | No | None. |
| `GET /clubs` | No | None; `verified=true` is a data filter. |
| `GET /clubs/{clubId}` | No | None. |
| `GET /clubs/{clubId}/events` | No | None. |
| `GET /events` | No | None. |
| `GET /events/{eventId}` | No | None. |
| `GET /me` | Yes | Scopes the student read to the JWT `sub`. |
| `GET /me/clubs` | Yes | Scopes membership rows to the JWT `sub`. |
| `GET /me/events` | Yes | Requires an authenticated student and filters through that student's club memberships. |
| `POST /clubs` | Yes | Ensures the JWT student exists; no admin or role check. |
| `POST /clubs/{clubId}/events` | Yes | Uses the JWT `sub` as author; no membership, e-board, owner, or admin check. |
| `POST /clubs/{clubId}/members/me` | Yes | Restricts the inserted student ID to the JWT `sub`; assigns regular-member flags. |
| `DELETE /clubs/{clubId}/members/me` | No | Handler requires JWT claims, but the route supplies no authorizer context, so it returns `400` before the delete. |
| `POST /clubs/{clubId}/thumbnails` | No | None. |
| `GET /clubs/{clubId}/events/{eventId}/images` | No | None. |
| `POST /clubs/{clubId}/events/{eventId}/images` | No | None. |
| `POST /clubs/{clubId}/events/{eventId}/images/confirm` | No | None. |
| `POST /clubs/{clubId}/events/{eventId}/thumbnails` | No | None. |

The event-image `POST` paths also register unauthenticated `OPTIONS` methods against their handlers. See the [API route matrix](../api/README.md#active-route-matrix).

## Write-operation semantics

### Club creation

[`POST /clubs`](../api/clubs.md#post-clubs) can be called by any user whose Cognito JWT passes the route authorizer. It inserts `clubs` and `club_info` rows but does not create a `club_members` row for the caller, set an owner flag, or verify the club. The `admins` table is not consulted.

### Event creation

[`POST /clubs/{clubId}/events`](../api/events.md#post-clubsclubidevents) authenticates the user and records the JWT `sub` as `fk_author_id`. It does not check that the caller belongs to `{clubId}`, has an e-board/owner flag, or has authority over associate clubs. Its current SQL fails before a successful response, as detailed in the endpoint reference; that execution failure is separate from the missing role check.

### Self-membership

[`POST /clubs/{clubId}/members/me`](../api/clubs.md#post-clubsclubidmembersme) is caller-scoped because the student ID comes only from the verified `sub`. It inserts `member_is_eboard = false` and `member_is_owner = false`, and it cannot add a different student through request data.

The corresponding `DELETE` handler contains a SQL guard that prevents deletion of an owner membership. The active route omits its Cognito authorizer, however, so ordinary API requests do not reach that SQL guard.

### Image operations

All signed-upload, confirmation, and gallery-read routes are public at API Gateway and contain no caller or database-role checks. `{clubId}` is ignored by event image handlers, and event ownership is not verified. The confirmation handler trusts client-provided image metadata and object keys. See [Image API](../api/images.md) and [Image upload architecture](image-uploads.md).

## Cognito, database roles, and IAM

These mechanisms answer different questions:

| Mechanism | Question answered |
| --- | --- |
| Cognito JWT authorizer | Is this request carrying a valid token from the configured user pool/client, and what is its `sub`? |
| MySQL role flags and relations | Is that student a member, e-board member, owner, admin, or associated with a particular club/event? |
| Lambda IAM role | May this Lambda read the secret, connect to RDS, or access an S3 key prefix? |
| S3 signed URL | May the URL bearer perform this specific S3 operation until the signature expires? |

Lambda IAM permissions constrain AWS service calls made by the function; they do not inspect the application's Cognito student or database role. Similarly, a signed URL delegates its one S3 operation to whoever holds it, independently of Cognito after issuance.

## Failure boundary

On a Cognito-attached route, a missing or invalid token is rejected by API Gateway before Lambda. On a public route, a bearer header is not parsed into verified claims. Handler-level input errors generally return `400`, while the active leave-club mismatch returns `400` because `requestContext.authorizer` is absent.

The backend does not use Cognito groups, custom authorization claims, route scopes, or the `admins` table in active request authorization. Any authorization attributed to those mechanisms would not describe the current handlers.
