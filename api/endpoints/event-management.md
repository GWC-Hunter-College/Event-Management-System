# Event management (`/auth/events`)

Planned protected routes for an event's managers. The PDF's revamp moved these off `/clubs/{clubId}/events/{eventId}` because an event can belong to several clubs, so authorization shouldn't depend on a single `clubId` in the path (PDF p. 18). The PDF's original list had them as `GET`, `PATCH`, and `DELETE /clubs/{clubId}/events/{eventId}` plus `.../images` and `.../thumbnails` (pp. 11–12).

**None of these routes exist yet.** Some of the behavior exists elsewhere, with no auth:

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

**Response:** not defined. **Proposed:** start from the public [event object](events.md#event-object).

**Status:** ⬜ Not built. The public read and the event role check both exist but aren't combined into a route.

**Open:** which statuses managers may see, and whether an event's author gets any rights of their own.

**Plan:** PDF "Endpoints Revamp", p. 18 · [query group 24](../../database/README.md#24-authorized-event-read), reusing groups [11](../../database/README.md#11-public-and-composite-event-read) and [14](../../database/README.md#14-event-authorization)

## 🔴 GET `/auth/events/{eventId}/images`

Returns an event's images, including for drafts (PDF: "JWT so u can use on drafted events").

**Auth:** 🔴 JWT + e-board or owner of a linked club.

**Path params:** `eventId`.

**Response:** not defined.

**Status:** ⬜ Not built. The public club-scoped read ([`clubs/events/images/get/get.go`](../../infrastructure/legacy/lambda/api/clubs/events/images/get/get.go)) is missing its SQL, has no role check, and ignores `{clubId}`, so don't copy it for this route.

**Plan:** PDF "Endpoints Revamp", p. 18 · query groups [16](../../database/README.md#16-event-image-list) and [14](../../database/README.md#14-event-authorization)

## 🔴 PATCH `/auth/events/{eventId}`

Updates event fields from a JSON body. It's also how an event gets published; the PDF's example is `{ "status": "POSTED" }`. The Edit and Cancel buttons in the frontend's event manager bar have no endpoint behind them yet.

**Auth:** 🔴 JWT + e-board or owner of a linked club, plus field validation and a status-transition policy.

**Path params:** `eventId`.

**Request body (PDF):** JSON containing the fields to change.

**Response:** not defined.

**Status:** ⬜ Not built. No route, handler, update query, or transaction.

**Open:** patch semantics, which fields may change, concurrency control, changing associate clubs, and publishing rules. The frontend also has a `cancelled` status that the schema's `status` enum (`drafted`, `posted`, `archived`) doesn't have.

**Plan:** PDF "Endpoints Revamp", p. 18 · [query group 25](../../database/README.md#25-event-update-and-publish)

## 🔴 POST `/auth/events/{eventId}/thumbnails`

Returns a presigned upload URL for an event thumbnail (PDF: `?filetype=<FILETYPE>&filename=<FILENAME>`), followed by the [upload flow](images.md#how-image-uploads-work).

**Auth:** 🔴 JWT + e-board or owner of a linked club.

**Status:** ⬜ Not built. The signer exists at the public [`POST /clubs/{clubId}/events/{eventId}/thumbnails`](images.md#-post-clubsclubideventseventidthumbnails) with no auth, and there's no confirm step.

**Plan:** PDF "Endpoints Revamp", p. 18 · [query group 27](../../database/README.md#27-club-and-event-thumbnail-metadata-assignment) for the confirm step

## 🔴 POST `/auth/events/{eventId}/images`

Returns a presigned upload URL for an event gallery image; the body includes `filename` and `mimetype` (PDF).

**Auth:** 🔴 JWT + e-board or owner of a linked club.

**Status:** ⬜ Not built. The signer and confirm steps exist at public club-scoped routes ([sign](images.md#-post-clubsclubideventseventidimages), [confirm](images.md#-post-clubsclubideventseventidimagesconfirm)) with no auth. The confirm step trusts client-supplied metadata and doesn't check the object exists in S3.

**Plan:** PDF "Endpoints Revamp", p. 18 · [query group 15](../../database/README.md#15-event-image-metadata-confirmation)

## 🔴 DELETE `/auth/events/{eventId}`

Archives an event (PDF: "set status of event to archived"), or deletes it, depending on the product's retention policy.

**Auth:** 🔴 JWT + e-board or owner of a linked club.

**Path params:** `eventId`.

**Status:** ⬜ Not built. No route, handler, archive or delete query, or cleanup.

**Open:** the schema supports both `status = 'archived'` and `deleted_at`, and public reads don't filter on `deleted_at`. Pick one lifecycle before building this.

**Plan:** PDF "Endpoints Revamp", p. 18 · [query group 26](../../database/README.md#26-event-deletion)
