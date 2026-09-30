# Club bulletin board

**Proposed, 2026-09-30.** Nothing here is built. Every route is ⬜ and **Proposed**, in the [api/ reference](../../api/README.md#legend) style.

**Design:** Figma file `i19JIkiEiAfqoTPNBUUcKe`, page "Bulletin Board — concept", [node 4277:2](https://www.figma.com/design/i19JIkiEiAfqoTPNBUUcKe/GWC-Club-Website--Copy-?node-id=4277-2). Read for this proposal through the Figma MCP:

| Frame | Node | What it shows |
| --- | --- | --- |
| Desktop / Board — Girls Who Code (today) | `4277:4` | The board: 19 pins (sticky notes, captioned polaroids, flyers, event tickets) overlapping on a grid. Header actions **REVIEW · 3**, share, **+ ADD TO THE BOARD**. A timeline with a play button, one tick per pin from Fall '24 to Fall '26, and "Showing Today · 19 pins since Sep 2024". |
| Desktop / Board — member adding a photo | `4283:60` | The "Add to the board" panel: 1 Pick a photo, 2 Place and rotate it on the board, 3 Send it to the e-board to review. An optional caption, "Or add a note instead of a photo", and "The e-board reviews every pin before it appears. You'll get a notification when it's up." The photo sits on the board with resize handles and the tag "DRAG TO PLACE · YOURS GOES ON TOP". |
| Desktop / Board — e-board reviewing submissions | `4283:336` | The review drawer, "3 waiting": each with a thumbnail or NOTE badge, its caption, "A member · 2h ago", APPROVE and DECLINE. The selected one is previewed in place, tagged "PENDING · SUBMITTED BY A MEMBER". Footer: "Approved pins go on top with today's date. Declined ones are never shown; the member gets a short note." |
| Mobile / Board — Girls Who Code (today) | `4284:136` | "GWC BOARD, 19 pins · since Sep 2024". The board is larger than the screen: "PINCH TO ZOOM · DRAG TO LOOK AROUND", a minimap, a compact timeline (SEP '24 → NOW), and a **+ ADD** button. |
| Explainer / How the board works | `4285:145` | The four steps, "Later, not now", and "What each pin needs to store" (quoted under [Data model](#data-model)). |

The page also has "Desktop / Board — rewound to May 2025 (timeline)" (`4282:29`), which wasn't in the list to read.

**Depends on, and doesn't exist yet** (details in [Dependencies](#dependencies)): **member photo uploads**, **albums**, and **notifications**, plus image renditions and working event-gallery uploads.

## Problem

A club's photos today live in event galleries (when those work) and on members' phones. There's no place that feels like the club's own: what it did, who showed up, the inside jokes. The Board tab on the club page shows "The club board is coming soon" ([coverage.md](../../api/coverage.md#clubclubid-club)).

The board is one shared collage per club, separate from photo albums. From the explainer: "Photos is the organized archive of albums; the board is a collage the whole club builds together, one pin at a time."

**Who uses it and how:**

1. **The e-board pins** event tickets and photos straight from the club's albums, and notes of their own. No review.
2. **Members add**: pick a photo or write a note, drag it where they want it, rotate and resize it, and send it for review.
3. **The e-board approves or declines** each submission from a review queue, previewing it in place. Approved pins go on top with the approval date. Declined ones are never shown, and the member gets a short note.
4. **It stacks up.** New pins cover old ones. Anyone can drag the timeline to see the board on any date, or press play to watch it grow: the board at date *t* is the approved pins whose `approved_at` is before *t*, in stack order.

**Later, not now** (from the design): boards live on club pages until clubs fill them up; then a "Popular boards" row on Discover (or a nav item); and the timeline stays hidden until a board has at least a semester of pins.

## Concepts

**The canvas.** Every board is a fixed logical canvas of **1600 × 1000 board units**, about the proportions of the desktop board. Positions are stored in board units, never pixels: desktop scales the canvas to fit its width, and mobile shows part of it with pan, zoom, and a minimap. One scale factor (`pixels = units × viewportWidth / 1600`) converts in both directions.

**A pin's geometry.** `x` and `y` are the pin's center; `width` and `height` are its box; `rotation` is degrees clockwise about the center. Center-based coordinates keep rotation independent of size. The design's single "size" becomes a box, because a photo's height depends on its aspect ratio, and `images` doesn't store dimensions ([I2](../../database/docs/schema-review.md#i2-owner-uploader-alt-text) deferred them). The client knows the ratio when the pin is placed, so it sends both.

**Kinds and styles.** The Figma layers name four looks: `note`, `polaroid`, `flyer`, and `ticket`. Those are three kinds of content and a presentation choice:

| Kind | Content | Styles | Who can add it |
| --- | --- | --- | --- |
| `photo` | An existing `images` row, plus an optional caption ("kickoff! 32 of us") | `polaroid` (framed, with the caption), `flyer` (bare poster) | E-board; members (from phase 2) |
| `note` | Up to 280 characters | `yellow`, `lilac` | E-board; members (from phase 2) |
| `ticket` | An event the club hosts or co-hosts; the API renders its current title and date | — | E-board only |

Tape, pushpin colors, and the handwritten caption font are decoration: the client derives them from the pin id, so they're stable without being stored.

**Stack order.** Each approved pin gets the next number on its board when it's approved (for e-board pins, when they're added). The number only grows, so "new pins cover old ones" and "approved pins go on top" are the same rule. Pending and declined pins have none. A removed pin keeps its number, so restoring it puts it back where it was.

**The timeline** needs no route of its own. [`GET /clubs/{clubId}/board`](#-get-clubsclubidboard) returns every approved pin with `approvedAt` and `stackOrder`. The client draws one tick per pin, filters `approvedAt <= t` while scrubbing, and replays by adding pins in stack order. Semester labels are computed in `America/New_York`. Removed pins disappear from every date: moderation beats history.

## Data model

The explainer's list, "club · image or note text · position (x, y) · rotation · size · stack order · added by · submitted at · status (pending / approved / declined) · approved at", maps to columns like this:

| Design field | Column |
| --- | --- |
| club | `fk_club_id` |
| image or note text | `kind`, with `fk_image_id`, `note_text`, or `fk_event_id` (tickets) |
| position (x, y) | `x`, `y` |
| rotation | `rotation` |
| size | `width`, `height` |
| stack order | `stack_order` |
| added by | `fk_added_by`, and `added_as` (whether it skipped review) |
| submitted at | `submitted_at` |
| status | `status` |
| approved at | `approved_at` |
| (the short note) | `decline_note`, with `fk_reviewed_by` and `reviewed_at` |
| (e-board removal) | `deleted_at` |

### DDL sketch

In the baseline's conventions, so it could go into the baseline while there's no live database:

```sql
-- Board pins: each club's shared collage. One board per club, so a pin names its
-- club and there's no boards table. Positions are board units on a fixed 1600 x 1000
-- canvas: x and y are the pin's center, width and height its box, rotation degrees
-- clockwise about the center.
-- kind picks the content. A photo points at an images row, reused and never copied.
-- A note carries note_text. A ticket points at an event. The CHECK keeps exactly one.
-- status is the review. approved_at and stack_order are set together on approval,
-- and stack_order only grows, so new pins cover old ones. added_as records whether
-- the pin skipped review (eboard) or went through it (member).
-- deleted_at is a removal by the e-board: hidden everywhere, restorable for 30 days.
CREATE TABLE `board_pins` (
  `id` INT NOT NULL AUTO_INCREMENT,
  `fk_club_id` INT NOT NULL,
  `kind` ENUM('photo', 'note', 'ticket') NOT NULL,
  `fk_image_id` CHAR(36) NULL,
  `fk_event_id` INT NULL,
  `note_text` VARCHAR(280) NULL,
  `caption` VARCHAR(80) NULL,
  `style` VARCHAR(20) NULL,
  `x` SMALLINT NOT NULL,
  `y` SMALLINT NOT NULL,
  `width` SMALLINT NOT NULL,
  `height` SMALLINT NOT NULL,
  `rotation` DECIMAL(4,1) NOT NULL DEFAULT 0,
  `stack_order` INT NULL,
  `status` ENUM('pending', 'approved', 'declined') NOT NULL DEFAULT 'pending',
  `added_as` ENUM('eboard', 'member') NOT NULL,
  `fk_added_by` CHAR(36) NULL,
  `submitted_at` TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
  `fk_reviewed_by` CHAR(36) NULL,
  `reviewed_at` TIMESTAMP NULL,
  `approved_at` TIMESTAMP NULL,
  `decline_note` VARCHAR(280) NULL,
  `deleted_at` TIMESTAMP NULL,
  PRIMARY KEY (`id`),
  UNIQUE KEY `uq_board_pins_club_stack` (`fk_club_id`, `stack_order`),
  KEY `idx_board_pins_club_status` (`fk_club_id`, `status`, `submitted_at`),
  KEY `idx_board_pins_image` (`fk_image_id`),
  KEY `idx_board_pins_event` (`fk_event_id`),
  KEY `idx_board_pins_added_by` (`fk_added_by`),
  KEY `idx_board_pins_reviewed_by` (`fk_reviewed_by`),
  KEY `idx_board_pins_deleted` (`deleted_at`),
  CONSTRAINT `chk_board_pins_content` CHECK (
    (`kind` = 'photo' AND `fk_image_id` IS NOT NULL AND `fk_event_id` IS NULL AND `note_text` IS NULL)
    OR (`kind` = 'note' AND `note_text` IS NOT NULL AND `fk_image_id` IS NULL AND `fk_event_id` IS NULL)
    OR (`kind` = 'ticket' AND `fk_event_id` IS NOT NULL AND `fk_image_id` IS NULL AND `note_text` IS NULL)
  ),
  CONSTRAINT `chk_board_pins_approved` CHECK ((`status` = 'approved') = (`approved_at` IS NOT NULL AND `stack_order` IS NOT NULL)),
  CONSTRAINT `chk_board_pins_reviewed` CHECK ((`status` = 'pending') = (`reviewed_at` IS NULL)),
  CONSTRAINT `chk_board_pins_rotation` CHECK (`rotation` BETWEEN -180 AND 180),
  CONSTRAINT `fk_board_pins_club` FOREIGN KEY (`fk_club_id`) REFERENCES `clubs` (`id`),
  CONSTRAINT `fk_board_pins_image` FOREIGN KEY (`fk_image_id`) REFERENCES `images` (`id`),
  CONSTRAINT `fk_board_pins_event` FOREIGN KEY (`fk_event_id`) REFERENCES `events` (`id`) ON DELETE CASCADE,
  CONSTRAINT `fk_board_pins_added_by` FOREIGN KEY (`fk_added_by`) REFERENCES `students` (`id`) ON DELETE SET NULL,
  CONSTRAINT `fk_board_pins_reviewed_by` FOREIGN KEY (`fk_reviewed_by`) REFERENCES `students` (`id`) ON DELETE SET NULL
);
```

**Foreign keys** ([F1](../../database/docs/schema-review.md#f1-on-delete-behavior)):

- `fk_image_id` is `RESTRICT`: a pin is a **use** of a shared file, like `event_images`, so the database refuses to delete a file a pin still shows. That's the in-use check [I4](../../database/docs/schema-review.md#i4-deleting-an-image) relies on.
- `fk_event_id` is `ON DELETE CASCADE`: only the purge hard-deletes an event, and a ticket for an event that's gone for good has nothing to show, so it goes with it. `SET NULL` isn't an option: MySQL refuses `SET NULL` and `ON UPDATE CASCADE` on a column a `CHECK` uses (error 3823, tested), and a ticket with no event would break `chk_board_pins_content` anyway.
- `fk_club_id` is `RESTRICT` until club deletion is designed. The two student columns are attribution, `SET NULL`.

**Indexes.** The board read filters `(fk_club_id, status)` and orders by `stack_order`; `uq_board_pins_club_stack` serves the order and guarantees no two pins share a place. The review queue reads `(fk_club_id, status = 'pending')` by `submitted_at`. The image and event indexes serve the in-use checks.

**A read view,** `board_pin_details`, like the other three ([V1](../../database/docs/schema-review.md#v1-read-views)): approved, not removed pins, joined to the image's `object_key` and `alt_text` and to `event_details` for tickets. A ticket whose event is soft-deleted drops out of the view, and comes back if the event is restored. Every public read selects from it, so those rules are written once.

**Assigning stack order.** In the approve (or e-board add) transaction: lock the club row (`SELECT id FROM clubs WHERE id = ? FOR UPDATE`), take `COALESCE(MAX(stack_order), 0) + 1` for the club, and write it with `approved_at`. The lock makes approvals on one board take turns; the UNIQUE key is the backstop.

**Row size.** About 200 bytes a pin. A busy board with 300 pins is 60 KB of rows.

## Endpoints

All Proposed, all ⬜. **Member** means any `club_members` row for the club; **e-board** means `role IN ('eboard', 'owner')`.

| Access | Method | Path | What it does | Role | Phase |
| --- | --- | --- | --- | --- | --- |
| 🟢 | GET | `/clubs/{clubId}/board` | Approved pins in stack order, for the board and its timeline. | Public | 1 |
| 🔴 | GET | `/clubs/{clubId}/photos` | Photos the caller may pin: the picker. | Member (public photos only) or e-board (all the club's photos) | 1 |
| 🔴 | POST | `/clubs/{clubId}/board/pins` | Adds a pin. From the e-board it's approved at once; from a member it's pending. | E-board (phase 1); member (phase 2) | 1 |
| 🔴 | DELETE | `/clubs/{clubId}/board/pins/{pinId}` | Removes a pin (soft delete), or withdraws the caller's own pending pin. | E-board or admin; the submitter for their own pending pin | 1 |
| 🔴 | POST | `/clubs/{clubId}/board/pins/{pinId}/restore` | Restores a pin removed less than 30 days ago. | E-board or admin | 1 |
| 🔴 | GET | `/clubs/{clubId}/board/review` | The pending queue, oldest first, with the count for **REVIEW · n**. | E-board | 2 |
| 🔴 | POST | `/clubs/{clubId}/board/pins/{pinId}/approve` | Approves a pending pin, optionally nudging its placement. Notifies the member. | E-board | 2 |
| 🔴 | POST | `/clubs/{clubId}/board/pins/{pinId}/decline` | Declines a pending pin with a short note. Notifies the member. | E-board | 2 |
| 🔴 | GET | `/clubs/{clubId}/board/pins/mine` | The caller's pending pins, and declined ones from the last 30 days with their notes. | Member | 2 |
| 🔴 | POST | `/clubs/{clubId}/board/uploads` | Presigned upload URL for a member's own photo. | Member | 3 |
| 🔴 | POST | `/clubs/{clubId}/board/uploads/confirm` | Records the uploaded photo; returns its `imageId` for a pin. | Member | 3 |

**Why the review count isn't on the board read.** On 🟢 routes API Gateway passes no claims ([Auth](../../api/README.md#sending-a-token)), so the public board can't tell who's looking. The e-board's count comes from the 🔴 review route, and a member's own pending pins from `/mine`. The frontend shows the member's pending pins in place, tagged "pending", only to them.

### 🟢 GET `/clubs/{clubId}/board`

**Auth:** 🟢 Public, like event galleries and announcements: every pin on it was put there or approved by the e-board. Whether the board should be members-only is an [open question](#open-questions).

**Response `200` (Proposed):**

```json
{
  "message": "Successfully fetched the board",
  "board": { "clubId": 2, "width": 1600, "height": 1000, "pinCount": 19, "firstApprovedAt": "2024-09-10T16:00:00Z" },
  "pins": [
    {
      "id": 101, "kind": "note", "noteText": "Welcome to our board! Pin your favorite GWC moments ✦", "style": "yellow",
      "addedAs": "eboard", "x": 230, "y": 190, "width": 240, "height": 230, "rotation": -4.0,
      "stackOrder": 1, "approvedAt": "2024-09-10T16:00:00Z"
    },
    {
      "id": 102, "kind": "photo", "imageUrl": "<signed-rendition-url>", "fullImageUrl": "<signed-original-url>",
      "altText": "Members at the fall kickoff", "caption": "kickoff! 32 of us", "style": "polaroid",
      "addedAs": "member", "x": 480, "y": 590, "width": 270, "height": 290, "rotation": 2.5,
      "stackOrder": 2, "approvedAt": "2024-10-15T19:30:00Z"
    },
    {
      "id": 110, "kind": "ticket",
      "event": { "id": 42, "title": "Hack Night", "startDate": "2025-11-14T23:00:00Z", "timezone": "America/New_York", "status": "posted" },
      "addedAs": "eboard", "x": 520, "y": 800, "width": 240, "height": 100, "rotation": 1.0,
      "stackOrder": 11, "approvedAt": "2025-11-14T15:00:00Z"
    }
  ]
}
```

- Every approved, non-removed pin, in `stackOrder`. No paging: a board is bounded (see the pin cap question), and the timeline needs them all.
- No submitter or reviewer in the public read. `addedAs` lets the client sign e-board notes "– the e-board", as the design does.
- `firstApprovedAt` drives "since Sep 2024" and the later rule that hides the timeline until a board has a semester of pins.
- `imageUrl` is a signed URL for a resized rendition, `fullImageUrl` for the original ([Images and storage](#images-and-storage)). Both expire in about an hour (schema [decision 7](../../database/docs/schema-review.md#decisions)).

**Errors:** `400` bad `clubId`; `404` unknown club; `500`.

**Queries (proposed):** `EXISTS_club.sql` (`404`) → `clubs/board/list/SELECT_board_pins.sql` (read from `board_pin_details`, ordered by `stack_order`).

### 🔴 GET `/clubs/{clubId}/photos`

The "Pick a photo" step. Until albums exist, "the club's photos" are the images already attached to it:

- the flyers and gallery images of events linked to the club (`events_to_clubs`), and
- images the club owns (`images.fk_club_id`): its board uploads (phase 3) and announcement images.

Members see only what's public anyway: photos of `posted` or `cancelled` events, and approved board photos. The e-board also sees photos of drafts. Soft-deleted images and events never appear.

**Auth:** 🔴 JWT + member of `clubId` (`403` otherwise).

**Query params:** `eventId` (optional: one event's photos), `limit` (default 60, max 100), `before` (cursor).

**Response `200` (Proposed):** `{ "message": "…", "photos": [ { "imageId", "thumbnailUrl", "altText", "event": { "id", "title", "startDate" } | null, "addedAt" } ] }`, newest first, grouped by event on screen.

**Queries (proposed):** `EXISTS_club.sql` → `IS_club_member.sql` and `IS_student_authorized_club.sql` → `clubs/photos/list/SELECT_club_photos.sql` (read, with a `members_only_public` flag).

### 🔴 POST `/clubs/{clubId}/board/pins`

**Auth:** 🔴 JWT + member of `clubId`. In phase 1, e-board only. E-board pins are approved at once (`added_as = 'eboard'`); member pins are `pending` (`added_as = 'member'`).

**Request body (Proposed):**

```json
{ "kind": "photo", "imageId": "11111111-2222-3333-4444-555555555555", "caption": "my first GWC meeting!", "style": "polaroid",
  "x": 1180, "y": 560, "width": 260, "height": 300, "rotation": 3.5 }
```

| Field | Rule |
| --- | --- |
| `kind` | `photo`, `note`, or `ticket`. Members can't pin tickets. |
| `imageId` | `photo` only. Must be in the caller's [`/photos`](#-get-clubsclubidphotos) set (or, phase 3, the caller's own upload for this board). |
| `noteText` | `note` only. 1–280 characters, plain text. |
| `eventId` | `ticket` only. An event linked to the club, `posted` or `cancelled`. |
| `caption` | `photo` only, optional, up to 80 characters. |
| `style` | Optional; one of the styles for the kind. |
| `x`, `y` | `0`–`1600` and `0`–`1000`. |
| `width`, `height` | `60`–`600` each. |
| `rotation` | `-45`–`45` in the API (the `CHECK` allows ±180). |

**Limits:** a member may have at most 3 pending pins per board (`409` beyond that), so the queue can't be flooded.

**Response `200` (Proposed):** `{ "message": "Pin sent for review", "pin": { …, "status": "pending" } }`, or `"Pin added"` with `"status": "approved"` and its `stackOrder` for the e-board.

**Errors:** `400` validation (with `"errors": [{"field", "message"}]`), `401`, `403` not a member (or not e-board, for a ticket or in phase 1), `404` unknown club, image, or event, `409` pending limit, `500`.

**Writes (one transaction):** e-board: lock the club row → next stack order → insert approved. Member: insert pending → `Notify` `board.pin_submitted` to the club's e-board, grouped per club ([notifications](notifications.md#types)).

### 🔴 GET `/clubs/{clubId}/board/review`

**Auth:** 🔴 JWT + e-board of `clubId`.

**Response `200` (Proposed):** `{ "message": "…", "count": 3, "pending": [ { "id", "kind", "imageUrl" | "noteText", "caption", "x", "y", "width", "height", "rotation", "submittedAt", "submittedBy": { "studentId", "email", "firstName", "lastName" } } ] }`, oldest first. `submittedBy` follows [decision 6](../../api/README.md#decisions): the e-board may see members' emails. Names are usually `null` because nothing writes `student_info`, which is why the design says "A member".

### 🔴 POST `/clubs/{clubId}/board/pins/{pinId}/approve`

**Auth:** 🔴 JWT + e-board of `clubId`.

**Request body (Proposed, optional):** `{ "x", "y", "width", "height", "rotation" }`, if the reviewer nudges it while previewing in place.

**Writes (one transaction):** lock the pin (`SELECT … FOR UPDATE`; `404` unknown or not on this board; `409` not pending) → lock the club row → next stack order → `UPDATE … SET status = 'approved', approved_at = CURRENT_TIMESTAMP, reviewed_at = CURRENT_TIMESTAMP, fk_reviewed_by = ?, stack_order = ?` (guarded by `status = 'pending'`) → `Notify` `board.pin_approved` to the submitter.

**Response `200` (Proposed):** `{ "message": "Pin approved", "pin": { … } }`.

### 🔴 POST `/clubs/{clubId}/board/pins/{pinId}/decline`

**Auth:** 🔴 JWT + e-board of `clubId`.

**Request body (Proposed):** `{ "note": "Please keep photos on topic" }`, 1–280 characters. The UI offers a few canned notes and free text.

**Writes (one transaction):** `UPDATE … SET status = 'declined', reviewed_at = CURRENT_TIMESTAMP, fk_reviewed_by = ?, decline_note = ? WHERE id = ? AND fk_club_id = ? AND status = 'pending'` (0 rows: `404` or `409`) → `Notify` `board.pin_declined`. The pin is never shown on the board. A photo the member uploaded for it stays until the [purge](#images-and-storage).

### 🔴 DELETE `/clubs/{clubId}/board/pins/{pinId}` and `/restore`

- **The e-board (or an admin)** removes any pin: sets `deleted_at`. It disappears from the board and the timeline. `POST …/restore` within 30 days brings it back at its old stack order; `409` if not removed, `410` if 30 days or more, as for events.
- **A member** can withdraw their own **pending** pin: it's deleted outright (it was never shown, so there's nothing to restore), and a photo they uploaded only for it is released in the same transaction.

### Phase 3: member uploads

`POST /clubs/{clubId}/board/uploads` and `/confirm` follow the [image upload flow](../../api/endpoints/images.md#how-image-uploads-work): the signer returns `{ uploadUrl, imageId, objectKey }` with a key under `clubs/{clubId}/board/`; the confirm checks the object exists, inserts an `images` row with `purpose = 'board-photo'`, the club as owner, the member as uploader, and their alt text, and returns `imageId`. The member then creates the pin with it. JPEG, PNG and WebP only, 10 MB at most. iPhone HEIC photos are converted in the browser before upload or refused.

## Authorization

| Who | Can |
| --- | --- |
| Anyone | See approved pins and replay the timeline. |
| Member of the club | Pick from the club's public photos; submit photos (phase 2: existing ones; phase 3: their own) and notes; see and withdraw their pending pins; see their declined pins' notes. |
| E-board or owner | Everything a member can; pin tickets, photos (including drafts' photos) and notes with no review; review, approve (with a nudge), decline with a note; remove and restore any pin. |
| Admin | Remove and restore any pin; take down any image ([`DELETE /images/{imageId}`](../../api/endpoints/images.md#-delete-imagesimageid)), which removes its pins. |
| Non-member | Read only. `403` on every write. |

Moving or resizing an approved pin isn't in the design and isn't proposed: the timeline would then need a history of positions. Remove and re-pin instead.

## Images and storage

**Reuse, not copies.** A photo pin points at an existing `images` row (schema [decision 1](../../database/docs/schema-review.md#decisions)): "Pin images can reuse album photos, so nothing is uploaded twice." Phases 1 and 2 store no new files at all. Phase 3's member uploads are ordinary `images` rows owned by the club, so the e-board can take them down, and a later export finds them.

**Changes to existing image and purge queries,** so I4's rules keep holding with pins as a new kind of use:

| Query | Change |
| --- | --- |
| [`UPDATE_image_soft_delete_if_unused.sql`](../../database/queries/images/release/UPDATE_image_soft_delete_if_unused.sql), [`SELECT_purgeable_images.sql`](../../database/queries/images/purge/SELECT_purgeable_images.sql) | One more `NOT EXISTS (SELECT 1 FROM board_pins p WHERE p.fk_image_id = i.id)`. Any pin row counts, removed or declined too, because the `RESTRICT` key would refuse the delete anyway. |
| Takedown ([`DELETE /images/{imageId}`](../../api/endpoints/images.md#-delete-imagesimageid)) | A new `DELETE_board_pins_using_image.sql` in its transaction, beside the logo, flyer and gallery clears. A takedown removes every use, pins included, outright; their notifications cascade. |
| [`UPDATE_images_of_purgeable_events.sql`](../../database/queries/events/purge/UPDATE_images_of_purgeable_events.sql) | One more `NOT EXISTS` for board pins, so a purged event's photo that's pinned stays alive on the board. |
| Event purge | No new step: [`DELETE_purgeable_events.sql`](../../database/queries/events/purge/DELETE_purgeable_events.sql) removes a purged event's tickets through `fk_event_id`'s `ON DELETE CASCADE`, and their notifications cascade from the pins. Its header comment gains a line saying so. |
| Pin purge (new) | In the purge's first transaction: `UPDATE_images_of_purgeable_pins.sql` (soft-delete files whose only uses are pins removed or declined more than 30 days ago, taking the pin's removal or decline time), then `DELETE_purgeable_pins.sql`. The image batches that follow delete those files, as they do for events. |

The invariant stays as written: a pending, approved, or removed pin keeps its file alive, and only the purge hands it over.

**Renditions are needed.** A board shows 20–60 photos at once, each a few hundred pixels wide. Today every read would sign the original upload, a 2–4 MB phone photo, so one board view would move 50–150 MB. Proposed: an S3-triggered resize Lambda, outside the VPC because it only needs S3, writes a 640-pixel JPEG beside each upload (`{object_key}.w640.jpg`, about 60–120 KB). The API signs the rendition for board tiles, gallery tiles, and announcement cards, and the original for full view. No schema change: the rendition key is derived from `object_key`, and the purge deletes both objects. Galleries, the Home hero, and announcement images benefit too.

## Rough AWS cost

Using the [shared assumptions](README.md#scale-and-price-assumptions), with about 3,000 board views a month and 40 photos per view.

| Item | Monthly |
| --- | --- |
| Rows | Under 1 MB a year. $0. |
| Board reads (API Gateway + Lambda) | 3,000 requests: under $0.01. |
| S3 GETs for images | 120,000 × $0.0004 per 1,000: ~$0.05. |
| Data transfer, **with renditions** (~100 KB each) | ~12 GB: inside the 100 GB monthly free allowance, else ~$1.10. |
| Data transfer, **without renditions** (~2.5 MB each) | ~300 GB: ~$18, and slow pages on phones. |
| Resize Lambda | ~5,000 new images × ~1 GB-s: inside the free tier, else ~$0.10. |
| Rendition storage | +3–5% of image storage: cents. |
| Phase 3 uploads | ~50 active boards × 30 photos × 2.5 MB = ~3.75 GB a semester: ~$0.09 a month more each semester. |
| **Total** | **Under $1 a month with renditions** |

## Dependencies

What the board relies on that doesn't exist yet:

| Dependency | State today | Needed from | Stand-in until then |
| --- | --- | --- | --- |
| **Notifications** | ⬜ [Proposed](notifications.md) | Phase 2: "your pin is up", "declined", "waiting for review" | Phase 1 has no review, so it needs none. |
| **Member photo uploads** | ⬜ None. Every upload route is for managers (and has no auth at all today); nothing checks type, size, or content. | Phase 3 | Phase 2 members pick photos the club already has. |
| **Albums** (the Photos tab) | ⬜ Not designed. The only photo collections are event galleries. | Phase 4 | [`GET /clubs/{clubId}/photos`](#-get-clubsclubidphotos) over flyers and event galleries. |
| **Event gallery uploads** | 🟨/⬜ The confirm fails on warm Lambdas, the list route is broken, and the protected routes are Later ([images.md](../../api/endpoints/images.md)). | Phase 1, to have photos worth pinning | Flyers, which are Needed now. |
| **Image renditions** | ⬜ | Phase 1 | A small pilot can live with originals. |
| **Student display names** | `student_info` exists, nothing writes it | The review queue | "A member", as the design shows, and the email for the e-board. |
| **Role checks in handlers** | 🟨 The queries exist; no handler calls them | Everything | — |

## Changes to existing tables and endpoints

- **Tables:** none altered. `board_pins` and `board_pin_details` are new; `notifications` gains `fk_board_pin_id` ([notifications.md](notifications.md#ddl-sketch)).
- **Queries:** the five image and purge changes in [Images and storage](#images-and-storage).
- **Endpoints:** [`DELETE /images/{imageId}`](../../api/endpoints/images.md#-delete-imagesimageid) and [`POST /admins/purge`](../../api/endpoints/admins.md#-post-adminspurge) gain steps; the purge's response gains `pinsPurged`. No other existing route changes.
- **`images.purpose`:** a new value, `board-photo` (phase 3).
- **Frontend:** the Board tab replaces SOON; coverage.md's Board row gets endpoints.
- **Export:** approved pins join the export format later ([import-export.md](import-export.md#phased-plan)).

## Open questions

1. **Public or members-only?** Proposed public, like galleries. The design places it under "My Community".
2. **Does the e-board see who submitted?** Proposed yes (email, per decision 6), for moderation; the design shows "A member".
3. **Is the decline note required?** Proposed yes, with canned choices.
4. **Can members pin event tickets?** Proposed no.
5. **When is a board "full"?** The design's "Later" depends on it. A cap of 300 approved pins?
6. **Photos of other people:** a member may upload a photo of classmates. Is the e-board's review enough, or should there be a "report this pin" flow for anyone in a photo?
7. **Leaving the club:** withdraw that member's pending pins automatically? Approved pins stay.
8. **Unverified clubs:** do they get a board?
9. **Sharing a date:** should the share button link to a rewound board (`?at=2025-05-01`)? That's client-only.
10. **Popular boards:** ranked by pins approved in the last 30 days? That's one 🟢 read over `board_pins` when the time comes.

## Phased plan

| Phase | Scope | Size |
| --- | --- | --- |
| **1. First version: the e-board's board** | `board_pins` and its view; the board read (🟢), the photo picker, e-board pins (photos from flyers and galleries, notes, tickets, approved at once), remove and restore; the image and purge query changes; renditions; the Board tab with the timeline and replay. No review, no member input, no notifications. | Medium: 1 table, 5 routes, ~10 queries, 1 resize Lambda. |
| 2 | Member submissions of existing club photos and notes; the review drawer; approve, decline, `/mine`; the three `board.*` notifications. | Medium. Needs notifications phase 1. |
| 3 | Member photo uploads with type and size checks. | Small once the upload flow is protected. |
| 4 | Albums, when a Photos tab exists; the "Popular boards" row; hiding the timeline until a semester of pins; a nav item. | — |
