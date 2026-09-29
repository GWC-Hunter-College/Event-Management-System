# Clubs

Club discovery, detail, and creation. Related routes live elsewhere: joining and leaving in [memberships.md](memberships.md), a club's events in [club-events.md](club-events.md), logo uploads in [images.md](images.md), and verification writes in [verification.md](verification.md).

Routes are registered in [`club_routes.go`](../../infrastructure/legacy/gateway/routes/club_routes.go). Response shapes come from [`models.Club` and `models.ClubDetailed`](../../infrastructure/legacy/database/models/club.go); example values are made up, following the [implementation's club docs](../../infrastructure/legacy/docs/api/clubs.md).

**Club object:** list responses return `id`, `name`, and `thumbnailUrl`. The detail response adds `website_url` and `description`, in snake_case while the other keys are camelCase. `thumbnailUrl` is the logo's S3 object key, not a URL, and it's omitted when the club has no logo.

## 🟢 GET `/clubs`

Lists clubs; with `?verified=true`, only clubs in `verified_clubs`. The frontend calls `GET /clubs?verified=true` from the Clubs page (`/clubs`), and from Home and Events to look up host club names, because event responses leave club names empty.

**Auth:** 🟢 Public. Verification is a data filter, not an authorization rule.

**Query params:**

| Name | Rule |
| --- | --- |
| `verified` | Only the exact strings `true` and `TRUE` filter to verified clubs. Any other value, or none, returns all clubs. |

**Response `200`:**

```json
{
  "message": "Succesfully fetched 1 clubs",
  "clubs": [
    { "id": 7, "name": "Example Club", "thumbnailUrl": "clubs/7/thumbnails/example.png" }
  ]
}
```

No clubs returns `200` with `"clubs": []`.

**Errors:** `500` `{"error": "Could not fetch clubs: <db error>"}`.

**Status:** ✅ Implemented. This one route covers both the PDF's "list clubs" read and `GET /clubs?verified=true`, which the PDF's first endpoint list writes as `GET /clubs/verified=true`.

**Known issues:**

- `thumbnailUrl` is an S3 object key, not a URL; the frontend maps it straight to its `logo` field.

**Code:** route [`club_routes.go`](../../infrastructure/legacy/gateway/routes/club_routes.go) · handler [`clubs/get.go`](../../infrastructure/legacy/lambda/api/clubs/get.go) · SQL [`clubs/SELECT_clubs.sql`](../../infrastructure/legacy/utils/query_client/queries/clubs/SELECT_clubs.sql) · [query group 5](../../database/README.md#5-club-list-and-verified-filter)

## 🔴 POST `/clubs`

Creates a club and its `club_info` row in one transaction (`ExecInsertQuery`). The frontend's New club page (`/club/create`) calls it.

**Auth:** 🔴 JWT. Any signed-in user can create a club; no admin, membership, or ownership check runs. Whether creating a club should require an admin is an [open question](../README.md#open-questions).

**Request body** (the handler's `RequestBodySchema`, which embeds `models.ClubDetailed`):

```json
{
  "club": {
    "name": "Example Club",
    "website_url": "https://club.example.edu",
    "description": "A sanitized example club."
  }
}
```

`club.name` is required and non-empty. `website_url` and `description` are optional, with no URL or length validation. `id` and `thumbnailUrl` are accepted but ignored.

**Response `200`:**

```json
{ "message": "Successfully inserted club into database", "clubId": 7 }
```

**Errors:**

| Status | When |
| --- | --- |
| `401` | Token missing or invalid (from API Gateway). |
| `400` | No claims or `sub`, `RequireStudent` failed, the body isn't JSON, or validation failed. Validation failures add `"errors": [{"field": "Name", "message": "<namespace> is required"}]`. |
| `500` | The insert transaction failed, including a duplicate name (`clubs.name` is unique). |

**Status:** ✅ Implemented.

**Known issues:**

- The creator isn't added to `club_members` and no owner is set, so nobody can manage the new club. The frontend's `docs/api.md` assumes the creator becomes the owner. See [query group 29](../../database/README.md#29-club-creator-ownership) and the [open questions](../README.md#open-questions).
- New clubs aren't verified, so they don't appear in `GET /clubs?verified=true`.
- The frontend also sends `logo` (a `blob:` URL in design mode) and `tags`; the handler ignores both.

**Code:** route [`club_routes.go`](../../infrastructure/legacy/gateway/routes/club_routes.go) · handler [`clubs/post/post.go`](../../infrastructure/legacy/lambda/api/clubs/post/post.go) · SQL [`clubs/INSERT_club.sql`](../../infrastructure/legacy/utils/query_client/queries/clubs/INSERT_club.sql) and [`clubs/INSERT_club_info.sql`](../../infrastructure/legacy/utils/query_client/queries/clubs/INSERT_club_info.sql) · [query group 7](../../database/README.md#7-club-creation)

## 🟢 GET `/clubs/{clubId}`

Returns one club with its logo key, website, and description. It returns unverified clubs too. The frontend's Club (`/club/:clubId`), Event (for the host club), and New event pages call it.

**Auth:** 🟢 Public.

**Path params:** `clubId`, passed to SQL as received, with no integer check.

**Response `200`:**

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

`website_url` and `description` can be `null`.

**Errors:**

| Status | Body |
| --- | --- |
| `400` | `{"error": "Club with id <clubId> not found"}` |
| `500` | `{"error": "Could not fetch club: <db error>"}` |

**Status:** ✅ Implemented.

**Known issues:**

- A missing club returns `400`, not `404`. The frontend's `docs/api.md` expects `404`.

**Code:** route [`club_routes.go`](../../infrastructure/legacy/gateway/routes/club_routes.go) · handler [`clubs/clubId/get.go`](../../infrastructure/legacy/lambda/api/clubs/clubId/get.go) · SQL [`clubs/SELECT_club.sql`](../../infrastructure/legacy/utils/query_client/queries/clubs/SELECT_club.sql) · [query group 6](../../database/README.md#6-club-detail)
