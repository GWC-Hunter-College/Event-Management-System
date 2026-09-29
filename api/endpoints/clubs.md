# Clubs

Club discovery, detail, and creation. Related routes live elsewhere: joining and leaving in [memberships.md](memberships.md), a club's events in [club-events.md](club-events.md), logo uploads in [images.md](images.md), and verification writes in [verification.md](verification.md).

Routes are registered in [`club_routes.go`](../../infrastructure/legacy/gateway/routes/club_routes.go). Response shapes come from [`models.Club` and `models.ClubDetailed`](../../infrastructure/legacy/database/models/club.go); example values are made up, following the [implementation's club docs](../../infrastructure/legacy/docs/api/clubs.md).

**Club object:** list responses return `id`, `name`, and `thumbnailUrl`. The detail response adds `website_url` and `description`, in snake_case while the other keys are camelCase. `thumbnailUrl` is the logo's S3 object key, not a URL, and it's omitted when the club has no logo.

## Proposed club object

**Proposed.** Need: **Now**. What club reads would return so the Clubs, Club, My Clubs, and Event pages show everything they already have UI for. It keeps `thumbnailUrl`, which the frontend's `fromJsonClub` already maps to its `logo`. Fixes [M6](../coverage.md#m6-no-usable-image-urls) and [M13](../coverage.md#m13-club-fields-description-tags-member-count-logo).

```json
{
  "id": 7,
  "name": "Example Club",
  "thumbnailUrl": "<readable-logo-url>",
  "description": "A sanitized example club.",
  "tags": ["technology", "community"],
  "memberCount": 41,
  "verified": true
}
```

| Field | Returned by | Notes |
| --- | --- | --- |
| `id`, `name` | Every club read | Unchanged. |
| `thumbnailUrl` | Every club read | A readable URL instead of the S3 key; absent when there's no logo. |
| `description` | `GET /clubs`, `GET /clubs/{clubId}` | Newly in the list: club cards show it. `null` when unset. |
| `tags` | `GET /clubs`, `GET /clubs/{clubId}` | New. Topic keys from the fixed list below, lowercase, at most 3; `[]` when none. The UI uppercases them for display. Needs a `club_tags` table ([schema needs](../coverage.md#schema-needs)). |
| `memberCount` | `GET /clubs`, `GET /clubs/{clubId}` | New. Count of `club_members` rows. The frontend hides the count when it's absent, so it can ship later than the rest. |
| `verified` | `GET /clubs/{clubId}` | New. Lets a club page show whether it's listed; no current screen needs it, so it's optional. |
| `website_url` | `GET /clubs/{clubId}` | Unchanged (snake_case). No screen shows it. |
| `role` | `GET /me/clubs` only | Unchanged. |

**Topics** ([decision 5](../README.md#decisions), 2026-09-29): a fixed list, stored and returned as lowercase keys, at most 3 per club. The list is taken to be the New club form's six topics, lowercased: `technology`, `arts`, `community`, `women in stem`, `career`, `sports`. The exact key spelling (for example `women in stem` or `women-in-stem`) is still an [open question](../README.md#open-questions). The form sends its uppercase labels today, and the design-mode fixtures use other, title-case tags; both change to the keys ([frontend changes](../coverage.md#frontend-changes)).

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
- The Clubs page's cards show description, topic tags, and member count, none of which the list returns.

**Proposed changes** (Need: **Now**; screens: Clubs; host-club lookups on Home and Events until [M8](../coverage.md#m8-event-objects-have-empty-club-names) is fixed): return the [proposed club object](#proposed-club-object) in the same `{"message", "clubs"}` envelope. Access stays 🟢 and `verified` works as today.

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

**Proposed contract** (Need: **Now**; screen: New club form). Fixes [M4](../coverage.md#m4-the-creator-doesnt-become-the-clubs-owner) and [M13](../coverage.md#m13-club-fields-description-tags-member-count-logo).

- **Auth:** 🔴 JWT. Whether it also needs an admin is [open](../README.md#open-questions).
- **Request body:**

  ```json
  {
    "club": { "name": "Example Club", "description": "A sanitized example club.", "website_url": "https://club.example.edu" },
    "tags": ["technology", "community"]
  }
  ```

  `tags` is optional: at most 3 keys from the [topic list](#proposed-club-object); an unknown key or a fourth tag is `400`. `logo` isn't accepted: the form uploads it after creation through the [logo flow](images.md#how-image-uploads-work).
- **Response `200`:** unchanged, `{"message": "Successfully inserted club into database", "clubId": 7}`.
- **Writes:** `clubs`, `club_info`, the tags, and an owner membership for the caller (`role = 'owner'`), in one transaction.
- **Errors:** as today, plus `409` for a duplicate name instead of `500`.

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
- `clubId` isn't checked as an integer.

**Proposed changes** (Need: **Now**; screens: Club page header, Event page host, New event form chip; later the GWC website): return the [proposed club object](#proposed-club-object) under `club`; `404` `{"error": "Club with id <clubId> not found"}` for a missing club and `400` for a non-integer `clubId` ([M3](../coverage.md#m3-400-where-the-frontend-expects-401-403-or-404)). Access stays 🟢; unverified clubs stay readable.

**Code:** route [`club_routes.go`](../../infrastructure/legacy/gateway/routes/club_routes.go) · handler [`clubs/clubId/get.go`](../../infrastructure/legacy/lambda/api/clubs/clubId/get.go) · SQL [`clubs/SELECT_club.sql`](../../infrastructure/legacy/utils/query_client/queries/clubs/SELECT_club.sql) · [query group 6](../../database/README.md#6-club-detail)

## 🔴 PATCH `/clubs/{clubId}`

Updates a club's name, description, website, or tags. **Proposed**; not in the planning PDF. Need: **Later**: no screen edits a club yet (logo changes go through the [logo upload flow](images.md#how-image-uploads-work)).

**Auth:** 🔴 JWT + owner or e-board member of the club ([decision 6](../README.md#decisions)). Deleting a club is owner-only under the same decision, but no endpoint deletes a club and none is proposed ([open question](../README.md#open-questions)).

**Path params:** `clubId`, an integer.

**Request body (Proposed):** any subset of `{ "club": { "name", "description", "website_url" }, "tags": [] }`; omitted fields are unchanged, and `tags` replaces the list.

**Response `200` (Proposed):** `{"message": "Successfully updated club 7", "club": { /* proposed club object */ }}`.

**Errors (Proposed):** `400` validation (including topic keys), `401`, `403` not an owner or e-board member, `404` unknown club, `409` duplicate name, `500`.

**Status:** ⬜ Not built. No route, handler, or update query.
