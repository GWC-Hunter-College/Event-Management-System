# Announcements

Updates a club's e-board posts to its members: the club page's Announcements tab, which shows "SOON" today ("Later, the e-board will be able to post updates here"). **Every route here is planned (⬜) and Proposed**: none is in the planning PDF, and nothing is built. The design follows [events](events.md): the same soft delete, the same 30-day restore window and admin purge ([decisions 2, 8 and 9](../README.md#decisions)), and the same club links, so one announcement can belong to several clubs (for example HunterHacks for the CS clubs).

Need: **Later** for every route: the Announcements tab is SOON ([coverage.md](../coverage.md#clubclubid-club)).

The SQL for every route exists in the database module ([query group 32](../../database/README.md#32-announcements)): the `announcements` and `announcements_to_clubs` tables and the `announcement_details` view in the [baseline schema](../../database/migrations/schema/2026_09_29_baseline_up.sql), and the queries under [`database/queries/announcements/`](../../database/queries/announcements/) and [`clubs/announcements/`](../../database/queries/clubs/announcements/).

## Proposed announcement object

**Proposed.** Modelled on the [proposed event object](events.md#proposed-event-object): the same field names where the meaning is the same, and the same `owners` structure.

```json
{
  "id": 7,
  "authorId": "aaaaaaaa-bbbb-cccc-dddd-eeeeeeeeeeee",
  "title": "HunterHacks registration is open",
  "body": "Sign up by Friday. Teams of up to four.",
  "status": "posted",
  "postedAt": "2026-09-28T14:00:00Z",
  "createdAt": "2026-09-27T16:00:00Z",
  "updatedAt": "2026-09-27T16:00:00Z",
  "owners": {
    "owner": { "id": 3, "name": "Hunter CS Club", "thumbnailUrl": "<readable-logo-url>" },
    "associates": [{ "id": 2, "name": "Girls Who Code @ Hunter", "thumbnailUrl": "<readable-logo-url>" }]
  }
}
```

| Field | Notes |
| --- | --- |
| `body` | Plain text. `null` on a draft that has none yet; posting requires one. |
| `status` | `draft` or `posted`. See [Status](#status). |
| `postedAt` | When it was posted: ISO 8601 in UTC with `Z`. `null` on a draft. Posted lists are ordered by it, newest first. |
| `createdAt`, `updatedAt` | ISO 8601 in UTC with `Z`, as events. A draft written a week before it's posted still lists by `postedAt`. |
| `owners.owner` | The club that posted it (`club_is_announcement_owner`). |
| `owners.associates` | Other clubs it belongs to, sorted by id; `[]` when none. |

No images or tags yet. When they come, they follow events: `thumbnailUrl` and `altText` for a flyer, `imageCount` and `images` for a gallery, and a tags list, each backed by the same kind of table events use (see the comment in the baseline).

## Status

| Status | Who sees it | Meaning |
| --- | --- | --- |
| `draft` | The e-board and owners of any club it belongs to, through the [drafts list](#-get-clubsclubidannouncementsdrafts) and [`GET /auth/announcements/{announcementId}`](#-get-authannouncementsannouncementid). | Not published. The default when it's created. |
| `posted` | Everyone. | Published. |

**Allowed transition:** `draft → posted` only, through [`PATCH`](#-patch-authannouncementsannouncementid) with `"status": "posted"`, and only when there's a body. Posting sets `postedAt` to the current time; creating an announcement as `posted` sets it at creation. The database keeps `postedAt` set exactly when `status` is `posted`, and a restore keeps the original `postedAt`. Nothing goes back to `draft`, and there's no `cancelled`. The SQL enforces it: the posting `UPDATE` requires `status = 'draft'`, so any other change affects no rows.

**Deleting** is a soft delete, as for events: [`DELETE`](#-delete-authannouncementsannouncementid) sets `deleted_at`, and every read treats the announcement as not found. It can be [restored](#-post-authannouncementsannouncementidrestore) for 30 days; after that, an admin's [purge](admins.md#-post-adminspurge) removes it for good.

**Who manages an announcement:** the e-board and owners of **any** club it belongs to, as for [events](event-management.md) ([query](../../database/queries/authorization/announcements/can_manage/IS_student_authorized_announcement.sql)). Restoring is narrower: the owning club's e-board and owners, or an admin.

## 🟢 GET `/clubs/{clubId}/announcements`

Lists a club's posted announcements, most recently posted first (by `postedAt`). The club page's Announcements tab calls it, and so does the [GWC website](../coverage.md#gwc-website), with its configured club ID.

**Auth:** 🟢 Public.

**Path params:** `clubId`, an integer.

**Query params:** `limit` (default 50, `1`–`100`) and `page` (default 0), as in the [event lists](events.md#proposed-list-parameters). There's no `when` or date range.

**Response `200` (Proposed):**

```jsonc
{ "message": "Succesfully fetched 2 announcements", "announcements": [ /* announcement objects */ ] }
```

Each announcement lists every club it belongs to, not only this one.

**Errors:** `400` bad `clubId`, `limit`, or `page`; `404` `{"error": "Club with id <clubId> not found"}`; `500`.

**Status:** ⬜ Not built. SQL: [`SELECT_club_announcements.sql`](../../database/queries/clubs/announcements/list/SELECT_club_announcements.sql), with [`EXISTS_club.sql`](../../database/queries/clubs/get/EXISTS_club.sql) for the `404`.

## 🟢 GET `/announcements/{announcementId}`

Returns one posted announcement. For a link to one announcement, or a GWC-website read.

**Auth:** 🟢 Public. Drafts and deleted announcements are `404`.

**Response `200` (Proposed):** `{ "message": "Succesfully fetched announcement 7", "announcement": { /* announcement object */ } }`.

**Errors:** `400` non-integer id, `404`, `500`.

**Status:** ⬜ Not built. SQL: [`SELECT_announcement.sql`](../../database/queries/announcements/read/SELECT_announcement.sql) with `public_only` `TRUE`.

## 🔴 GET `/clubs/{clubId}/announcements/drafts`

Lists a club's draft announcements for its e-board and owners, most recently updated first.

**Auth:** 🔴 JWT + e-board or owner of `clubId`. `403` otherwise; `404` for an unknown club.

**Query params:** `limit` and `page`, as above.

**Response `200` (Proposed):** the list envelope, all with `"status": "draft"`.

**Status:** ⬜ Not built. SQL: [`SELECT_club_announcement_drafts.sql`](../../database/queries/clubs/announcements/drafts/SELECT_club_announcement_drafts.sql).

## 🔴 GET `/auth/announcements/{announcementId}`

Returns an announcement in either status, for its managers: the edit form's prefill.

**Auth:** 🔴 JWT + e-board or owner of any club it belongs to. `403` otherwise; `404` for an unknown or deleted announcement.

**Response `200` (Proposed):** `{ "message": "Succesfully fetched announcement 7", "announcement": { /* announcement object */ } }`.

**Status:** ⬜ Not built. SQL: [`SELECT_announcement.sql`](../../database/queries/announcements/read/SELECT_announcement.sql) with `public_only` `FALSE`, after the [manager check](../../database/queries/authorization/announcements/can_manage/IS_student_authorized_announcement.sql).

## 🔴 POST `/clubs/{clubId}/announcements`

Creates an announcement as `clubId`, the owner club, optionally shared with other clubs.

**Auth:** 🔴 JWT + e-board or owner of `clubId`. `403` otherwise; `404` for an unknown club.

**Request body (Proposed):**

```json
{
  "announcement": { "title": "HunterHacks registration is open", "body": "Sign up by Friday. Teams of up to four." },
  "status": "posted",
  "associates": [2, 6]
}
```

| Field | Rule |
| --- | --- |
| `announcement.title` | Required and non-empty. |
| `announcement.body` | Optional for a draft; required when `status` is `posted`. |
| `status` | Optional, `draft` or `posted`; omitted means `draft`, as for events. `posted` sets `postedAt` to now. |
| `associates` | Optional club ids, default `[]`; duplicates and `clubId` itself are dropped. |

**Response `200` (Proposed):** `{"message": "Successfully created announcement", "announcementId": 7}`, as event creation returns `eventId`.

**Errors:** `400` validation (with `"errors": [{"field", "message"}]`) or an unknown associate club, `401`, `403`, `404`, `500`.

**Writes:** the announcement, the owner link, and the associate links, in one transaction ([order](../../database/README.md#transactions)).

**Status:** ⬜ Not built.

## 🔴 PATCH `/auth/announcements/{announcementId}`

Edits an announcement, or posts a draft.

**Auth:** 🔴 JWT + e-board or owner of any club it belongs to. `403` otherwise; `404` for an unknown or deleted announcement.

**Request body (Proposed):** any subset of the create body's fields; omitted fields are unchanged. `status`, when present, must be `posted`, and only a draft with a body can move to it; `"status": "draft"` on a draft changes nothing. `associates`, when present, replaces the list. Last write wins, as for events.

```json
{ "announcement": { "body": "Sign up by Friday. Teams of up to four. Free food." }, "status": "posted" }
```

**Response `200` (Proposed):** `{ "message": "Successfully updated announcement 7", "announcement": { /* announcement object */ } }`.

**Errors:** `400` validation or a transition other than `draft → posted` (such as posting an announcement with no body, or `posted → draft`); `401`, `403`, `404`, `500`.

**Status:** ⬜ Not built. SQL under [`announcements/update/`](../../database/queries/announcements/update/).

## 🔴 DELETE `/auth/announcements/{announcementId}`

Deletes an announcement (soft delete).

**Auth:** 🔴 JWT + e-board or owner of any club it belongs to.

**Response `200` (Proposed):** `{"message": "Successfully deleted announcement 7", "announcementId": 7}`.

**Errors:** `401`, `403`, `404` (unknown or already deleted), `500`.

**Status:** ⬜ Not built. SQL: [`UPDATE_announcement_soft_delete.sql`](../../database/queries/announcements/delete/UPDATE_announcement_soft_delete.sql).

## 🔴 POST `/auth/announcements/{announcementId}/restore`

Restores an announcement deleted less than 30 days ago, as [event restore](event-management.md#-post-autheventseventidrestore) does.

**Auth:** 🔴 JWT + e-board or owner of the announcement's **owning** club, or an admin.

**Response `200` (Proposed):** `{ "message": "Successfully restored announcement 7", "announcement": { /* announcement object */ } }`, with the status it had when it was deleted.

**Errors:**

| Status | When |
| --- | --- |
| `401`, `403` | No token; not on the owning club's e-board, not an owner, and not an admin. |
| `404` | No such announcement, or it has been purged. |
| `409` | It isn't deleted. |
| `410` | It was deleted 30 or more days ago. |
| `500` | Database failure. |

**Status:** ⬜ Not built. SQL under [`announcements/restore/`](../../database/queries/announcements/restore/), with the [owning-club check](../../database/queries/authorization/announcements/manages_owner_club/IS_student_announcement_owner_club_manager.sql).

## Future work

- Notifying members when an announcement is posted (email or push). Not now: it depends on a notification system, which the bulletin board needs too. See the [open questions](../README.md#open-questions).
