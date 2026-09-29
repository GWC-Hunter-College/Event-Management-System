# Database module

The provider-independent MySQL database module: schema history, database-facing models, the application's SQL, and a Go client. It holds five model files and 21 active SQL queries copied unchanged from legacy, eight schema/database history files, one seed file, and an independent Go client. Nothing calls it: every handler and caller still runs from [`infrastructure/legacy/`](../infrastructure/legacy/), and the module has no migration runner.

This file is also the map of every query the API needs, as [29 query groups](#query-groups), each linked to its SQL, its callers, and the endpoints it serves.

## Boundary

**MySQL is the application database technology. Amazon RDS is only the current hosting environment.** The schema and query behavior target any MySQL host:

- Amazon RDS for MySQL;
- MySQL on Amazon EC2;
- another compatible MySQL host; and
- local MySQL.

Provider independence here means independence from AWS hosting, credentials, networking, and deployment APIs. It doesn't mean converting the MySQL schema or SQL dialect to another database engine.

The module owns MySQL schema history, migrations, query behavior, transactions, and database-facing models. AWS Secrets Manager lookup, RDS endpoints, VPC attachment, Lambda configuration, and IAM permissions aren't in it; they sit behind adapters on the API/infrastructure side.

## Sources

The legacy code is authoritative for status. The query groups were checked against:

- the initializer's selected DDL and seed migrations;
- all 21 embedded application SQL files under [`utils/query_client/queries/`](../infrastructure/legacy/utils/query_client/queries/);
- every current handler query call;
- query-client connection, loading, and transaction code;
- inactive endpoint stub SQL; and
- schema-implied operations the PDF's endpoint list needs.

The historical `GWC Website Documentation.pdf` supplied entity and endpoint intent only. Its older conceptual schema, July-era SQL assumptions, and project-management checkmarks do not override the current November DDL or current callers.

## Status legend

- ✅ **SQL exists** — the SQL and its callers are identifiable and portable, even when the caller has a separately documented API/auth issue.
- 🟨 **Partial** — query behavior is broken, missing one required part, spread across stale/inactive sources, or only reusable primitives exist.
- ⬜ **No SQL** — the schema or a planned endpoint needs the behavior, but no implementation exists.
- ❓ **Needs a decision** — product/data semantics have to be decided before the query group can be defined.

## Query group summary

There are **29 query groups**:

| Status | Count |
| --- | ---: |
| ✅ SQL exists | 14 |
| 🟨 Partial or spread across sources | 7 |
| ⬜ No SQL | 7 |
| ❓ Needs a decision | 1 |
| **Total** | **29** |

The current query surface accounts for 16 responsibilities: 14 ✅ groups and 2 🟨 groups. The 21 embedded SQL files form 15 of those groups; the sixteenth is the active event-image read whose required SQL file is absent. The other 13 groups cover needs from the PDF or inactive stubs, schema gaps, and one unresolved ownership decision.

Legacy has no query-specific unit or integration tests. This module has offline query-loading, transaction, and model-scan unit tests. A ✅ means the SQL is identifiable and portable, not that its runtime behavior has been proven against MySQL.

The SQL for groups 1–15 is copied to its module path, including the known broken event-creation SQL. A group's status describes its behavior and callers, not just whether a file exists. Missing queries, inactive stubs, SQL repairs, and application transaction changes aren't in the module.

## Directory layout

The module's layout. Only the paths holding copied SQL exist, plus `models/`, `client/`, and the independent `go.mod`/`go.sum`. Every other query path is created with the first behavior it holds, so there are no empty directories.

```text
database/
├── go.mod
├── go.sum
├── models/
├── migrations/
│   ├── schema/
│   └── seed/
├── queries/
│   ├── students/
│   │   └── ensure/
│   ├── me/
│   │   ├── get/
│   │   ├── clubs/
│   │   │   ├── list/
│   │   │   └── eboard/
│   │   │       └── list/
│   │   └── events/
│   │       └── list/
│   ├── clubs/
│   │   ├── list/
│   │   ├── get/
│   │   ├── create/
│   │   ├── members/
│   │   │   ├── list/
│   │   │   ├── create/
│   │   │   ├── leave/
│   │   │   └── update_role/
│   │   ├── events/
│   │   │   └── list/
│   │   ├── verification/
│   │   │   ├── create/
│   │   │   └── delete/
│   │   └── thumbnails/
│   │       └── confirm/
│   ├── events/
│   │   ├── read/
│   │   ├── create/
│   │   ├── update/
│   │   ├── delete/
│   │   ├── images/
│   │   │   ├── list/
│   │   │   └── confirm/
│   │   └── thumbnails/
│   │       └── confirm/
│   ├── authorization/
│   │   ├── clubs/
│   │   │   └── can_manage/
│   │   └── events/
│   │       └── can_manage/
│   ├── admins/
│   │   ├── list/
│   │   ├── get/
│   │   ├── create/
│   │   ├── delete/
│   │   └── is_admin/
│   └── images/
│       └── delete/
└── client/
```

Route-shaped query paths are for finding behavior, not a reason to duplicate SQL: the public and authorized event routes share the event-read queries and compose them with visibility and authorization policies. See the [API module layout](../api/LAYOUT.md#directory-layout).

## Current authoritative schema

The initializer explicitly selects [`11_04_2025_create_core_tables_up.sql`](../infrastructure/legacy/lambda/internal/database/init/migrations/11_04_2025_create_core_tables_up.sql) as the current core DDL. If this image or older documentation differs from that file, the SQL file wins.

![Current legacy database schema](assets/database-schema.png)

See the [detailed legacy database overview](../infrastructure/legacy/docs/database/README.md) and [table-by-table schema reference](../infrastructure/legacy/docs/database/schema.md).

### Current tables

| Table | Current purpose | Query coverage |
| --- | --- | --- |
| `students` | Cognito `sub` identity and optional email | Ensure/upsert and current-student reads exist. |
| `student_info` | Optional username/name profile | No current application query. |
| `clubs` | Club identity, name, and logo image FK | Read/create queries exist; logo assignment does not. |
| `club_info` | Website URL and description | Read/create queries exist. |
| `club_members` | Student/club membership with e-board/owner flags | Self-join/leave/read/auth queries exist; lists and role updates do not. |
| `verified_clubs` | Public-discovery approval relation | Read filter exists; writes do not. |
| `admins` | Global administrator relation | Only stale hard-coded stub SQL exists. |
| `events` | Author, thumbnail, details, status, dates, and lifecycle timestamps | Reads exist; create SQL is broken; update/delete do not exist. |
| `event_descriptions` | One description per event | Read exists; create SQL is broken. |
| `event_tags` | Event tags | No current application query. |
| `images` | Storage metadata | Event-image insert exists; delete and thumbnail assignment do not. |
| `events_to_clubs` | Many-to-many event/club relation and owner-club flag | Read, create-link, and event-authorization queries exist. |
| `event_images` | Event/gallery-image relation | Insert exists; active read query is missing. |

### Initialization sources

The current initializer at [`lambda/internal/database/init/main.go`](../infrastructure/legacy/lambda/internal/database/init/main.go):

1. runs [`07_11_2025_create_databases_up.sql`](../infrastructure/legacy/lambda/internal/database/init/migrations/07_11_2025_create_databases_up.sql) to create `STAGING` and `PRODUCTION`;
2. applies the November core DDL to both databases; and
3. applies [`09_14_2025_seed_tables.sql`](../infrastructure/legacy/lambda/internal/database/init/migrations/09_14_2025_seed_tables.sql) to `STAGING`.

The runner splits files on semicolons and executes statements without a migration-history table or encompassing transaction. Foreign-key ALTER statements and seed inserts can leave a partially initialized database, and the seed is not safely repeatable. There is no down migration matching the current November DDL. These are migration-runner issues, not schema defects. The module's copies of the migration files are history only; the module has never executed one.

## Shared application query access

Current access is implemented by [`query_client.go`](../infrastructure/legacy/utils/query_client/query_client.go), [`helpers.go`](../infrastructure/legacy/utils/query_client/helpers.go), and [`types.go`](../infrastructure/legacy/utils/query_client/types.go):

- Go embeds every `queries/**/*.sql` file into each consuming binary.
- `sqlx` and the MySQL driver execute `Get`, `Select`, `Exec`, and transaction helpers.
- active handlers use `NewClientFromHost`, which loads credentials from AWS Secrets Manager and receives an RDS host/schema through environment variables.
- `loadSQLFromFile` allocates exactly 4096 bytes and ignores the byte count returned by `Read`, so every query risks trailing null bytes and queries over 4 KiB risk truncation.
- `Get`, `Select`, and `Exec` do not use context-aware database methods.
- `NewClient` accepts but does not apply `dbName`; active code uses `NewClientFromHost` instead.
- the unused `ChangeDatabase` helper concatenates `USE ` with a caller-supplied database name, an unvalidated identifier; the module doesn't include it.

The new [`database/client/`](client/) accepts caller-supplied `mysql.Config` and propagates request contexts through `Get`, `Select`, `Exec`, and transaction operations. `Open` creates a lazy pool without dialing; `Ping(ctx)` explicitly checks connectivity. Callers supply credentials, `Net`/`Addr`, `DBName`, and TLS options using the MySQL driver's configuration defaults. AWS secret retrieval and RDS endpoint resolution are outside this module. There's no AWS adapter, and no legacy caller uses the module.

[`queries.Load`](queries/queries.go) embeds only the 21 copied application queries and reads their complete bytes. `ExecMulti` commits a statement batch; `ExecInsertQuery` also prepends the initial insert ID to selected later statements. Both propagate commit/rollback failures and return results only after successful commit. No application workflow is recomposed. Raw-row `Query`/`QueryRow` APIs, the unsafe `ChangeDatabase`/`QueryMulti` helpers, and initialization/migration execution are not ported.

Current SQL intentionally uses MySQL features such as `AUTO_INCREMENT`, `ENUM`, `BOOL`, backticks, `ON DUPLICATE KEY UPDATE`, and MySQL nullable/unique semantics. The module keeps those semantics; changing them would be a database redesign.

The DDL also permits null in many non-primary-key columns while several current Go club/event model fields use non-pointer values. Moving a query to the module needs scan-parity fixtures beyond the known nullable-student-email case.

## Query groups

### 1. Student existence and upsert

**Status:** ✅ Legacy queries exist and can be ported.

**Purpose:** Ensures a Cognito `sub` has a `students` row, optionally updates email, and supports both identity-trigger and request-time synchronization.

**Used by:** Internal Cognito student sync and these handler paths: [`GET /me`](../infrastructure/legacy/lambda/api/me/get.go), [`GET /me/clubs`](../infrastructure/legacy/lambda/api/me/clubs/get.go), [`GET /me/events`](../infrastructure/legacy/lambda/api/me/events/get.go), [`POST /clubs`](../infrastructure/legacy/lambda/api/clubs/post/post.go), [`POST /clubs/{clubId}/events`](../infrastructure/legacy/lambda/api/clubs/clubId/events/post/post.go), [`POST /clubs/{clubId}/members/me`](../infrastructure/legacy/lambda/api/clubs/clubId/members/me/post/post.go), and the currently miswired [`DELETE /clubs/{clubId}/members/me`](../infrastructure/legacy/lambda/api/clubs/clubId/members/me/delete/delete.go). Their route context is documented under [Internal functions](../api/endpoints/internal.md), [Me](../api/endpoints/me.md), [Clubs](../api/endpoints/clubs.md), [Memberships](../api/endpoints/memberships.md), and [Club events](../api/endpoints/club-events.md).

**Tables:** `students`.

**Legacy queries:** [`EXISTS_student_by_sub.sql`](../infrastructure/legacy/utils/query_client/queries/students/EXISTS_student_by_sub.sql), [`UPSERT_student.sql`](../infrastructure/legacy/utils/query_client/queries/students/UPSERT_student.sql), and [`UPSERT_student_sub_only.sql`](../infrastructure/legacy/utils/query_client/queries/students/UPSERT_student_sub_only.sql).

**Legacy callers:** [`utils/auth/ensure_student.go`](../infrastructure/legacy/utils/auth/ensure_student.go) and [`lambda/internal/auth/postConfirm/upsert.go`](../infrastructure/legacy/lambda/internal/auth/postConfirm/upsert.go).

**Module path:** `database/queries/students/ensure/` (SQL copied).

**Portable:** Yes. Preserve MySQL `ON DUPLICATE KEY UPDATE`; inject connection credentials rather than loading AWS Secrets Manager in the query module.

**Notes:** The existence check and upsert are separate request-time operations, but the upsert remains safe under a race. Request-time synchronization doesn't update a changed email when the row already exists; whether it should is an [open question](#open-questions).

### 2. Current student

**Status:** ✅ Legacy query exists and can be ported.

**Purpose:** Reads the student matching the verified caller identity.

**Used by:** [`GET /me`](../api/endpoints/me.md#-get-me).

**Tables:** `students`.

**Legacy query:** [`SELECT_student_by_sub.sql`](../infrastructure/legacy/utils/query_client/queries/students/SELECT_student_by_sub.sql).

**Legacy caller:** [`lambda/api/me/get.go`](../infrastructure/legacy/lambda/api/me/get.go).

**Module path:** `database/queries/me/get/` (SQL copied).

**Portable:** Yes.

**Notes:** `students.email` is nullable, but the current Go model uses a non-nullable string. The module's copy preserves that model exactly, and a parity test documents the NULL scan failure. An explicit null representation is needed before API cutover.

### 3. Current student's clubs and roles

**Status:** ✅ Legacy query exists and can be ported.

**Purpose:** Lists the caller's club memberships, logo object key, and a computed role with owner precedence over e-board and member.

**Used by:** [`GET /me/clubs`](../api/endpoints/me.md#-get-meclubs).

**Tables:** `club_members`, `clubs`, `images`.

**Legacy query:** [`SELECT_student_clubs.sql`](../infrastructure/legacy/utils/query_client/queries/students/SELECT_student_clubs.sql).

**Legacy caller:** [`lambda/api/me/clubs/get.go`](../infrastructure/legacy/lambda/api/me/clubs/get.go).

**Module path:** `database/queries/me/clubs/list/` (SQL copied).

**Portable:** Yes.

**Notes:** The selected `images.object_key` is exposed by the current API as `thumbnailUrl`. Storage URL generation stays outside the query; whether to keep that naming quirk is an API-layer decision.

### 4. Current student's events

**Status:** ✅ Legacy query exists and can be ported.

**Purpose:** Reads posted events associated with any club joined by the caller, including description and linked-club rows.

**Used by:** Active [`GET /me/events`](../api/endpoints/me.md#-get-meevents), which replaced historical `GET /me/clubs/events`.

**Tables:** `events`, `events_to_clubs`, `clubs`, `images`, `event_descriptions`, `club_members`.

**Legacy query:** [`SELECT_student_events.sql`](../infrastructure/legacy/utils/query_client/queries/students/SELECT_student_events.sql).

**Legacy caller:** [`lambda/api/me/events/get.go`](../infrastructure/legacy/lambda/api/me/events/get.go).

**Module path:** `database/queries/me/events/list/` (SQL copied).

**Portable:** Yes.

**Notes:** SQL applies strict date bounds and paginates joined association rows, not distinct events. It returns only linked clubs the student joined, which can omit the owner club from an otherwise qualifying event. Keeping this behavior needs explicit contract tests.

### 5. Club list and verified filter

**Status:** ✅ Legacy query exists and can be ported.

**Purpose:** Lists club summaries and optionally restricts them to rows present in `verified_clubs`.

**Used by:** [`GET /clubs`](../api/endpoints/clubs.md#-get-clubs), including `?verified=true`.

**Tables:** `clubs`, `images`, `verified_clubs`.

**Legacy query:** [`SELECT_clubs.sql`](../infrastructure/legacy/utils/query_client/queries/clubs/SELECT_clubs.sql).

**Legacy caller:** [`lambda/api/clubs/get.go`](../infrastructure/legacy/lambda/api/clubs/get.go).

**Module path:** `database/queries/clubs/list/` (SQL copied).

**Portable:** Yes.

**Notes:** The boolean argument implements “all versus verified-only” in one query. The route accepts only exact `true`/`TRUE`; HTTP string parsing stays in the handler, not the query code.

### 6. Club detail

**Status:** ✅ Legacy query exists and can be ported.

**Purpose:** Reads one club with optional image object key and club information.

**Used by:** [`GET /clubs/{clubId}`](../api/endpoints/clubs.md#-get-clubsclubid).

**Tables:** `clubs`, `images`, `club_info`.

**Legacy query:** [`SELECT_club.sql`](../infrastructure/legacy/utils/query_client/queries/clubs/SELECT_club.sql).

**Legacy caller:** [`lambda/api/clubs/clubId/get.go`](../infrastructure/legacy/lambda/api/clubs/clubId/get.go).

**Module path:** `database/queries/clubs/get/` (SQL copied).

**Portable:** Yes.

**Notes:** Keep HTTP path validation and not-found response mapping outside the query. Optional `club_info` and logo rows must remain nullable in the data model.

### 7. Club creation

**Status:** ✅ Legacy queries exist and can be ported.

**Purpose:** Inserts the main club row, obtains its auto-increment ID, and inserts the matching `club_info` row atomically.

**Used by:** [`POST /clubs`](../api/endpoints/clubs.md#-post-clubs).

**Tables:** `clubs`, `club_info`.

**Legacy queries:** [`INSERT_club.sql`](../infrastructure/legacy/utils/query_client/queries/clubs/INSERT_club.sql) and [`INSERT_club_info.sql`](../infrastructure/legacy/utils/query_client/queries/clubs/INSERT_club_info.sql).

**Legacy caller/transaction:** [`lambda/api/clubs/post/post.go`](../infrastructure/legacy/lambda/api/clubs/post/post.go) through [`QueryClient.ExecInsertQuery`](../infrastructure/legacy/utils/query_client/query_client.go).

**Module path:** `database/queries/clubs/create/` (SQL copied).

**Portable:** Yes; retain a MySQL transaction and last-insert-ID behavior behind a database interface.

**Notes:** This transaction does **not** create an owner membership or verification row. That unresolved behavior is tracked as [group 29](#29-club-creator-ownership).

### 8. Join caller to club

**Status:** ✅ Legacy query exists and can be ported.

**Purpose:** Inserts a regular membership only when both student and club exist and the membership does not already exist.

**Used by:** Active self-service [`POST /clubs/{clubId}/members/me`](../api/endpoints/memberships.md#-post-clubsclubidmembersme); it partially satisfies historical `POST /clubs/{clubId}/members`.

**Tables:** `students`, `clubs`, `club_members`.

**Legacy query:** [`INSERT_club_member.sql`](../infrastructure/legacy/utils/query_client/queries/clubs/INSERT_club_member.sql).

**Legacy caller:** [`lambda/api/clubs/clubId/members/me/post/post.go`](../infrastructure/legacy/lambda/api/clubs/clubId/members/me/post/post.go).

**Module path:** Shared `database/queries/clubs/members/create/`, with caller-scoping enforced by the API service (SQL copied).

**Portable:** Yes.

**Notes:** The query accepts a student ID parameter and can be reused for an authorized “add specified member” operation, but the current handler supplies only the verified caller and always assigns both role flags false. It does not require a verified club.

### 9. Leave caller's club

**Status:** ✅ Legacy query exists and can be ported.

**Purpose:** Deletes a membership only when it belongs to the caller/club pair and is not marked owner.

**Used by:** [`DELETE /clubs/{clubId}/members/me`](../api/endpoints/memberships.md#-delete-clubsclubidmembersme).

**Tables:** `club_members`.

**Legacy query:** [`DELETE_club_member.sql`](../infrastructure/legacy/utils/query_client/queries/clubs/DELETE_club_member.sql).

**Legacy caller:** [`lambda/api/clubs/clubId/members/me/delete/delete.go`](../infrastructure/legacy/lambda/api/clubs/clubId/members/me/delete/delete.go).

**Module path:** `database/queries/clubs/members/leave/` (SQL copied).

**Portable:** Yes.

**Notes:** The SQL owner guard is valid behavior, but the active route never supplies the JWT context its handler requires. That is an API wiring defect, not a missing query. Also review nullable `member_is_owner`: `= FALSE` does not match null.

### 10. Club event list

**Status:** ✅ Legacy query exists and can be ported.

**Purpose:** Reads status/date-filtered events linked to one club, with descriptions and club-logo object keys.

**Used by:** Public [`GET /clubs/{clubId}/events`](../api/endpoints/club-events.md#-get-clubsclubidevents); its status parameter also serves the planned authorized draft list ([group 21](#21-club-draft-events)).

**Tables:** `events`, `events_to_clubs`, `clubs`, `images`, `event_descriptions`.

**Legacy query:** [`SELECT_club_events.sql`](../infrastructure/legacy/utils/query_client/queries/clubs/SELECT_club_events.sql).

**Legacy caller:** [`lambda/api/clubs/clubId/events/get.go`](../infrastructure/legacy/lambda/api/clubs/clubId/events/get.go).

**Module path:** `database/queries/clubs/events/list/` (SQL copied).

**Portable:** Yes.

**Notes:** The current handler supplies `posted`, `%`, strict date bounds, limit, and offset. SQL paginates joined rows, and filtering to the requested club prevents the response from reconstructing all event associations.

### 11. Public and composite event read

**Status:** ✅ Legacy query exists and can be ported.

**Purpose:** Reads posted event rows, descriptions, and owner/associate club relations for both list and single-event responses.

**Used by:** [`GET /events`](../api/endpoints/events.md#-get-events), [`GET /events/{eventId}`](../api/endpoints/events.md#-get-eventseventid), and the historical description/club subresources now combined into event detail.

**Tables:** `events`, `events_to_clubs`, `clubs`, `images`, `event_descriptions`.

**Legacy query:** [`SELECT_events.sql`](../infrastructure/legacy/utils/query_client/queries/events/SELECT_events.sql).

**Legacy callers:** [`lambda/api/events/get.go`](../infrastructure/legacy/lambda/api/events/get.go) and [`lambda/api/events/eventId/get.go`](../infrastructure/legacy/lambda/api/events/eventId/get.go).

**Module path:** Shared `database/queries/events/read/`, with list/get wrappers rather than duplicated SQL (SQL copied).

**Portable:** Yes.

**Notes:** The query uses `LIKE` for an integer event ID, strict date bounds, row-level pagination, and no `deleted_at` predicate. The detail handler caps joined rows at 100. The inactive stub event/description SQL is superseded by this composite query and isn't in the module.

### 12. Create event draft

**Status:** 🟨 Query group exists but is deterministically broken and split across transaction boundaries.

**Purpose:** Inserts a drafted event, its owner/associate club links, and description.

**Used by:** [`POST /clubs/{clubId}/events`](../api/endpoints/club-events.md#-post-clubsclubidevents).

**Tables:** `events`, `events_to_clubs`, `event_descriptions`; foreign keys also depend on `students` and `clubs`.

**Legacy queries:** [`INSERT_event.sql`](../infrastructure/legacy/utils/query_client/queries/events/INSERT_event.sql), [`INSERT_event_club_link.sql`](../infrastructure/legacy/utils/query_client/queries/events/INSERT_event_club_link.sql), and [`INSERT_event_description.sql`](../infrastructure/legacy/utils/query_client/queries/events/INSERT_event_description.sql).

**Legacy caller:** [`lambda/api/clubs/clubId/events/post/post.go`](../infrastructure/legacy/lambda/api/clubs/clubId/events/post/post.go).

**Module path:** `database/queries/events/create/` as one transaction (SQL copied).

**Portable:** Partial.

**Notes:** `INSERT_event.sql` names 11 columns but supplies 10 values and has no value expression for `rsvp_link` while the handler passes seven arguments. `INSERT_event_description.sql` has a trailing comma in its column list. The initial event insert uses `Exec` before link/description `ExecMulti`, so a later failure can orphan the draft. The module's copy preserves these SQL bytes unchanged, so it's broken too. It needs repair and tests before any caller uses it.

### 13. Club authorization

**Status:** ✅ Legacy query/helper exists and can be ported, but no active handler calls it.

**Purpose:** Returns whether a student is an e-board member or owner of a specified club.

**Used by:** Intended dependency for protected club/event writes and member administration documented under the [club role check](../api/endpoints/internal.md#-club-role-check).

**Tables:** `club_members`.

**Legacy query:** [`IS_student_authorized_club.sql`](../infrastructure/legacy/utils/query_client/queries/authorization/IS_student_authorized_club.sql).

**Legacy callers/helpers:** [`utils/auth/club_authorization.go`](../infrastructure/legacy/utils/auth/club_authorization.go) and a duplicate [`lambda/internal` package](../infrastructure/legacy/lambda/internal/auth/club_authorization/club_authorization.go); neither is called by active handlers.

**Module path:** `database/queries/authorization/clubs/can_manage/` (SQL copied).

**Portable:** Yes.

**Notes:** Port one implementation and expose it through a shared application authorization policy. Owner-only operations such as historical role promotion require a stricter query/policy than this e-board-or-owner check.

### 14. Event authorization

**Status:** ✅ Legacy query/helper exists and can be ported, but no active handler calls it.

**Purpose:** Returns whether a student is an e-board member or owner of any club linked to an event.

**Used by:** Intended dependency for historical [authorized event routes](../api/endpoints/event-management.md).

**Tables:** `events_to_clubs`, `club_members`.

**Legacy query:** [`IS_student_authorized_event.sql`](../infrastructure/legacy/utils/query_client/queries/authorization/IS_student_authorized_event.sql).

**Legacy callers/helpers:** [`utils/auth/event_authorization.go`](../infrastructure/legacy/utils/auth/event_authorization.go) and duplicate [`lambda/internal` package](../infrastructure/legacy/lambda/internal/auth/event_authorization/event_authorization.go); neither is called by active handlers.

**Module path:** `database/queries/authorization/events/can_manage/` (SQL copied).

**Portable:** Yes.

**Notes:** The “any associated club” rule matches the historical rationale for `/auth/events/{eventId}`. Whether owner-club authority, event authorship, or admins take precedence is an [open question](#open-questions).

### 15. Event-image metadata confirmation

**Status:** ✅ Legacy insert queries exist and can be ported.

**Purpose:** Inserts image metadata and links the new image to an event after a direct storage upload.

**Used by:** [`POST /clubs/{clubId}/events/{eventId}/images/confirm`](../api/endpoints/images.md#-post-clubsclubideventseventidimagesconfirm); the generic historical internal image write can reuse the metadata insert.

**Tables:** `images`, `event_images`; the event foreign key also depends on `events`.

**Legacy queries:** [`INSERT_image.sql`](../infrastructure/legacy/utils/query_client/queries/images/INSERT_image.sql) and [`INSERT_event_image.sql`](../infrastructure/legacy/utils/query_client/queries/images/INSERT_event_image.sql).

**Legacy caller:** [`lambda/api/clubs/events/images/confirm/post.go`](../infrastructure/legacy/lambda/api/clubs/events/images/confirm/post.go).

**Module path:** `database/queries/events/images/confirm/`, reusing a shared image-metadata insert primitive (SQL copied).

**Portable:** Yes.

**Notes:** The current handler performs two independent `Exec` calls, so a failed association leaves an orphan `images` row. It trusts client-provided identifiers/keys and does not check object existence. Both database writes belong in one transaction; storage verification belongs in the application/storage adapter. The handler also closes a package-level client after each request, which can break warm Lambda reuse.

### 16. Event-image list

**Status:** 🟨 Active handler exists, but its required query file is absent.

**Purpose:** Reads image metadata for an event so the API/storage adapter can produce download URLs.

**Used by:** Active but broken club-scoped gallery read and historical public/protected image reads under [Images](../api/endpoints/images.md) and [Event management](../api/endpoints/event-management.md).

**Tables:** Expected `event_images` joined to `images`.

**Legacy query:** **Missing:** `utils/query_client/queries/images/SELECT_event_images.sql`.

**Legacy caller:** [`lambda/api/clubs/events/images/get/get.go`](../infrastructure/legacy/lambda/api/clubs/events/images/get/get.go).

**Module path:** Shared `database/queries/events/images/list/` (not created).

**Portable:** Partial.

**Notes:** Define and test the selected fields against the handler's nested `sqlx` scan shape. The handler also closes its package-level database connection after every request, risking failure on warm Lambda reuse. Posted-versus-draft visibility and URL signing are API/policy/storage concerns, not reasons to duplicate the metadata query.

### 17. My e-board clubs

**Status:** 🟨 Inactive handler and stale stub SQL only; active role-list behavior can be reused.

**Purpose:** Lists clubs where the caller is an e-board member or owner.

**Used by:** Historical [`GET /me/clubs/eboard`](../api/endpoints/me.md#-get-meclubseboard).

**Tables:** `club_members`, `clubs`, and optionally `images`.

**Legacy query:** [`stub/lambda/me/clubs/eboard/eboard.sql`](../infrastructure/legacy/stub/lambda/me/clubs/eboard/eboard.sql).

**Legacy caller:** Hard-coded inactive [`stub/lambda/me/clubs/eboard/get.go`](../infrastructure/legacy/stub/lambda/me/clubs/eboard/get.go).

**Module path:** `database/queries/me/clubs/eboard/list/`, or reuse group 3 with a role filter (not created).

**Portable:** Partial.

**Notes:** The stub SQL selects student IDs rather than clubs and lacks parentheses around its `AND`/`OR` role condition. Prefer filtering/reusing the current club-and-role query over porting this SQL.

### 18. Club member and e-board listing

**Status:** ⬜ Query required but not found.

**Purpose:** Lists club memberships, optionally filtered to e-board/owner roles, for historical member administration routes.

**Used by:** Historical `GET /clubs/{clubId}/members` and `GET /clubs/{clubId}/eboard` under [Memberships](../api/endpoints/memberships.md).

**Tables:** Expected `club_members`, `students`, and possibly `student_info`.

**Legacy query/caller:** None; only placeholder `.txt` files exist under [`stub/lambda/clubs/clubId/members/`](../infrastructure/legacy/stub/lambda/clubs/clubId/members/).

**Module path:** Shared `database/queries/clubs/members/list/` with an explicit role filter (not created).

**Portable:** N/A until implemented.

**Notes:** Define visibility, profile fields, pagination, and ordering before implementation. Do not create separate duplicated SQL for members and e-board if one filtered query suffices.

### 19. Add a specified club member

**Status:** 🟨 The insert primitive exists, but no authorized endpoint adds a student other than the caller.

**Purpose:** Adds a chosen student to a club, distinct from self-service join.

**Used by:** Historical `POST /clubs/{clubId}/members`; the active route is caller-only `.../members/me`.

**Tables:** `students`, `clubs`, `club_members`.

**Legacy query:** Reusable [`INSERT_club_member.sql`](../infrastructure/legacy/utils/query_client/queries/clubs/INSERT_club_member.sql).

**Legacy caller:** Only the self-join handler [`lambda/api/clubs/clubId/members/me/post/post.go`](../infrastructure/legacy/lambda/api/clubs/clubId/members/me/post/post.go).

**Module path:** Shared `database/queries/clubs/members/create/` with API-level caller/target authorization (SQL copied, shared with group 8).

**Portable:** Partial; SQL is portable, endpoint/policy behavior is missing.

**Notes:** Define who can add another student, how the target is identified, and whether initial roles can be supplied. Do not let a route body bypass caller/owner policy.

### 20. Update member roles

**Status:** ⬜ Query required but not found.

**Purpose:** Promotes/demotes a club member's e-board/owner flags.

**Used by:** Historical `PUT /clubs/{clubId}/members/roles` under [Memberships](../api/endpoints/memberships.md#-put-clubsclubidmembersroles).

**Tables:** `club_members`.

**Legacy query/caller:** No embedded application `UPDATE club_members` query or active handler found. A local [`stub/environment/populate.sql`](../infrastructure/legacy/stub/environment/populate/populate.sql) fixture contains hard-coded UPDATE statements but is not endpoint behavior.

**Module path:** `database/queries/clubs/members/update_role/` (not created).

**Portable:** N/A until implemented.

**Notes:** The historical owner-only rule is stricter than the current club authorization query. Define invariants for one/multiple owners, self-promotion, self-demotion, and nullable role values.

### 21. Club draft events

**Status:** 🟨 Reusable status-parameterized SQL exists, but no draft route/handler composes it with authorization.

**Purpose:** Lists drafted events linked to a club for authorized managers.

**Used by:** Historical `GET /clubs/{clubId}/events/drafts` under [Club events](../api/endpoints/club-events.md#-get-clubsclubideventsdrafts).

**Tables:** Same as [group 10](#10-club-event-list): `events`, `events_to_clubs`, `clubs`, `images`, `event_descriptions`.

**Legacy query:** Reusable [`SELECT_club_events.sql`](../infrastructure/legacy/utils/query_client/queries/clubs/SELECT_club_events.sql); active caller hardcodes `posted`.

**Legacy caller:** [`lambda/api/clubs/clubId/events/get.go`](../infrastructure/legacy/lambda/api/clubs/clubId/events/get.go) for public posted reads only.

**Module path:** Reuse `database/queries/clubs/events/list/`; pass an application-approved status after club authorization (SQL copied, shared with group 10).

**Portable:** Partial.

**Notes:** Do not expose the query's status argument directly to an unauthenticated caller.

### 22. Admin CRUD

**Status:** 🟨 Schema plus hard-coded GET/DELETE stub SQL only.

**Purpose:** Lists, reads, creates, and removes global administrator records.

**Used by:** Historical [Admin routes](../api/endpoints/admins.md).

**Tables:** `admins`, with `students` needed for validated identities and richer responses.

**Legacy queries:** Unwired [`GET_admins_studentId.sql`](../infrastructure/legacy/stub/lambda/admins/studentId/GET_admins_studentId.sql) and [`DELETE_admins_studentId.sql`](../infrastructure/legacy/stub/lambda/admins/studentId/DELETE_admins_studentId.sql).

**Legacy caller:** None. No list/create SQL or active handler exists.

**Module path:** `database/queries/admins/{list,get,create,delete}/` plus a reusable admin check (not created).

**Portable:** Partial.

**Notes:** Both stubs hardcode integer ID `2`, while the current schema uses Cognito `CHAR(36)` student IDs. Treat the stubs as intent only. Admin bootstrap and last-admin rules belong in application policy/transactions.

### 23. Club verification writes

**Status:** ⬜ Queries required but not found.

**Purpose:** Adds/removes a club from public verification state.

**Used by:** Historical verification write routes under [Verification](../api/endpoints/verification.md).

**Tables:** `verified_clubs`, with `clubs` for target validation.

**Legacy query/caller:** No insert/delete query or handler. Group 5 reads verification state only.

**Module path:** `database/queries/clubs/verification/create/` and `delete/` (not created).

**Portable:** N/A until implemented.

**Notes:** Define duplicate/missing-row idempotency and the admin/club-administration policy outside the query.

### 24. Authorized event read

**Status:** 🟨 Read and authorization primitives exist separately; no protected/draft-capable composition exists.

**Purpose:** Reads a non-public event only when the caller can manage at least one associated club.

**Used by:** Historical `GET /auth/events/{eventId}` and related protected reads under [Event management](../api/endpoints/event-management.md#-get-autheventseventid).

**Tables:** The union of group 11 (`events`, descriptions, clubs/images/links) and group 14 (`events_to_clubs`, `club_members`).

**Legacy queries:** Reuse [`SELECT_events.sql`](../infrastructure/legacy/utils/query_client/queries/events/SELECT_events.sql) and [`IS_student_authorized_event.sql`](../infrastructure/legacy/utils/query_client/queries/authorization/IS_student_authorized_event.sql).

**Legacy caller:** No composed handler; public event detail hardcodes `posted` and role helper is unused.

**Module path:** Reuse `database/queries/events/read/` plus `database/queries/authorization/events/can_manage/` (both SQL copied; nothing composes them).

**Portable:** Partial.

**Notes:** Keep authorization and read primitives reusable, but execute them through one application policy that avoids time-of-check/time-of-use ambiguity where material. Define which event statuses authorized callers may read.

### 25. Event update and publish

**Status:** ⬜ Queries required but not found.

**Purpose:** Updates allowed event fields, description, associations, and status transitions such as drafted to posted.

**Used by:** Historical [`PATCH /auth/events/{eventId}`](../api/endpoints/event-management.md#-patch-autheventseventid).

**Tables:** At minimum `events`; potentially `event_descriptions`, `events_to_clubs`, and `event_tags` depending on the approved patch contract.

**Legacy query/caller:** No event `UPDATE` SQL or patch handler found.

**Module path:** `database/queries/events/update/` as a transaction assembled from reusable field-specific statements (not created).

**Portable:** N/A until implemented.

**Notes:** Define patch semantics, allowed fields, status transitions, timestamp updates, associate-club authority, and concurrency behavior before SQL is written. Event creation always produces `drafted`, so nothing can publish an event through the deployed API.

### 26. Event deletion

**Status:** ⬜ Query required but not found.

**Purpose:** Archives, soft-deletes, or physically deletes an event according to an explicit lifecycle policy.

**Used by:** Historical [`DELETE /auth/events/{eventId}`](../api/endpoints/event-management.md#-delete-autheventseventid).

**Tables:** `events` and, for physical cleanup, `event_descriptions`, `event_tags`, `events_to_clubs`, `event_images`, and possibly `images`.

**Legacy query/caller:** No archive/delete SQL or handler found.

**Module path:** `database/queries/events/delete/` (not created).

**Portable:** N/A until implemented.

**Notes:** The schema offers both `status='archived'` and nullable `deleted_at`; current reads filter status but not `deleted_at`. Foreign keys have no documented cascade policy. Choose soft-delete/read behavior and storage cleanup before implementing a transaction.

### 27. Club and event thumbnail metadata assignment

**Status:** ⬜ Queries required but not found.

**Purpose:** Confirms an uploaded thumbnail, creates image metadata, and assigns the image ID to `clubs.fk_logo_id` or `events.fk_thumbnail_id`.

**Used by:** Missing club/event thumbnail confirmation steps documented under [Images](../api/endpoints/images.md#how-image-uploads-work).

**Tables:** `images`, plus `clubs` or `events`.

**Legacy query/caller:** Presign handlers exist, and group 15 provides a reusable image insert, but no thumbnail confirmation handler or FK-update SQL exists.

**Module path:** `database/queries/clubs/thumbnails/confirm/` and `database/queries/events/thumbnails/confirm/`, sharing an image insert primitive (not created).

**Portable:** N/A until implemented as complete transactions.

**Notes:** Define replacement semantics and old-image cleanup. Both thumbnail foreign keys are unique, so duplicate/reassignment behavior and transaction ordering must be tested.

### 28. Image deletion

**Status:** ⬜ Query required but not found.

**Purpose:** Removes image links/metadata safely and coordinates deletion of the storage object.

**Used by:** Historical [`DELETE /images/{imageId}`](../api/endpoints/images.md#-delete-imagesimageid).

**Tables:** Depending on purpose: `event_images`, `events`, `clubs`, and `images`.

**Legacy query/caller:** No deletion SQL, handler, or storage cleanup flow found.

**Module path:** `database/queries/images/delete/`, with storage deletion behind the API's storage adapter (not created).

**Portable:** N/A until implemented.

**Notes:** Define whether database unlink/delete precedes object deletion, how retries recover partial failure, whether shared images are allowed, and how orphaned uploads are collected. S3 operations themselves do not belong in this database module.

### 29. Club creator ownership

**Status:** ❓ Manual product/data review required.

**Purpose:** Would determine whether creating a club atomically makes the authenticated creator its owner.

**Used by:** [`POST /clubs`](../api/endpoints/clubs.md#-post-clubs) and subsequent club-management authorization.

**Tables:** `clubs`, `club_info`, `club_members`, and `students`.

**Legacy query/caller:** Group 7 creates `clubs` and `club_info`; the handler ensures the student exists but never passes `sub` into the transaction or inserts an owner membership. Seed data demonstrates owner memberships, but not the intended creation rule.

**Module path:** If approved, extend the `database/queries/clubs/create/` transaction rather than adding a disconnected follow-up (group 7's SQL copied; no ownership insert).

**Portable:** Unknown until the rule is decided.

**Notes:** The historical description says “create a club under the signed in user,” which suggests ownership, but the code doesn't implement it. A mechanical port leaves it unchanged.

## SQL source accounting

### Embedded application SQL

All 21 files embedded by `utils/query_client` are accounted for in the inventory:

| Legacy directory | Files | Query groups |
| --- | ---: | --- |
| [`queries/students/`](../infrastructure/legacy/utils/query_client/queries/students/) | 6 | 1–4 |
| [`queries/clubs/`](../infrastructure/legacy/utils/query_client/queries/clubs/) | 7 | 5–10 |
| [`queries/events/`](../infrastructure/legacy/utils/query_client/queries/events/) | 4 | 11–12 |
| [`queries/authorization/`](../infrastructure/legacy/utils/query_client/queries/authorization/) | 2 | 13–14 |
| [`queries/images/`](../infrastructure/legacy/utils/query_client/queries/images/) | 2 | 15; group 16 identifies the missing read file |

### Inactive endpoint-stub SQL

Seven SQL files under `stub/lambda/` are historical/supporting evidence, not active queries:

- [`GET_me_clubs_events.sql`](../infrastructure/legacy/stub/lambda/me/clubs/events/GET_me_clubs_events.sql) is superseded by group 4.
- [`eboard.sql`](../infrastructure/legacy/stub/lambda/me/clubs/eboard/eboard.sql) is recorded in group 17.
- [`GET_events.sql`](../infrastructure/legacy/stub/lambda/events/GET_events.sql), [`eventsId.sql`](../infrastructure/legacy/stub/lambda/events/eventId/eventsId.sql), and [`description.sql`](../infrastructure/legacy/stub/lambda/events/eventId/description/description.sql) are superseded by group 11.
- [`GET_admins_studentId.sql`](../infrastructure/legacy/stub/lambda/admins/studentId/GET_admins_studentId.sql) and [`DELETE_admins_studentId.sql`](../infrastructure/legacy/stub/lambda/admins/studentId/DELETE_admins_studentId.sql) are recorded in group 22.

### Schema, seed, and local-stub SQL

Nine migration files remain preserved in legacy and are copied byte-for-byte here: eight in [`migrations/schema/`](migrations/schema/) and `09_14_2025_seed_tables.sql` in [`migrations/seed/`](migrations/seed/). Only the database-creation migration, November core DDL, and September seed are selected by the current initializer. These earlier files are historical, not authoritative:

- [`07_11_2025_create_core_tables_up.sql`](../infrastructure/legacy/lambda/internal/database/init/migrations/07_11_2025_create_core_tables_up.sql) and its [`down`](../infrastructure/legacy/lambda/internal/database/init/migrations/07_11_2025_create_core_tables_down.sql);
- the July member-form [`up`](../infrastructure/legacy/lambda/internal/database/init/migrations/07_11_2025_create_member_form_migration_table_up.sql) and [`down`](../infrastructure/legacy/lambda/internal/database/init/migrations/07_11_2025_create_member_form_migration_table_down.sql); and
- the September core [`up`](../infrastructure/legacy/lambda/internal/database/init/migrations/09_08_2025_create_core_tables_up.sql) and [`down`](../infrastructure/legacy/lambda/internal/database/init/migrations/09_08_2025_create_core_tables_down.sql).

The three [`stub/environment/`](../infrastructure/legacy/stub/environment/) SQL files are local experiments/fixtures and are not part of the current initializer or query client. `SELECT 1 + 1` is embedded only in the dormant database-test handler and is not an application query group.

## Current transaction boundaries

| Workflow | Current boundary | Issue |
| --- | --- | --- |
| Student ensure | Existence check followed by an upsert when missing | Separate calls, but duplicate-key upsert handles races; email-update behavior differs by path. |
| Club creation | `INSERT_club` + `INSERT_club_info` in `ExecInsertQuery` transaction | Portable; creator ownership waits on the group 29 decision. |
| Event creation | Event insert commits first; links + description use a later `ExecMulti` transaction | Broken SQL and possible orphan event; needs one transaction. |
| Event-image confirmation | Two independent `Exec` calls | Possible orphan metadata; needs one transaction. |
| QueryClient `ExecMulti` | One transaction across supplied statements | Reusable concept. The module's version propagates rollback/commit errors and is context-aware. |
| QueryClient `QueryMulti` | Begins a transaction and returns row handles after commit | Unused and unsafe to port without redesign. |
| Database initialization | Statements split and executed one-by-one | No migration history, rollback, or safe retry. The module has no provider-neutral runner to replace it yet. |

## Module status

What the module has and lacks, in dependency order. Statuses use the [status legend](#status-legend).

| # | Item | Status | Detail |
| ---: | --- | --- | --- |
| 1 | Versioned MySQL migration baseline from the November DDL | ⬜ | Must not re-run destructive or non-idempotent initialization against existing databases. |
| 2 | Injected MySQL configuration instead of AWS-coupled connection creation | 🟨 | The independent client exists. The AWS Secrets Manager/RDS adapter and the caller cutover don't. |
| 3 | Full-read SQL loader and context-aware query execution | 🟨 | In the module only; legacy is unchanged. |
| 4 | Query/transaction integration tests against compatible MySQL | ⬜ | Needed before any handler moves. |
| 5 | Groups 1–11 and 13–15 ported with parity fixtures for null handling, date bounds, pagination, role precedence, and response mapping | 🟨 | SQL copied; no callers or parity fixtures. |
| 6 | Groups 12 and 16 repaired and tested | ⬜ | Needed before their API equivalents are exposed. |
| 7 | Group 29 and the authorization/product questions in the API reference resolved | ❓ | See [Open questions](#open-questions) and the [API's open questions](../api/README.md#open-questions). |
| 8 | Missing groups 18, 20, 23, and 25–28 | ⬜ | Each is written with its endpoint and policy, not ahead of it; no unused SQL or empty directories. |
| 9 | The same query suite run against local MySQL and AWS-hosted compatible MySQL | ⬜ | This is what proves hosting independence. |
| 10 | API callers cut over | ⬜ | Incrementally. [`infrastructure/legacy/`](../infrastructure/legacy/) stays the working reference until parity and rollback plans are reviewed. |

## Open questions

Product and data decisions the code doesn't settle:

- Whether request-time student sync updates a changed email on an existing row.
- Whether owner-club authority, event authorship, or admins take precedence in event authorization.
- Club creator ownership and initial verification behavior.
- Admin bootstrap, last-admin protection, and admin-to-club response semantics.
- Owner-only versus e-board-or-owner role administration.
- Draft visibility, event publication transitions, archival versus `deleted_at`, and physical cleanup.
- Pagination by joined row versus distinct event, deterministic ordering, and association completeness.
- Nullable role/email fields and current Go model compatibility.
- Image/thumbnail metadata ownership, transaction boundaries, replacement, deletion, and storage failure recovery.
- Whether unused `student_info` and `event_tags` are requirements or unserved schema. Nothing proposes deleting them.
- How to baseline existing `STAGING`/`PRODUCTION` databases into a real migration history without recreating resources or data.

## Files

- `database/README.md`: this reference: schema, query groups, and module status.
- `database/assets/database-schema.png`: reused current-schema visual.
- `database/go.mod` and `go.sum`: independent module using the existing Go 1.23.0 directive and Go 1.24.3 toolchain; existing modules are unchanged and no workspace is introduced.
- `database/models/`: all five source models copied exactly, including fields, tags, types, and existing nullable-field limitations.
- `database/queries/`: all 21 active SQL files copied exactly into their module paths; both image inserts are in `events/images/confirm/`. No inactive stub or missing query is included.
- `database/migrations/`: all nine SQL history/seed files copied exactly, without a runner or execution.
- `database/client/`: provider-independent pool, context-aware query methods, and transaction helpers; no active callers.
- Offline tests: all 21 query byte lengths/hashes and absence of null padding, reads beyond 4096 bytes, mock transaction success/failure and insert IDs, and model scan compatibility (including preserved NULL-to-string failures). These do not establish SQL correctness or MySQL integration parity.
- [`infrastructure/legacy/utils/query_client/`](../infrastructure/legacy/utils/query_client/): current application SQL/client source.
- [`infrastructure/legacy/lambda/internal/database/init/`](../infrastructure/legacy/lambda/internal/database/init/): current initialization/migration source.
- [`api/README.md`](../api/README.md): endpoint reference index; [`api/LAYOUT.md`](../api/LAYOUT.md): where each endpoint's code goes in the `api/` module.

Run the independent module's offline unit tests from `database/` with `go test ./...`. Tests use an in-memory SQL mock and a stub dialer; they do not connect to MySQL or AWS. The module has no migration execution, live integration tests, API cutover, or SQL repairs, and the [open questions](#open-questions) are unresolved; see [Module status](#module-status).

The module is source only and unused. It changes no active schema, SQL or API behavior, credentials, database hosting, or deployed resource.
