# Admins

All planned. The PDF says **every** admin route must first check that the caller is an admin. The schema has an `admins (fk_student_id)` table, but there's no admin route, handler, or [admin check](internal.md#-admin-check). According to the stub note in [`stub/lambda/admins/admins.txt`](../../infrastructure/legacy/stub/lambda/admins/admins.txt), admins are the people who verify clubs, so random users can't flood the directory with fake ones.

The PDF shows ✅ next to two of these routes. Those marks tracked task progress; no admin route exists in the code.

Student IDs are Cognito `sub` UUID strings (`students.id` is `CHAR(36)`).

**Need: Later.** The frontend has no admin page or admin UI, so nothing calls these routes yet; see [coverage.md](../coverage.md#admin). When an admin page exists, it can check [`isAdmin` on `GET /me`](me.md#-get-me) (Proposed) before showing itself.

## 🔴 GET `/admins`

Lists admins, "their id and clubs" (PDF).

**Auth:** 🔴 JWT + admin.

**Response:** not defined. Open: whether "clubs" means clubs they manage, clubs they've joined, or a separate response.

**Status:** ⬜ Not built. Only planning text and the schema exist.

**Plan:** PDF "Endpoints Revamp", p. 18 · [query group 22](../../database/README.md#22-admin-crud)

## 🔴 GET `/admins/{studentId}`

Returns one admin's record, if there is one.

**Auth:** 🔴 JWT + admin.

**Path params:** `studentId`.

**Response:** not defined.

**Status:** ⬜ Not built. There's only an unwired SQL stub, [`GET_admins_studentId.sql`](../../infrastructure/legacy/stub/lambda/admins/studentId/GET_admins_studentId.sql), which hard-codes the obsolete numeric student ID `2`. Treat it as design notes, not working SQL.

**Plan:** PDF "Endpoints Revamp", p. 19 · [query group 22](../../database/README.md#22-admin-crud)

## 🔴 POST `/admins`

Promotes a student to admin by inserting their ID into `admins`.

**Auth:** 🔴 JWT + admin. It must also check that the target student exists.

**Request body:** not defined.

**Response:** not defined.

**Status:** ⬜ Not built. No route, handler, insert query, or admin check.

**Rules** ([decision 12](../README.md#decisions)): any admin may add another admin. The first admin is created by a developer with a one-off SQL insert, not through this route ([how](../../database/README.md#the-first-admin)).

**Proposed** (Need: **Later**): body `{ "studentId": "<sub>" }`. `404` when there's no such student, `409` when they're already an admin. The response isn't defined yet.

**Plan:** PDF "Endpoints Revamp", p. 19 · [query group 22](../../database/README.md#22-admin-crud)

## 🔴 DELETE `/admins/{studentId}`

Demotes an admin by deleting their `admins` row.

**Auth:** 🔴 JWT + admin.

**Path params:** `studentId`.

**Response:** not defined.

**Status:** ⬜ Not built. There's only an unwired SQL stub, [`DELETE_admins_studentId.sql`](../../infrastructure/legacy/stub/lambda/admins/studentId/DELETE_admins_studentId.sql), with the same hard-coded ID `2`.

**Rules** ([decision 12](../README.md#decisions)): any admin may remove another admin, or themselves, except the last admin. Removing the last admin affects no rows.

**Proposed** (Need: **Later**): `404` when the student isn't an admin, `409` when they're the last admin. The response isn't defined yet.

**Plan:** PDF "Endpoints Revamp", p. 19 · [query group 22](../../database/README.md#22-admin-crud)

## 🔴 POST `/admins/purge`

Permanently removes what was deleted more than 30 days ago. **Proposed**; not in the PDF. It comes from [decision 8](../README.md#decisions): purging is done by admins on demand, not by a scheduled job.

**Auth:** 🔴 JWT + admin.

**What it removes:**

- events deleted more than 30 days ago, with their descriptions, tags, club links, and gallery links;
- the flyers and gallery images those events used, unless a club logo or another event still uses them; and
- images deleted more than 30 days ago that nothing references (replaced logos and flyers, and takedowns), with their S3 files; and
- [announcements](announcements.md) deleted more than 30 days ago, with their club links.

Nothing deleted less than 30 days ago is touched, so everything that can still be [restored](event-management.md#-post-autheventseventidrestore) survives: the purge and the restore use the same 30-day cutoff ([decision 9](../README.md#decisions)).

**Status:** ⬜ Not built. The SQL exists ([query group 31](../../database/README.md#31-purge-job)); the order of the steps is in [Transactions](../../database/README.md#transactions).

**Proposed contract** (Need: **Later**: there's no admin page yet).

- **Request body:** none. The 30 days are fixed, not a parameter.
- **Response `200`:**

  ```json
  { "message": "Purge complete", "eventsPurged": 3, "announcementsPurged": 1, "imagesPurged": 7, "imageFilesFailed": 0 }
  ```

  `imageFilesFailed` counts S3 deletes that failed after their database row was removed. Those leave orphan files that nothing references, which is harmless; a later cleanup can list the bucket.
- **Errors:** `401`, `403` not an admin, `500`. Running it twice is safe; the second run finds nothing.

**Plan:** [decision 8](../README.md#decisions) · [query group 31](../../database/README.md#31-purge-job)
