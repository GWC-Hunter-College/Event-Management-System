# Edit history

**Proposed, 2026-09-30.** Nothing here is built. Every route is ⬜ and **Proposed**, in the [api/ reference](../../api/README.md#legend) style.

**Depends on:** the write endpoints it records (club, event, and announcement create, edit, delete, restore), which are Needed now or Later in [coverage.md](../../api/coverage.md#needed-now). Nothing else.
**Needed by:** nothing, but [import and export](import-export.md#privacy) uses it as the record of who exported member data.

## Problem

A club's e-board is several people sharing one set of controls, and it changes every year. When an event's time is wrong, a draft vanishes, or an announcement is edited after posting, nobody can say who did it or what it said before. [coverage.md](../../api/coverage.md#not-covered-by-the-api-yet) lists edit history as not covered, and the schema review kept room for it ([H1](../../database/docs/schema-review.md#h1-event-edit-history)).

**Who uses it and how:**

- **E-board and owners** open Club → Manage → **History**: a list like "Ada L. changed the start time of Hack Night · Sep 28, 3:14 PM", which expands to show before and after. Filters: events, announcements, the club. On an event or announcement, the manager bar links to that item's own history. A deleted item shows "Restore" while it's still inside its 30 days.
- **Admins** see every club's history, to sort out disputes and abuse.
- **Members and the public** see none of it.

**What it must record:** who **created, edited, cancelled, deleted, or restored** a club's **events**, its **announcements**, and **the club itself**.

## The two designs

### A. Append-only audit log

One table for every kind of thing. Each change writes one row, in the same transaction as the change: the entity, its id, the action, the actor, the time, and the fields that changed, before and after, as JSON.

### B. Per-entity version tables

The planning PDF's `Event_Versions`, which the July 2025 schema implemented as `EVENT_VERSIONS` ([history file](../../database/migrations/history/07_11_2025_create_core_tables_up.sql)): a full copy of the event per version (author, name, image, status, date, `type` create/edit/delete, timestamp), with `EVENTS.current_version_id` and `original_version_id` pointing into it. Brought up to the baseline, it would look like this, repeated as `announcement_versions` and `club_versions`:

```sql
CREATE TABLE `event_versions` (
  `id` BIGINT NOT NULL AUTO_INCREMENT,
  `fk_event_id` INT NOT NULL,
  `version` INT NOT NULL,
  `type` ENUM('create', 'edit', 'cancel', 'delete', 'restore') NOT NULL,
  `fk_editor_id` CHAR(36) NULL,
  `title` VARCHAR(255) NOT NULL,
  `location` VARCHAR(255) NULL,
  `rsvp_link` VARCHAR(255) NULL,
  `status` ENUM('draft', 'posted', 'cancelled') NOT NULL,
  `start_date` DATETIME NOT NULL,
  `end_date` DATETIME NOT NULL,
  `timezone` VARCHAR(60) NOT NULL,
  `fk_thumbnail_id` CHAR(36) NULL,
  `description` TEXT NULL,
  `associates` JSON NOT NULL,
  `created_at` TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
  PRIMARY KEY (`id`),
  UNIQUE KEY `uq_event_versions_event_version` (`fk_event_id`, `version`)
);
```

(The July table had no event id at all, so versions between the original and the current one couldn't be found for an event. Any version design needs `fk_event_id`.)

### Comparison

| | A. Audit log | B. Version tables |
| --- | --- | --- |
| Answers "who did what, when" | Directly: one row per action, with the action named. | Indirectly: diff consecutive versions to see what changed; the `type` says only create, edit, or delete. |
| Answers "what did it look like on date X" | By replaying rows from the created snapshot. Rarely needed. | Directly. |
| Tables | One, for every entity. | One per entity: three now, more for pins or memberships. |
| Schema drift | None: rows hold API field names in JSON. A new event column needs no history change. | Every column added to `events` must be added to `event_versions` too, forever. |
| Child tables | Diffed at the API object level, so `description` (in `event_descriptions`), `associates` (in `events_to_clubs`), and the flyer are ordinary fields. | Must be copied into each version, or changes to them go unrecorded. |
| A club-wide feed | One query on one table. | A `UNION` across three tables with different columns. |
| Storage | Only changed fields: a time change is a few dozen bytes. | A full copy per edit, description included. |
| Revert | Possible: apply `old_values`. | Easy: copy a version back. |
| Purge and privacy | One statement removes a purged entity's rows. | Same, per table. |
| Effort | One table, one Go helper, one insert per write handler. | Three tables, three copy routines, kept in step with every schema change. |

**Recommendation: A, the audit log.** The question the e-board asks is "who did this", which the log answers row by row. Version tables answer "what did it look like", which nobody has asked for, at the price of a table per entity that has to shadow every future column. The log also covers the club itself and later kinds of change (role changes, verification, exports) without new tables. It fits what the baseline already does: rows are updated in place with stable ids (E1), deletes are soft (S2), and the July indirection through `current_version_id` stays gone, as H1 asked.

Also rejected:

- **MySQL triggers** writing the log. The schema review already avoided triggers because "a trigger would hide the rule" (S3). A trigger doesn't know who the API caller is unless every handler sets a session variable, and it sees tables, not the event object: an edit to the description and co-hosts would be three unrelated trigger rows.
- **Binlog change capture** (DMS or Debezium). Infrastructure and cost for a small feature, and no actor either.
- **CloudWatch logs.** Not queryable per club, expire, and aren't written in the transaction, so a crash can lose the record of a committed change.

## Data model

```sql
-- Edit history: one row per change to a club, an event or an announcement, written
-- in the same transaction as the change. Append-only. Nothing updates a row, and only
-- the purge deletes, replacing a purged entity's rows with one purged row.
-- entity_id has no foreign key, so history outlives what it describes.
-- entity_label is the title or name at the time, for lists and for purged entities.
-- club_ids lists every club whose e-board may read the row: the owner club and the
-- co-hosts, before and after the change. The multi-valued index serves MEMBER OF.
-- old_values and new_values hold only the fields that changed, by their API names.
-- A created row has new_values only. Delete, restore and purge rows have neither.
CREATE TABLE `audit_log` (
  `id` BIGINT NOT NULL AUTO_INCREMENT,
  `entity_type` ENUM('club', 'event', 'announcement') NOT NULL,
  `entity_id` INT NOT NULL,
  `entity_label` VARCHAR(255) NOT NULL,
  `action` ENUM('created', 'updated', 'posted', 'cancelled', 'deleted', 'restored', 'purged') NOT NULL,
  `fk_actor_id` CHAR(36) NULL,
  `actor_role` ENUM('owner', 'eboard', 'admin') NULL,
  `club_ids` JSON NOT NULL,
  `old_values` JSON NULL,
  `new_values` JSON NULL,
  `created_at` TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
  PRIMARY KEY (`id`),
  KEY `idx_audit_log_entity` (`entity_type`, `entity_id`, `id`),
  KEY `idx_audit_log_club_ids` ((CAST(`club_ids` AS UNSIGNED ARRAY))),
  KEY `idx_audit_log_actor` (`fk_actor_id`),
  CONSTRAINT `fk_audit_log_actor` FOREIGN KEY (`fk_actor_id`) REFERENCES `students` (`id`) ON DELETE SET NULL
);
```

Notes:

- **Names.** `old_values` and `new_values`, because `BEFORE` is a reserved word in MySQL.
- **`club_ids` as JSON** with a multi-valued index (MySQL 8.0.17 and later, which every [hosting option](../../infrastructure/aws/hosting-options.md) provides): a club's feed is `WHERE ? MEMBER OF (club_ids) ORDER BY id DESC LIMIT ?`. Recording the clubs on the row, rather than joining `events_to_clubs` at read time, keeps a co-host's view of history after the event is purged or the co-host is removed. The alternative is an `audit_log_clubs` link table, in the style of `events_to_clubs`; it costs a second insert per change for the same result.
- **`actor_role`** is the actor's role at the time, so "an admin cancelled this" reads correctly after roles change. `fk_actor_id` is `SET NULL` ([F1](../../database/docs/schema-review.md#f1-on-delete-behavior) attribution).
- **`entity_id` is `INT`**, which fits clubs, events, announcements, and board pins. Recording images (UUIDs) later would widen it to `VARCHAR(36)`.
- **`entity_type` and `action` are ENUMs,** like the status columns: a new kind of entry comes with code anyway.

### What gets recorded

One row per request that changes something. When a request changes fields and status together, the row's `action` is the status change and its values carry every changed field.

| Action | Event | Announcement | Club |
| --- | --- | --- | --- |
| `created` | `POST /clubs/{clubId}/events` (a create as `posted` is one row, with `status` in its values) | `POST /clubs/{clubId}/announcements` | `POST /clubs` |
| `updated` | `PATCH /auth/events/{eventId}` without a status change; the flyer confirm | `PATCH /auth/announcements/{announcementId}`; the image confirm (announcements phase 3) | `PATCH /clubs/{clubId}`; the logo confirm |
| `posted` | `PATCH` `draft → posted`, or `cancelled → posted` (shown as "posted again") | `PATCH` `draft → posted` | — |
| `cancelled` | `PATCH` `posted → cancelled` | — | — |
| `deleted` | `DELETE /auth/events/{eventId}` | `DELETE /auth/announcements/{announcementId}` | No route deletes a club yet. |
| `restored` | `POST /auth/events/{eventId}/restore` | `POST /auth/announcements/{announcementId}/restore` | — |
| `purged` | `POST /admins/purge` | `POST /admins/purge` | — |

**Fields diffed**, by API name: events `title`, `location`, `rsvpLink`, `startDate`, `endDate`, `timezone`, `status`, `description`, `associates` (club ids), `thumbnailId`; announcements `title`, `body`, `status`, `associates`, and later `thumbnailId`; clubs `name`, `description`, `websiteUrl`, `tags`, `thumbnailId`. `updatedAt` never counts as a change.

Example, an e-board member moving an event by an hour:

```json
{
  "entityType": "event", "entityId": 42, "entityLabel": "Hack Night", "action": "updated",
  "actor": { "studentId": "aaaaaaaa-…", "role": "eboard" }, "clubIds": [2, 3],
  "oldValues": { "startDate": "2026-10-14T22:00:00Z", "endDate": "2026-10-15T01:00:00Z" },
  "newValues": { "startDate": "2026-10-14T23:00:00Z", "endDate": "2026-10-15T02:00:00Z" }
}
```

### Writing it

A Go helper in the API module, called **inside the handler's transaction as its last write**:

```go
audit.Record(ctx, tx, audit.Entry{
    EntityType: audit.Event, EntityID: 42, Label: after.Title, Action: audit.Updated,
    ActorID: sub, ActorRole: role, ClubIDs: union(before.ClubIDs, after.ClubIDs),
    Old: before, New: after, // the helper keeps only the fields that differ
})
```

- The `PATCH` handlers already lock the row (`SELECT_event_for_update.sql`, `SELECT_announcement_for_update.sql`) and merge the request, so they hold `before` and the merged `after` before they write: no extra read.
- If the audit insert fails, the change rolls back, so the history is complete by construction.
- Single-statement writes (the soft deletes and restores) become two statements in one transaction.
- The helper is a convention for every write handler in the [api/ layout](../../api/LAYOUT.md). Adding it while the handlers are first written is much cheaper than retrofitting them.

**Queries (proposed):** `audit/create/INSERT_audit_entry.sql` (write), `audit/list/SELECT_club_history.sql`, `audit/list/SELECT_entity_history.sql` (reads), and `audit/purge/DELETE_audit_of_purged.sql` with `INSERT_audit_purged.sql` for the purge.

## Endpoints

All Proposed, all ⬜.

| Access | Method | Path | What it does | Role | Phase |
| --- | --- | --- | --- | --- | --- |
| 🔴 | GET | `/clubs/{clubId}/history` | The club's feed: its events, announcements, and itself, newest first. | E-board or owner of `clubId`, or an admin | 1 |
| 🔴 | GET | `/auth/events/{eventId}/history` | One event's history, deleted events included. | E-board or owner of any club linked to it, or an admin | 1 |
| 🔴 | GET | `/auth/announcements/{announcementId}/history` | One announcement's history. | E-board or owner of any club it belongs to, or an admin | 2 |
| 🔴 | GET | `/admins/history` | Every club's history, filterable by actor and club. | Admin | 3 |

### 🔴 GET `/clubs/{clubId}/history`

**Auth:** 🔴 JWT + e-board or owner of `clubId`, or an admin. `403` otherwise; `404` for an unknown club.

**Query params:** `entityType` (`event`, `announcement`, or `club`; optional), `limit` (default 50, `1`–`100`), `before` (an entry id, for "load more").

**Response `200` (Proposed):**

```json
{
  "message": "Successfully fetched 2 history entries",
  "entries": [
    {
      "id": 5012, "at": "2026-09-28T19:14:03Z", "action": "cancelled",
      "entity": { "type": "event", "id": 42, "label": "Hack Night", "exists": true, "deletedAt": null },
      "actor": { "studentId": "aaaaaaaa-…", "displayName": "ada@myhunter.cuny.edu", "role": "eboard" },
      "changes": [ { "field": "status", "from": "posted", "to": "cancelled" } ]
    },
    {
      "id": 4990, "at": "2026-09-20T13:02:11Z", "action": "deleted",
      "entity": { "type": "announcement", "id": 7, "label": "Room change", "exists": true, "deletedAt": "2026-09-20T13:02:11Z", "restorableUntil": "2026-10-20T13:02:11Z" },
      "actor": { "studentId": "bbbbbbbb-…", "displayName": "An admin", "role": "admin" },
      "changes": []
    }
  ]
}
```

- `changes` is `oldValues` and `newValues` turned into a list. Long text (`description`, `body`) is returned whole; the UI shows a word diff.
- `actor.displayName` is the student's name when `student_info` has one, else their email ([decision 6](../../api/README.md#decisions): the e-board may see members' emails), else "Former member" when the actor was removed. Admins acting outside their own clubs show as "An admin" (an [open question](#open-questions)).
- `entity.exists` and `deletedAt` come from a left join to the live table, so the UI can offer Restore within the window; a purged entity has `exists: false`.

**Queries (proposed):** `EXISTS_club.sql` (`404`) → `IS_student_authorized_club.sql` or `IS_admin.sql` (`403`) → `audit/list/SELECT_club_history.sql` (read).

### 🔴 GET `/auth/events/{eventId}/history`

**Auth:** 🔴 JWT + e-board or owner of any club linked to the event, or an admin. The existing [`IS_student_authorized_event.sql`](../../database/queries/authorization/events/can_manage/IS_student_authorized_event.sql) reads `events_to_clubs` directly, so it already works for a soft-deleted event, which is when history matters most. After a purge the links are gone; the event's single `purged` row is still in the club feed.

**Response `200` (Proposed):** the same entry shape, oldest first, for one event.

## Images and storage

**No change to image retention.** A flyer or logo change is recorded as `thumbnailId` from and to: ids, never files. The history shows the old image only while it still exists: a replaced flyer is released at once and purged after 30 days ([I4](../../database/docs/schema-review.md#i4-deleting-an-image)), and after that the entry reads "previous flyer (no longer available)". A taken-down image is never shown, whatever the history says.

That answers I4's note ("if edit history should show past flyers, a replaced flyer must be kept while a revision references it"): it shouldn't. Keeping old flyers alive for history would make every takedown and purge check the log too.

## Purge and privacy

- **Purge.** `POST /admins/purge` hard-deletes events and announcements deleted more than 30 days ago. Their history rows hold their content (titles, descriptions), so the purge deletes those rows too, in its first transaction, and writes one `purged` row per entity: label, clubs, and the admin who ran it, no values. "Removed for good" stays true, and the club can still see that it happened.
- **Students.** Account deletion isn't designed. When it is, `SET NULL` on `fk_actor_id` anonymizes the actor, but values that name a student (a future role change) need scrubbing too.
- **Who reads it.** Only the e-board and owners of a club in `club_ids`, and admins. History may contain draft text and emails, so it never appears on a 🟢 route.

## Rough AWS cost

About nothing, and no new AWS service. Using the [shared assumptions](README.md#scale-and-price-assumptions): 50 active clubs × (15 events × ~5 changes + 15 announcements × 2 + a few club edits) ≈ 5,500 rows a semester. At ~1.5 KB a row (descriptions make some large), that's about 8 MB a semester of RDS storage. Reads are rare; each is one indexed query.

## Changes to existing tables and endpoints

- **Tables:** none altered. `audit_log` is new.
- **Endpoints:** every write handler for clubs, events, and announcements gains one `INSERT_audit_entry` in its transaction, and its **Queries** line in `api/endpoints/` grows by one: `POST /clubs`, `PATCH /clubs/{clubId}`, the logo and flyer confirms, `POST /clubs/{clubId}/events`, `PATCH`, `DELETE`, and restore on `/auth/events/{eventId}`, and the four announcement writes. The soft deletes and restores stop being single statements, which [Transactions](../../database/README.md#transactions) records.
- **Purge:** `POST /admins/purge` gains two statements in its first transaction. Its request and response don't change.
- **Docs, once accepted:** H1 in the schema review points here, and coverage.md's "Edit history" leaves "Not covered".

## Open questions

1. **Admins in a club's history:** name them, or show "An admin"?
2. **Memberships:** record joins and leaves (high volume, and members may not expect it), or only role changes by owners (proposed for phase 2)?
3. **Public "edited" marks:** should posted events show "Updated 2h ago" to everyone? That needs only `updatedAt`, not this log.
4. **Revert:** a "Restore this version" button that applies `oldValues`? Proposed for later, if asked for.
5. **Retention:** keep history indefinitely (proposed, apart from the purge rule), or trim it after some years?
6. **Event authors:** if [the open question](../../api/README.md#open-questions) gives an event's author rights beyond the e-board, `actor_role` needs an `author` value.

## Phased plan

| Phase | Scope | Size |
| --- | --- | --- |
| **1. First version** | `audit_log`, the `audit.Record` helper as a convention in every club, event, and announcement write handler as they're built, the purge rule, `GET /clubs/{clubId}/history` and `GET /auth/events/{eventId}/history`, and Manage → History. | Small: 1 table, 2 routes, ~5 queries, one line per write handler. |
| 2 | The announcement history route; "Recently deleted" with Restore buttons built from the feed; role changes, verification, and image takedowns in the log; export and import entries ([import-export.md](import-export.md)). `action` gains `role_changed`, `verified`, `unverified`, `image_taken_down`, `exported`, and `imported`. | Small. |
| 3 | `GET /admins/history` with filters; revert, if wanted. | Small. |
