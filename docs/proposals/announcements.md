# Announcements: what's left

**Proposed, 2026-09-30.** The schema, queries, and eight endpoint contracts already exist; this proposal covers what they don't: the Announcements tab itself, and whether announcements get images, tags, and notifications. New and changed routes are ⬜ and **Proposed**.

**What exists** (don't reopen): the `announcements` and `announcements_to_clubs` tables and the `announcement_details` view in the [baseline](../../database/migrations/schema/2026_09_29_baseline_up.sql), [query group 32](../../database/README.md#32-announcements), and the contracts in [api/endpoints/announcements.md](../../api/endpoints/announcements.md): `draft → posted` only, `posted_at` ordering, soft delete with a 30-day restore and the admin purge, and multi-club announcements with one owner club.

**Depends on:** the core API's club role check and `GET /me/clubs`; [notifications](notifications.md) phase 1 for "a new post from your club". Images depend on the image confirm flow ([images.md](../../api/endpoints/images.md#how-image-uploads-work)).

## Problem

The club page's Announcements tab shows "SOON" ([coverage.md](../../api/coverage.md#clubclubid-club)). E-boards post updates in group chats and on Instagram today, where members who aren't following miss them, and the GWC website has nowhere to read them from ([decision 7](../../api/README.md#decisions)).

**Who uses it and how:**

- **Anyone, signed in or not** reads a club's posted announcements on its page, newest first. The GWC website reads the same public list for its configured club.
- **Members** get told about new posts ([Notifications](#notifications)) and, later, see posts from all their clubs on My Clubs.
- **E-board and owners** write drafts, post them, edit them, delete them, and restore one deleted by mistake.

## The Announcements tab

### Public and member view

```text
┌ Girls Who Code @ Hunter ─────────────────────────────────────────┐
│ EVENTS   ANNOUNCEMENTS   BOARD   MANAGE                           │
│                                                                   │
│ HunterHacks registration is open                    SEP 28, 2026 │
│ Girls Who Code @ Hunter · with Hunter CS Club                    │
│ Sign up by Friday. Teams of up to four.                          │
│ https://hunterhacks.example  (linked)                   edited    │
│ ───────────────────────────────────────────────────────────────── │
│ First meeting moved to room 1001                     SEP 3, 2026 │
│ …                                                                 │
│                        [ LOAD MORE ]                              │
└───────────────────────────────────────────────────────────────────┘
```

| Element | Source | Notes |
| --- | --- | --- |
| List, newest first | [`GET /clubs/{clubId}/announcements`](../../api/endpoints/announcements.md#-get-clubsclubidannouncements) | Pages of 20 (`limit=20&page=n`); "Load more" asks for the next page. |
| Title | `title` | Up to 255 characters, as the column. |
| Date | `postedAt` | Shown as a date in `America/New_York`, like event dates. Relative ("2h ago") for the last day. |
| Clubs | `owners.owner`, `owners.associates` | "Girls Who Code @ Hunter · with Hunter CS Club" when it's shared. Each name links to its club. |
| Body | `body` | Plain text. Line breaks kept, `http(s)` URLs made into links, nothing else rendered. Long bodies collapse after about eight lines with "Show more". |
| "edited" | `updatedAt` vs `postedAt` | Shown when `updatedAt` is more than a minute after `postedAt`. |
| Empty state | empty list | "No announcements yet." For managers: "Post your first announcement." |
| Error | any failure | "We couldn't load announcements." with a retry; the rest of the club page still works. |

**No author name** is shown. Posts speak for the club, `student_info` names are usually empty, and the list is public, including on the GWC website. `authorId` stays in the object for managers.

### Manager view

Shown when [`GET /me/clubs`](../../api/endpoints/me.md#-get-meclubs) says the viewer is `eboard` or `owner` of this club.

- **"+ NEW ANNOUNCEMENT"** at the top of the tab opens the composer.
- **Drafts** section above the posted list, from [`GET /clubs/{clubId}/announcements/drafts`](../../api/endpoints/announcements.md#-get-clubsclubidannouncementsdrafts), each with "Edit" and "Post". Hidden when there are none.
- **On each posted announcement,** a menu with "Edit" and "Delete". Delete confirms: "Delete this announcement? You can restore it for 30 days."
- **Recently deleted** (collapsed, at the bottom): announcements deleted in the last 30 days, each with "Restore". This needs a list route that doesn't exist yet: [`GET /clubs/{clubId}/announcements/deleted`](#-get-clubsclubidannouncementsdeleted). Events have the same gap: [restore](../../api/endpoints/event-management.md#-post-autheventseventidrestore) exists, but nothing lists deleted events.

### Composer

A page at `/club/:clubId/announcement/new`, and `/club/:clubId/announcement/new?draft=:announcementId` to resume a draft, following the New event form ([coverage.md](../../api/coverage.md#clubclubideventnew-new-event-form)).

| Field | Rule | Sent as |
| --- | --- | --- |
| Title | Required, 1–255 characters. | `announcement.title` |
| Body | Optional on a draft, required to post. Plain text, up to **5,000 characters** (a new API limit; the `TEXT` column allows about 64 KB). | `announcement.body` |
| "Also post to" | Other clubs. Phase 4, together with events' co-hosting picker, which is SOON too. | `associates` |
| Image | Phase 3 ([Images](#images)). | the image confirm route |

Buttons: **SAVE DRAFT** (create or `PATCH` without `status`) and **POST** (create with `"status": "posted"`, or `PATCH` a draft with it). A live preview shows the card as members will see it. A resumed draft always uses [`PATCH`](../../api/endpoints/announcements.md#-patch-authannouncementsannouncementid), never a second create; that's the lesson of [M12](../../api/coverage.md#m12-resuming-a-draft-creates-a-second-event).

### One announcement

A page at `/announcement/:announcementId`, like `/event/:eventId`, reading [`GET /announcements/{announcementId}`](../../api/endpoints/announcements.md#-get-announcementsannouncementid). Notifications and "Share" link here, and a deleted or draft announcement shows "We couldn't find that announcement". Managers can open drafts here through [`GET /auth/announcements/{announcementId}`](../../api/endpoints/announcements.md#-get-authannouncementsannouncementid).

### Elsewhere

- **My Clubs:** "From your clubs" can interleave announcements with events. That needs [`GET /me/announcements`](#-get-meannouncements) (phase 2).
- **Home and Discover:** no announcements. A campus-wide feed would need moderation and, at that point, [tags](#tags).

## Images

**Recommendation: yes, one optional image per announcement, in phase 3. No gallery.**

An announcement often is a flyer ("GBM Thursday, free pizza"), so one image carries most of the value. A gallery belongs to events, which already have one. The baseline comment anticipated exactly this: a nullable `fk_thumbnail_id` to `images`, like an event's flyer.

```sql
-- Phase 3. An announcement's image, as events.fk_thumbnail_id. A use of a shared file,
-- so RESTRICT, which is what the in-use checks rely on (I4).
ALTER TABLE `announcements`
  ADD COLUMN `fk_thumbnail_id` CHAR(36) NULL AFTER `fk_author_id`,
  ADD KEY `idx_announcements_thumbnail` (`fk_thumbnail_id`),
  ADD CONSTRAINT `fk_announcements_thumbnail` FOREIGN KEY (`fk_thumbnail_id`) REFERENCES `images` (`id`);
```

- **Upload:** the same presign → `PUT` → confirm flow as a flyer: [`POST /auth/announcements/{announcementId}/thumbnails`](#-post-authannouncementsannouncementidthumbnails-and-confirm) and its `/confirm`. The object key is `announcements/{announcementId}/thumbnails/{uuid}.{ext}`, the image's `purpose` is `announcement-thumbnail`, its owning club is the announcement's owner club, and `alt_text` is required when posting with an image.
- **Reuse** (decision 1): the confirm route also accepts `{"imageId": "…"}` with no upload, to reuse a file the owner club already has (an event's flyer, say). The API checks the image isn't deleted and its owning club is one the caller manages.
- **Removing it:** [`DELETE /auth/announcements/{announcementId}/thumbnail`](#-delete-authannouncementsannouncementidthumbnail) clears the pointer and releases the file (soft-deletes it only if nothing else uses it).
- **Reads:** `announcement_details` gains `thumbnail_object_key` and `alt_text`, and the object gains `thumbnailUrl` and `altText`, the same names events use ([proposed event object](../../api/endpoints/events.md#proposed-event-object)).

**Changes to the image queries**, so the rules in [I4](../../database/docs/schema-review.md#i4-deleting-an-image) keep holding:

| Query | Change |
| --- | --- |
| [`UPDATE_image_soft_delete_if_unused.sql`](../../database/queries/images/release/UPDATE_image_soft_delete_if_unused.sql), [`SELECT_purgeable_images.sql`](../../database/queries/images/purge/SELECT_purgeable_images.sql) | One more `NOT EXISTS` for `announcements.fk_thumbnail_id`. |
| Takedown ([`DELETE /images/{imageId}`](../../api/endpoints/images.md#-delete-imagesimageid)) | A new `UPDATE_announcements_clear_thumbnail.sql` in its transaction, next to the logo and flyer clears. |
| Purge | A new `UPDATE_images_of_purgeable_announcements.sql` before [`DELETE_purgeable_announcements.sql`](../../database/queries/announcements/purge/DELETE_purgeable_announcements.sql), mirroring the event version, and that file's "no images yet" comment goes. |
| Invariant | "A referenced image never has `deleted_at` set, except through an event **or announcement** that is itself soft-deleted." |

**Storage:** if a third of posts carry an image of about 2 MB, 50 active clubs posting weekly add ~120 MB a month, about $0.003 a month more in S3 each month. Display should use the resized renditions the [board](bulletin-board.md#images-and-storage) needs, not the original.

## Tags

**Recommendation: no tags.** Announcements are read on one club's page or in one student's feed, where there's nothing to filter: a club posts a few a week. Tags would add a picker to the composer, a table, and a filter UI for no reader. Event tags are unused for the same reason ([H2](../../database/docs/schema-review.md#h2-event-tags)).

If a campus-wide feed arrives, add `announcement_tags (fk_announcement_id, tag)` with `tag` referencing the fixed [`topics`](../../database/docs/schema-review.md#c1-club-topics) list, cascading from `announcements`, as the baseline comment sketches. A **pinned** announcement (kept at the top of the tab) is likely more useful than tags; see [Open questions](#open-questions).

## Notifications

**Recommendation: yes,** with [notifications](notifications.md) phase 1, type `announcement.posted`.

- **When:** an announcement becomes `posted`, either created as posted or moved from draft. The `Notify` insert runs in the same transaction, after the status write, so a post that fails notifies nobody.
- **Who:** members of every role in the **owner club**, plus members of **associate clubs on whose e-board or ownership the author also sits**, deduplicated, minus the author. The e-board and owners of any other associate club get `announcement.shared` instead. Without that rule, any club's e-board could push notifications to another club's members by listing it as an associate, and associate links need no consent from the other club today. The same question applies to event co-hosts.
- **Not on edits.** Editing a posted announcement doesn't notify again. Deleting hides its notifications; restoring shows them again, without new ones.
- **Email** (notifications phase 3): on by default for this type, subject to that proposal's open question.

The recipients query, in the style of the existing files:

```sql
-- notifications/create/INSERT_announcement_posted.sql (proposed)
INSERT INTO notifications (fk_recipient_id, type, fk_club_id, fk_actor_id, fk_announcement_id)
SELECT DISTINCT cm.fk_student_id, 'announcement.posted', ?, ?, ?
FROM club_members cm
WHERE cm.fk_club_id IN (/* owner club, and associates the author manages */)
  AND cm.fk_student_id <> ?;
```

## Endpoints

New and changed routes. The eight existing contracts are in [announcements.md](../../api/endpoints/announcements.md).

| Access | Method | Path | What it does | Role | Phase |
| --- | --- | --- | --- | --- | --- |
| 🔴 | POST | `/clubs/{clubId}/announcements` | **Changed:** notifies when created as `posted`; body limit 5,000. | E-board or owner of `clubId` | 1 |
| 🔴 | PATCH | `/auth/announcements/{announcementId}` | **Changed:** notifies on `draft → posted`; body limit 5,000. | E-board or owner of any club it belongs to | 1 |
| 🔴 | GET | `/me/announcements` | Posted announcements from the caller's clubs, newest first. | Any signed-in student | 2 |
| 🔴 | GET | `/clubs/{clubId}/announcements/deleted` | Announcements deleted in the last 30 days, for restoring. | E-board or owner of `clubId`, or an admin | 2 |
| 🔴 | POST | `/auth/announcements/{announcementId}/thumbnails` | Presigned upload URL for the image. | E-board or owner of any club it belongs to | 3 |
| 🔴 | POST | `/auth/announcements/{announcementId}/thumbnails/confirm` | Attaches an uploaded or reused image. | Same | 3 |
| 🔴 | DELETE | `/auth/announcements/{announcementId}/thumbnail` | Removes the image. | Same | 3 |

### 🔴 GET `/me/announcements`

**Auth:** 🔴 JWT. Only announcements of clubs the caller is a member of (any role).

**Query params:** `limit` (default 20, `1`–`50`), `page`.

**Response `200` (Proposed):** the list envelope of [announcement objects](../../api/endpoints/announcements.md#proposed-announcement-object), posted only, by `postedAt` descending. An announcement shared by two of the caller's clubs appears once.

**Queries (proposed):** `me/announcements/SELECT_my_announcements.sql` (read, from `announcement_details` joined to `announcements_to_clubs` and the caller's `club_members`).

### 🔴 GET `/clubs/{clubId}/announcements/deleted`

**Auth:** 🔴 JWT + e-board or owner of `clubId`, or an admin. Limited to announcements whose **owner** club is `clubId`, because only that club's managers can [restore](../../api/endpoints/announcements.md#-post-authannouncementsannouncementidrestore).

**Response `200` (Proposed):** the list envelope, each object with `deletedAt` and `restorableUntil` (`deletedAt` + 30 days). Ordered by `deletedAt` descending.

**Queries (proposed):** `EXISTS_club.sql` (`404`) → the club or admin check (`403`) → `clubs/announcements/deleted/SELECT_club_deleted_announcements.sql` (read from the base table, since the view hides deleted rows; `deleted_at > CURRENT_TIMESTAMP - INTERVAL 30 DAY`, the same hard-coded window as the restore).

### 🔴 POST `/auth/announcements/{announcementId}/thumbnails` and `/confirm`

The event flyer contracts, with announcement paths ([signer](../../api/endpoints/images.md#-post-clubsclubideventseventidthumbnails), [confirm](../../api/endpoints/images.md#-post-clubsclubideventseventidthumbnailsconfirm)).

- **Signer body:** `{ "filename", "mimetype" }`; response `{ "uploadUrl", "imageId", "objectKey" }`. `mimetype` must be `image/jpeg`, `image/png`, or `image/webp`.
- **Confirm body:** either `{ "filename", "mimetype", "imageId", "objectKey", "altText" }` after an upload, or `{ "imageId", "altText" }` to reuse a file. The confirm locks the announcement, inserts the image (upload only), sets `fk_thumbnail_id`, and releases the replaced image, in one transaction.
- **Response `200`:** `{ "message": "Image saved", "announcementId": 7, "thumbnailUrl": "<readable-url>" }`.

### 🔴 DELETE `/auth/announcements/{announcementId}/thumbnail`

Clears `fk_thumbnail_id` and runs the release query on the old file, in one transaction. `200` `{ "message": "Image removed", "announcementId": 7 }`; `404` if there's no image.

## Authorization

Unchanged from [announcements.md](../../api/endpoints/announcements.md#status): reading posted ones is 🟢; managing is the e-board and owners of **any** club an announcement belongs to; restoring is the **owning** club's e-board and owners, or an admin. New here: `GET /me/announcements` is scoped to the caller's memberships, and the deleted list follows the restore rule.

## Rough AWS cost

Negligible. The tab's reads are one API call per club page view plus "load more"; at 10,000 club page views a month that's nothing extra on a server, or about $0.02 of API Gateway and Lambda serverless. Images add the ~$0.003 a month of storage above, plus reads through signed URLs. Notifications are costed in [notifications.md](notifications.md#rough-aws-cost).

## Changes to existing tables and endpoints

- **Tables:** none until phase 3, which adds `announcements.fk_thumbnail_id` and two columns to `announcement_details`.
- **Endpoints:** `POST /clubs/{clubId}/announcements` and `PATCH /auth/announcements/{announcementId}` gain the `Notify` step (their Queries lines each grow by one) and the 5,000-character body limit.
- **Queries:** the image queries in [Images](#images), in phase 3.
- **Docs:** coverage.md's Announcements tab row moves from **Later** to **Now** when the frontend starts on it; the "Future work" note in announcements.md is answered.
- **Frontend:** the tab, composer, and single-announcement page; routes `/club/:clubId/announcement/new` and `/announcement/:announcementId`.

## Open questions

1. **Body format:** plain text with linked URLs (recommended), or Markdown with a sanitizer?
2. **Associate consent:** should listing another club require that club's acceptance, for announcements and for events? Until it does, the [notification rule](#notifications) keeps other clubs' members out.
3. **Pinned announcements:** a `pinned_at` column so a club can keep one post on top?
4. **Author display:** show "Posted by {name}" to members, or never?
5. **Scheduled posts** ("post Monday at 9"): they'd need a scheduled job, which [decision 8](../../api/README.md#decisions) avoided for the purge.
6. **Limits:** is 5,000 characters right for a body?

## Phased plan

| Phase | Scope | Size |
| --- | --- | --- |
| **1. First version** | The tab (public list, drafts, composer, edit, delete), the 5,000-character limit, and `announcement.posted` notifications. All on existing routes and queries, plus one notification insert. No images, no associates picker. | Small: frontend-heavy; 1 new query. |
| 2 | `/announcement/:id` page, `GET /me/announcements` on My Clubs, and "Recently deleted" with `GET /clubs/{clubId}/announcements/deleted`. | Small: 2 routes, 2 queries. |
| 3 | One image per announcement: column, view, three routes, image-query changes. | Medium. |
| 4 | "Also post to" picker, together with event co-hosting, after the consent question is settled. | Small once decided. |
