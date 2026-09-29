# Schema review

A review of the current schema ([`11_04_2025_create_core_tables_up.sql`](../migrations/schema/11_04_2025_create_core_tables_up.sql)) against the API contracts in [api/README.md](../../api/README.md) and the frontend's needs in [api/coverage.md](../../api/coverage.md). It settles the schema before the Phase 3 queries are written. The "change now" findings are in a new migration pair, [`09_29_2026_align_schema_with_api_up.sql`](../migrations/schema/09_29_2026_align_schema_with_api_up.sql) and [`_down.sql`](../migrations/schema/09_29_2026_align_schema_with_api_down.sql).

Each finding gives what's there now, the problem, a recommendation (**keep**, **change now**, or **defer**), and the endpoints that depend on it. Finding IDs (S1, T1, I1, …) match the section comments in the migration.

## Decisions needed

Numbered in the order they block work. The migration already assumes my recommendation for 1–3; if you decide otherwise, the statements to remove are named.

1. **Images: one table for stored files, with uses as references.** Keep `images` as "a file we stored". Each use (club logo, event flyer, gallery photo, and later board pins) points at the row. One file may have several uses. An image belongs to the club it was uploaded for; event images belong to the event's owner club. The owning club's managers (and admins) may delete the file itself; other clubs may only reference it.
   **Recommend: yes.** The migration drops the three single-use UNIQUE indexes (I1) and backfills the owning club on that basis (I2). If you decide no, remove I1 and the I2 backfill.
2. **Convert existing event times to UTC.** Every existing row stores wall-clock time in its `timezone` column's zone. Evidence: the seed's hackathon is `09:00–20:00` `America/New_York`, which as UTC would be 4 a.m.–3 p.m. From now on, rows store UTC (T1).
   **Recommend: yes, convert once, in the migration.** Before running it, check how many `events` rows `PRODUCTION` has (the create route has been broken since November, so probably only manual inserts), and confirm that the RDS instance has time zone tables. If it doesn't, the migration stops; it never converts some rows and leaves others (see [T1](#t1-store-start-and-end-in-utc)).
3. **Topic keys.** The migration seeds `technology`, `arts`, `community`, `women in stem`, `career` and `sports`, keeping the spaces.
   **Recommend: keep the spaces.** Decision 5 has the UI uppercase the key for display, which gives `WOMEN IN STEM`; a slug would display as `WOMEN-IN-STEM`. If you want slugs anyway, one `UPDATE topics` renames a key everywhere, because the foreign key cascades on update ([C1](#c1-club-topics)).
4. **Drop `events.deleted_at`.** Decision 2 makes `archived` the removal state, and nothing writes `deleted_at`. Two soft-delete mechanisms mean every read has to check both.
   **Recommend: drop it in a later migration.** It isn't in this one, because dropping a column is destructive. Until you decide, public reads filter `deleted_at IS NULL` as well ([S2](#s2-archived-and-deleted_at)).
5. **What deleting an image means** (Need: Later, for `DELETE /images/{imageId}`). Removing an image from a gallery, or replacing a logo or flyer, removes the reference. The file row is deleted only when nothing references it anymore; the current `RESTRICT` foreign keys make that check for free. A takedown by the owning club or an admin removes every reference and then the file, in one transaction.
   **Recommend: both, as described.** This doesn't block the migration ([I4](#i4-deleting-an-image)).
6. **Concurrent edits to an event** (`PATCH /auth/events/{eventId}`).
   **Recommend: last write wins for now.** Only a draft's managers edit it, and "EDIT" on posted events is Later. Add a `version` column when the edit form ships ([E1](#e1-updating-a-draft-in-place-m12)).
7. **How the API serves image URLs** (M6). Doesn't affect the schema, because the database stores only `object_key` either way.
   **Recommend: signed GET URLs with about a one-hour expiry now.** Today's one minute expires while a page is still open. A CDN can come later ([I5](#i5-from-row-to-url-m6)).

## Summary

| ID | Finding | Recommendation | Endpoints |
| --- | --- | --- | --- |
| [I1](#i1-one-images-table-uses-as-references) | One `images` table; single-use UNIQUEs block reuse | Keep the table; change now: drop the UNIQUEs | Thumbnail confirms, event reads, later pins |
| [I2](#i2-owner-uploader-alt-text) | No owning club, uploader, or alt text | Change now | Thumbnail confirms, event reads, `DELETE /images` |
| [I3](#i3-required-image-columns) | Every image column nullable; `object_key` not unique | Change now | Every read that returns a URL |
| [I4](#i4-deleting-an-image) | No deletion semantics; no `deleted_at` | Defer (decision 5) | `DELETE /images`, logo/flyer replacement |
| [I5](#i5-from-row-to-url-m6) | Reads return `object_key` or placeholders, not URLs | No schema change | Every club and event read |
| [S1](#s1-one-status-vocabulary) | `drafted` vs `draft`; no `cancelled`; nullable | Change now | Every event read, create, `PATCH`, drafts list |
| [S2](#s2-archived-and-deleted_at) | `archived` and `deleted_at` overlap | Defer (decision 4) | Public event reads, `DELETE /auth/events` |
| [S3](#s3-allowed-transitions) | Transitions not enforced anywhere | Enforce in the `UPDATE` | `PATCH`, `DELETE /auth/events` |
| [T1](#t1-store-start-and-end-in-utc) | Wall-clock `DATETIME` read as browser-local time | Change now: UTC | Every event read and write |
| [T2](#t2-timestamp-defaults) | `created_at`/`updated_at` have no defaults | Change now | Drafts list order, `PATCH` |
| [C1](#c1-club-topics) | No club topics | Change now | `GET /clubs`, `GET /clubs/{clubId}`, `POST /clubs`, `PATCH /clubs` |
| [C2](#c2-member-count) | No member count | Derive with `COUNT` | `GET /clubs`, `GET /clubs/{clubId}` |
| [C3](#c3-editing-a-club) | Editing a club | Keep | `PATCH /clubs/{clubId}` (Later) |
| [E1](#e1-updating-a-draft-in-place-m12) | Updating a draft in place | Keep, plus T2 and M1 | `PATCH /auth/events/{eventId}` |
| [M1](#m1-missing-primary-keys) | `verified_clubs`, `admins`, `event_descriptions` have no primary key | Change now | `GET /clubs?verified=true`, admin check, event reads, `PATCH` |
| [M2](#m2-nullable-role-flags) | Nullable role flags | Change now | Leave club, role checks, `GET /me/clubs` |
| [M3](#m3-one-owner-club-per-event) | Nothing stops two owner clubs per event | Change now | Every event read (`owners.owner`) |
| [M4](#m4-other-nullable-columns) | Other nullable columns | Change now for `clubs.name` and event columns | Club and event reads |
| [K1](#k1-indexes) | Indexes | Change now: `events (status, start_date)` | Event lists, drafts list, member count |
| [F1](#f1-on-delete-behavior) | No `ON DELETE` actions | Keep `RESTRICT` for existing FKs; set actions on new ones | Future club, account, and image deletion |
| [H1](#h1-event-edit-history), [H2](#h2-event-tags) | Older designs: edit history, event tags | Note only | Not covered by the API yet |

## Images

### I1. One `images` table, uses as references

**Now.** `images` records each stored file (`id`, `purpose`, `object_key`, `filename`, `mimetype`, `created_at`). Uses point at it: `clubs.fk_logo_id`, `events.fk_thumbnail_id`, and the `event_images` link table for galleries. Each of the three is UNIQUE, so a file can have at most one use of each kind.

**The hypothesis holds.** Keep one table meaning "a file we stored", and express how an image is used through references:

- Every purpose has the same lifecycle: sign an upload, `PUT` to S3, confirm. One insert (`INSERT_image.sql` is already shared), one URL function over `object_key` (I5), and one place to authorize, export, and garbage-collect files.
- A table per image type would duplicate the file columns and give three URL paths. Reuse (a board pin showing an event photo) would then mean copying the row, and copying the S3 object or sharing a key between rows. Sharing a key breaks deletion: removing one row's object removes the other's.
- What differs by use belongs on the reference, not the file: gallery order, a pin's note, or an alt-text override if a use ever needs its own.
- Single-valued uses stay pointer columns (`fk_logo_id`, `fk_thumbnail_id`), so the database enforces "at most one flyer" for free. Multi-valued uses stay link tables (`event_images`, later board pins).

**The UNIQUE constraints.** They don't block the named reuse, because board pins would be their own link table. They do block other reuse the model allows:

- `events.fk_thumbnail_id UNIQUE`: the same flyer on two events (a weekly meeting, or a future "duplicate event").
- `event_images.fk_image_id UNIQUE`: one photo in two events' galleries (a co-hosted series, a recap). The primary key `(fk_event_id, fk_image_id)` already prevents duplicates within one gallery.
- `clubs.fk_logo_id UNIQUE`: two clubs sharing a logo. Unlikely, but the constraint encodes "one file, one use", which contradicts the model.

They also invite delete code to assume a file has a single owner. Dropping them is cheap either way; doing it now means the thumbnail-confirm and gallery queries are written against the model they'll live with.

**Recommendation: keep the table; change now: drop the three UNIQUEs** and replace each with a plain index, which its foreign key needs (decision 1). Tested: after the migration one image serves as the flyer of two events and sits in two galleries, and deleting it is refused while it's referenced.

**Endpoints:** `POST /clubs/{clubId}/thumbnails/confirm`, `POST /clubs/{clubId}/events/{eventId}/thumbnails/confirm` (Now); gallery confirm and reads (Later); future board pins.

### I2. Owner, uploader, alt text

**Now.** `images` has no owner, no uploader, and no alt text.

- **Owning club (`fk_club_id`), change now.** Authorization needs it: may this caller attach, replace, or delete this file? It's also how a per-club export finds a club's files. Every current upload has a club: the logo's club, or the event's owner club. The migration backfills existing rows from their current use while each file has exactly one use. After reuse starts, that derivation becomes ambiguous. The column is nullable for a future upload with no club (an avatar, an admin banner).
- **Uploader (`fk_uploaded_by`), change now.** The confirm routes that are Needed now are 🔴, so the caller's `sub` is available. Attribution can't be reconstructed later. Existing rows stay `NULL`, because the gallery confirm route is public and never knew the caller. `ON DELETE SET NULL`: removing a student keeps their uploads.
- **Alt text (`alt_text VARCHAR(1000)`), change now.** Decision 4, needed by the flyer confirm and `altText` on event reads. Alt text describes the image's content, so it stays with the file across reuse; if one use ever needs different wording, add an override on that reference.
- **`deleted_at`, defer.** See I4.
- **`purpose`, keep.** Redefined as "which upload flow created the file" (`club-thumbnail`, `event-thumbnail`, `event-image`), set once at confirm. Display decisions come from the references, never from `purpose`: a gallery photo pinned to a board still says `event-image`. Don't constrain the values yet; a new upload flow brings its own migration anyway.
- `width`, `height`, and byte size would help layout and limits: defer until the upload flow checks files.

**Endpoints:** the thumbnail confirms (Now), event reads' `altText` (Now), `DELETE /images/{imageId}` and export (Later).

### I3. Required image columns

**Now.** Every `images` column except `id` is nullable; `object_key` isn't unique; `created_at` has no default.

**Problem.** A row without `object_key` can't produce a URL. Two rows with the same key break the one-row-per-file rule: deleting either would delete the object the other still uses.

**Recommendation: change now.** `purpose`, `object_key`, and `mimetype` become `NOT NULL`; the gallery confirm, the only writer so far, already requires all of them. Add `UNIQUE (object_key)` and default `created_at` to the current time. `filename` stays nullable; a server-generated file may not have one.

**Endpoints:** every read that returns `thumbnailUrl` or `images` (Now).

### I4. Deleting an image

**Now.** No deletion code exists. Foreign keys are `RESTRICT`, so deleting a referenced image fails.

**What it should mean** once several things can reference a file (decision 5):

- **Removing a use** (take a photo out of a gallery, replace a logo or flyer, unpin): delete or null the reference only. Then try to delete the file row if nothing else references it: `RESTRICT` makes the attempt fail while any reference exists, so the database does the reference count. Delete the S3 object only after the row is gone.
- **Takedown** (the owning club or an admin removes the file everywhere, for example someone asks to be taken out of a photo): in one transaction, remove every reference, then the row; delete the object after commit.
- **Failed object deletes** leave an orphan object, which is harmless. A cleanup sweep must check **both** `STAGING` and `PRODUCTION`, because the dev and prod APIs share one bucket with no environment prefix. A sweep that checks only one database would delete the other environment's files.
- **`deleted_at`: defer.** It's only useful as a tombstone that queues S3 deletion for retry, and only `DELETE /images` (Later) would write it. Adding it now would repeat `events.deleted_at`: a column some reads filter and others forget.
- If edit history (H1) should show past flyers, a replaced flyer must be kept while a revision references it. The `RESTRICT` check above handles that automatically once revisions reference images.

**Endpoints:** `DELETE /images/{imageId}` (Later), logo and flyer replacement in the confirms (Now: the old row is just left in place until this is built).

### I5. From row to URL (M6)

**Now.** Club reads return `images.object_key` as `thumbnailUrl`. Event objects carry only `thumbnailId`, and `owners.*.thumbnailUrl` is a hard-coded external placeholder. The gallery read signs one-minute GET URLs.

**Recommendation: no schema change.** Store only `object_key`, never a URL: signed URLs expire, and a CDN domain differs by environment. The API's storage adapter turns `object_key` into a URL when it builds the response. So the queries must return, for every event:

- the flyer's `object_key` and `alt_text` (`LEFT JOIN images f ON f.id = e.fk_thumbnail_id`);
- each linked club's name and logo `object_key` (already joined, but the name isn't copied: M8);
- `imageCount`: gallery links excluding the flyer (`COUNT(*) FROM event_images WHERE fk_event_id = e.id AND fk_image_id <> COALESCE(e.fk_thumbnail_id, '')`), since the flyer may also sit in the gallery now; and
- for `GET /events/{eventId}` only, the gallery keys in upload order (`images.created_at`, which is correct until galleries can reuse files; then add a `created_at` to `event_images`).

Keep `thumbnailId` in the event object. The URL strategy is decision 7.

**Endpoints:** `GET /events`, `GET /events/{eventId}`, `GET /clubs/{clubId}/events`, `GET /me/events`, `GET /clubs`, `GET /clubs/{clubId}`, `GET /me/clubs` (all Now).

## Event status

### S1. One status vocabulary

**Now.** `events.status ENUM('drafted', 'posted', 'archived')`, nullable, no default. The API contract uses `draft`, `posted`, `cancelled`, `archived`.

**Problem.** A mapping layer between `drafted` and `draft` is where bugs come from: the frontend already maps any unknown status to `draft`. `cancelled` doesn't exist. A `NULL` status matches no read.

**Recommendation: change now.** Use the API's words in the database: `ENUM('draft', 'posted', 'cancelled', 'archived') NOT NULL DEFAULT 'draft'`. The migration renames existing `drafted` rows through a transitional ENUM and turns `NULL` into `draft`; no read ever showed a `NULL`-status row, so that doesn't change what anyone sees. The default matches decision 1. `cancelled` stays in public lists; `archived` never appears publicly.

No working code writes `drafted`: `INSERT_event.sql` does, but it's the broken create query and has to be rewritten anyway. After the migration, the rewrite must write `draft`, or omit `status` and take the default. Tested: `'drafted'` is now rejected.

**Endpoints:** every event read, `POST /clubs/{clubId}/events`, `GET /clubs/{clubId}/events/drafts`, `PATCH` and `DELETE /auth/events/{eventId}` (Now, except `DELETE`).

### S2. `archived` and `deleted_at`

**Now.** Both `status = 'archived'` and a nullable `deleted_at` exist. Nothing writes either, and public reads ignore `deleted_at`.

**Recommendation: defer (decision 4).** `DELETE /auth/events/{eventId}` sets `archived`, and nothing leaves `archived`, so `archived` is the soft delete. `deleted_at` is redundant; the recommendation is to drop it once you agree. Meanwhile every public read filters `status IN ('posted', 'cancelled') AND deleted_at IS NULL`, as [events.md](../../api/endpoints/events.md#event-status) already says. The archive time is `updated_at`, because nothing changes an archived row.

**Endpoints:** public event reads (Now), `DELETE /auth/events/{eventId}` (Later).

### S3. Allowed transitions

From [decision 2](../../api/README.md#decisions):

| From | To | Through |
| --- | --- | --- |
| `draft` | `posted` | `PATCH` with `"status": "posted"` |
| `posted` | `cancelled` | `PATCH` with `"status": "cancelled"` |
| `cancelled` | `posted` | `PATCH` with `"status": "posted"` |
| any | `archived` | `DELETE`; `PATCH` can't set it |

Nothing goes back to `draft`, and nothing leaves `archived`. Saving a draft's fields without `status` isn't a transition.

**Recommendation: enforce in the `UPDATE`, not the schema.** A `CHECK` can't see the old value, and a trigger body contains semicolons, which the legacy runner splits. Put the allowed source states in the `WHERE` clause so the check and the write are one atomic statement:

```sql
UPDATE events SET status = 'posted' WHERE id = ? AND status IN ('draft', 'cancelled');
UPDATE events SET status = 'cancelled' WHERE id = ? AND status = 'posted';
UPDATE events SET status = 'archived' WHERE id = ? AND status <> 'archived';
```

Zero affected rows means a disallowed transition (`400`) or an unknown or archived event (`404`); a follow-up read tells them apart.

**Endpoints:** `PATCH /auth/events/{eventId}` (Now), `DELETE /auth/events/{eventId}` (Later).

## Times (M14)

### T1. Store start and end in UTC

**Now.** `start_date` and `end_date` are `DATETIME` with no offset. Existing rows are wall-clock time in the row's `timezone` (`America/New_York` in every seed row). The API returns `YYYY-MM-DD HH:MM:SS`, which the frontend parses as the browser's local time. The legacy create handler passes the client's date strings straight to MySQL.

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

**Recommendation: change now.**

- **Storage.** `start_date` and `end_date` hold UTC. Keep `DATETIME` rather than `TIMESTAMP`: it's range-safe past 2038 and never converted by the session.
- **Writes.** The API parses ISO 8601 with an offset (`time.RFC3339Nano`), rejects values without an offset, and passes a `time.Time` in UTC. Never the raw string.
- **Reads.** The client config sets `ParseTime = true`, keeps `Loc = UTC`, and sets `Params["time_zone"] = "'+00:00'"`, so a local MySQL with a `SYSTEM` time zone behaves like RDS. Its default parameter group already runs in UTC. The API formats every event time as RFC 3339 with `Z`: `2026-09-15T21:00:00Z`.
- **"Now" in SQL** is `UTC_TIMESTAMP()`, never `NOW()`, which follows the session time zone. `when=upcoming` is `end_date > UTC_TIMESTAMP()`.
- **What `timezone` is for:** the IANA zone the event takes place in. The Event page uses it to show times, and `.ics` export writes it as `TZID`. It never changes how the stored value is read. `NOT NULL DEFAULT 'America/New_York'`, the default M2 already proposes. The API validates it with `time.LoadLocation`.
- **Date-only list bounds** (`startDate=2026-09-01`) need a zone. Treat them as midnight in `America/New_York`, or accept only full timestamps with an offset.
- **Existing rows** are converted once with `CONVERT_TZ(start_date, timezone, '+00:00')` (decision 2). Tested: September seed events move +4 hours and the November hackathon +5, so daylight saving is handled.
  - `CONVERT_TZ` needs the server's time zone tables and returns `NULL` without them. The migration makes both columns `NOT NULL` first, so a missing table or an unknown zone name fails the whole `UPDATE` (error 1048), and no rows change. That was tested with and without the tables.
  - RDS for MySQL supports named time zones, so the tables should be there, but I couldn't check the instance from here. Check with `SELECT CONVERT_TZ('2025-11-10 09:00:00', 'America/New_York', '+00:00')`, which should return `2025-11-10 14:00:00`.
  - The conversion must run exactly once. The down migration converts back.
- `end_date >= start_date` becomes a `CHECK`, a backstop for the API's own rule. Every seed row passes.

**Endpoints:** every event read, `POST /clubs/{clubId}/events`, `PATCH /auth/events/{eventId}` (all Now).

### T2. Timestamp defaults

**Now.** `events.created_at`, `events.updated_at`, and `images.created_at` are `TIMESTAMP` with no default. The stack uses MySQL 8.0's `explicit_defaults_for_timestamp = ON`, so nothing sets them automatically, and they're nullable.

**Recommendation: change now.** `created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP`. `events.updated_at` also gets `ON UPDATE CURRENT_TIMESTAMP`. Existing `NULL`s are filled from the other column, or the current time. `TIMESTAMP` stores UTC, whatever the session's zone. Its 2038 limit is the one reason to move these to `DATETIME` later; defer that.

Two things the queries must handle:

- `ON UPDATE` fires only when a column of `events` actually changes. A `PATCH` that changes only the description (in `event_descriptions`) or the co-hosts (in `events_to_clubs`) must also run `UPDATE events SET updated_at = CURRENT_TIMESTAMP WHERE id = ?`, or the drafts list order is wrong.
- `updated_at` has one-second precision, which is too coarse for optimistic locking. That's why decision 6 would use a `version` column.

`clubs` and `club_members` have no `created_at`. A join date can't be recovered later, but no endpoint needs it: defer.

**Endpoints:** `GET /clubs/{clubId}/events/drafts` (ordered by `updated_at`), `PATCH /auth/events/{eventId}`, every event read's `createdAt`/`updatedAt` (all Now).

## What the "Needed now" endpoints still require

### C1. Club topics

**Now.** Nothing stores club topics. `POST /clubs` ignores `tags` (M13).

**Recommendation: change now.** Two tables:

- `topics (tag)`: the fixed list, as data. Adding a topic is an `INSERT`, not an `ALTER` of an ENUM. Renaming a key is one `UPDATE`, because the foreign key cascades on update (decision 3). It's deliberately not club-specific, so event tags (H2) or student interests can reuse the list later.
- `club_tags (fk_club_id, slot, tag)`. The primary key is `(fk_club_id, slot)` with `CHECK (slot BETWEEN 1 AND 3)`, so the database itself enforces "at most 3 per club" without a trigger. `slot` also keeps the order the user picked. `UNIQUE (fk_club_id, tag)` blocks duplicates, and the club foreign key cascades on delete, because tags are part of the club.

Both `tag` columns use `utf8mb4_bin`. With the default case-insensitive collation, `'Women In STEM'` passed the foreign key and was stored uppercase (found in testing), so the "lowercase keys" rule wasn't enforced. Now only exact keys pass.

Writes replace the whole list, as `PATCH` defines: delete the club's rows, then insert with `slot` = position + 1, in the club transaction. Reads use `GROUP_CONCAT(tag ORDER BY slot)` or a second query. Tested: a fourth tag, an unknown key, and wrong case are all rejected.

**Endpoints:** `GET /clubs`, `GET /clubs/{clubId}`, `POST /clubs` (Now), `PATCH /clubs/{clubId}` (Later).

### C2. Member count

**Recommendation: derive with `COUNT`, don't store.** `club_members` already has an index on `fk_club_id`: MySQL created it for the foreign key, although the DDL doesn't declare one. InnoDB appends the primary key to it, so it's effectively `(fk_club_id, fk_student_id)`. `EXPLAIN` of the per-club `GROUP BY` shows an index-only scan (`Using index`). A stored counter would have to be kept right on every join, leave, and future admin removal, and would drift. `memberCount` counts every row: members, e-board, and owners.

**Endpoints:** `GET /clubs`, `GET /clubs/{clubId}` (Now).

### C3. Editing a club

`PATCH /clubs/{clubId}` is Later in the [endpoint table](../../api/README.md#endpoints), but `POST /clubs` needs the same fields now. Its name, description, and website are covered by `clubs` and `club_info`. Tags come from C1, and the owner membership from `club_members`, which already supports it (M4).

**Recommendation: keep.** Notes for the queries:

- A club may lack a `club_info` row (clubs created outside the API), so an edit upserts it with `INSERT … ON DUPLICATE KEY UPDATE`.
- `clubs.name` becomes `NOT NULL` (M4). Its UNIQUE index is case- and accent-insensitive under the default collation, so `Math Club` and `math club` collide. That's the right rule for club names; map the error to `409`.

**Endpoints:** `POST /clubs` (Now), `PATCH /clubs/{clubId}` (Later).

### E1. Updating a draft in place (M12)

**Recommendation: keep the tables; the changes it needs are elsewhere in this review.**

- It updates the `events` row in place, which keeps `events.id` stable. That's what H1 needs.
- `event_descriptions` gets a primary key (M1), so the description is an upsert.
- `updated_at` is maintained by T2, with the explicit bump for description-only edits.
- The status check is the guarded `UPDATE` from S3.
- `location` stays nullable, because a draft may have none; the API requires it when the result is `posted`.
- `fk_author_id` stays nullable, for a future `ON DELETE SET NULL`. The Go model's `Location` and `AuthorID` must become pointers, or scanning a `NULL` fails.
- Concurrency control is decision 6.

**Endpoints:** `PATCH /auth/events/{eventId}`, `GET /auth/events/{eventId}` (Now).

### Other schema needs from coverage.md

- `cancelled` status → S1. Alt text → I2. `club_tags` → C1.
- Owner membership on club creation (M4) needs no schema change.
- The drafts list uses K1.
- Paging by event instead of joined rows (M9) is a query change: page over `events` in a subquery, then join the clubs.
- Leaving a club works again for rows with `NULL` flags (M2).

## Integrity and performance

### M1. Missing primary keys

**Now.** `verified_clubs (fk_club_id UNIQUE)`, `admins (fk_student_id UNIQUE)`, and `event_descriptions (fk_event_id UNIQUE)` have no primary key. Each key column is nullable, and MySQL allows any number of `NULL` rows in a UNIQUE index.

**Recommendation: change now.** Make the column `NOT NULL` and the primary key. First delete rows whose key is `NULL`: they mark nothing, no query can reach them, and there are none in the seed. The pre-flight checks below count them before anything is deleted.

**Endpoints:** `GET /clubs?verified=true`, the `verified` field (Now); the admin check and admin routes (Later); event reads and `PATCH` (Now).

### M2. Nullable role flags

**Now.** `club_members.member_is_eboard`, `club_members.member_is_owner`, and `events_to_clubs.club_is_event_owner` are nullable `BOOL`s with no default. The leave query's `member_is_owner = FALSE` doesn't match `NULL`, so such a member can never leave.

**Recommendation: change now.** `NOT NULL DEFAULT FALSE`, with existing `NULL`s set to `FALSE`. The flags stay separate columns rather than one role ENUM: `GET /me/clubs` already computes one role with owner first, and the proposed `PUT .../roles` takes one `role` value and writes both flags.

**Endpoints:** `DELETE /clubs/{clubId}/members/me`, the club and event role checks, `GET /me/clubs` (Now).

### M3. One owner club per event

**Now.** Nothing stops two `events_to_clubs` rows with `club_is_event_owner = TRUE` for one event. The read handlers then let "the last owner row processed" win.

**Recommendation: change now.** Add a functional UNIQUE index on `IF(club_is_event_owner, fk_event_id, NULL)`. Only owner rows get a non-`NULL` key, so there's at most one per event. Every seed event has exactly one owner. "At least one" can't be enforced without a trigger; the create transaction always inserts the owner link. Tested: making a second owner fails with a duplicate-key error.

**Endpoints:** every event read (`owners.owner`), the image-owner backfill in I2 (Now).

### M4. Other nullable columns

The DDL writes `NOT NULL` nowhere.

- **Change now:**
  - `clubs.name` (every create sets it; a `NULL` name escapes the UNIQUE index).
  - `events.title`, `start_date`, `end_date`, `timezone`, `status`, `created_at`, `updated_at` (every create sets or defaults them).
  - `images.purpose`, `object_key`, `mimetype`, `created_at` (I3).
  - The primary-key columns in M1 and the flags in M2.
  - Each of these fails loudly if an existing row is `NULL`, rather than inventing data. `status` and `timezone` are the exceptions: they get the defaults described in S1 and T1.
- **Keep nullable:**
  - `students.email` (M5: the access token has no email; the Go model and `GET /me` need a `null` email).
  - `student_info.*`, `club_info.*`.
  - `events.location` (drafts), `rsvp_link`, `fk_thumbnail_id`, `deleted_at`.
  - `events.fk_author_id` (a future account deletion sets it to `NULL`).
  - `images.filename`, `fk_club_id`, `fk_uploaded_by`, `alt_text`.

After the migration, the Go `Event` model's non-pointer `Title`, `Status`, dates, `Timezone`, `CreatedAt`, and `UpdatedAt` can no longer receive a `NULL`. `Location` and `AuthorID` still can (E1).

### K1. Indexes

Checked with `information_schema.statistics` after loading the November DDL:

| Filter | Index today | Recommendation |
| --- | --- | --- |
| `events` by `status`, ordered or ranged by `start_date` | None | **Change now:** `idx_events_status_start (status, start_date)`. Serves every public list, `when=past` ordering, and the drafts list's status filter. |
| `events_to_clubs.fk_club_id` | Exists: `fk_club_id`, created implicitly for the foreign key | Keep. InnoDB appends the primary key, so it covers `(fk_club_id, fk_event_id)`. |
| `club_members.fk_club_id` | Exists: `fk_club_id`, implicit | Keep. Member count is index-only (C2). |
| `event_images` by event | Primary key `(fk_event_id, fk_image_id)` | Keep. |
| `images.fk_club_id`, `images.fk_uploaded_by` | Created by the new foreign keys | — |
| `events.fk_thumbnail_id`, `clubs.fk_logo_id`, `event_images.fk_image_id` | UNIQUE today | Plain indexes after I1. The foreign keys need them. |

`when=upcoming` filters on `end_date` and orders by `start_date`. At this data size `(status, start_date)` is enough; revisit if `EXPLAIN` shows scans.

### F1. `ON DELETE` behavior

**Now.** No foreign key declares an action, so every one is `RESTRICT`. Nothing in the API hard-deletes a student, club, or event: events are archived, and there's no club or account deletion.

**Recommendation: keep `RESTRICT` on the existing foreign keys** until each delete workflow is designed. While nothing deletes those parents, `RESTRICT` turns a forgotten cleanup step, or a stray manual `DELETE`, into an error instead of silent loss. For images, `RESTRICT` is the in-use check that I4 relies on. The target policy, applied now to the new foreign keys and to old ones in the migration that ships each workflow:

| Kind of reference | `ON DELETE` | Examples |
| --- | --- | --- |
| Part of its parent | `CASCADE` | `club_tags` (**set now**), `club_info`, `verified_clubs`, `event_descriptions`, `event_tags` |
| Attribution | `SET NULL` | `images.fk_uploaded_by` (**set now**), `events.fk_author_id` |
| Use of a shared file | `RESTRICT` (in use), or explicit unlinking for a takedown (I4) | `clubs.fk_logo_id`, `events.fk_thumbnail_id`, `event_images.fk_image_id` |
| Membership and ownership | `RESTRICT` until club and account deletion are designed | `club_members`, `admins`, `events_to_clubs`, `images.fk_club_id` |

The new foreign keys are named (`fk_images_club`, `fk_club_tags_topic`, …) so a later migration can alter them. The November ones got generated names (`events_ibfk_1`, …), which depend on statement order.

## Older designs

Noted only, so today's changes don't block them. Nothing here is designed.

### H1. Event edit history

The July 2025 schema (`07_11_2025_create_core_tables_up.sql`, from the planning PDF's era) had `EVENT_VERSIONS` (author, name, image, status, date, `type` create/edit/delete, timestamp), with `EVENTS` pointing at `current_version_id` and `original_version_id`. coverage.md lists edit history under "Not covered by the API yet".

Today's shape stays compatible with an append-only `event_revisions` table, written in the same transaction as each `PATCH`, `DELETE`, or status change. What keeps that possible:

- Updating in place with a stable `events.id` (E1).
- Archiving instead of deleting (S2).
- `fk_author_id` meaning the creator (a revision would record each editor).
- Not hard-deleting a replaced flyer while a revision references it (I4).

Avoid the July indirection, where every read goes through `current_version_id`.

### H2. Event tags

`event_tags (fk_event_id, tag VARCHAR(255))` exists with free-text tags, and nothing reads or writes it. The frontend has no tag input (M11). If event tags use a fixed list later, point `event_tags.tag` at `topics` (C1): that's why the topic table isn't club-specific. Don't drop `event_tags`. The July schema's `INTERESTS` (career, community, volunteer) overlaps the same list, if student interests return.

## The migration

[`09_29_2026_align_schema_with_api_up.sql`](../migrations/schema/09_29_2026_align_schema_with_api_up.sql) applies M1–M4, S1, T1, T2, K1, I1–I3 and C1, in that order: M3 has to exist before the I2 backfill joins on owner clubs. [`_down.sql`](../migrations/schema/09_29_2026_align_schema_with_api_down.sql) reverses it in reverse order.

### How it was tested

On MySQL 8.0.46 (RDS runs 8.0.37) with the default strict `sql_mode`, starting from the November DDL plus the September seed. Each file ran through a copy of the legacy initializer's `runMigration`, which splits on semicolons.

- **Up** runs as 37 statements. Tested without time zone tables (stops at T1 with no rows changed) and with them (all events converted with the right daylight-saving offsets).
- **Constraints** after up, each exercised:
  - check constraints: a fourth tag, `end_date` before `start_date`;
  - topic foreign key: an unknown key, wrong case;
  - uniqueness: a second owner club, a duplicate `object_key`;
  - the old `'drafted'` status;
  - image reuse allowed, and deleting a referenced image refused.
- **Down, then compare.** Schema dumps before up and after down are identical except the order of one index in `clubs`. Event times return to their original values; the only other data difference is `created_at`, because the test reseeded. Up then runs again cleanly.
- **Down fails** if an image has been given a second use, at restoring its UNIQUE index. That's intended: resolve the reuse first.
- The runner test caught a semicolon inside a comment, which the split turned into a broken statement. No comment in either file contains one now, and no chunk is only a comment (MySQL rejects an empty query).

### Pre-flight checks

Run against each database before up. Every count should be 0, and the conversion check should return `2025-11-10 14:00:00`.

```sql
SELECT COUNT(*) FROM verified_clubs WHERE fk_club_id IS NULL;        -- deleted by M1
SELECT COUNT(*) FROM admins WHERE fk_student_id IS NULL;             -- deleted by M1
SELECT COUNT(*) FROM event_descriptions WHERE fk_event_id IS NULL;   -- deleted by M1
SELECT COUNT(*) FROM clubs WHERE name IS NULL;                       -- M4 fails
SELECT COUNT(*) FROM events WHERE title IS NULL OR start_date IS NULL OR end_date IS NULL;  -- T1 fails
SELECT COUNT(*) FROM events WHERE end_date < start_date;             -- T2 check fails
SELECT COUNT(*) FROM (SELECT fk_event_id FROM events_to_clubs WHERE club_is_event_owner
  GROUP BY fk_event_id HAVING COUNT(*) > 1) x;                       -- M3 fails
SELECT COUNT(*) FROM images WHERE purpose IS NULL OR object_key IS NULL OR mimetype IS NULL;  -- I3 fails
SELECT COUNT(*) FROM (SELECT object_key FROM images GROUP BY object_key HAVING COUNT(*) > 1) x;  -- I3 fails
SELECT DISTINCT timezone FROM events;                                -- every value must be a valid zone name
SELECT CONVERT_TZ('2025-11-10 09:00:00', 'America/New_York', '+00:00');
```

MySQL commits each DDL statement on its own, so a failure partway leaves the earlier statements applied. The pre-flight checks exist so up doesn't fail partway.

### Wiring it into the legacy initializer

The initializer at [`lambda/internal/database/init/main.go`](../../infrastructure/legacy/lambda/internal/database/init/main.go) hard-codes `11_04_2025_create_core_tables_up.sql`. This migration isn't wired in. Appending it to `initTableMigrationFiles` would be wrong, because the initializer:

- runs on every create, update, and delete of the custom resource;
- re-runs every listed file each time, with no record of what's applied; and
- re-seeds `STAGING` with non-idempotent inserts, so a second run already fails.

This file converts times once (T1), so a re-run would shift them again, and its `ADD COLUMN`s would fail. Wiring it in would take:

1. **A history table** (`schema_migrations (version, applied_at, checksum)`), and a runner that applies only unapplied files in an explicit order, recording each file after its last statement succeeds.
2. **A baseline:** insert history rows for the November DDL (and the seed, in `STAGING`) on the existing databases without running them.
3. **The seed out of the migration list:** run it only on an empty `STAGING`, or make it idempotent.
4. **Custom-resource changes:** run only on `Create` and `Update`, never `Delete`, and read the request type.
5. **The file in the image:** copy it into [`lambda/internal/database/init/migrations/`](../../infrastructure/legacy/lambda/internal/database/init/migrations/), the directory the Docker image copies. The module's `database/migrations/` isn't packaged.
6. **An explicit order, or new names.** The `MM_DD_YYYY` names don't sort by date: `09_29_2026…` sorts before `11_04_2025…`. A runner that orders by filename (golang-migrate, goose) would apply this file before the table it alters. Keep an explicit list, or switch to `YYYY_MM_DD` names when adopting a tool.
7. **Time zone tables** on the target server, checked by the pre-flight query above.

## Notes for the queries

Collected from the findings above. None of these is a schema change.

- Pass times as `time.Time` in UTC, and compare with `UTC_TIMESTAMP()`. Client config: `ParseTime = true`, `Loc = UTC`, `time_zone = '+00:00'` (T1).
- Public reads: `status IN ('posted', 'cancelled') AND deleted_at IS NULL`. Match `eventId` with `=`, not `LIKE` (S2).
- Status changes use a guarded `UPDATE` (S3). Bump `events.updated_at` explicitly when only a description or co-host changes (T2).
- The rewritten `INSERT_event.sql` writes `draft` or takes the default (S1).
- `imageCount` excludes the flyer. The gallery is ordered by `images.created_at` (I5).
- Topics are replaced as a whole list, with `slot` = position + 1 (C1). Member count is a grouped `COUNT` (C2).
- Upsert `club_info` and `event_descriptions` (C3, M1).
- Thumbnail confirms insert the image with `fk_club_id`, `fk_uploaded_by`, and `alt_text`, and set the pointer, in one transaction. The old flyer or logo row is left in place until I4 is built.
- Model changes: `Student.Email`, `Event.Location`, `Event.AuthorID` become pointers. `Image` gains `ClubID`, `UploadedBy`, `AltText`.
