# HTTP API reference

The CDK application creates two API Gateway HTTP APIs with the same active business routes:

| API | MySQL schema | CDK construct |
| --- | --- | --- |
| `ClubEventApiDev` | `STAGING` | [`NewDevApiStack`](../../internal/stack/developmentApi.go) |
| `ClubEventApiProd` | `PRODUCTION` | [`NewProdApiStack`](../../internal/stack/productionApi.go) |

Each stack exports its AWS-assigned API endpoint as `myHttpApiEndpoint`. The source does not define a custom API domain, path prefix, or version prefix, so the paths below are relative to that output.

The route inventory is derived from the route functions called by both API stacks. Stub handlers, the `DatabaseRoutes` helper, and commented-out routes are not included because neither active API stack registers them.

## Calling the API

Public route:

```bash
curl "https://api.example.test/events?startDate=2026-09-01&endDate=2026-10-01"
```

Route with the Cognito authorizer attached:

```bash
curl \
  -H "Authorization: Bearer <cognito-jwt>" \
  "https://api.example.test/me"
```

API Gateway verifies the JWT only for routes whose CDK registration supplies an `Authorizer`. A bearer header on any other route does not cause API Gateway to populate `requestContext.authorizer.jwt.claims`. See [Authentication](../architecture/authentication.md) and [Authorization](../architecture/authorization.md).

## Active route matrix

“JWT attached” describes the API Gateway route configuration. “Handler authorization” describes the checks made after the request reaches Lambda.

| Method and path | JWT attached | Handler authorization actually enforced | Reference |
| --- | --- | --- | --- |
| `GET /health` | No | None | [Health](#health) |
| `GET /me` | Yes | Uses JWT `sub`; ensures and reads that student | [Students](students.md#get-me) |
| `GET /me/clubs` | Yes | Uses JWT `sub`; returns that student's memberships | [Students](students.md#get-meclubs) |
| `GET /me/events` | Yes | Uses JWT `sub`; returns events linked to that student's clubs | [Students](students.md#get-meevents) |
| `GET /clubs` | No | None | [Clubs](clubs.md#get-clubs) |
| `POST /clubs` | Yes | Authenticated identity only; no admin or club-role check | [Clubs](clubs.md#post-clubs) |
| `GET /clubs/{clubId}` | No | None | [Clubs](clubs.md#get-clubsclubid) |
| `GET /clubs/{clubId}/events` | No | None | [Clubs](clubs.md#get-clubsclubidevents) |
| `POST /clubs/{clubId}/events` | Yes | Authenticated identity only; no membership, e-board, or owner check | [Events](events.md#post-clubsclubidevents) |
| `POST /clubs/{clubId}/members/me` | Yes | Acts on the JWT student; inserts a regular membership | [Clubs](clubs.md#post-clubsclubidmembersme) |
| `DELETE /clubs/{clubId}/members/me` | No | Handler requires JWT claims, but this route does not provide them; deletion is not reached through the active route | [Clubs](clubs.md#delete-clubsclubidmembersme) |
| `POST /clubs/{clubId}/thumbnails` | No | None | [Images](images.md#post-clubsclubidthumbnails) |
| `GET /events` | No | None | [Events](events.md#get-events) |
| `GET /events/{eventId}` | No | None | [Events](events.md#get-eventseventid) |
| `GET /clubs/{clubId}/events/{eventId}/images` | No | None | [Images](images.md#get-clubsclubideventseventidimages) |
| `POST /clubs/{clubId}/events/{eventId}/images` | No | None | [Images](images.md#post-clubsclubideventseventidimages) |
| `POST /clubs/{clubId}/events/{eventId}/images/confirm` | No | None | [Images](images.md#post-clubsclubideventseventidimagesconfirm) |
| `POST /clubs/{clubId}/events/{eventId}/thumbnails` | No | None | [Images](images.md#post-clubsclubideventseventidthumbnails) |

The three event-image `POST` registrations also register `OPTIONS` against the same Lambda integration. The handlers contain no separate `OPTIONS` branch. Both HTTP APIs additionally have API-level CORS configuration.

## Response conventions

Most database-backed handlers use [`gateway/helpers/responses.go`](../../gateway/helpers/responses.go):

| Status | Body pattern | Meaning in active handlers |
| --- | --- | --- |
| `200` | `{"message":"...", ...}` | Successful reads and writes; create and join operations also use `200`, not `201`. |
| `400` | `{"error":"...", ...}` | Invalid input, missing handler-visible claims, duplicate membership, or several not-found branches. |
| `404` | `{"error":"..."}` | Used by `GET /events/{eventId}` when its query returns no rows. |
| `500` | `{"error":"...", ...}` | Database, query, or serialization failures. Some handlers include the underlying error text. |

The image handlers construct responses directly and use different field names and error-body formatting; their exact contracts are in [Images](images.md). A missing or invalid token on a JWT-attached route is rejected by API Gateway before Lambda, normally as `401 Unauthorized`.

Successful helper responses set `Content-Type: application/json` and permissive CORS headers. Image and health responses do not all set `Content-Type` consistently.

Database-backed handlers share the embedded-SQL loader described in the [database overview](../database/README.md#application-query-access). It reads into a fixed 4 KiB buffer without applying the returned byte count, so a loaded query can contain trailing zero bytes or be truncated. The endpoint pages document each handler's response construction and any additional deterministic query-file or SQL defects visible in source; a deployed database integration should still be exercised before treating a success example as a runtime guarantee.

## CORS

Both APIs allow all origins and all request headers. The production API advertises `GET`, `POST`, `OPTIONS`, `PATCH`, and `DELETE`; the development API advertises the same list without `DELETE`. No active application route uses `PATCH`.

## Health

### `GET /health`

- Authentication: none.
- Handler: [`lambda/api/test/ping/main.go`](../../lambda/api/test/ping/main.go).
- Parameters and body: none.
- `200` response: `{"greeting":"Server running"}`.
- A JSON-marshalling failure follows the handler's `400` path.

## Topic guides

- [Student and current-identity routes](students.md)
- [Club routes](clubs.md)
- [Event routes and event representation](events.md)
- [Image and signed-URL routes](images.md)
- [Authentication architecture](../architecture/authentication.md)
- [Authorization architecture](../architecture/authorization.md)
