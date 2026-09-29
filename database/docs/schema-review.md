# Schema review

A review of the November schema ([`history/11_04_2025_create_core_tables_up.sql`](../migrations/history/11_04_2025_create_core_tables_up.sql)) against the API contracts in [api/README.md](../../api/README.md) and the frontend's needs in [api/coverage.md](../../api/coverage.md), and the design of the schema that replaces it.

**The database is new.** There's no live database: AWS isn't running and the legacy system isn't deployed, so there are no rows anywhere to preserve or convert. The result of this review is therefore not a migration but one baseline that creates the whole schema from scratch: [`schema/2026_09_29_baseline_up.sql`](../migrations/schema/2026_09_29_baseline_up.sql), with [`_down.sql`](../migrations/schema/2026_09_29_baseline_down.sql), which drops everything. The older schema files are in [`migrations/history/`](../migrations/history/) as a read-only record and are never applied. Once a real database exists, every change goes in a new dated migration after the baseline.

Each finding gives what the November schema had, the problem, and what the baseline does. Finding IDs (S1, T1, I1, …) match the comments in the baseline where they apply.

## Decisions

Taken on 2026-09-29, in answer to this review's questions.

1. **Images: one table for stored files, with uses as references.** `images` is "a file we stored". Club logos, event flyers, gallery photos, and later board pins reference a row, and one file may have several uses. There are no UNIQUE constraints on the uses. Each image has an owning club, an uploader, and alt text ([I1](#i1-one-images-table-uses-as-references), [I2](#i2-owner-uploader-alt-text)).
2. **Times are stored in UTC.** There's no existing data, so nothing is converted ([T1](#t1-store-start-and-end-in-utc)).
3. **Topic keys keep their spaces** (`women in stem`), not slugs ([C1](#c1-club-topics)).
4. **`events.deleted_at` is the delete marker, and there's no `archived` status** ([S2](#s2-deleting-an-event)).
   - Statuses are `draft`, `posted`, and `cancelled`: the life of an event that still exists.
   - `DELETE /auth/events/{eventId}` sets `deleted_at` to now, and every read filters on `deleted_at IS NULL`. Nothing hard-deletes on a request.
   - Rows deleted long enough ago are hard-deleted later, with their S3 files. (Decisions 8 and 9 set how: an admin's purge, and a 30-day window.)
   - `images` has a `deleted_at` too, so the same purge cleans up S3 files.
   - api/README.md's [decision 2](../../api/README.md#decisions) and the [event status docs](../../api/endpoints/events.md#event-status) are updated to match.
5. **Deleting an image** (Later). Removing a use (take a photo out of a gallery, replace a logo or flyer) deletes only the reference, and the file goes once nothing references it. A takedown by the owning club or an admin removes every reference, then the file ([I4](#i4-deleting-an-image)).
6. **Concurrent edits to an event:** last write wins for now. Add a `version` column when the edit form ships ([E1](#e1-updating-a-draft-in-place-m12)).
7. **Image URLs:** signed GET URLs with about a one-hour expiry. A CDN can come later ([I5](#i5-from-row-to-url-m6)).

Follow-up decisions, also taken on 2026-09-29, in answer to the queries' open questions. They're [API decisions 8–12](../../api/README.md#decisions).

8. **Purging is done by admins, on demand.** `POST /admins/purge` (🔴, admin) runs the purge queries. There's no scheduled job. It only removes events and images deleted more than 30 days ago ([S2](#s2-deleting-an-event)).
9. **Deleted events can be restored for 30 days.** `POST /auth/events/{eventId}/restore` (🔴, e-board or owner of the event's owning club, or an admin) clears `deleted_at` only when the event was deleted less than 30 days ago. The purge uses the same cutoff, so nothing restorable is ever purged ([S2](#s2-deleting-an-event)).
10. **Date-only list bounds cover whole days in `America/New_York`**, with the end date inclusive (an exclusive bound at 00:00 the next day), and an event matches when it overlaps the range. Full timestamps with an offset are accepted too. Stored times stay UTC ([T1](#t1-store-start-and-end-in-utc)).
11. **One role per member:** `club_members.role ENUM('member', 'eboard', 'owner') NOT NULL DEFAULT 'member'` replaces the two flags. The e-board list is `role IN ('eboard', 'owner')`. A club's last owner can't be demoted and can't leave ([M2](#m2-one-role-per-member)).
12. **Admins:** admins add and remove other admins, but the last admin can't be removed, including by themselves. A developer creates the first admin with a one-off SQL insert ([database/README.md](../README.md#the-first-admin)).

Still open: whether an event's author, or an admin, gets management rights beyond the linked clubs' e-board and owners.

## Summary

| ID | Finding | In the baseline | Endpoints |
| --- | --- | --- | --- |
| [I1](#i1-one-images-table-uses-as-references) | One `images` table; single-use UNIQUEs block reuse | No UNIQUE on any use; plain indexes | Thumbnail confirms, event reads, later pins |
| [I2](#i2-owner-uploader-alt-text) | No owning club, uploader, or alt text | `fk_club_id`, `fk_uploaded_by`, `alt_text` | Thumbnail confirms, event reads, `DELETE /images` |
| [I3](#i3-required-image-columns) | Every image column nullable; `object_key` not unique | `NOT NULL`, `UNIQUE (object_key)` | Every read that returns a URL |
| [I4](#i4-deleting-an-image) | No deletion semantics | `images.deleted_at`; release, takedown, and purge queries | `DELETE /images`, logo and flyer replacement |
| [I5](#i5-from-row-to-url-m6) | Reads return `object_key` or placeholders, not URLs | No schema change; views return every key an event needs | Every club and event read |
| [S1](#s1-one-status-vocabulary) | `drafted` vs `draft`; no `cancelled`; nullable | `ENUM('draft', 'posted', 'cancelled') NOT NULL DEFAULT 'draft'` | Every event read, create, `PATCH`, drafts list |
| [S2](#s2-deleting-an-event) | `archived` and `deleted_at` overlap | `deleted_at` only; 30-day restore; admin purge | Every event read, `DELETE /auth/events`, restore, `POST /admins/purge` |
| [S3](#s3-allowed-transitions) | Transitions not enforced anywhere | Enforced in each status `UPDATE` | `PATCH /auth/events` |
| [T1](#t1-store-start-and-end-in-utc) | Wall-clock `DATETIME` read as browser-local time | UTC `DATETIME` | Every event read and write |
| [T2](#t2-timestamp-defaults) | `created_at`/`updated_at` have no defaults | Defaults and `ON UPDATE` | Drafts list order, `PATCH` |
| [C1](#c1-club-topics) | No club topics | `topics`, `club_tags` | `GET /clubs`, `GET /clubs/{clubId}`, `POST /clubs`, `PATCH /clubs` |
| [C2](#c2-member-count) | No member count | Derived with `COUNT` in `club_details` | `GET /clubs`, `GET /clubs/{clubId}` |
| [C3](#c3-editing-a-club) | Editing a club | No change needed | `PATCH /clubs/{clubId}` (Later) |
| [E1](#e1-updating-a-draft-in-place-m12) | Updating a draft in place | Covered by T2 and M1 | `PATCH /auth/events/{eventId}` |
| [M1](#m1-missing-primary-keys) | `verified_clubs`, `admins`, `event_descriptions` have no primary key | Primary keys | Verified filter, admin check, event reads, `PATCH` |
| [M2](#m2-one-role-per-member) | Two nullable role flags | One `role` column | Leave club, role checks, member list, `GET /me/clubs` |
| [M3](#m3-one-owner-club-per-event) | Nothing stops two owner clubs per event | Functional UNIQUE index | Every event read (`owners.owner`) |
| [M4](#m4-other-nullable-columns) | Other nullable columns | `NOT NULL` where every write sets a value | Club and event reads |
| [K1](#k1-indexes) | Indexes | `events (status, start_date)` and named indexes for every foreign key | Event lists, drafts list, member count |
| [F1](#f1-on-delete-behavior) | No `ON DELETE` actions | The target policy, applied | Admin purge; future club and account deletion |
| [V1](#v1-read-views) | Every event read repeats the same joins | Views `event_details`, `club_details`, and `announcement_details` | Every club, event, and announcement read |
| [A1](#a1-announcements) | No announcements | `announcements`, `announcements_to_clubs` | The Announcements tab (Later) |
| [H1](#h1-event-edit-history), [H2](#h2-event-tags) | Older designs: edit history, event tags | Note only | Not covered by the API yet |

## Images

### I1. One `images` table, uses as references

**November schema.** `images` records each stored file (`id`, `purpose`, `object_key`, `filename`, `mimetype`, `created_at`). Uses point at it: `clubs.fk_logo_id`, `events.fk_thumbnail_id`, and the `event_images` link table for galleries. Each of the three was UNIQUE, so a file could have at most one use of each kind.

**Why one table.** Keep one table meaning "a file we stored", and express how an image is used through references (decision 1):

- Every purpose has the same lifecycle: sign an upload, `PUT` to S3, confirm. One insert ([`INSERT_image.sql`](../queries/images/create/INSERT_image.sql)), one URL function over `object_key` (I5), and one place to authorize, export, and garbage-collect files.
- A table per image type would duplicate the file columns and give three URL paths. Reuse (a board pin showing an event photo) would then mean copying the row, and copying the S3 object or sharing a key between rows. Sharing a key breaks deletion: removing one row's object removes the other's.
- What differs by use belongs on the reference, not the file: gallery order (`event_images.created_at`), a pin's note, or an alt-text override if a use ever needs its own.
- Single-valued uses stay pointer columns (`fk_logo_id`, `fk_thumbnail_id`), so the database enforces "at most one flyer" for free. Multi-valued uses stay link tables (`event_images`, later board pins).

**Baseline.** No UNIQUE constraint on any use, only the plain index each foreign key needs. The primary key `(fk_event_id, fk_image_id)` still prevents duplicates within one gallery. Because one file can now sit in several galleries, `event_images` has its own `created_at`, which orders each gallery. Tested: one image serves in two events' galleries, and deleting it is refused while it's referenced.

**Endpoints:** `POST /clubs/{clubId}/thumbnails/confirm`, `POST /clubs/{clubId}/events/{eventId}/thumbnails/confirm` (Now); gallery confirm and reads (Later); future board pins.

### I2. Owner, uploader, alt text

- **Owning club (`fk_club_id`).** Authorization needs it: may this caller take the file down? It's also how a per-club export finds a club's files. Every upload has a club: the logo's club, or the event's owner club. Nullable, for a future upload with no club (an avatar, an admin banner).
- **Uploader (`fk_uploaded_by`).** The confirm routes are 🔴, so the caller's `sub` is available. `ON DELETE SET NULL`: removing a student keeps their uploads.
- **Alt text (`alt_text VARCHAR(1000)`).** Needed by the flyer confirm and `altText` on event reads. Alt text describes the image's content, so it stays with the file across reuse; if one use ever needs different wording, add an override on that reference.
- **`deleted_at`.** See I4.
- **`purpose`.** "Which upload flow created the file" (`club-thumbnail`, `event-thumbnail`, `event-image`), set once at confirm. Display decisions come from the references, never from `purpose`: a gallery photo pinned to a board still says `event-image`. Values aren't constrained; a new upload flow brings its own migration anyway.
- `width`, `height`, and byte size would help layout and limits. Deferred until the upload flow checks files.

**Endpoints:** the thumbnail confirms (Now), event reads' `altText` (Now), `DELETE /images/{imageId}` and export (Later).

### I3. Required image columns

A row without `object_key` can't produce a URL, and two rows with the same key break the one-row-per-file rule: deleting either would delete the object the other still uses. So `purpose`, `object_key`, and `mimetype` are `NOT NULL`, `object_key` is UNIQUE, and `created_at` defaults to the current time. `filename` stays nullable; a server-generated file may not have one.

**Endpoints:** every read that returns `thumbnailUrl` or `images` (Now).

### I4. Deleting an image

Decision 5, with `images.deleted_at` from decision 4. Foreign keys from uses to `images` are `RESTRICT`, so the database refuses to delete a file that's still referenced, and does the reference count for free.

- **Removing a use** (replace a logo or flyer; later, take a photo out of a gallery or unpin it): remove the reference, then soft-delete the file only if nothing else references it ([`UPDATE_image_soft_delete_if_unused.sql`](../queries/images/release/UPDATE_image_soft_delete_if_unused.sql)). Both thumbnail confirms do this for the file they replace.
- **Takedown** (`DELETE /images/{imageId}`, by the owning club's managers or an admin; for example someone asks to be taken out of a photo): in one transaction, clear every logo and flyer pointer, delete every gallery link, then soft-delete the file ([`images/delete/`](../queries/images/delete/)). The file stops being served at once, because no read reaches it any more, and signed URLs already handed out expire within the hour (decision 7).
- **Invariant:** a referenced image never has `deleted_at` set, except through an event that is itself soft-deleted.
- **Purge** (`POST /admins/purge`, decision 8): once a file has been deleted for more than 30 days, hard-deletes the row, then deletes the S3 object. See [S2](#s2-deleting-an-event) for the order.
- **Failed object deletes** leave an orphan object, which is harmless: the row is already gone, so nothing points at it. The row is deleted first so a failure can never leave a row pointing at a missing object.
- If edit history (H1) should show past flyers, a replaced flyer must be kept while a revision references it. The release query's reference check handles that once revisions reference images.

**Endpoints:** `DELETE /images/{imageId}` (Later), logo and flyer replacement in the confirms (Now).

### I5. From row to URL (M6)

No schema change. Store only `object_key`, never a URL: signed URLs expire, and a CDN domain differs by environment. The API's storage adapter turns each `object_key` into a signed GET URL with about a one-hour expiry (decision 7) when it builds the response. The `event_details` view (V1) returns, for every event:

- the flyer's `object_key` and `alt_text`;
- the owner club's name and logo `object_key`, and each associate club's name and logo `object_key`;
- `image_count`: gallery images other than the flyer, since the flyer may also sit in the gallery.

[`SELECT_event_images.sql`](../queries/events/images/list/SELECT_event_images.sql) returns the gallery keys, flyer excluded, in `event_images.created_at` order, for `GET /events/{eventId}` and the gallery routes. The event object keeps `thumbnailId`.

**Endpoints:** `GET /events`, `GET /events/{eventId}`, `GET /clubs/{clubId}/events`, `GET /me/events`, `GET /clubs`, `GET /clubs/{clubId}`, `GET /me/clubs` (all Now).

## Event status

### S1. One status vocabulary

**November schema.** `events.status ENUM('drafted', 'posted', 'archived')`, nullable, no default. The API contract uses `draft`, `posted`, `cancelled`.

**Baseline.** The API's words, in the database: `ENUM('draft', 'posted', 'cancelled') NOT NULL DEFAULT 'draft'`. No mapping layer, which is where bugs come from; the frontend already maps any unknown status to `draft`. The default matches [API decision 1](../../api/README.md#decisions). `cancelled` stays in public lists. Tested: `'archived'` and `'drafted'` are rejected.

**Endpoints:** every event read, `POST /clubs/{clubId}/events`, `GET /clubs/{clubId}/events/drafts`, `PATCH /auth/events/{eventId}`.

### S2. Deleting an event

**November schema.** Both `status = 'archived'` and a nullable `deleted_at` existed. Nothing wrote either, and public reads ignored `deleted_at`.

**Baseline** (decision 4). `deleted_at` is the only delete marker, and `archived` is gone.

- `DELETE /auth/events/{eventId}` runs [`UPDATE_event_soft_delete.sql`](../queries/events/delete/UPDATE_event_soft_delete.sql): `deleted_at = CURRENT_TIMESTAMP` for an event in any status. `status` isn't touched.
- Every read goes through the `event_details` view, which filters `deleted_at IS NULL` in one place, so no query can forget it. The locking reads and status updates filter it too. A deleted event is `404` everywhere, for managers as well.
- **Restore** (decision 9): [`UPDATE_event_restore.sql`](../queries/events/restore/UPDATE_event_restore.sql) clears `deleted_at` only when `deleted_at > CURRENT_TIMESTAMP - INTERVAL 30 DAY`. Deleting never touches the event's status, club links, or images, so a restored event comes back as it was.
- **The purge** (decision 8, `POST /admins/purge`, [query group 31](../README.md#31-purge-job)) runs when an admin asks, and only touches rows deleted more than 30 days ago:
  1. In one transaction: soft-delete the flyers and gallery images of those events, unless a club logo or an event outside the purge still uses them; each image takes its event's `deleted_at`. Then hard-delete those events. Their description, tags, club links, and gallery links go with them (`ON DELETE CASCADE`, F1).
  2. Find images deleted more than 30 days ago that nothing references, oldest first, in batches.
  3. For each: delete the row (the `RESTRICT` foreign keys refuse it if something references the file again), then delete the S3 object. A failed object delete leaves a harmless orphan (I4).
- **One cutoff.** Every restore and purge query has `INTERVAL 30 DAY` written into it rather than taking a parameter, so a caller can't purge with a shorter window than the restore allows. Restore needs `deleted_at` newer than the cutoff and the purge older, so no row is both restorable and purgeable.

**Endpoints:** every event read (Now); `DELETE /auth/events/{eventId}`, `POST /auth/events/{eventId}/restore`, and `POST /admins/purge` (Later).

### S3. Allowed transitions

From [API decision 2](../../api/README.md#decisions):

| From | To | Through |
| --- | --- | --- |
| `draft` | `posted` | `PATCH` with `"status": "posted"`; the event needs a location |
| `posted` | `cancelled` | `PATCH` with `"status": "cancelled"` |
| `cancelled` | `posted` | `PATCH` with `"status": "posted"` |

Nothing goes back to `draft`. Saving a draft's fields without `status` isn't a transition. Deleting isn't a status change.

**Enforced in the `UPDATE`, not the schema.** A `CHECK` can't see the old value, and a trigger would hide the rule. Each status query lists its allowed starting statuses in its `WHERE` clause, so the check and the write are one atomic statement, and an illegal change affects 0 rows:

```sql
UPDATE events SET status = 'posted'    WHERE id = ? AND deleted_at IS NULL AND status IN ('draft', 'cancelled') AND location IS NOT NULL AND location <> '';
UPDATE events SET status = 'cancelled' WHERE id = ? AND deleted_at IS NULL AND status = 'posted';
```

Zero affected rows means a disallowed transition (`400`) or an unknown or deleted event (`404`); [`SELECT_event_status.sql`](../queries/events/read/SELECT_event_status.sql) tells them apart. Tested: `draft → cancelled`, `posted → posted`, `cancelled → cancelled`, posting a draft with no location, and changing a deleted event all affect 0 rows.

**Endpoints:** `PATCH /auth/events/{eventId}` (Now).

## Times (M14)

### T1. Store start and end in UTC

**November schema.** `start_date` and `end_date` were `DATETIME` with no offset, holding wall-clock time in the row's `timezone`. The API returned `YYYY-MM-DD HH:MM:SS`, which the frontend parses as the browser's local time.

**What MySQL does with a `Z` suffix.** Tested on MySQL 8.0.46 with the default `STRICT_TRANS_TABLES`, both through the `mysql` client and through `go-sql-driver/mysql` with a bound string parameter:

| Value inserted into a `DATETIME` | Result |
| --- | --- |
| `2026-09-15T21:00:00.000Z` (what the frontend sends) | **Error 1292**, "Incorrect datetime value". Without strict mode it's accepted with warning 1265, and the `Z` is dropped. |
| `2026-09-15T21:00:00Z` | Error 1292 |
| `2026-09-15T21:00:00+00:00` | Stored as `21:00:00` when the session time zone is UTC |
| `2026-09-15T17:00:00-04:00` | Stored as `21:00:00` when the session time zone is UTC |
| `2026-09-15T21:00:00+00:00` in a `-04:00` session | Stored as `17:00:00`: MySQL (8.0.19 and later) converts an offset to the **session** time zone |
| `time.Time` from Go, driver default `Loc = UTC` | Stored as UTC; reads back as `2026-09-15T21:00:00Z` with `ParseTime = true` |

So passing the frontend's string through fails, and passing an offset string depends on the session time zone.

**Baseline** (decision 2):

- **Storage.** `start_date` and `end_date` hold UTC. `DATETIME` rather than `TIMESTAMP`: it's range-safe past 2038 and never converted by the session.
- **Writes.** The API parses ISO 8601 with an offset (`time.RFC3339Nano`), rejects values without an offset, and passes a `time.Time` in UTC. Never the raw string.
- **Reads.** The client config sets `ParseTime = true`, keeps `Loc = UTC`, and sets `Params["time_zone"] = "'+00:00'"`, so a local MySQL with a `SYSTEM` time zone behaves like RDS. The API formats every event time as RFC 3339 with `Z`: `2026-09-15T21:00:00Z`.
- **"Now" in SQL.** `UTC_TIMESTAMP()` for the UTC `DATETIME` columns (`when=upcoming` is `end_date > UTC_TIMESTAMP()`); `CURRENT_TIMESTAMP` for the `TIMESTAMP` columns (`created_at`, `updated_at`, `deleted_at`), because both follow the session time zone. Never `NOW()` against a UTC `DATETIME`.
- **What `timezone` is for:** the IANA zone the event takes place in. The Event page uses it to show times, and `.ics` export writes it as `TZID`. It never changes how the stored value is read. `NOT NULL DEFAULT 'America/New_York'`. The API validates it with `time.LoadLocation`.
- **Date-only list bounds** (decision 10): `startDate=2026-10-01` is 00:00 on October 1 in `America/New_York`, and `endDate=2026-10-31` is an exclusive bound at 00:00 on November 1 there, both converted to UTC by the API. A full timestamp with an offset is converted as given. The list queries take the UTC range `[start, end)` and match events that overlap it.
- `end_date >= start_date` is a `CHECK`, a backstop for the API's own rule.

**Endpoints:** every event read, `POST /clubs/{clubId}/events`, `PATCH /auth/events/{eventId}` (all Now).

### T2. Timestamp defaults

**Baseline.** `created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP` on `events`, `images`, and `event_images`; `events.updated_at` also has `ON UPDATE CURRENT_TIMESTAMP`. `TIMESTAMP` stores UTC whatever the session's zone; its 2038 limit is the one reason to move these to `DATETIME` later.

Two things the queries handle:

- `ON UPDATE` fires only when a column of `events` actually changes. [`UPDATE_event.sql`](../queries/events/update/UPDATE_event.sql) sets `updated_at = CURRENT_TIMESTAMP` explicitly, so a `PATCH` that changes only the description or the co-hosts still moves the draft to the top of the drafts list.
- `updated_at` has one-second precision, which is too coarse for optimistic locking. That's why decision 6 will use a `version` column.

`clubs` and `club_members` have no `created_at`. A join date can't be recovered later, but no endpoint needs it; add it with the first feature that does.

**Endpoints:** `GET /clubs/{clubId}/events/drafts` (ordered by `updated_at`), `PATCH /auth/events/{eventId}`, every event read's `createdAt`/`updatedAt` (all Now).

## What the "Needed now" endpoints require

### C1. Club topics

Two tables (decision 3):

- `topics (tag)`: the fixed list, as data, loaded by the baseline: `technology`, `arts`, `community`, `women in stem`, `career`, `sports`. Adding a topic is an `INSERT`, not an `ALTER` of an ENUM. Renaming a key is one `UPDATE`, because the foreign key cascades on update. It's deliberately not club-specific, so event tags (H2) or student interests can reuse the list later.
- `club_tags (fk_club_id, slot, tag)`. The primary key is `(fk_club_id, slot)` with `CHECK (slot BETWEEN 1 AND 3)`, so the database itself enforces "at most 3 per club" without a trigger. `slot` also keeps the order the user picked. `UNIQUE (fk_club_id, tag)` blocks duplicates, and the club foreign key cascades on delete, because tags are part of the club.

Both `tag` columns use `utf8mb4_bin`. With the default case-insensitive collation, `'Women In STEM'` passed the foreign key and was stored uppercase, so the "lowercase keys" rule wasn't enforced. Now only exact keys pass.

Writes replace the whole list: delete the club's rows, then insert with `slot` = position + 1, in the club transaction. Reads come from `club_details.tags`, a JSON array in slot order. Tested: a fourth tag, an unknown key, wrong case, and a duplicate are all rejected.

**Endpoints:** `GET /clubs`, `GET /clubs/{clubId}`, `POST /clubs` (Now), `PATCH /clubs/{clubId}` (Later).

### C2. Member count

**Derived with `COUNT`, not stored.** `club_members` has an index on `fk_club_id`; InnoDB appends the primary key to it, so the per-club count is an index-only scan. A stored counter would have to be kept right on every join, leave, and future admin removal, and would drift. `member_count` counts every row: members, e-board, and owners.

**Endpoints:** `GET /clubs`, `GET /clubs/{clubId}` (Now).

### C3. Editing a club

`PATCH /clubs/{clubId}` is Later, but `POST /clubs` needs the same fields now: `clubs`, `club_info`, `club_tags`, and the owner membership in `club_members`. No schema change.

- A club may lack a `club_info` row, so an edit upserts it.
- `clubs.name` is `NOT NULL` and UNIQUE under the default case- and accent-insensitive collation, so `Math Club` and `math club` collide. That's the right rule for club names; map the error to `409`.

**Endpoints:** `POST /clubs` (Now), `PATCH /clubs/{clubId}` (Later).

### E1. Updating a draft in place (M12)

No schema change beyond what's elsewhere here:

- It updates the `events` row in place, which keeps `events.id` stable. That's what H1 needs.
- `event_descriptions` has a primary key (M1), so the description is an upsert.
- `updated_at` is maintained by T2, with the explicit bump.
- Status changes are the guarded `UPDATE`s from S3.
- `location` stays nullable, because a draft may have none; posting requires one.
- `fk_author_id` stays nullable, for `ON DELETE SET NULL`. The Go model's `Location` and `AuthorID` must become pointers, or scanning a `NULL` fails.
- Concurrency: last write wins (decision 6). The `PATCH` transaction locks the row with [`SELECT_event_for_update.sql`](../queries/events/update/SELECT_event_for_update.sql) while it merges and writes, so two requests can't interleave inside one update, but the later request still overwrites the earlier one.

**Endpoints:** `PATCH /auth/events/{eventId}`, `GET /auth/events/{eventId}` (Now).

### Other schema needs from coverage.md

- `cancelled` status → S1. Alt text → I2. `club_tags` → C1.
- Owner membership on club creation (M4) needs no schema change.
- The drafts list uses K1.
- Paging by event instead of joined rows (M9) is a query change: every list selects from `event_details`, one row per event, so `LIMIT` counts events.

## Integrity and performance

### M1. Missing primary keys

**November schema.** `verified_clubs (fk_club_id UNIQUE)`, `admins (fk_student_id UNIQUE)`, and `event_descriptions (fk_event_id UNIQUE)` had no primary key, and each key column was nullable. MySQL allows any number of `NULL` rows in a UNIQUE index.

**Baseline.** Each column is `NOT NULL` and the primary key.

**Endpoints:** `GET /clubs?verified=true`, the `verified` field (Now); the admin check and admin routes (Later); event reads and `PATCH` (Now).

### M2. One role per member

**November schema.** `club_members.member_is_eboard`, `club_members.member_is_owner`, and `events_to_clubs.club_is_event_owner` were nullable `BOOL`s. The leave query's `member_is_owner = FALSE` didn't match `NULL`, so such a member could never leave. Two flags also allowed four combinations for three roles, so every read had to rank them (owner over e-board over member).

**Baseline** (decision 11). `club_members.role ENUM('member', 'eboard', 'owner') NOT NULL DEFAULT 'member'`: exactly one role per member, with nothing to rank. The index `(fk_club_id, role)` serves the owner counts and the e-board list (`role IN ('eboard', 'owner')`). `events_to_clubs.club_is_event_owner` stays a `BOOL NOT NULL DEFAULT FALSE`; it marks an event's host club, not a person's role.

Two rules the queries enforce, because a `CHECK` can't count rows:

- **The last owner can't be demoted.** `UPDATE_club_member_role.sql` counts the club's owners in the same statement.
- **The last owner can't leave.** `DELETE_club_member.sql` does the same. An owner can leave while another owner remains. The deployed legacy SQL refuses every owner.

Both run after `SELECT_club_owners_for_update.sql` locks the club's owner rows, so two owners can't demote each other, or both leave, at once.

**Endpoints:** `DELETE /clubs/{clubId}/members/me`, the club and event role checks, `GET /me/clubs` (Now); `PUT /clubs/{clubId}/members/roles`, the member list (Later).

### M3. One owner club per event

A functional UNIQUE index on `IF(club_is_event_owner, fk_event_id, NULL)`: only owner rows get a non-`NULL` key, so there's at most one per event. "At least one" can't be enforced without a trigger; the create transaction always inserts the owner link. Tested: a second owner fails with a duplicate-key error.

**Endpoints:** every event read (`owners.owner`).

### M4. Other nullable columns

The November DDL wrote `NOT NULL` nowhere.

- **`NOT NULL` in the baseline:** `clubs.name`; `events.title`, `status`, `start_date`, `end_date`, `timezone`, `created_at`, `updated_at`; `images.purpose`, `object_key`, `mimetype`, `created_at`; the primary-key columns in M1 and `club_members.role` in M2.
- **Nullable:**
  - `students.email` (the access token has no email; `GET /me` needs a `null` email).
  - `student_info.*`, `club_info.*`.
  - `events.location` (drafts), `rsvp_link`, `fk_thumbnail_id`, `deleted_at`.
  - `events.fk_author_id` (`ON DELETE SET NULL`).
  - `images.fk_club_id`, `fk_uploaded_by`, `filename`, `alt_text`, `deleted_at`.

The Go models in [`models/`](../models/) are still the legacy copies. `Student.Email`, `Event.Location`, and `Event.AuthorID` must become pointers, and `Image` needs `ClubID`, `UploadedBy`, and `AltText`, before a handler reads the baseline through them.

### K1. Indexes

| Filter | Index |
| --- | --- |
| `events` by `status`, ordered or ranged by `start_date` | `idx_events_status_start (status, start_date)`. Serves every public list and the drafts list's status filter. |
| `events.deleted_at`, `images.deleted_at` | `idx_events_deleted`, `idx_images_deleted`, for the purge. |
| `events_to_clubs.fk_club_id`, `club_members (fk_club_id, role)` | Named indexes. InnoDB appends the primary key, so they cover `(fk_club_id, fk_event_id)` and `(fk_club_id, role, fk_student_id)`. Member count is index-only. |
| `event_images` by event | Primary key `(fk_event_id, fk_image_id)`. |
| `events.fk_thumbnail_id`, `clubs.fk_logo_id`, `event_images.fk_image_id`, `images.fk_club_id`, `events.fk_author_id` | Plain named indexes, which their foreign keys need. |

`when=upcoming` filters on `end_date` and orders by `start_date`. At this data size `(status, start_date)` is enough; revisit if `EXPLAIN` shows scans.

### F1. `ON DELETE` behavior

The November schema's foreign keys declared no action, so every one was `RESTRICT`, and they had generated names (`events_ibfk_1`, …). The baseline names every foreign key and applies the target policy directly, since there's no existing constraint to alter:

| Kind of reference | `ON DELETE` | Foreign keys |
| --- | --- | --- |
| Part of its parent | `CASCADE` | `student_info`, `club_info`, `verified_clubs`, `club_tags`, `event_descriptions`, `event_tags`, and the event side of `events_to_clubs` and `event_images` |
| Attribution | `SET NULL` | `images.fk_uploaded_by`, `events.fk_author_id` |
| Use of a shared file | `RESTRICT` (the in-use check I4 relies on) | `clubs.fk_logo_id`, `events.fk_thumbnail_id`, `event_images.fk_image_id` |
| Membership and ownership | `RESTRICT` until club and account deletion are designed | `club_members`, `admins`, the club side of `events_to_clubs`, `images.fk_club_id` |

The event-side cascades are what let the purge hard-delete an event with one `DELETE`. Nothing deletes a student or a club yet, so `RESTRICT` turns a stray manual `DELETE` into an error instead of silent loss.

### V1. Read views

Every event read needs the same columns: the event, its flyer's key and alt text, its description, its photo count, its owner club, and its co-hosts. The legacy queries each repeated those joins, returned one row per event-to-club link, and so paged by links instead of events. The baseline defines three views:

- **`event_details`:** one row per event that isn't soft-deleted. The owner club is three plain columns; the co-hosts are a JSON array (`[{id, name, logoObjectKey}]`, in no guaranteed order, so the API sorts it by id). Every list and single read selects from it, so `LIMIT` counts events and the soft-delete filter is written once.
- **`club_details`:** one row per club, with the logo key, `club_info`, `tags` as a JSON array in slot order, `member_count`, and `verified`.
- **`announcement_details`:** one row per announcement that isn't soft-deleted, with the owner club and co-owning clubs in the same shape as `event_details` (A1).

They are plain views (no stored data), so they can't go stale. They are part of the schema, so a change to a view is a migration like any other.

## Announcements

### A1. Announcements

Added after the review, for the club page's Announcements tab (Later). Modelled on events, so every rule above that applies to events applies here:

- **`announcements`:** `id`, `fk_author_id` (`ON DELETE SET NULL`), `title` (`NOT NULL`), `body` (`TEXT`, nullable, because a draft may not have one yet; posting requires one), `status ENUM('draft', 'posted') NOT NULL DEFAULT 'draft'`, `created_at`, `updated_at` (`ON UPDATE`), and `deleted_at`. Indexes on `(status, created_at)` for the posted lists, the author, and `deleted_at`.
- **Status:** `draft → posted` is the only transition, enforced in the posting `UPDATE`'s `WHERE` clause as for events (S3). There's no `cancelled`.
- **Deleting:** the same soft delete, 30-day restore, and admin purge as events (S2), with the same hard-coded 30 days.
- **`announcements_to_clubs`:** `(fk_announcement_id, fk_club_id, club_is_announcement_owner)`, with the same one-owner functional index as `events_to_clubs` (M3), so one announcement can belong to several clubs (HunterHacks for the CS clubs). The announcement side cascades on delete, so the purge removes the links with it; the club side is `RESTRICT` (F1).
- **Room to grow:** no images or tags yet. A flyer would be a nullable `fk_thumbnail_id` to `images`, a gallery an `announcement_images` link table, and tags an `announcement_tags` table, each exactly like its event counterpart. The purge would then release images the way `UPDATE_images_of_purgeable_events.sql` does.
- **Ordering:** posted lists are ordered by `created_at`. A draft written long before it's posted sorts by when it was written; a `posted_at` column would change that, and is an [open question](../../api/endpoints/announcements.md#open-questions).

## Older designs

Noted only, so today's schema doesn't block them. Nothing here is designed.

### H1. Event edit history

The July 2025 schema ([`history/07_11_2025_create_core_tables_up.sql`](../migrations/history/07_11_2025_create_core_tables_up.sql)) had `EVENT_VERSIONS` (author, name, image, status, date, `type` create/edit/delete, timestamp), with `EVENTS` pointing at `current_version_id` and `original_version_id`. coverage.md lists edit history under "Not covered by the API yet".

The baseline stays compatible with an append-only `event_revisions` table, written in the same transaction as each `PATCH`, delete, or status change. What keeps that possible:

- Updating in place with a stable `events.id` (E1).
- Soft-deleting instead of deleting on the request (S2). A revision table would need its own answer for the purge, which hard-deletes.
- `fk_author_id` meaning the creator (a revision would record each editor).
- Not hard-deleting a replaced flyer while a revision references it (I4).

Avoid the July indirection, where every read goes through `current_version_id`.

### H2. Event tags

`event_tags (fk_event_id, tag VARCHAR(255))` exists with free-text tags, and nothing reads or writes it. The frontend has no tag input (M11). If event tags use a fixed list later, point `event_tags.tag` at `topics` (C1): that's why the topic table isn't club-specific. The July schema's `INTERESTS` (career, community, volunteer) overlaps the same list, if student interests return.

## The baseline

[`2026_09_29_baseline_up.sql`](../migrations/schema/2026_09_29_baseline_up.sql) creates every table, loads the six topics, and creates the three views. `clubs.fk_logo_id` and `images.fk_club_id` form a cycle, so `images` is created first and its foreign key to `clubs` is added with one `ALTER` after `clubs` exists. [`_down.sql`](../migrations/schema/2026_09_29_baseline_down.sql) drops the views, drops that foreign key, and drops every table, children first. It destroys all data and is only for rebuilding a local or test database.

The seed, [`seed/2026_09_29_baseline_seed.sql`](../migrations/seed/2026_09_29_baseline_seed.sql), is development data. It uses explicit ids and times relative to the day it's loaded, so it always has past, ongoing, and upcoming events. It covers drafts (one with no location), a cancelled event, multi-club events, an event with no description, an event soft-deleted 3 days ago (restorable) and one 45 days ago (purgeable), a flyer that's also in its gallery, one photo in two galleries, soft-deleted images, a club with no `club_info`, a club with two owners, a student with a `NULL` email, and announcements: posted, a draft with no body, one shared by three clubs, one soft-deleted 5 days ago, and one 40 days ago.

### How it was tested

In a throwaway MySQL 8.0.46 container on a local machine (RDS runs 8.0.37), with the default strict `sql_mode` and the server time zone at `+00:00`. The SQL was mounted read-only from `database/`; nothing from `infrastructure/legacy` was used, and no AWS credentials or remote databases were involved.

- Baseline, then seed, applied cleanly.
- Every file in [`queries/`](../queries/) ran with bound parameters (server-side `PREPARE … EXECUTE USING`), with the results checked against the seed. See [database/README.md](../README.md#how-the-queries-were-tested).
- Constraints exercised: a fourth topic, wrong-case and unknown topics, a duplicate topic, a duplicate club name in different case, `end_date` before `start_date`, the old `'archived'` status, a duplicate `object_key`, and deleting a referenced image.
- The down file, run on the seeded database after every query had changed it, left the database empty. The baseline and seed then ran again, and the schema dump was identical to the first one.
- The files avoid semicolons inside comments, and no statement is only a comment, so a runner that splits on semicolons can apply them.

## Future work: a migration runner

For when a live database exists. Until then the baseline is applied by hand to an empty database.

The legacy initializer at [`lambda/internal/database/init/main.go`](../../infrastructure/legacy/lambda/internal/database/init/main.go) hard-codes the November DDL and the September seed. It isn't suitable for the baseline or for any migration after it, because it:

- runs on every create, update, and delete of the custom resource;
- re-runs every listed file each time, with no record of what's applied; and
- re-seeds `STAGING` with non-idempotent inserts, so a second run fails.

A runner needs:

1. **A history table** (`schema_migrations (version, applied_at, checksum)`), and a runner that applies only unapplied files in order, recording each file after its last statement succeeds.
2. **The baseline as version 1.** A fresh database applies it; later changes are new dated files after it.
3. **The seed out of the migration list:** run it only on an empty development database.
4. **Custom-resource changes,** if it stays a CloudFormation custom resource: run only on `Create` and `Update`, never `Delete`, and read the request type.
5. **The files in the image.** The Docker image copies [`lambda/internal/database/init/migrations/`](../../infrastructure/legacy/lambda/internal/database/init/migrations/); the module's `database/migrations/` isn't packaged.
6. **Date-sortable names.** The baseline uses `YYYY_MM_DD`, so tools that order by filename (golang-migrate, goose) apply files in date order. The history files' `MM_DD_YYYY` names don't sort, which is one more reason they're never applied.
7. **Partial failure.** MySQL commits each DDL statement on its own, so a failure partway through a file leaves the earlier statements applied. Keep each migration small, and check its preconditions first.

## Notes for the queries

Collected from the findings above. The queries in [`queries/`](../queries/) follow them.

- Pass times as `time.Time` in UTC, and compare with `UTC_TIMESTAMP()`; compare `TIMESTAMP` columns with `CURRENT_TIMESTAMP`. Client config: `ParseTime = true`, `Loc = UTC`, `time_zone = '+00:00'` (T1).
- Read events and clubs through `event_details` and `club_details`, which already filter soft-deleted events (V1, S2). Match ids with `=`, never `LIKE`.
- Status changes use a guarded `UPDATE` (S3). `UPDATE_event.sql` bumps `updated_at` explicitly (T2).
- Topics are replaced as a whole list, with `slot` = position + 1 (C1). Member count is a `COUNT` (C2).
- Upsert `club_info` and `event_descriptions` (C3, M1).
- Thumbnail confirms insert the image with `fk_club_id`, `fk_uploaded_by`, and `alt_text`, set the pointer, and release the replaced file, in one transaction (I4).
- Model changes before cutover: `Student.Email`, `Event.Location`, `Event.AuthorID` become pointers; `Image` gains `ClubID`, `UploadedBy`, `AltText`, `DeletedAt` (M4).

## Open questions

- Whether an event's author, or an admin, gets management rights beyond the linked clubs' e-board and owners. Nothing in the schema depends on the answer: `fk_author_id` and `admins` already exist.
