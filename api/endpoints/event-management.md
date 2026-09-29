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

Returns an event in any status (drafted, posted, and so on) with its description, for the event's managers (PDF: "protected for getting drafts").

**Auth:** 🔴 JWT + e-board or owner of a linked club.

**Path params:** `eventId`.

**Response:** not defined.

**Status:** ⬜ Not built. The public read and the event role check both exist but aren't combined into a route.

**Proposed contract** (Need: **Now**, for the New event form's `?draft=:eventId` prefill. The edit form behind "EDIT EVENT" and the Manage tab's "EDIT" would use it too, but that's **Later**.) Fixes the prefill half of [M1](../coverage.md#m1-drafts-in-public-lists).

- **Auth:** 🔴 JWT + e-board or owner of any linked club. `403` otherwise; `404` for an unknown event.
- **Response `200`:** the same envelope as the public read, with a [proposed event object](events.md#proposed-event-object) in any [status](events.md#event-status), including `draft` and `archived`:

  ```jsonc
  { "message": "Succesfully fetched event 42", "event": { /* event object */ } }
  ```

**Open:** which statuses managers may see, and whether an event's author gets any rights of their own.

**Plan:** PDF "Endpoints Revamp", p. 18 · [query group 24](../../database/README.md#24-authorized-event-read), reusing groups [11](../../database/README.md#11-public-and-composite-event-read) and [14](../../database/README.md#14-event-authorization)

## 🔴 GET `/auth/events/{eventId}/images`

Returns an event's images, including for drafts (PDF: "JWT so u can use on drafted events").

**Auth:** 🔴 JWT + e-board or owner of a linked club.

**Path params:** `eventId`.

**Response:** not defined.

**Status:** ⬜ Not built. The public club-scoped read ([`clubs/events/images/get/get.go`](../../infrastructure/legacy/lambda/api/clubs/events/images/get/get.go)) is missing its SQL, has no role check, and ignores `{clubId}`, so don't copy it for this route.

**Need: Later.** No screen manages gallery images yet; managers see images through `images` in [`GET /auth/events/{eventId}`](#-get-autheventseventid).

**Plan:** PDF "Endpoints Revamp", p. 18 · query groups [16](../../database/README.md#16-event-image-list) and [14](../../database/README.md#14-event-authorization)

## 🔴 PATCH `/auth/events/{eventId}`

Updates event fields from a JSON body. It's also how an event gets published; the PDF's example is `{ "status": "POSTED" }`. The Edit and Cancel buttons in the frontend's event manager bar have no endpoint behind them.

**Auth:** 🔴 JWT + e-board or owner of a linked club, plus field validation and a status-transition policy.

**Path params:** `eventId`.

**Request body (PDF):** JSON containing the fields to change.

**Response:** not defined.

**Status:** ⬜ Not built. No route, handler, update query, or transaction.

**Open:** patch semantics, which fields may change, concurrency control, changing associate clubs, and publishing rules. The frontend also has a `cancelled` status that the schema's `status` enum (`drafted`, `posted`, `archived`) doesn't have.

**Proposed contract.** Need: **Now**, because resuming a draft on the New event form (`?draft=:eventId`, then "SAVE DRAFT" or "POST EVENT") has to update that draft instead of creating a second event ([M12](../coverage.md#m12-resuming-a-draft-creates-a-second-event)). The Event page's "CANCEL EVENT" also uses it now (the cancel half of [M10](../coverage.md#m10-event-status-vocabulary-and-cancelled)). The "EDIT EVENT" and Manage tab "EDIT" buttons use it too, but editing a posted event is **Later**.

- **Auth:** 🔴 JWT + e-board or owner of any linked club. `403` otherwise; `404` for an unknown or archived event.
- **Request body:** any subset of the [create body](club-events.md#-post-clubsclubidevents)'s fields; omitted fields are unchanged. `status`, when present, is `posted` or `cancelled` and must be an [allowed transition](events.md#event-status) ([decision 2](../README.md#decisions)): `draft → posted`, `posted → cancelled`, or `cancelled → posted`. There's no way back to `draft`; saving a resumed draft sends its fields without `status` (or with `"status": "draft"` on a draft, which changes nothing). `archived` is only reachable through [`DELETE`](#-delete-autheventseventid). `associates`, when present, replaces the co-host list. Alt text isn't a field here: it's on the flyer's image row and changes with a new [flyer confirm](images.md#-post-clubsclubideventseventidthumbnailsconfirm).

  ```json
  { "event": { "title": "Example Event (moved)", "location": "Room 101" }, "status": "posted" }
  ```

  Cancelling sends only `{ "status": "cancelled" }`.
- **Response `200`:** the updated [event object](events.md#proposed-event-object), so the page can re-render without a second read:

  ```jsonc
  { "message": "Successfully updated event 42", "event": { /* event object */ } }
  ```

- **Errors:** `400` validation (same rules as create, including `location` when the result is `posted`) or a transition the table doesn't allow, such as `posted → draft` or `draft → cancelled`; `401`, `403`, `404`, `500`.

**Plan:** PDF "Endpoints Revamp", p. 18 · [query group 25](../../database/README.md#25-event-update-and-publish)

## 🔴 POST `/auth/events/{eventId}/thumbnails`

Returns a presigned upload URL for an event thumbnail (PDF: `?filetype=<FILETYPE>&filename=<FILENAME>`), followed by the [upload flow](images.md#how-image-uploads-work).

**Auth:** 🔴 JWT + e-board or owner of a linked club.

**Status:** ⬜ Not built. The signer exists at the public [`POST /clubs/{clubId}/events/{eventId}/thumbnails`](images.md#-post-clubsclubideventseventidthumbnails) with no auth, and there's no confirm step.

**Proposed contract** (Need: **Now**; screen: New event form flyer dropzone, called after the event is created or updated). Only one signer should exist: this route, or the club-scoped one with an authorizer added; which is an [open question](../README.md#open-questions).

- **Request body:** `{ "filename": "poster.png", "mimetype": "image/png" }`, as the built signer takes; not the PDF's query string.
- **Response `200`:** unchanged from the built signer, `{ "uploadUrl", "imageId", "objectKey" }`.
- **Next steps:** `PUT` the file to `uploadUrl`, then confirm with [`POST /clubs/{clubId}/events/{eventId}/thumbnails/confirm`](images.md#-post-clubsclubideventseventidthumbnailsconfirm) (or its `/auth/events` equivalent), which also records the alt text.

**Plan:** PDF "Endpoints Revamp", p. 18 · [query group 27](../../database/README.md#27-club-and-event-thumbnail-metadata-assignment) for the confirm step

## 🔴 POST `/auth/events/{eventId}/images`

Returns a presigned upload URL for an event gallery image; the body includes `filename` and `mimetype` (PDF).

**Auth:** 🔴 JWT + e-board or owner of a linked club.

**Status:** ⬜ Not built. The signer and confirm steps exist at public club-scoped routes ([sign](images.md#-post-clubsclubideventseventidimages), [confirm](images.md#-post-clubsclubideventseventidimagesconfirm)) with no auth. The confirm step trusts client-supplied metadata and doesn't check the object exists in S3.

**Need: Later.** The New event form's "More photos for this event" row and the Event page's manager add-photo tile both show "SOON". When built, it takes the same body and returns the same shape as the built club-scoped signer.

**Plan:** PDF "Endpoints Revamp", p. 18 · [query group 15](../../database/README.md#15-event-image-metadata-confirmation)

## 🔴 DELETE `/auth/events/{eventId}`

Archives an event (PDF: "set status of event to archived"), or deletes it, depending on the product's retention policy.

**Auth:** 🔴 JWT + e-board or owner of a linked club.

**Path params:** `eventId`.

**Status:** ⬜ Not built. No route, handler, archive or delete query, or cleanup.

**Open:** the schema supports both `status = 'archived'` and `deleted_at`, and public reads don't filter on `deleted_at`. Pick one lifecycle before building this.

**Proposed contract** (Need: **Later**: the manager bar and Manage tab have no delete button). Archiving is separate from cancelling: a cancelled event stays listed, an archived one disappears from every public read ([Event status](events.md#event-status)). Any status can be archived, and nothing leaves `archived` ([decision 2](../README.md#decisions)).

- **Auth:** 🔴 JWT + e-board or owner of any linked club.
- **Request body:** none.
- **Response `200`:** `{"message": "Successfully archived event 42", "eventId": 42}`.
- **Errors:** `401`, `403`, `404` (unknown or already archived), `500`.

**Plan:** PDF "Endpoints Revamp", p. 18 · [query group 26](../../database/README.md#26-event-deletion)
