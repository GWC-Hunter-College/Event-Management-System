# Database module

The provider-independent MySQL database module: the baseline schema, development seed data, database-facing models, the application's SQL, and a Go client. It holds one baseline schema (up and down), one seed, 83 SQL query files covering every documented endpoint that's needed Now or Later, and an independent Go client. Nothing calls it yet: every deployed handler still runs from [`infrastructure/legacy/`](../infrastructure/legacy/), and the module has no migration runner.

This file is also the map of every query the API needs, as [32 query groups](#query-groups), each linked to its SQL and the endpoints it serves. The [directory layout](#directory-layout) lists every file, and [Transactions](#transactions) gives the order and boundary of every endpoint that runs more than one statement.

## Boundary

**MySQL is the application database technology. Amazon RDS is only the current hosting environment.** The schema and query behavior target any MySQL host:

- Amazon RDS for MySQL;
- MySQL on Amazon EC2;
- another compatible MySQL host; and
- local MySQL.

Provider independence here means independence from AWS hosting, credentials, networking, and deployment APIs. It doesn't mean converting the MySQL schema or SQL dialect to another database engine.

The module owns MySQL schema, migrations, query behavior, transactions, and database-facing models. AWS Secrets Manager lookup, RDS endpoints, VPC attachment, Lambda configuration, and IAM permissions aren't in it; they sit behind adapters on the API/infrastructure side.

## Sources

The legacy code is authoritative for what's deployed. The query groups were checked against:

- the initializer's selected DDL and seed migrations;
- all 21 embedded application SQL files under [`utils/query_client/queries/`](../infrastructure/legacy/utils/query_client/queries/);
- every current handler query call;
- query-client connection, loading, and transaction code;
- inactive endpoint stub SQL; and
- the endpoint reference in [api/README.md](../api/README.md), including its proposed contracts and decisions.

The historical `GWC Website Documentation.pdf` supplied entity and endpoint intent only. Its older conceptual schema, July-era SQL assumptions, and project-management checkmarks don't override the code or the API reference.

## Status legend

- ✅ **SQL written and run:** every statement the group needs exists in [`queries/`](queries/) and was executed against MySQL 8.0.46 with the seed ([how](#how-the-queries-were-tested)). Callers, handlers, and route changes are separate work in the API.
- 🟨 **Partial:** only reusable primitives exist, or no endpoint uses the SQL yet.
- ⬜ **No SQL:** not written, because no endpoint that's needed Now or Later calls for it.

## Query group summary

There are **32 query groups**:

| Status | Count |
| --- | ---: |
| ✅ SQL written and run | 30 |
| 🟨 Partial | 1 |
| ⬜ No SQL | 1 |
| **Total** | **32** |

The 🟨 group is [19](#19-add-a-specified-club-member) (adding a student other than the caller: the insert exists, no endpoint does it). The ⬜ group is [17](#17-my-e-board-clubs) (`GET /me/clubs/eboard`, frontend need "—": the frontend filters `GET /me/clubs`).

Groups 30 (club update) and 31 (admin purge) are new in Phase 4, and group 32 (announcements) after it. The future features still listed as not covered by the API (board, edit history, import and export) have no queries.

## Baseline schema

There's no live database, so the schema is one baseline that creates everything from scratch, not a chain of migrations:

- [`migrations/schema/2026_09_29_baseline_up.sql`](migrations/schema/2026_09_29_baseline_up.sql): every table, the six topics, and the three read views.
- [`migrations/schema/2026_09_29_baseline_down.sql`](migrations/schema/2026_09_29_baseline_down.sql): drops everything. Only for rebuilding a local or test database.
- [`migrations/seed/2026_09_29_baseline_seed.sql`](migrations/seed/2026_09_29_baseline_seed.sql): development data with explicit ids and times relative to the load date. Never load it into production.
- [`migrations/history/`](migrations/history/): the July, September, and November schema files and the September seed, kept as a read-only record. They're never applied. The deployed legacy system still uses its own copies under [`infrastructure/legacy/lambda/internal/database/init/migrations/`](../infrastructure/legacy/lambda/internal/database/init/migrations/), which are unchanged.

Once a real database exists, changes go in new dated migrations after the baseline. Why the baseline looks the way it does, and the decisions behind it, are in the [schema review](docs/schema-review.md). What a migration runner will need is [future work](docs/schema-review.md#future-work-a-migration-runner).

![Legacy (November) database schema](assets/database-schema.png)

The diagram shows the November schema that legacy deploys. The baseline adds `topics` and `club_tags`, the image columns, `club_members.role` in place of the two flags, the announcement tables, and the three views, and removes `archived`; see the table below.

### Tables and views

| Table | Purpose | Queries |
| --- | --- | --- |
| `students` | Cognito `sub` identity and optional email | Ensure/upsert, current student, member and admin lists. |
| `student_info` | Optional username and name | Read by the member and admin lists. Nothing writes it yet. |
| `clubs` | Club identity, unique name, logo image | Every club read; create, rename, logo confirm, takedown. |
| `club_info` | Website URL and description | Create; upserted by club update. |
| `club_members` | Membership, with one `role`: `member`, `eboard`, or `owner` | Join, leave (never the last owner), owner on create, member list, role update, authorization. |
| `verified_clubs` | Directory approval | Verified filter; verify and unverify. |
| `admins` | Global administrators; never empty once the first is created | Admin check and admin CRUD; the last admin can't be removed. |
| `topics` | The fixed topic list, loaded by the baseline | Referenced by `club_tags`. |
| `club_tags` | Up to three topics per club, in order | Replaced by create and update; read through `club_details`. |
| `events` | Event fields; `status` is `draft`, `posted`, or `cancelled`; UTC times; `deleted_at` soft delete | Create, update, status transitions, soft delete, restore (30 days), admin purge; read through `event_details`. |
| `event_descriptions` | One description per event | Create; upserted by update. |
| `event_tags` | Free-text event tags | Nothing reads or writes them yet. |
| `images` | One row per stored file, with owning club, uploader, alt text, and `deleted_at` | Insert on confirm, release, takedown, purge. |
| `events_to_clubs` | Event-to-club links, at most one owner per event | Create, co-host replacement, lists, event authorization. |
| `event_images` | Gallery links, ordered by their own `created_at` | Gallery confirm and list, takedown. |
| `announcements` | Club updates: title, body, `status` `draft` or `posted`, `posted_at` (set exactly when posted), `deleted_at` soft delete | Create, update, post, soft delete, restore (30 days), admin purge; read through `announcement_details`. |
| `announcements_to_clubs` | Announcement-to-club links, at most one owner club per announcement | Create, associate replacement, club lists, announcement authorization. |
| `event_details` (view) | One row per event that isn't soft-deleted, with flyer, description, photo count, owner club, and co-hosts | Every event read. |
| `announcement_details` (view) | One row per announcement that isn't soft-deleted, with owner club and co-owning clubs | Every announcement read. |
| `club_details` (view) | One row per club, with logo, `club_info`, tags, member count, and verified | Club list and detail. |

## Directory layout

Every file under [`queries/`](queries/), grouped by resource. **Legacy** files are byte-for-byte copies of `infrastructure/legacy/utils/query_client/queries/`, pinned by [`inventory_test.go`](queries/inventory_test.go). **Fixed** files are legacy copies changed on purpose in Phase 4; their pins are the new bytes. Everything else is **new**. Each new and fixed file starts with a comment giving what it does, the endpoints that use it, its `?` parameters in order, and what it returns; a test checks that header. Legacy files keep their original bytes, so their parameters are listed here.

```text
queries/
├── students/ensure/
│   ├── EXISTS_student_by_sub.sql           legacy   (sub)
│   ├── UPSERT_student.sql                  legacy   (sub, email)
│   └── UPSERT_student_sub_only.sql         legacy   (sub)
├── me/
│   ├── get/SELECT_student_by_sub.sql       legacy   (sub)
│   ├── clubs/list/SELECT_student_clubs.sql fixed
│   └── events/list/SELECT_student_events.sql fixed
├── clubs/
│   ├── list/SELECT_clubs.sql               fixed
│   ├── get/
│   │   ├── SELECT_club.sql                 fixed
│   │   └── EXISTS_club.sql
│   ├── create/
│   │   ├── INSERT_club.sql                 legacy   (name)
│   │   ├── INSERT_club_info.sql            legacy   (club id, website url, description)
│   │   └── INSERT_club_owner.sql
│   ├── update/
│   │   ├── UPDATE_club.sql
│   │   └── UPSERT_club_info.sql
│   ├── tags/replace/
│   │   ├── DELETE_club_tags.sql
│   │   └── INSERT_club_tag.sql
│   ├── members/
│   │   ├── create/INSERT_club_member.sql   fixed
│   │   ├── leave/DELETE_club_member.sql    fixed
│   │   ├── list/SELECT_club_members.sql
│   │   └── update_role/
│   │       ├── SELECT_club_owners_for_update.sql
│   │       └── UPDATE_club_member_role.sql
│   ├── events/
│   │   ├── list/SELECT_club_events.sql     fixed
│   │   └── drafts/SELECT_club_drafts.sql
│   ├── thumbnails/confirm/
│   │   ├── SELECT_club_logo_for_update.sql
│   │   └── UPDATE_club_logo.sql
│   ├── announcements/
│   │   ├── list/SELECT_club_announcements.sql
│   │   └── drafts/SELECT_club_announcement_drafts.sql
│   └── verification/
│       ├── create/INSERT_verified_club.sql
│       └── delete/DELETE_verified_club.sql
├── announcements/
│   ├── read/
│   │   ├── SELECT_announcement.sql
│   │   └── SELECT_announcement_status.sql
│   ├── create/
│   │   ├── INSERT_announcement.sql
│   │   └── INSERT_announcement_club_link.sql
│   ├── update/
│   │   ├── SELECT_announcement_for_update.sql
│   │   ├── UPDATE_announcement.sql
│   │   ├── DELETE_announcement_associates.sql
│   │   └── UPDATE_announcement_status_posted.sql
│   ├── delete/UPDATE_announcement_soft_delete.sql
│   ├── restore/
│   │   ├── UPDATE_announcement_restore.sql
│   │   └── SELECT_announcement_deletion.sql
│   └── purge/DELETE_purgeable_announcements.sql       POST /admins/purge
├── events/
│   ├── read/
│   │   ├── SELECT_events.sql               fixed
│   │   ├── SELECT_event.sql
│   │   └── SELECT_event_status.sql
│   ├── create/
│   │   ├── INSERT_event.sql                fixed
│   │   ├── INSERT_event_description.sql    fixed
│   │   └── INSERT_event_club_link.sql      legacy   (event id, club id, is owner)
│   ├── update/
│   │   ├── SELECT_event_for_update.sql
│   │   ├── UPDATE_event.sql
│   │   ├── UPSERT_event_description.sql
│   │   ├── DELETE_event_associates.sql
│   │   ├── UPDATE_event_status_posted.sql
│   │   └── UPDATE_event_status_cancelled.sql
│   ├── delete/UPDATE_event_soft_delete.sql
│   ├── restore/
│   │   ├── UPDATE_event_restore.sql
│   │   └── SELECT_event_deletion.sql
│   ├── thumbnails/confirm/UPDATE_event_thumbnail.sql
│   ├── images/
│   │   ├── list/SELECT_event_images.sql
│   │   └── confirm/INSERT_event_image.sql  legacy   (event id, image id)
│   └── purge/                              POST /admins/purge
│       ├── UPDATE_images_of_purgeable_events.sql
│       └── DELETE_purgeable_events.sql
├── images/
│   ├── create/INSERT_image.sql             fixed, moved from events/images/confirm/
│   ├── get/SELECT_image.sql
│   ├── release/UPDATE_image_soft_delete_if_unused.sql
│   ├── delete/
│   │   ├── UPDATE_clubs_clear_logo.sql
│   │   ├── UPDATE_events_clear_thumbnail.sql
│   │   ├── DELETE_event_image_links.sql
│   │   └── UPDATE_image_soft_delete.sql
│   └── purge/                              POST /admins/purge
│       ├── SELECT_purgeable_images.sql
│       └── DELETE_purged_image.sql
├── authorization/
│   ├── clubs/
│   │   ├── can_manage/IS_student_authorized_club.sql   fixed
│   │   ├── is_member/IS_club_member.sql
│   │   └── is_owner/IS_club_owner.sql
│   ├── events/
│   │   ├── can_manage/IS_student_authorized_event.sql  fixed
│   │   └── manages_owner_club/IS_student_owner_club_manager.sql
│   ├── announcements/
│   │   ├── can_manage/IS_student_authorized_announcement.sql
│   │   └── manages_owner_club/IS_student_announcement_owner_club_manager.sql
│   └── admins/is_admin/IS_admin.sql
└── admins/
    ├── list/SELECT_admins.sql
    ├── get/SELECT_admin.sql
    ├── create/INSERT_admin.sql
    └── delete/
        ├── SELECT_admins_for_update.sql
        └── DELETE_admin.sql
```

Totals: 83 files: 8 legacy, 13 fixed, 62 new. Names follow the existing prefixes: `SELECT_`, `INSERT_`, `UPDATE_`, `DELETE_`, `UPSERT_`, `EXISTS_`, `IS_`. The soft deletes and the restore are `UPDATE_…` because the statement is an `UPDATE`; `DELETE_` is for the purge, link removals, leaving a club, and removing an admin.

Note the legacy parameter orders: `IS_student_authorized_event.sql` takes the event id first, while every other authorization query takes the student id first, and `INSERT_club_member.sql` takes the club id first.

**Authorization** is its own reusable queries, never copied into an endpoint's SQL. A handler runs the check, then the endpoint's query:

| Check | Query | Rule |
| --- | --- | --- |
| Is member | `authorization/clubs/is_member/IS_club_member.sql` | Any membership row. |
| Can manage club | `authorization/clubs/can_manage/IS_student_authorized_club.sql` | `role IN ('eboard', 'owner')` in the club. |
| Can manage event | `authorization/events/can_manage/IS_student_authorized_event.sql` | E-board or owner of **any** club linked to the event. |
| Manages the event's owning club | `authorization/events/manages_owner_club/IS_student_owner_club_manager.sql` | E-board or owner of the event's owning club only. Used by restore, together with the admin check. |
| Can manage announcement | `authorization/announcements/can_manage/IS_student_authorized_announcement.sql` | E-board or owner of **any** club the announcement belongs to. |
| Manages the announcement's owning club | `authorization/announcements/manages_owner_club/IS_student_announcement_owner_club_manager.sql` | E-board or owner of the owning club only. Used by restore, with the admin check. |
| Is owner | `authorization/clubs/is_owner/IS_club_owner.sql` | `role = 'owner'` in the club. |
| Is admin | `authorization/admins/is_admin/IS_admin.sql` | Has an `admins` row. |

### Fixed defects

Each fix is in the module's copy only; `infrastructure/legacy` is unchanged.

| File | Defect | Fix |
| --- | --- | --- |
| `events/create/INSERT_event.sql` | 11 columns, 10 values; no value for `rsvp_link`; wrote `drafted` and `NOW()` | 8 columns, 8 parameters, status is a parameter, timestamps take their defaults. |
| `events/create/INSERT_event_description.sql` | Trailing comma in the column list | Removed. |
| `events/images/list/SELECT_event_images.sql` | Missing (the legacy gallery read loads a file that doesn't exist) | Written: gallery without the flyer, in upload order. |
| `events/read/SELECT_events.sql`, `clubs/events/list/SELECT_club_events.sql`, `me/events/list/SELECT_student_events.sql` | Paged event-to-club rows, not events; matched ids and status with `LIKE`; strict date bounds; posted only; `/me/events` and club lists dropped other linked clubs | Select from `event_details`, one row per event, so `LIMIT` counts events; `=` and `IN` instead of `LIKE`; `posted` and `cancelled`; every linked club; the proposed list parameters. |
| `clubs/list/SELECT_clubs.sql`, `clubs/get/SELECT_club.sql` | No description (list), tags, member count, or verified flag | Select from `club_details`. |
| `images/create/INSERT_image.sql` | No owning club, uploader, or alt text; `NOW()` for `created_at` | Added the three columns; `created_at` takes its default. Moved to `images/create/`, because every confirm shares it. |
| `authorization/clubs/can_manage/IS_student_authorized_club.sql`, `authorization/events/can_manage/IS_student_authorized_event.sql`, `clubs/members/create/INSERT_club_member.sql`, `me/clubs/list/SELECT_student_clubs.sql` | Read or wrote the two role flags, which the baseline replaces with one `role` column ([decision 11](../api/README.md#decisions)) | Read and write `role`. `SELECT_student_clubs.sql` also orders by club name. Parameter orders are unchanged. |
| `clubs/members/leave/DELETE_club_member.sql` | Refused every owner, and `member_is_owner = FALSE` didn't match `NULL` | Refuses only the club's last owner, counted in the same statement. |

**List parameters** ([proposed](../api/endpoints/events.md#proposed-list-parameters)), shared by the three public lists: `when` is `upcoming` (not ended yet, `end_date > UTC_TIMESTAMP()`, start ascending), `past` (ended, start descending), or `NULL` (start ascending). `startDate` and `endDate` become a half-open UTC range `[start, end)`, and an event matches when it overlaps it: `start_date < end AND (end_date > start OR start_date >= start)`, so a multi-day event that started earlier still shows up. The API converts a date to 00:00 that day in `America/New_York` (for `endDate`, 00:00 the next day, so the whole day is included) and a timestamp with an offset as given, both to UTC ([decision 10](../api/README.md#decisions)). The last clause keeps a zero-length event that starts exactly at the range start. `limit` (default 50, maximum 100) and `page` are validated by the API and passed as `LIMIT` and `OFFSET`. The drafts list takes only `limit` and `page` and orders by `updated_at` descending.

## Transactions

Endpoints that run more than one statement. Each "in one transaction" workflow uses the client's transaction helpers, and runs its statements in this order.

| Endpoint | Order | Boundary |
| --- | --- | --- |
| `POST /clubs` | `INSERT_club` → `INSERT_club_info` → `INSERT_club_owner` (caller) → `INSERT_club_tag` once per topic, `slot` = position + 1 | One transaction. A duplicate name (`409`), an unknown topic, or a fourth topic (`400`) rolls everything back. |
| `PATCH /clubs/{clubId}` | `SELECT_club` (merge omitted fields) → `UPDATE_club` → `UPSERT_club_info` → if `tags` sent: `DELETE_club_tags` → `INSERT_club_tag` × n → `SELECT_club` for the response | One transaction, after `EXISTS_club` (`404`) and the can-manage-club check (`403`). Last write wins. |
| `POST /clubs/{clubId}/events` | `INSERT_event` (last insert id is the event id) → `INSERT_event_description` if there's a description → `INSERT_event_club_link` (owner, `TRUE`) → `INSERT_event_club_link` (`FALSE`) per associate, deduplicated | One transaction, after `EXISTS_club` (`404`) and the can-manage-club check (`403`). A bad associate id fails its foreign key and rolls back. |
| `PATCH /auth/events/{eventId}` | `SELECT_event_for_update` (locks the row; no row is `404`) → merge → `UPDATE_event` → `UPSERT_event_description` if sent → if `associates` sent: `DELETE_event_associates` → `INSERT_event_club_link` × n → if `status` sent and differs from the current one: `UPDATE_event_status_posted` or `UPDATE_event_status_cancelled` (0 rows is `400`) → commit → `SELECT_event` (`public_only` `FALSE`) and `SELECT_event_images` for the response | One transaction, after the can-manage-event check (`403`). A `status` equal to the current one skips the status update. |
| `DELETE /auth/events/{eventId}` | `UPDATE_event_soft_delete` | One statement. 0 rows is `404`. |
| `POST /auth/events/{eventId}/restore` | `IS_student_owner_club_manager` or `IS_admin` (`403`) → `UPDATE_event_restore`. On 0 rows, `SELECT_event_deletion`: no row is `404`, not deleted is `409`, deleted 30 or more days ago is `410`. Then `SELECT_event` (`public_only` `FALSE`) and `SELECT_event_images` for the response | One statement does the restore; the window check is in its `WHERE` clause. |
| `POST /clubs/{clubId}/events/{eventId}/thumbnails/confirm` | `SELECT_event_for_update` (no row is `404`) → `INSERT_image` (`event-thumbnail`, owner club, caller, alt text) → `UPDATE_event_thumbnail` → `UPDATE_image_soft_delete_if_unused` for the old flyer, if there was one | One transaction, after the can-manage-event check and the S3 existence check. |
| `POST /clubs/{clubId}/thumbnails/confirm` | `SELECT_club_logo_for_update` (no row is `404`) → `INSERT_image` (`club-thumbnail`) → `UPDATE_club_logo` → `UPDATE_image_soft_delete_if_unused` for the old logo, if any | One transaction, after the can-manage-club check and the S3 existence check. |
| `POST /clubs/{clubId}/events/{eventId}/images/confirm` | `SELECT_event_for_update` (for `owner_club_id`; no row is `404`) → `INSERT_image` (`event-image`) → `INSERT_event_image` | One transaction. Replaces the legacy handler's two separate commits. |
| `DELETE /images/{imageId}` (takedown) | `SELECT_image` (no row is `404`; `fk_club_id` for the check) → `UPDATE_clubs_clear_logo` → `UPDATE_events_clear_thumbnail` → `DELETE_event_image_links` → `UPDATE_image_soft_delete` | One transaction, after the can-manage-club check on the owning club, or `IS_admin`. An admin's purge deletes the S3 object once the file has been deleted for more than 30 days. |
| `PUT /clubs/{clubId}/members/roles` | `SELECT_club_owners_for_update` → `UPDATE_club_member_role`. On 0 rows: `IS_club_member` (0 is `404`); otherwise, if the member already has that role, `200`; else `409` (last owner) | One transaction, after `EXISTS_club` and `IS_club_owner` for the caller (`403`). Locking the owner rows stops two owners demoting each other at once. |
| `POST /clubs/{clubId}/members/me` | `EXISTS_club` (`404`) → `INSERT_club_member` (0 rows is `409`) | Two statements, no transaction needed: the insert is guarded on its own. |
| `POST /clubs/{clubId}/announcements` | `INSERT_announcement` (last insert id is the announcement id) → `INSERT_announcement_club_link` (owner, `TRUE`) → `INSERT_announcement_club_link` (`FALSE`) per associate, deduplicated | One transaction, after `EXISTS_club` (`404`) and the can-manage-club check (`403`). |
| `PATCH /auth/announcements/{announcementId}` | `SELECT_announcement_for_update` (no row is `404`) → merge → `UPDATE_announcement` → if `associates` sent: `DELETE_announcement_associates` → `INSERT_announcement_club_link` × n → if `status` is `posted` and the current status is `draft`: `UPDATE_announcement_status_posted` (0 rows is `400`: no body) → commit → `SELECT_announcement` (`public_only` `FALSE`) | One transaction, after the can-manage-announcement check (`403`). |
| `DELETE /auth/announcements/{announcementId}`, `POST .../restore` | As for events: `UPDATE_announcement_soft_delete` (0 rows is `404`); restore is `IS_student_announcement_owner_club_manager` or `IS_admin` → `UPDATE_announcement_restore`, and on 0 rows `SELECT_announcement_deletion` (`404`, `409`, or `410`) | One statement each. |
| `DELETE /clubs/{clubId}/members/me` | `SELECT_club_owners_for_update` → `DELETE_club_member`; on 0 rows `IS_club_member` (0 is `404`, 1 is `403`: the caller is the last owner) | One transaction. Locking the owner rows stops the last two owners leaving at once. |
| `POST /admins`, `POST /clubs/{clubId}/verification` | `INSERT_admin` or `INSERT_verified_club`; on 0 rows, `SELECT_student_by_sub` or `EXISTS_club` tells missing (`404`) from already done | No transaction needed. |
| `DELETE /admins/{studentId}` | `SELECT_admins_for_update` → `DELETE_admin`; on 0 rows `IS_admin` (0 is `404`, 1 is `409`: the last admin) | One transaction, after the caller's admin check. Locking every admin row stops two admins removing each other at once. |
| `POST /admins/purge` | `IS_admin` (`403`). **Transaction 1:** `UPDATE_images_of_purgeable_events` → `DELETE_purgeable_events` → `DELETE_purgeable_announcements`. **Then, in batches until one comes back empty:** `SELECT_purgeable_images` → for each row: `DELETE_purged_image` (commit) → delete the S3 object | Transaction 1 is atomic. Each image is its own step, and its row goes before its object, so a failed S3 delete leaves only a harmless orphan object. Every purge query hard-codes the 30 days that `UPDATE_event_restore` uses. |

Reads with a follow-up check, not a transaction: `GET /events/{eventId}` is `SELECT_event` (`public_only` `TRUE`, no row is `404`) then `SELECT_event_images`. `GET /auth/events/{eventId}` is the can-manage-event check, then the same two with `public_only` `FALSE`. The gallery routes run `SELECT_event_status` first (public routes need `posted` or `cancelled`).

## How the queries were tested

On a local machine, in a throwaway MySQL 8.0.46 Docker container (RDS runs 8.0.37) with the default strict `sql_mode` and the server time zone at `+00:00`. Only files under `database/` were mounted, read-only. Nothing from `infrastructure/legacy` was run, built, or imported, and no AWS credentials, `.env` files, or remote databases were used. The container was removed afterwards.

1. Applied the baseline, then the seed.
2. Ran every one of the 83 query files with bound parameters: each file's text was sent to `PREPARE … FROM` unchanged, with `EXECUTE … USING` for the parameters, the same binding the Go driver uses. The results were checked against the seed. Among them:
   - Paging with `limit` 2 returned four pages of distinct events. The three-club hackathon counted once.
   - `when=upcoming` included the ongoing event and ordered by start ascending. `when=past` ordered descending. Cancelled events appeared; drafts and soft-deleted events didn't.
   - Every illegal status change affected 0 rows: `draft → cancelled`, `posted → posted`, `cancelled → cancelled`, posting a draft with no location, and changing a soft-deleted event. The allowed `draft → posted → cancelled → posted` chain affected 1 row each.
   - The role update refused to demote a club's last owner, and allowed demoting one of two. An owner could leave while a second owner remained; the last owner couldn't. The e-board filter returned e-board members and owners.
   - Date ranges matched by overlap: a range starting the day after the hackathon began still returned it, and an event starting exactly at the (exclusive) range end was left out.
   - Restore brought back an event deleted 3 days ago, and one deleted 29 days 23 hours 59 minutes ago. It refused one deleted 45 days ago, an event that wasn't deleted, and an unknown id, and `SELECT_event_deletion` told those apart.
   - The last admin couldn't be removed, even by themselves, and removing one of two admins worked.
   - Announcements: club lists returned posted announcements by `posted_at`, newest first. The seed's announcement drafted 25 days ago but posted 12 hours ago came first, and a just-posted one went above it. Lists included one owned by another club; drafts and deleted ones stayed out. Creating as posted set `posted_at`, a draft left it `NULL`, and the CHECK refused a posted row without `posted_at` and a draft with one. A restore kept the original `posted_at`, and a 256-character title was rejected, not truncated. Posting a draft with no body, and posting a posted one, affected 0 rows. A second owner club and a `cancelled` status were refused. Delete, restore (5 days: yes; 40 days: no), and the purge (the 40-day one only, with its club links) behaved as for events.
   - A flyer that's also in its gallery was excluded from `image_count` and the gallery. The release query kept a photo that was still in two galleries; the takedown removed both links, then soft-deleted it.
   - The purge removed the event soft-deleted 45 days ago, kept the one deleted 29 days 23 hours ago, and cascaded its links. It then soft-deleted and purged that event's unshared flyer and photo. Deleting a still-referenced image was refused.
   - Constraint errors: a fourth topic, wrong-case and unknown topics, a duplicate topic, a duplicate club name in different case, end before start, statuses `'archived'` and `'drafted'`, and a duplicate `object_key`.
3. Ran the down file on the changed database, which left it empty. Then ran the baseline and seed again, and the schema dump matched the first one exactly.

`go test ./...` in `database/` checks the pins for the 21 copied files, the header of every new or fixed file, the naming prefixes, that every `.sql` file on disk is in the `go:embed` list, complete loading of each file, the client's transactions, and model scans. It doesn't connect to MySQL; the container run above is the only execution against a database, and it isn't automated yet.

## The first admin

There's no endpoint for creating the first admin: `POST /admins` needs an admin to call it ([decision 12](../api/README.md#decisions)). A developer creates the first one by hand, once per database:

1. The person signs in to the site once, so the Cognito trigger (or `RequireStudent`) creates their `students` row. Their `sub` is the user's `sub` attribute in the Cognito user pool console; `GET /me` also returns it as `student.id`.
2. Connect to the database with an account that can write to it, and run:

   ```sql
   INSERT INTO admins (fk_student_id)
   SELECT id FROM students WHERE id = '<their sub>';
   ```

   It inserts one row. 0 rows means the `sub` has no `students` row yet: they haven't signed in, or the `sub` is wrong.
3. Check it: `SELECT fk_student_id FROM admins;`

From then on, admins add and remove each other through `POST /admins` and `DELETE /admins/{studentId}`, and [`DELETE_admin.sql`](queries/admins/delete/DELETE_admin.sql) never removes the last one.

## Shared application query access

Legacy access is implemented by [`query_client.go`](../infrastructure/legacy/utils/query_client/query_client.go), [`helpers.go`](../infrastructure/legacy/utils/query_client/helpers.go), and [`types.go`](../infrastructure/legacy/utils/query_client/types.go):

- Go embeds every `queries/**/*.sql` file into each consuming binary.
- `sqlx` and the MySQL driver execute `Get`, `Select`, `Exec`, and transaction helpers.
- Active handlers use `NewClientFromHost`, which loads credentials from AWS Secrets Manager and receives an RDS host/schema through environment variables.
- `loadSQLFromFile` allocates exactly 4096 bytes and ignores the byte count returned by `Read`, so every query risks trailing null bytes and queries over 4 KiB risk truncation.
- `Get`, `Select`, and `Exec` don't use context-aware database methods.
- The unused `ChangeDatabase` helper concatenates `USE ` with a caller-supplied database name; the module doesn't include it.

The module's [`client/`](client/) accepts caller-supplied `mysql.Config` and propagates request contexts through `Get`, `Select`, `Exec`, and transaction operations. `Open` creates a lazy pool without dialing; `Ping(ctx)` explicitly checks connectivity. AWS secret retrieval and RDS endpoint resolution are outside this module. [`queries.Load`](queries/queries.go) embeds all 63 query files and reads their complete bytes.

Client configuration the baseline needs ([T1](docs/schema-review.md#t1-store-start-and-end-in-utc)): `ParseTime = true`, `Loc = UTC`, and `Params["time_zone"] = "'+00:00'"`. The go-sql-driver default `clientFoundRows = false` is assumed: affected-row counts are *changed* rows, which the 0-row checks above rely on.

The Go models in [`models/`](models/) are still the legacy copies and don't match the baseline yet: `Student.Email`, `Event.Location`, and `Event.AuthorID` need to be pointers, `ClubWithRole` reads `role` from the column now, `Event.Status` values change, and `Image` needs `ClubID`, `UploadedBy`, `AltText`, and `DeletedAt`. The view columns (`image_count`, `owner_*`, the JSON `associates` and `tags`) need model structs too. That's API cutover work; this phase changed docs and SQL only.

## Query groups

Each group lists its endpoints (the access marker and status are the endpoint's, from [api/README.md](../api/README.md#endpoints)), its files, and anything a handler must know.

### 1. Student existence and upsert

**Status:** ✅ · **Endpoints:** 🔵 Cognito student sync trigger, 🔵 `RequireStudent` (before every 🔴 route).
**Files:** `students/ensure/EXISTS_student_by_sub.sql`, `UPSERT_student.sql`, `UPSERT_student_sub_only.sql` (legacy, unchanged).
**Notes:** Request-time sync doesn't update a changed email on an existing row; whether it should is an [open question](#open-questions).

### 2. Current student

**Status:** ✅ · **Endpoints:** 🔴 [`GET /me`](../api/endpoints/me.md#-get-me).
**Files:** `me/get/SELECT_student_by_sub.sql` (legacy); the proposed `isAdmin` comes from `authorization/admins/is_admin/IS_admin.sql`.
**Notes:** `email` may be `NULL`; the Go model must allow it.

### 3. Current student's clubs and roles

**Status:** ✅ · **Endpoints:** 🔴 [`GET /me/clubs`](../api/endpoints/me.md#-get-meclubs).
**Files:** `me/clubs/list/SELECT_student_clubs.sql` (fixed: reads `club_members.role`). `thumbnail_url` is the logo's object key; the API signs it.

### 4. Current student's events

**Status:** ✅ · **Endpoints:** 🔴 [`GET /me/events`](../api/endpoints/me.md#-get-meevents).
**Files:** `me/events/list/SELECT_student_events.sql` (fixed).
**Notes:** Posted and cancelled events linked to any club the caller joined, with every linked club, one row per event.

### 5. Club list and verified filter

**Status:** ✅ · **Endpoints:** 🟢 [`GET /clubs`](../api/endpoints/clubs.md#-get-clubs).
**Files:** `clubs/list/SELECT_clubs.sql` (fixed): the proposed club object's columns, ordered by name.

### 6. Club detail

**Status:** ✅ · **Endpoints:** 🟢 [`GET /clubs/{clubId}`](../api/endpoints/clubs.md#-get-clubsclubid); `EXISTS_club` serves every club-scoped route's `404`.
**Files:** `clubs/get/SELECT_club.sql` (fixed), `clubs/get/EXISTS_club.sql`.

### 7. Club creation

**Status:** ✅ · **Endpoints:** 🔴 [`POST /clubs`](../api/endpoints/clubs.md#-post-clubs).
**Files:** `clubs/create/INSERT_club.sql`, `INSERT_club_info.sql` (legacy), `INSERT_club_owner.sql`, `clubs/tags/replace/INSERT_club_tag.sql`. Order in [Transactions](#transactions).

### 8. Join caller to club

**Status:** ✅ · **Endpoints:** 🔴 [`POST /clubs/{clubId}/members/me`](../api/endpoints/memberships.md#-post-clubsclubidmembersme).
**Files:** `clubs/members/create/INSERT_club_member.sql` (legacy), `clubs/get/EXISTS_club.sql`.

### 9. Leave caller's club

**Status:** ✅ · **Endpoints:** 🟨 [`DELETE /clubs/{clubId}/members/me`](../api/endpoints/memberships.md#-delete-clubsclubidmembersme) (the route still needs its authorizer).
**Files:** `clubs/members/update_role/SELECT_club_owners_for_update.sql`, `clubs/members/leave/DELETE_club_member.sql` (fixed), `authorization/clubs/is_member/IS_club_member.sql`.
**Notes:** Only a club's last owner can't leave ([decision 11](../api/README.md#decisions)); the deployed SQL refuses every owner.

### 10. Club event list

**Status:** ✅ · **Endpoints:** 🟢 [`GET /clubs/{clubId}/events`](../api/endpoints/club-events.md#-get-clubsclubidevents).
**Files:** `clubs/events/list/SELECT_club_events.sql` (fixed), `clubs/get/EXISTS_club.sql`.

### 11. Public and composite event read

**Status:** ✅ · **Endpoints:** 🟢 [`GET /events`](../api/endpoints/events.md#-get-events), 🟢 [`GET /events/{eventId}`](../api/endpoints/events.md#-get-eventseventid).
**Files:** `events/read/SELECT_events.sql` (fixed), `SELECT_event.sql`, `SELECT_event_status.sql`, and `events/images/list/SELECT_event_images.sql` for the detail's `images`.

### 12. Create event draft

**Status:** ✅ · **Endpoints:** 🟨 [`POST /clubs/{clubId}/events`](../api/endpoints/club-events.md#-post-clubsclubidevents).
**Files:** `events/create/INSERT_event.sql`, `INSERT_event_description.sql` (both fixed), `INSERT_event_club_link.sql` (legacy). One transaction; see [Transactions](#transactions).

### 13. Club authorization

**Status:** ✅ · **Endpoints:** 🔵 club role check, used by every club write.
**Files:** `authorization/clubs/can_manage/IS_student_authorized_club.sql` (legacy), `is_member/IS_club_member.sql`, `is_owner/IS_club_owner.sql`.

### 14. Event authorization

**Status:** ✅ · **Endpoints:** 🔵 event role check, used by every `/auth/events` route and the event image routes.
**Files:** `authorization/events/can_manage/IS_student_authorized_event.sql` (legacy). Whether an event's author or an admin gets rights of their own is an [open question](#open-questions).

### 15. Event-image metadata confirmation

**Status:** ✅ · **Endpoints:** 🔵 image metadata write; 🟨 [`POST /clubs/{clubId}/events/{eventId}/images/confirm`](../api/endpoints/images.md#-post-clubsclubideventseventidimagesconfirm), and the two thumbnail confirms (group 27).
**Files:** `images/create/INSERT_image.sql` (fixed, moved), `events/images/confirm/INSERT_event_image.sql` (legacy).

### 16. Event-image list

**Status:** ✅ · **Endpoints:** 🟨 [`GET /clubs/{clubId}/events/{eventId}/images`](../api/endpoints/images.md#-get-clubsclubideventseventidimages), ⬜ [`GET /events/{eventId}/images`](../api/endpoints/events.md#-get-eventseventidimages), ⬜ [`GET /auth/events/{eventId}/images`](../api/endpoints/event-management.md#-get-autheventseventidimages), and the `images` of the event detail reads.
**Files:** `events/images/list/SELECT_event_images.sql` (the file legacy is missing), `events/read/SELECT_event_status.sql` for visibility.
**Notes:** Columns are `image_id`, `mimetype`, `object_key`, `alt_text`. The legacy handler's nested `SelectQuerySchema` scan doesn't match them; it's rewritten with the API anyway.

### 17. My e-board clubs

**Status:** ⬜ · **Endpoints:** 🔴 [`GET /me/clubs/eboard`](../api/endpoints/me.md#-get-meclubseboard), frontend need "—".
**Notes:** Not written: the frontend filters group 3 by role. If the route is built, filter `SELECT_student_clubs.sql`'s rows by role rather than porting the stale stub SQL.

### 18. Club member and e-board listing

**Status:** ✅ · **Endpoints:** ⬜ [`GET /clubs/{clubId}/members`](../api/endpoints/memberships.md#-get-clubsclubidmembers), ⬜ [`GET /clubs/{clubId}/eboard`](../api/endpoints/memberships.md#-get-clubsclubideboard) (as `role=eboard`).
**Files:** `clubs/members/list/SELECT_club_members.sql`.
**Notes:** `eboard` returns e-board members and owners, as memberships.md defines the e-board list; `owner` and `member` match the computed role.

### 19. Add a specified club member

**Status:** 🟨 · **Endpoints:** none. The PDF's `POST /clubs/{clubId}/members` is designed as self-join only.
**Notes:** `INSERT_club_member.sql` takes any student id, so it would serve if an endpoint is ever added.

### 20. Update member roles

**Status:** ✅ · **Endpoints:** ⬜ [`PUT /clubs/{clubId}/members/roles`](../api/endpoints/memberships.md#-put-clubsclubidmembersroles).
**Files:** `clubs/members/update_role/SELECT_club_owners_for_update.sql`, `UPDATE_club_member_role.sql`, `authorization/clubs/is_owner/IS_club_owner.sql`.
**Notes:** Sets the one `role` column. A club's last owner can't be demoted.

### 21. Club draft events

**Status:** ✅ · **Endpoints:** ⬜ [`GET /clubs/{clubId}/events/drafts`](../api/endpoints/club-events.md#-get-clubsclubideventsdrafts).
**Files:** `clubs/events/drafts/SELECT_club_drafts.sql`. The status is fixed in the SQL, so no caller can choose it.

### 22. Admin CRUD

**Status:** ✅ · **Endpoints:** ⬜ [every `/admins` route](../api/endpoints/admins.md), 🔵 admin check.
**Files:** `admins/list/SELECT_admins.sql`, `get/SELECT_admin.sql`, `create/INSERT_admin.sql`, `delete/SELECT_admins_for_update.sql`, `delete/DELETE_admin.sql`, `authorization/admins/is_admin/IS_admin.sql`.
**Notes:** The admin responses aren't defined; the queries return student id, email, and name. `DELETE_admin.sql` never removes the last admin, including when an admin removes themselves ([decision 12](../api/README.md#decisions)). The first admin is created by hand: see [The first admin](#the-first-admin). What "their clubs" means is [open](#open-questions).

### 23. Club verification writes

**Status:** ✅ · **Endpoints:** ⬜ [`POST` and `DELETE /clubs/{clubId}/verification`](../api/endpoints/verification.md).
**Files:** `clubs/verification/create/INSERT_verified_club.sql`, `delete/DELETE_verified_club.sql`. Verifying twice is a no-op.

### 24. Authorized event read

**Status:** ✅ · **Endpoints:** ⬜ [`GET /auth/events/{eventId}`](../api/endpoints/event-management.md#-get-autheventseventid).
**Files:** the event role check (group 14), then `events/read/SELECT_event.sql` with `public_only` `FALSE` and `SELECT_event_images.sql`.

### 25. Event update and publish

**Status:** ✅ · **Endpoints:** ⬜ [`PATCH /auth/events/{eventId}`](../api/endpoints/event-management.md#-patch-autheventseventid).
**Files:** `events/update/*` (six files), `events/create/INSERT_event_club_link.sql`, `events/read/SELECT_event_status.sql`. Order in [Transactions](#transactions); allowed transitions in the [schema review](docs/schema-review.md#s3-allowed-transitions).

### 26. Event deletion

**Status:** ✅ · **Endpoints:** ⬜ [`DELETE /auth/events/{eventId}`](../api/endpoints/event-management.md#-delete-autheventseventid), ⬜ [`POST /auth/events/{eventId}/restore`](../api/endpoints/event-management.md#-post-autheventseventidrestore).
**Files:** `events/delete/UPDATE_event_soft_delete.sql`; `events/restore/UPDATE_event_restore.sql`, `SELECT_event_deletion.sql`, and `authorization/events/manages_owner_club/IS_student_owner_club_manager.sql`.
**Notes:** A soft delete ([decision 4](docs/schema-review.md#decisions)). Restore works for 30 days ([decision 9](../api/README.md#decisions)); group 31 removes the event after that, with the same 30 days hard-coded.

### 27. Club and event thumbnail metadata assignment

**Status:** ✅ · **Endpoints:** ⬜ [`POST /clubs/{clubId}/thumbnails/confirm`](../api/endpoints/images.md#-post-clubsclubidthumbnailsconfirm), ⬜ [`POST /clubs/{clubId}/events/{eventId}/thumbnails/confirm`](../api/endpoints/images.md#-post-clubsclubideventseventidthumbnailsconfirm). The signers (✅ `POST /clubs/{clubId}/thumbnails`, ✅ `.../events/{eventId}/thumbnails`, ⬜ `POST /auth/events/{eventId}/thumbnails`) need only the role checks and `EXISTS_club` or `SELECT_event_status`.
**Files:** `clubs/thumbnails/confirm/*`, `events/thumbnails/confirm/UPDATE_event_thumbnail.sql`, `events/update/SELECT_event_for_update.sql`, `images/create/INSERT_image.sql`, `images/release/UPDATE_image_soft_delete_if_unused.sql`.

### 28. Image deletion

**Status:** ✅ · **Endpoints:** ⬜ [`DELETE /images/{imageId}`](../api/endpoints/images.md#-delete-imagesimageid).
**Files:** `images/get/SELECT_image.sql`, `images/delete/*` (four files). Removing one use, as the confirms do, is `images/release/`.

### 29. Club creator ownership

**Status:** ✅ · **Endpoints:** 🔴 [`POST /clubs`](../api/endpoints/clubs.md#-post-clubs).
**Files:** `clubs/create/INSERT_club_owner.sql`, inside the club create transaction.
**Notes:** Written to the proposed `POST /clubs` contract, which makes the creator the owner. Whether creating a club should also require an admin is still an [API open question](../api/README.md#open-questions).

### 30. Club update

**Status:** ✅ · **Endpoints:** ⬜ [`PATCH /clubs/{clubId}`](../api/endpoints/clubs.md#-patch-clubsclubid).
**Files:** `clubs/update/UPDATE_club.sql`, `UPSERT_club_info.sql`, `clubs/tags/replace/*`, `clubs/get/SELECT_club.sql`.

### 31. Purge job

**Status:** ✅ · **Endpoints:** ⬜ 🔴 [`POST /admins/purge`](../api/endpoints/admins.md#-post-adminspurge), run by an admin on demand. There's no scheduled job ([decision 8](../api/README.md#decisions)).
**Files:** `events/purge/UPDATE_images_of_purgeable_events.sql`, `DELETE_purgeable_events.sql`, `images/purge/SELECT_purgeable_images.sql`, `DELETE_purged_image.sql`, and the admin check.
**Notes:** Removes only what was deleted more than 30 days ago: events, images, and announcements (`announcements/purge/DELETE_purgeable_announcements.sql`, in the same transaction as the events). The 30 days are written into each query, the same as in the restores, so nothing restorable is purged. The S3 deletes go through the API's storage adapter.

### 32. Announcements

**Status:** ✅ · **Endpoints:** ⬜ every route in [announcements.md](../api/endpoints/announcements.md): 🟢 `GET /clubs/{clubId}/announcements`, 🟢 `GET /announcements/{announcementId}`, 🔴 the drafts list, the manager read, create, `PATCH`, `DELETE`, and restore. All Proposed, Need Later (the club page's Announcements tab).
**Files:** `clubs/announcements/*` (two), `announcements/*` (eleven), `authorization/announcements/*` (two), and `announcements/purge/` for group 31.
**Notes:** Modelled on events: one row per announcement through the `announcement_details` view, the same soft delete and 30-day restore, owner and associate clubs, and the same role checks. The only transition is `draft → posted`, enforced in `UPDATE_announcement_status_posted.sql`, which also requires a body and sets `posted_at`. `INSERT_announcement.sql` sets `posted_at` when an announcement is created as posted. Posted lists are ordered by `posted_at`, newest first, using the `(status, posted_at)` index, and a CHECK keeps `posted_at` set exactly when `status` is `posted`. The GWC website reads the same public list.

## SQL source accounting

- **Embedded legacy SQL:** all 21 files under `utils/query_client/queries/` are in the module: 8 unchanged, 13 fixed (one of them moved). `infrastructure/legacy` itself is unchanged.
- **Inactive endpoint-stub SQL:** seven files under `stub/lambda/` are design notes only and aren't in the module. `GET_me_clubs_events.sql` is superseded by group 4, `eboard.sql` by group 17, `GET_events.sql`, `eventsId.sql`, and `description.sql` by group 11, and the two `admins/studentId` files by group 22.
- **Schema and seed history:** the nine older files are in [`migrations/history/`](migrations/history/), unchanged. The three `stub/environment/` SQL files are local experiments and aren't used.

## Module status

| # | Item | Status | Detail |
| ---: | --- | --- | --- |
| 1 | Baseline schema, down file, and seed | ✅ | Applied, dropped, and rebuilt on local MySQL 8.0.46. |
| 2 | Queries for every endpoint needed Now or Later | ✅ | 30 of 32 groups; see the [summary](#query-group-summary). |
| 3 | Injected MySQL configuration instead of AWS-coupled connection creation | 🟨 | The independent client exists. The AWS Secrets Manager/RDS adapter and the caller cutover don't. |
| 4 | Automated query tests against MySQL | ⬜ | The container run was manual. A test that starts MySQL and runs the same checks would keep them from regressing. |
| 5 | Go models matching the baseline | ⬜ | See [Shared application query access](#shared-application-query-access). |
| 6 | A migration runner | ⬜ | [Future work](docs/schema-review.md#future-work-a-migration-runner), for when a live database exists. |
| 7 | The purge and restore endpoints | ⬜ | Only their SQL exists; the purge's S3 deletes need the storage adapter. |
| 8 | API callers cut over | ⬜ | [`infrastructure/legacy/`](../infrastructure/legacy/) stays the working reference until parity and rollback plans are reviewed. |

## Open questions

Product and data decisions the code doesn't settle:

- Whether request-time student sync updates a changed email on an existing row.
- Whether an event's author or an admin gets management rights beyond the linked clubs' e-board and owners. (An admin can already restore an event and take down an image.)
- What "their clubs" means in `GET /admins`.
- Whether unused `student_info` and `event_tags` are requirements or unserved schema. Nothing proposes deleting them.

## Files

- `database/README.md`: this reference.
- `database/docs/schema-review.md`: the schema decisions, the findings behind the baseline, and future migration work.
- `database/assets/database-schema.png`: the legacy November schema diagram.
- `database/go.mod` and `go.sum`: the independent module (Go 1.23.0 directive, Go 1.24.3 toolchain).
- `database/migrations/schema/`: the baseline up and down files.
- `database/migrations/seed/`: the development seed.
- `database/migrations/history/`: the older schema and seed files, never applied.
- `database/models/`: the five legacy model files, unchanged.
- `database/queries/`: the 83 query files, `queries.go` (the embed list and `Load`), and the tests.
- `database/client/`: provider-independent pool, context-aware query methods, and transaction helpers; no callers yet.

Run the module's unit tests from `database/` with `go test ./...`. They don't connect to MySQL or AWS.
