# Event management (`/auth/events`)

Protected routes for an event's managers. The design puts them under `/auth/events/{eventId}` rather than `/clubs/{clubId}/events/{eventId}`: an event can belong to several clubs, so authorization doesn't depend on a single `clubId` in the path (PDF p. 18). The PDF's first endpoint list (pp. 11–12) shows the same operations under `/clubs/{clubId}/events/{eventId}`.

**None of these routes exist.** Some of the behavior exists elsewhere, with no auth:

| Planned route | What exists today |
| --- | --- |
| `GET /auth/events/{eventId}` | Public [`GET /events/{eventId}`](events.md#-get-eventseventid) (posted events only) and an unused [event role check](internal.md#-event-role-check). |
| `GET /auth/events/{eventId}/images` | Public and broken [`GET /clubs/{clubId}/events/{eventId}/images`](images.md#-get-clubsclubideventseventidimages). |
| `POST /auth/events/{eventId}/thumbnails` | Public [`POST /clubs/{clubId}/events/{eventId}/thumbnails`](images.md#-post-clubsclubideventseventidthumbnails). |
| `POST /auth/events/{eventId}/images` | Public [`POST /clubs/{clubId}/events/{eventId}/images`](images.md#-post-clubsclubideventseventidimages) and its [`/confirm`](images.md#-post-clubsclubideventseventidimagesconfirm) route. |
| `PATCH` and `DELETE /auth/events/{eventId}` | Nothing. |

Whether the protected media routes end up under `/auth/events` or stay club-scoped is an [open question](../README.md#open-questions). Either way, keep a single upload signer rather than two independent ones.

Every route below is 🔴 and needs a JWT plus e-board or owner of **any** club linked to the event (PDF), using the [event role check](internal.md#-event-role-check).

## 🔴 GET `/auth/events/{eventId}`

Returns an event in any status (draft, posted, or cancelled) with its description, for the event's managers (PDF: "protected for getting drafts").

**Auth:** 🔴 JWT + e-board or owner of a linked club.

**Path params:** `eventId`.

**Response:** not defined.

**Status:** ⬜ Not built. The public read and the event role check both exist but aren't combined into a route.

**Proposed contract** (Need: **Now**, for the New event form's `?draft=:eventId` prefill. The edit form behind "EDIT EVENT" and the Manage tab's "EDIT" would use it too, but that's **Later**.) Fixes the prefill half of [M1](../coverage.md#m1-drafts-in-public-lists).

- **Auth:** 🔴 JWT + e-board or owner of any linked club. `403` otherwise; `404` for an unknown event.
- **Response `200`:** the same envelope as the public read, with a [proposed event object](events.md#proposed-event-object) in any [status](events.md#event-status), including `draft`. A deleted event is `404`:

  ```jsonc
  { "message": "Succesfully fetched event 42", "event": { /* event object */ } }
  ```

**Open:** which statuses managers may see, and whether an event's author gets any rights of their own.

**Queries** (3): 1. [`authorization/events/can_manage/IS_student_authorized_event.sql`](../../database/queries/authorization/events/can_manage/IS_student_authorized_event.sql) (auth, `403`) → 2. [`events/read/SELECT_event.sql`](../../database/queries/events/read/SELECT_event.sql) (read, `public_only` `FALSE`; no row is `404`) → 3. [`events/images/list/SELECT_event_images.sql`](../../database/queries/events/images/list/SELECT_event_images.sql) (read, `images`)

**Plan:** PDF "Endpoints Revamp", p. 18 · [query group 24](../../database/README.md#24-authorized-event-read), reusing groups [11](../../database/README.md#11-public-and-composite-event-read) and [14](../../database/README.md#14-event-authorization)

## 🔴 GET `/auth/events/{eventId}/images`

Returns an event's images, including for drafts (PDF: "JWT so u can use on drafted events").

**Auth:** 🔴 JWT + e-board or owner of a linked club.

**Path params:** `eventId`.

**Response:** not defined.

**Status:** ⬜ Not built. The public club-scoped read ([`clubs/events/images/get/get.go`](../../infrastructure/legacy/lambda/api/clubs/events/images/get/get.go)) is missing its SQL, has no role check, and ignores `{clubId}`, so don't copy it for this route.

**Need: Later.** No screen manages gallery images yet; managers see images through `images` in [`GET /auth/events/{eventId}`](#-get-autheventseventid).

**Queries** (3): 1. [`authorization/events/can_manage/IS_student_authorized_event.sql`](../../database/queries/authorization/events/can_manage/IS_student_authorized_event.sql) (auth, `403`) → 2. [`events/read/SELECT_event_status.sql`](../../database/queries/events/read/SELECT_event_status.sql) (read, no row is `404`; any status) → 3. [`events/images/list/SELECT_event_images.sql`](../../database/queries/events/images/list/SELECT_event_images.sql) (read)

**Plan:** PDF "Endpoints Revamp", p. 18 · query groups [16](../../database/README.md#16-event-image-list) and [14](../../database/README.md#14-event-authorization)

## 🔴 PATCH `/auth/events/{eventId}`

Updates event fields from a JSON body. It's also how an event gets published; the PDF's example is `{ "status": "POSTED" }`. The Edit and Cancel buttons in the frontend's event manager bar have no endpoint behind them.

**Auth:** 🔴 JWT + e-board or owner of a linked club, plus field validation and a status-transition policy.

**Path params:** `eventId`.

**Request body (PDF):** JSON containing the fields to change.

**Response:** not defined.

**Status:** ⬜ Not built. No route or handler. The SQL and its transaction order exist ([query group 25](../../database/README.md#25-event-update-and-publish)).

**Open:** patch semantics, which fields may change, changing associate clubs, and publishing rules. Concurrent edits are last-write-wins until the edit form ships ([schema decision 6](../../database/docs/schema-review.md#decisions)). The legacy schema's `status` enum (`drafted`, `posted`, `archived`) has no `cancelled`; the [baseline schema](../../database/migrations/schema/2026_09_29_baseline_up.sql) has `draft`, `posted`, and `cancelled`.

**Proposed contract.** Need: **Now**, because resuming a draft on the New event form (`?draft=:eventId`, then "SAVE DRAFT" or "POST EVENT") has to update that draft instead of creating a second event ([M12](../coverage.md#m12-resuming-a-draft-creates-a-second-event)). The Event page's "CANCEL EVENT" also uses it now (the cancel half of [M10](../coverage.md#m10-event-status-vocabulary-and-cancelled)). The "EDIT EVENT" and Manage tab "EDIT" buttons use it too, but editing a posted event is **Later**.

- **Auth:** 🔴 JWT + e-board or owner of any linked club. `403` otherwise; `404` for an unknown or deleted event.
- **Request body:** any subset of the [create body](club-events.md#-post-clubsclubidevents)'s fields; omitted fields are unchanged. `status`, when present, is `posted` or `cancelled` and must be an [allowed transition](events.md#event-status) ([decision 2](../README.md#decisions)): `draft → posted`, `posted → cancelled`, or `cancelled → posted`. There's no way back to `draft`; saving a resumed draft sends its fields without `status` (or with `"status": "draft"` on a draft, which changes nothing). There's no `archived` status; removing an event is [`DELETE`](#-delete-autheventseventid). `associates`, when present, replaces the co-host list. Alt text isn't a field here: it's on the flyer's image row and changes with a new [flyer confirm](images.md#-post-clubsclubideventseventidthumbnailsconfirm).

  ```json
  { "event": { "title": "Example Event (moved)", "location": "Room 101" }, "status": "posted" }
  ```

  Cancelling sends only `{ "status": "cancelled" }`.
- **Response `200`:** the updated [event object](events.md#proposed-event-object), so the page can re-render without a second read:

  ```jsonc
  { "message": "Successfully updated event 42", "event": { /* event object */ } }
  ```

- **Errors:** `400` validation (same rules as create, including `location` when the result is `posted`) or a transition the table doesn't allow, such as `posted → draft` or `draft → cancelled`; `401`, `403`, `404`, `500`.

**Queries** (10; steps 2–8 in one transaction): 1. [`authorization/events/can_manage/IS_student_authorized_event.sql`](../../database/queries/authorization/events/can_manage/IS_student_authorized_event.sql) (auth, `403`) → 2. [`events/update/SELECT_event_for_update.sql`](../../database/queries/events/update/SELECT_event_for_update.sql) (read, locks the row; no row is `404`; the API merges the body) → 3. [`events/update/UPDATE_event.sql`](../../database/queries/events/update/UPDATE_event.sql) (write) → 4. [`events/update/UPSERT_event_description.sql`](../../database/queries/events/update/UPSERT_event_description.sql) (write, if `description` is sent) → 5. [`events/update/DELETE_event_associates.sql`](../../database/queries/events/update/DELETE_event_associates.sql) (write, if `associates` is sent) → 6. [`events/create/INSERT_event_club_link.sql`](../../database/queries/events/create/INSERT_event_club_link.sql) (write, once per associate, if `associates` is sent) → 7. [`events/update/UPDATE_event_status_posted.sql`](../../database/queries/events/update/UPDATE_event_status_posted.sql) (write, if `status` is `posted` and differs; 0 rows is `400`) → 8. [`events/update/UPDATE_event_status_cancelled.sql`](../../database/queries/events/update/UPDATE_event_status_cancelled.sql) (write, if `status` is `cancelled` and differs; 0 rows is `400`) → 9. [`events/read/SELECT_event.sql`](../../database/queries/events/read/SELECT_event.sql) (read, the response, `public_only` `FALSE`) → 10. [`events/images/list/SELECT_event_images.sql`](../../database/queries/events/images/list/SELECT_event_images.sql) (read, the response)

**Plan:** PDF "Endpoints Revamp", p. 18 · [query group 25](../../database/README.md#25-event-update-and-publish)

## 🔴 POST `/auth/events/{eventId}/thumbnails`

Returns a presigned upload URL for an event thumbnail (PDF: `?filetype=<FILETYPE>&filename=<FILENAME>`), followed by the [upload flow](images.md#how-image-uploads-work).

**Auth:** 🔴 JWT + e-board or owner of a linked club.

**Status:** ⬜ Not built. The signer exists at the public [`POST /clubs/{clubId}/events/{eventId}/thumbnails`](images.md#-post-clubsclubideventseventidthumbnails) with no auth, and there's no confirm step.

**Proposed contract** (Need: **Now**; screen: New event form flyer dropzone, called after the event is created or updated). Only one signer should exist: this route, or the club-scoped one with an authorizer added; which is an [open question](../README.md#open-questions).

- **Request body:** `{ "filename": "poster.png", "mimetype": "image/png" }`, as the built signer takes; not the PDF's query string.
- **Response `200`:** unchanged from the built signer, `{ "uploadUrl", "imageId", "objectKey" }`.
- **Next steps:** `PUT` the file to `uploadUrl`, then confirm with [`POST /clubs/{clubId}/events/{eventId}/thumbnails/confirm`](images.md#-post-clubsclubideventseventidthumbnailsconfirm) (or its `/auth/events` equivalent), which also records the alt text.

**Queries** (2): 1. [`authorization/events/can_manage/IS_student_authorized_event.sql`](../../database/queries/authorization/events/can_manage/IS_student_authorized_event.sql) (auth, `403`) → 2. [`events/read/SELECT_event_status.sql`](../../database/queries/events/read/SELECT_event_status.sql) (read, no row is `404`) → then the S3 presign (no query). Its confirm step is [`POST /clubs/{clubId}/events/{eventId}/thumbnails/confirm`](images.md#-post-clubsclubideventseventidthumbnailsconfirm).

**Plan:** PDF "Endpoints Revamp", p. 18 · [query group 27](../../database/README.md#27-club-and-event-thumbnail-metadata-assignment) for the confirm step

## 🔴 POST `/auth/events/{eventId}/images`

Returns a presigned upload URL for an event gallery image; the body includes `filename` and `mimetype` (PDF).

**Auth:** 🔴 JWT + e-board or owner of a linked club.

**Status:** ⬜ Not built. The signer and confirm steps exist at public club-scoped routes ([sign](images.md#-post-clubsclubideventseventidimages), [confirm](images.md#-post-clubsclubideventseventidimagesconfirm)) with no auth. The confirm step trusts client-supplied metadata and doesn't check the object exists in S3.

**Need: Later.** The New event form's "More photos for this event" row and the Event page's manager add-photo tile both show "SOON". When built, it takes the same body and returns the same shape as the built club-scoped signer.

**Queries** (2): 1. [`authorization/events/can_manage/IS_student_authorized_event.sql`](../../database/queries/authorization/events/can_manage/IS_student_authorized_event.sql) (auth, `403`) → 2. [`events/read/SELECT_event_status.sql`](../../database/queries/events/read/SELECT_event_status.sql) (read, no row is `404`) → then the S3 presign (no query). Its confirm step is [`POST /clubs/{clubId}/events/{eventId}/images/confirm`](images.md#-post-clubsclubideventseventidimagesconfirm).

**Plan:** PDF "Endpoints Revamp", p. 18 · [query group 15](../../database/README.md#15-event-image-metadata-confirmation)

## 🔴 DELETE `/auth/events/{eventId}`

Deletes an event. The PDF says "set status of event to archived"; [decision 2](../README.md#decisions) replaces that with a soft delete through `deleted_at`, and there's no `archived` status.

**Auth:** 🔴 JWT + e-board or owner of a linked club.

**Path params:** `eventId`.

**Status:** ⬜ Not built. No route or handler. The SQL exists in the database module: the soft delete, the restore, and the admin purge ([query groups 26 and 31](../../database/README.md#26-event-deletion)).

**Proposed contract** (Need: **Later**: the manager bar and Manage tab have no delete button). Deleting is separate from cancelling: a cancelled event stays listed, a deleted one disappears from every read, public and manager ([Event status](events.md#event-status)). An event in any status can be deleted. The delete sets `deleted_at` and doesn't touch `status`. For 30 days the event can be [restored](#-post-autheventseventidrestore); after that, an admin's [purge](admins.md#-post-adminspurge) removes it and its unused S3 files for good ([decisions 2, 8 and 9](../README.md#decisions)).

- **Auth:** 🔴 JWT + e-board or owner of any linked club.
- **Request body:** none.
- **Response `200`:** `{"message": "Successfully deleted event 42", "eventId": 42}`.
- **Errors:** `401`, `403`, `404` (unknown or already deleted), `500`.

**Queries** (2): 1. [`authorization/events/can_manage/IS_student_authorized_event.sql`](../../database/queries/authorization/events/can_manage/IS_student_authorized_event.sql) (auth, `403`) → 2. [`events/delete/UPDATE_event_soft_delete.sql`](../../database/queries/events/delete/UPDATE_event_soft_delete.sql) (write, 0 rows is `404`)

**Plan:** PDF "Endpoints Revamp", p. 18 · [query group 26](../../database/README.md#26-event-deletion)

## 🔴 POST `/auth/events/{eventId}/restore`

Restores a deleted event. **Proposed**; not in the PDF. It comes from [decision 9](../README.md#decisions): a deleted event can be restored for 30 days.

**Auth:** 🔴 JWT + e-board or owner of the event's **owning** club (`club_is_event_owner`), or an admin. That's narrower than the other `/auth/events` routes, which accept any linked club: an associate club can't bring back an event the owner deleted.

**Path params:** `eventId`.

**Status:** ⬜ Not built. The SQL exists: [`events/restore/`](../../database/queries/events/restore/) and the [owning-club check](../../database/queries/authorization/events/manages_owner_club/IS_student_owner_club_manager.sql) ([query group 26](../../database/README.md#26-event-deletion)).

**Proposed contract** (Need: **Later**: no screen lists deleted events or offers an undo yet).

- **Request body:** none.
- **Response `200`:** the restored event in the manager read's shape, so the page can show it again:

  ```jsonc
  { "message": "Successfully restored event 42", "event": { /* proposed event object */ } }
  ```

  The event comes back with the status it had when it was deleted (`draft`, `posted`, or `cancelled`), its club links, and its images: deleting never touches those, and the purge only removes events deleted more than 30 days ago.
- **Errors:**

  | Status | When |
  | --- | --- |
  | `401` | No token (API Gateway). |
  | `403` | Not on the owning club's e-board, not an owner, and not an admin. |
  | `404` | No such event, or it has been purged. |
  | `409` | The event isn't deleted. |
  | `410` | The event was deleted 30 or more days ago. It can't be restored, and the next purge removes it. |
  | `500` | Database failure. |

**Queries** (6): 1. [`authorization/events/manages_owner_club/IS_student_owner_club_manager.sql`](../../database/queries/authorization/events/manages_owner_club/IS_student_owner_club_manager.sql) or [`authorization/admins/is_admin/IS_admin.sql`](../../database/queries/authorization/admins/is_admin/IS_admin.sql) (auth, `403` unless either passes) → 2. [`events/restore/UPDATE_event_restore.sql`](../../database/queries/events/restore/UPDATE_event_restore.sql) (write) → on 0 rows: 3. [`events/restore/SELECT_event_deletion.sql`](../../database/queries/events/restore/SELECT_event_deletion.sql) (read, no row `404`, not deleted `409`, 30 days or more `410`) → otherwise: 4. [`events/read/SELECT_event.sql`](../../database/queries/events/read/SELECT_event.sql) (read, the response, `public_only` `FALSE`) → 5. [`events/images/list/SELECT_event_images.sql`](../../database/queries/events/images/list/SELECT_event_images.sql) (read, the response). The 30-day window is in the restore's `WHERE` clause, so one statement does the restore.

**Plan:** [decision 9](../README.md#decisions) · [query group 26](../../database/README.md#26-event-deletion)
