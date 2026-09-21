# API migration map

This directory is the future boundary for the provider-independent backend API application. No handlers have been moved here yet. The current integrated implementation remains under [`infrastructure/legacy/`](../infrastructure/legacy/), and that code—not the historical planning document—is the source of truth for every status below.

The API remains the layer through which separately maintained frontend applications communicate with backend data:

```text
Frontend -> API -> Database
```

AWS Lambda, API Gateway, Cognito, S3, and RDS are current adapters or hosting choices. They must not define the future application's core endpoint, authorization, storage, or database behavior.

## Legends

### Access

- 🟢 **Public** — the active API Gateway route has no Cognito authorizer.
- 🔴 **Authenticated / Protected** — the active route has the Cognito JWT authorizer, or an inactive historical route was explicitly designed to require authentication.
- 🔵 **Internal** — a Lambda or application function that is not exposed through API Gateway.

For inactive historical routes, the detailed **Authentication** field distinguishes historical intent from actual active configuration. The marker never implies that an absent route is deployed.

### Implementation and migration status

- ✅ **Implemented** — a corresponding active legacy route, handler, and required supporting behavior are identifiable for migration.
- 🟨 **Partial** — only a stub or supporting primitive exists, the route is inactive, or a deterministic wiring/query defect prevents the intended behavior.
- ⬜ **Not found** — no corresponding route/handler/query implementation was found.
- ❓ **Ambiguous** — evidence conflicts or a product decision is required before status can be resolved.

### Portability

“Portable” describes reusable application behavior, not whether the current Lambda can be copied unchanged. AWS coupling normally means an adapter must be extracted; it does not make the behavior unusable.

## Migration summary

The counts are kept in two non-additive views because the historical list overlaps the active API and several routes changed shape.

### Active legacy API

| Status | Count |
| --- | ---: |
| ✅ Implemented | 15 |
| 🟨 Partial | 3 |
| ⬜ Not found | 0 |
| ❓ Ambiguous | 0 |
| **Active HTTP routes** | **18** |

The active access configuration is 12 🟢 public routes and 6 🔴 Cognito-JWT routes. There is no API-wide default authorizer.

### Historical “Endpoints Revamp” reconciliation

| Status | Count |
| --- | ---: |
| ✅ Implemented at the historical route/intent | 8 |
| 🟨 Partial, moved, stubbed, or incomplete | 16 |
| ⬜ No executable implementation found | 11 |
| ❓ Ambiguous | 0 |
| **Historical items** | **35** |

Two blue/internal functions are documented: the current Cognito student synchronization (✅) and the historical generic image metadata write (🟨). They are not active HTTP routes.

The three active partial routes are:

1. `POST /clubs/{clubId}/events`: wired, but its event and description INSERT statements are invalid.
2. `DELETE /clubs/{clubId}/members/me`: its handler requires JWT claims, but its active API Gateway route omits the authorizer.
3. `GET /clubs/{clubId}/events/{eventId}/images`: its handler references a SQL file that does not exist.

The active development and production APIs register the same 18 method/path contracts. Three event-image `POST` paths additionally register `OPTIONS` against their Lambda integration; those preflight registrations are not counted as application endpoints. See the [detailed legacy route matrix](../infrastructure/legacy/docs/api/README.md#active-route-matrix).

## Evidence and interpretation rules

The inventory was checked against:

- active route registration in [`gateway/routes/`](../infrastructure/legacy/gateway/routes/);
- both API stacks in [`productionApi.go`](../infrastructure/legacy/internal/stack/productionApi.go) and [`developmentApi.go`](../infrastructure/legacy/internal/stack/developmentApi.go);
- Lambda handlers under [`lambda/api/`](../infrastructure/legacy/lambda/api/);
- embedded SQL under [`utils/query_client/queries/`](../infrastructure/legacy/utils/query_client/queries/);
- authentication and authorization helpers under [`utils/auth/`](../infrastructure/legacy/utils/auth/);
- image/S3 integrations and Cognito trigger wiring; and
- the inactive stub stack and stub handlers, which count only as partial evidence.

The locally available pre-refactor stash copy of `GWC Website Documentation.pdf` supplied the route grouping, intent, and green/red/blue visual model; the PDF is not tracked in the current branch. Its checkmarks, assignments, dates, and project-management notes were not used as implementation evidence. Relevant historical sections are the original endpoint list (pages 14–18), “Endpoints Revamp” (pages 20–25), and image-upload flow (pages 33–35).

## Planned API directory structure

The following tree is a **representative planned skeleton only**. These directories do not exist yet and should be created incrementally as behavior is migrated and tested. Some leaf operations are collapsed for readability; the per-endpoint target paths below are the complete migration mapping.

```text
api/
├── internal/
│   ├── students/
│   │   └── sync/
│   └── images/
│       └── confirm/
├── me/
│   ├── get/
│   ├── clubs/
│   │   ├── get/
│   │   └── eboard/
│   └── events/
│       └── get/
├── clubs/
│   ├── list/
│   ├── create/
│   └── {clubId}/
│       ├── get/
│       ├── members/
│       │   ├── list/
│       │   ├── join/
│       │   ├── leave/
│       │   └── roles/
│       ├── events/
│       │   ├── list/
│       │   ├── drafts/
│       │   └── create/
│       ├── verification/
│       │   ├── create/
│       │   └── delete/
│       └── thumbnails/
│           ├── presign/
│           └── confirm/
├── events/
│   ├── list/
│   └── {eventId}/
│       ├── get/
│       └── images/
│           └── get/
├── auth/
│   └── events/
│       └── {eventId}/
│           ├── get/
│           ├── update/
│           ├── delete/
│           ├── images/
│           └── thumbnails/
├── admins/
│   ├── list/
│   ├── create/
│   └── {studentId}/
│       ├── get/
│       └── delete/
├── images/
│   └── {imageId}/
│       └── delete/
├── operations/
│   └── health/
├── shared/
│   ├── authentication/
│   ├── authorization/
│   ├── validation/
│   └── responses/
└── adapters/
    ├── http/
    ├── identity/
    ├── storage/
    └── database/
```

Route-shaped folders are navigation aids, not a reason to duplicate shared authorization, query, response, or storage logic. The corresponding **planned** query tree is documented in the [database migration map](../database/README.md#planned-database-directory-structure).

## Historical-to-current route changes

| Historical design | Current source-of-truth finding |
| --- | --- |
| `GET /me/clubs/events` | The active equivalent is `GET /me/events`; the old path survives only in the inactive stub stack. |
| `POST /clubs/{clubId}/members` | Self-join is active as `POST /clubs/{clubId}/members/me`; no endpoint can add a different student. |
| `POST /clubs/thumbnails` | The active path requires a club ID: `POST /clubs/{clubId}/thumbnails`. |
| `GET /clubs/verified=true` | Verification is a query parameter on `GET /clubs?verified=true`. |
| Separate event description and club subresources | The active `GET /events/{eventId}` response already joins description and associated-club IDs. The historical subresource routes are inactive stubs. |
| `/auth/events/{eventId}` revamp | No `/auth/events` route is active. Some image operations remain on older club-scoped paths and are currently public. |
| Generic `POST /images` | Event-gallery metadata is instead written by the active `POST .../images/confirm` route. No generic image route is active. |
| Club/event thumbnail confirmations | Presign handlers exist, but neither confirmation/metadata-assignment route exists. |
| Original club-scoped event detail/update/delete | The PDF later proposed `/auth/events`; neither protected form is active. Only club-scoped media paths remain. |

### Historical endpoint checklist

This compact view preserves every item in the historical “Endpoints Revamp.” The detailed sections below de-duplicate moved/reused behavior and provide the actual paths needed for migration.

| Historical route/function | Status | Current finding |
| --- | --- | --- |
| Register Lambda | ✅ | Cognito student upsert handler/query are active as an internal trigger. |
| `GET /me/clubs` | ✅ | Exact active protected route. |
| `GET /me/clubs/events` | 🟨 | Complete behavior moved to `GET /me/events`; exact path is an inactive stub. |
| `GET /me/clubs/eboard` | 🟨 | Inactive hard-coded stub and stale SQL only. |
| `POST /clubs` | ✅ | Exact active protected route. |
| `POST /clubs/thumbnails` | 🟨 | Active replacement is public `POST /clubs/{clubId}/thumbnails` with a JSON body. |
| `GET /clubs/{clubId}` | ✅ | Exact active public route. |
| `GET /clubs/{clubId}/members` | ⬜ | Placeholder text only; no route, handler, or query. |
| `GET /clubs/{clubId}/eboard` | ⬜ | No route, handler, or list query. |
| `POST /clubs/{clubId}/members` | 🟨 | Active replacement is caller-only `POST .../members/me`. |
| `PUT /clubs/{clubId}/members/roles` | ⬜ | No route, handler, owner-only check, or update query. |
| `DELETE /clubs/{clubId}/members/me` | 🟨 | Handler/query exist, but the active route omits the JWT authorizer. |
| `GET /clubs/{clubId}/events` | ✅ | Exact active public route. |
| `GET /clubs/{clubId}/events/drafts` | ⬜ | No active route/handler; only reusable status-filtered SQL support. |
| `POST /clubs/{clubId}/events` | 🟨 | Active route exists, but SQL is invalid and club-role authorization is absent. |
| `GET /events` | ✅ | Exact active public posted-event route. |
| `GET /events/{eventId}` | ✅ | Exact active public posted-event route. |
| `GET /events/{eventId}/images` | 🟨 | Exact path is an inactive stub; changed club-scoped active read lacks its SQL. |
| `GET /events/{eventId}/description` | 🟨 | Dedicated path inactive; real event detail includes description. |
| `GET /events/{eventId}/clubs` | 🟨 | Dedicated path inactive; real event detail includes owner/associate IDs. |
| `GET /auth/events/{eventId}` | 🟨 | Public read and unused authorization helper exist separately; no composed protected route. |
| `GET /auth/events/{eventId}/images` | 🟨 | Public changed-path handler exists but lacks SQL and authorization. |
| `PATCH /auth/events/{eventId}` | ⬜ | No update handler or query. |
| `POST /auth/events/{eventId}/thumbnails` | 🟨 | Public club-scoped S3 signer exists; authorization and confirmation do not. |
| `POST /auth/events/{eventId}/images` | 🟨 | Public club-scoped S3 signer/confirmation exist; authorization does not. |
| `DELETE /auth/events/{eventId}` | ⬜ | No archive/delete handler or query. |
| `GET /admins` | ⬜ | Planning text/schema only. |
| `GET /admins/{studentId}` | 🟨 | Unwired SQL stub hardcodes an obsolete numeric student ID. |
| `POST /admins` | ⬜ | No route, handler, insert query, or admin check. |
| `DELETE /admins/{studentId}` | 🟨 | Unwired SQL stub hardcodes an obsolete numeric student ID. |
| `GET /clubs?verified=true` | ✅ | Implemented by active `GET /clubs` query-parameter handling. |
| `POST /clubs/{clubId}/verification` | ⬜ | Schema/read support only; no write operation. |
| `DELETE /clubs/{clubId}/verification` | ⬜ | Schema/read support only; no write operation. |
| Internal `POST /images` | 🟨 | Generic function absent; metadata insertion exists in event-image confirmation. |
| `DELETE /images/{imageId}` | ⬜ | No database or storage deletion workflow. |

## Internal student registration

### 🔵 ✅ Cognito student synchronization

**Purpose:** Upserts a student from Cognito `sub` and optional email after sign-up confirmation and after authentication.

**Invocation:** Cognito `PostConfirmation_ConfirmSignUp` and `PostAuthentication_Authentication`; not API Gateway.

**Legacy implementation:** [`lambda/internal/auth/postConfirm/upsert.go`](../infrastructure/legacy/lambda/internal/auth/postConfirm/upsert.go), wired by [`authorization.go`](../infrastructure/legacy/internal/stack/authorization.go) and [`developmentApi.go`](../infrastructure/legacy/internal/stack/developmentApi.go).

**Database query:** [Student existence and upsert](../database/README.md#1-student-existence-and-upsert) — [`students/UPSERT_student.sql`](../infrastructure/legacy/utils/query_client/queries/students/UPSERT_student.sql).

**Target API location:** `api/internal/students/sync/` (**planned**).

**Target database query location:** `database/queries/students/ensure/` (**planned**).

**Portable:** Yes, after extracting Cognito event parsing and the AWS credential/connection adapter.

**Notes:** The handler logs and returns the original Cognito event when the database write fails, allowing authentication to continue. Protected HTTP handlers also call `RequireStudent` as a request-time fallback.

## Shared authentication and authorization dependencies

| Capability | Legacy source | Current use | Planned responsibility |
| --- | --- | --- | --- |
| Extract verified `sub` and optional email | [`extract_sub.go`](../infrastructure/legacy/utils/auth/extract_sub.go) | Club/event creation and membership handlers | HTTP/identity adapter plus application identity context |
| Ensure a student row exists | [`ensure_student.go`](../infrastructure/legacy/utils/auth/ensure_student.go) | All six JWT-attached handlers | Shared student synchronization service |
| Check club e-board/owner role | [`club_authorization.go`](../infrastructure/legacy/utils/auth/club_authorization.go) and duplicate [`lambda/internal` helper](../infrastructure/legacy/lambda/internal/auth/club_authorization/club_authorization.go) | **No active handler calls it** | Shared club authorization policy using [query group 13](../database/README.md#13-club-authorization) |
| Check event-associated club role | [`event_authorization.go`](../infrastructure/legacy/utils/auth/event_authorization.go) and duplicate [`lambda/internal` helper](../infrastructure/legacy/lambda/internal/auth/event_authorization/event_authorization.go) | **No active handler calls it** | Shared event authorization policy using [query group 14](../database/README.md#14-event-authorization) |
| Format validation errors | [`validation_error.go`](../infrastructure/legacy/utils/errors/validation_error.go) | Club and event create handlers | Shared transport-neutral validation result mapping |

A Cognito JWT establishes identity only. It does not establish admin, membership, e-board, owner, or event authority. The current enforcement gaps must be preserved as known migration work, not silently presented as implemented authorization.

## Me

### 🔴 ✅ GET `/me`

**Purpose:** Returns the student record associated with the caller's verified Cognito `sub`.

**Authentication:** Cognito JWT authorizer attached; identity-scoped only.

**Legacy route:** [`gateway/routes/student_routes.go`](../infrastructure/legacy/gateway/routes/student_routes.go).

**Legacy handler:** [`lambda/api/me/get.go`](../infrastructure/legacy/lambda/api/me/get.go).

**Database query:** [Current student](../database/README.md#2-current-student) — [`students/SELECT_student_by_sub.sql`](../infrastructure/legacy/utils/query_client/queries/students/SELECT_student_by_sub.sql), plus the shared student-existence/upsert group.

**Authorization dependencies:** Verified `sub`; `RequireStudent`. No role check is required because the lookup is caller-relative.

**Target API location:** `api/me/get/` (**planned**).

**Target database query location:** `database/queries/me/get/` (**planned**).

**Portable:** Yes, after extracting API Gateway/JWT parsing and database connection creation.

**Notes:** The database permits a null email while the current Go response model uses a non-nullable string; preserve the current contract intentionally or correct it in a separately reviewed migration.

### 🔴 ✅ GET `/me/clubs`

**Purpose:** Lists clubs joined by the caller and computes `owner`, `eboard`, or `member` for each membership.

**Authentication:** Cognito JWT authorizer attached.

**Legacy route:** [`gateway/routes/student_routes.go`](../infrastructure/legacy/gateway/routes/student_routes.go).

**Legacy handler:** [`lambda/api/me/clubs/get.go`](../infrastructure/legacy/lambda/api/me/clubs/get.go).

**Database query:** [Current student's clubs and roles](../database/README.md#3-current-students-clubs-and-roles) — [`students/SELECT_student_clubs.sql`](../infrastructure/legacy/utils/query_client/queries/students/SELECT_student_clubs.sql).

**Authorization dependencies:** Verified `sub`; `RequireStudent`; membership rows scope the result.

**Target API location:** `api/me/clubs/get/` (**planned**).

**Target database query location:** `database/queries/me/clubs/list/` (**planned**).

**Portable:** Yes, after extracting HTTP/JWT and database adapters.

**Notes:** The field named `thumbnailUrl` currently contains the database `object_key`, not a signed or public URL.

### 🔴 ✅ GET `/me/events`

**Purpose:** Lists posted events associated with clubs the caller has joined, with date and pagination filters.

**Authentication:** Cognito JWT authorizer attached.

**Legacy route:** [`gateway/routes/student_routes.go`](../infrastructure/legacy/gateway/routes/student_routes.go).

**Legacy handler:** [`lambda/api/me/events/get.go`](../infrastructure/legacy/lambda/api/me/events/get.go).

**Database query:** [Current student's events](../database/README.md#4-current-students-events) — [`students/SELECT_student_events.sql`](../infrastructure/legacy/utils/query_client/queries/students/SELECT_student_events.sql).

**Authorization dependencies:** Verified `sub`; `RequireStudent`; `club_members` scopes results.

**Target API location:** `api/me/events/get/` (**planned**).

**Target database query location:** `database/queries/me/events/list/` (**planned**).

**Portable:** Yes, after extracting HTTP/JWT and database adapters.

**Notes:** This is the active replacement for historical `GET /me/clubs/events`. SQL paginates joined rows before Go groups them by event, so one page can contain fewer distinct events and partial club associations.

### 🔴 🟨 GET `/me/clubs/events`

**Purpose:** Historical path for the caller's club events.

**Authentication:** Historically protected; no active route exists.

**Legacy implementation:** Defined only by the commented-out [`StubLambdaStack`](../infrastructure/legacy/internal/stack/stubLambda.go) with a hard-coded handler at [`stub/lambda/me/clubs/events/get.go`](../infrastructure/legacy/stub/lambda/me/clubs/events/get.go). The real implementation moved to `GET /me/events`.

**Database query:** [Current student's events](../database/README.md#4-current-students-events) is the active implementation. A stale hard-coded SQL example remains at [`GET_me_clubs_events.sql`](../infrastructure/legacy/stub/lambda/me/clubs/events/GET_me_clubs_events.sql).

**Authorization dependencies:** Same caller identity and membership scoping as `GET /me/events`.

**Target API location:** Reuse `api/me/events/get/` (**planned**); retain the old path only if an explicit compatibility requirement is approved.

**Target database query location:** `database/queries/me/events/list/` (**planned**).

**Portable:** Partial; the real replacement is portable, but the historical route itself is not active.

**Notes:** Do not migrate both paths as separate business logic.

### 🔴 🟨 GET `/me/clubs/eboard`

**Purpose:** Lists clubs for which the caller is an e-board member or owner.

**Authentication:** Historically protected; no active route exists.

**Legacy implementation:** Inactive stub route in [`stubLambda.go`](../infrastructure/legacy/internal/stack/stubLambda.go) and hard-coded handler [`stub/lambda/me/clubs/eboard/get.go`](../infrastructure/legacy/stub/lambda/me/clubs/eboard/get.go).

**Database query:** [My e-board clubs](../database/README.md#17-my-e-board-clubs) — stale stub SQL at [`eboard.sql`](../infrastructure/legacy/stub/lambda/me/clubs/eboard/eboard.sql); the active club/role query can likely be reused.

**Authorization dependencies:** Verified `sub`; membership with `member_is_eboard` or `member_is_owner`.

**Target API location:** `api/me/clubs/eboard/get/` (**planned**).

**Target database query location:** `database/queries/me/clubs/eboard/list/`, or reuse `database/queries/me/clubs/list/` with an explicit role filter (**planned**).

**Portable:** Partial.

**Notes:** The stub SQL has an `AND`/`OR` precedence defect and selects student IDs rather than clubs; it is design evidence, not migration-ready SQL.

## Clubs

### 🟢 ✅ GET `/clubs`

**Purpose:** Lists club summaries. With `verified=true` or `verified=TRUE`, returns only clubs represented in `verified_clubs`; otherwise returns all clubs.

**Authentication:** None.

**Legacy route:** [`gateway/routes/club_routes.go`](../infrastructure/legacy/gateway/routes/club_routes.go).

**Legacy handler:** [`lambda/api/clubs/get.go`](../infrastructure/legacy/lambda/api/clubs/get.go).

**Database query:** [Club list and verified filter](../database/README.md#5-club-list-and-verified-filter) — [`clubs/SELECT_clubs.sql`](../infrastructure/legacy/utils/query_client/queries/clubs/SELECT_clubs.sql).

**Authorization dependencies:** None; verification is a data filter, not an authorization grant.

**Target API location:** `api/clubs/list/` (**planned**).

**Target database query location:** `database/queries/clubs/list/` (**planned**).

**Portable:** Yes, after extracting the HTTP and database adapters.

**Notes:** This one active route covers both historical “all clubs” and `GET /clubs?verified=true` behavior.

### 🔴 ✅ POST `/clubs`

**Purpose:** Creates a club and its `club_info` row.

**Authentication:** Cognito JWT authorizer attached. No admin check is performed.

**Legacy route:** [`gateway/routes/club_routes.go`](../infrastructure/legacy/gateway/routes/club_routes.go).

**Legacy handler:** [`lambda/api/clubs/post/post.go`](../infrastructure/legacy/lambda/api/clubs/post/post.go).

**Database query:** [Club creation](../database/README.md#7-club-creation) — [`clubs/INSERT_club.sql`](../infrastructure/legacy/utils/query_client/queries/clubs/INSERT_club.sql) and [`clubs/INSERT_club_info.sql`](../infrastructure/legacy/utils/query_client/queries/clubs/INSERT_club_info.sql) in `ExecInsertQuery`.

**Authorization dependencies:** Verified `sub`; `RequireStudent`. No admin, membership, or ownership helper is called.

**Target API location:** `api/clubs/create/` (**planned**).

**Target database query location:** `database/queries/clubs/create/` (**planned**).

**Portable:** Yes, after extracting HTTP/JWT and database adapters.

**Notes:** The creator is not inserted into `club_members`, no owner is assigned, and the club is not verified. Whether creator ownership should be added is the database map's [manual-review item](../database/README.md#29-club-creator-ownership).

### 🟢 ✅ GET `/clubs/{clubId}`

**Purpose:** Returns one club with its name, logo object key, website URL, and description.

**Authentication:** None.

**Legacy route:** [`gateway/routes/club_routes.go`](../infrastructure/legacy/gateway/routes/club_routes.go).

**Legacy handler:** [`lambda/api/clubs/clubId/get.go`](../infrastructure/legacy/lambda/api/clubs/clubId/get.go).

**Database query:** [Club detail](../database/README.md#6-club-detail) — [`clubs/SELECT_club.sql`](../infrastructure/legacy/utils/query_client/queries/clubs/SELECT_club.sql).

**Authorization dependencies:** None.

**Target API location:** `api/clubs/{clubId}/get/` (**planned**).

**Target database query location:** `database/queries/clubs/get/` (**planned**).

**Portable:** Yes, after extracting the HTTP and database adapters.

**Notes:** The handler passes `clubId` to SQL without integer validation and reports a missing row through its `400` helper rather than `404`.

### 🔴 ⬜ GET `/clubs/{clubId}/members`

**Purpose:** Historical protected listing of a club's members.

**Authentication:** Historically limited to e-board members and owners; no active route exists.

**Legacy implementation:** No route, handler, or application query found; only placeholder text exists under [`stub/lambda/clubs/clubId/members/`](../infrastructure/legacy/stub/lambda/clubs/clubId/members/).

**Database query:** [Club member and e-board listing](../database/README.md#18-club-member-and-e-board-listing) — required but not implemented.

**Authorization dependencies:** Club e-board/owner check using the reusable club authorization policy.

**Target API location:** `api/clubs/{clubId}/members/list/` (**planned**).

**Target database query location:** `database/queries/clubs/members/list/` (**planned**).

**Portable:** N/A until implemented.

**Notes:** Decide response fields, pagination, and whether ordinary members may see any subset before implementation.

### 🔴 ⬜ GET `/clubs/{clubId}/eboard`

**Purpose:** Historical protected listing of owner/e-board memberships for a club.

**Authentication:** Historically limited to e-board members and owners; no active route exists.

**Legacy implementation:** No route, handler, or query found; only placeholder material exists.

**Database query:** Reuse [club member and e-board listing](../database/README.md#18-club-member-and-e-board-listing) with an explicit role filter.

**Authorization dependencies:** Club e-board/owner check.

**Target API location:** `api/clubs/{clubId}/members/list/` with a role filter, or `api/clubs/{clubId}/eboard/get/` if a distinct contract is retained (**planned**).

**Target database query location:** Prefer shared `database/queries/clubs/members/list/` (**planned**).

**Portable:** N/A until implemented.

**Notes:** Avoid duplicate member-list SQL solely to preserve two route names.

### 🔴 ✅ POST `/clubs/{clubId}/members/me`

**Purpose:** Joins the authenticated caller to a club as a regular member.

**Authentication:** Cognito JWT authorizer attached.

**Legacy route:** [`gateway/routes/club_routes.go`](../infrastructure/legacy/gateway/routes/club_routes.go).

**Legacy handler:** [`lambda/api/clubs/clubId/members/me/post/post.go`](../infrastructure/legacy/lambda/api/clubs/clubId/members/me/post/post.go).

**Database query:** [Join caller to club](../database/README.md#8-join-caller-to-club) — [`clubs/INSERT_club_member.sql`](../infrastructure/legacy/utils/query_client/queries/clubs/INSERT_club_member.sql).

**Authorization dependencies:** Verified `sub`; `RequireStudent`. The caller identity, not a body-supplied student ID, is inserted.

**Target API location:** `api/clubs/{clubId}/members/join/` (**planned**).

**Target database query location:** `database/queries/clubs/members/create/` or a caller-specific wrapper under `join/` (**planned**).

**Portable:** Yes, after extracting HTTP/JWT and database adapters.

**Notes:** This is the active self-service form of historical `POST /clubs/{clubId}/members`. It assigns both role flags false and does not require the club to be verified.

### 🔴 ⬜ PUT `/clubs/{clubId}/members/roles`

**Purpose:** Historical owner-only promotion or demotion of a member's e-board/owner flags.

**Authentication:** Historically protected; no active route exists.

**Legacy implementation:** No route, handler, or `UPDATE club_members` query found.

**Database query:** [Update member roles](../database/README.md#20-update-member-roles) — required but not implemented.

**Authorization dependencies:** Verified caller plus owner-only policy. The existing club helper allows e-board **or** owner and is therefore not sufficient for the historical owner-only rule without refinement.

**Target API location:** `api/clubs/{clubId}/members/roles/update/` (**planned**).

**Target database query location:** `database/queries/clubs/members/update_role/` (**planned**).

**Portable:** N/A until implemented.

**Notes:** The historical query parameters did not identify the member consistently; define the target-member contract before migration.

### 🟢 🟨 DELETE `/clubs/{clubId}/members/me`

**Purpose:** Leaves a club as the current student while preventing an owner membership from being deleted.

**Authentication:** **Misconfigured in legacy.** The active API Gateway route is public, but the handler requires verified JWT claims and therefore returns before executing the delete in ordinary calls.

**Legacy route:** [`gateway/routes/club_routes.go`](../infrastructure/legacy/gateway/routes/club_routes.go) — the registration omits `Authorizer`.

**Legacy handler:** [`lambda/api/clubs/clubId/members/me/delete/delete.go`](../infrastructure/legacy/lambda/api/clubs/clubId/members/me/delete/delete.go).

**Database query:** [Leave caller's club](../database/README.md#9-leave-callers-club) — [`clubs/DELETE_club_member.sql`](../infrastructure/legacy/utils/query_client/queries/clubs/DELETE_club_member.sql).

**Authorization dependencies:** Intended verified `sub`; `RequireStudent`; SQL owner guard.

**Target API location:** `api/clubs/{clubId}/members/leave/` (**planned**).

**Target database query location:** `database/queries/clubs/members/leave/` (**planned**).

**Portable:** Partial; the handler/query logic is reusable after correcting the transport authorization boundary.

**Notes:** This documentation records the defect; this task does not change route behavior. A nullable owner flag also requires review because `member_is_owner = FALSE` does not match null.

### 🟢 ✅ GET `/clubs/{clubId}/events`

**Purpose:** Lists posted events linked to a club with date and pagination filters.

**Authentication:** None.

**Legacy route:** [`gateway/routes/club_routes.go`](../infrastructure/legacy/gateway/routes/club_routes.go).

**Legacy handler:** [`lambda/api/clubs/clubId/events/get.go`](../infrastructure/legacy/lambda/api/clubs/clubId/events/get.go).

**Database query:** [Club event list](../database/README.md#10-club-event-list) — [`clubs/SELECT_club_events.sql`](../infrastructure/legacy/utils/query_client/queries/clubs/SELECT_club_events.sql).

**Authorization dependencies:** None; the handler hardcodes `posted` status.

**Target API location:** `api/clubs/{clubId}/events/list/` (**planned**).

**Target database query location:** `database/queries/clubs/events/list/` (**planned**).

**Portable:** Yes, after extracting the HTTP and database adapters.

**Notes:** Joined-row pagination, strict date comparisons, map-based response grouping, placeholder thumbnails, and partial association results are current behaviors requiring contract tests before migration.

### 🔴 ⬜ GET `/clubs/{clubId}/events/drafts`

**Purpose:** Historical e-board/owner listing of drafted events for one club.

**Authentication:** Historically protected; no active route exists.

**Legacy implementation:** No active route or handler. The current club-event SQL is status-parameterized, but its handler always supplies `posted`.

**Database query:** [Club draft events](../database/README.md#21-club-draft-events) — reusable query support exists, endpoint behavior does not.

**Authorization dependencies:** Club e-board/owner check.

**Target API location:** `api/clubs/{clubId}/events/drafts/get/` (**planned**) or an authenticated filter on the shared list service.

**Target database query location:** Reuse `database/queries/clubs/events/list/` with an explicit authorized status policy (**planned**).

**Portable:** N/A as an endpoint; the underlying read query is portable.

**Notes:** Do not expose a caller-controlled status filter without enforcing the role policy first.

### 🔴 🟨 POST `/clubs/{clubId}/events`

**Purpose:** Creates a drafted event, owner-club link, associate-club links, and description.

**Authentication:** Cognito JWT authorizer attached. Contrary to the historical design, no membership/e-board/owner authorization check runs.

**Legacy route:** [`gateway/routes/club_routes.go`](../infrastructure/legacy/gateway/routes/club_routes.go).

**Legacy handler:** [`lambda/api/clubs/clubId/events/post/post.go`](../infrastructure/legacy/lambda/api/clubs/clubId/events/post/post.go).

**Database query:** [Create event draft](../database/README.md#12-create-event-draft) — [`INSERT_event.sql`](../infrastructure/legacy/utils/query_client/queries/events/INSERT_event.sql), [`INSERT_event_club_link.sql`](../infrastructure/legacy/utils/query_client/queries/events/INSERT_event_club_link.sql), and [`INSERT_event_description.sql`](../infrastructure/legacy/utils/query_client/queries/events/INSERT_event_description.sql).

**Authorization dependencies:** Verified `sub`; `RequireStudent`; should call the club authorization policy but currently does not.

**Target API location:** `api/clubs/{clubId}/events/create/` (**planned**).

**Target database query location:** `database/queries/events/create/` (**planned**, shared across route adapters).

**Portable:** Partial; validation and orchestration are reusable after SQL repair, one transaction boundary, and authorization extraction.

**Notes:** `INSERT_event.sql` lists 11 columns but only 10 values and omits an RSVP value expression; `INSERT_event_description.sql` has a trailing comma. The initial event insert is also outside the later link/description transaction.

### 🟢 ✅ POST `/clubs/{clubId}/thumbnails`

**Purpose:** Returns a one-minute S3 presigned PUT URL for a club-thumbnail object key.

**Authentication:** None in the active route.

**Legacy route:** [`gateway/routes/club_image_routes.go`](../infrastructure/legacy/gateway/routes/club_image_routes.go).

**Legacy handler:** [`lambda/api/clubs/thumbnails/post/post.go`](../infrastructure/legacy/lambda/api/clubs/thumbnails/post/post.go).

**Database query:** N/A. The endpoint only signs storage access; [thumbnail metadata assignment](../database/README.md#27-club-and-event-thumbnail-metadata-assignment) is not implemented.

**Authorization dependencies:** None currently; historical intent was protected and should be resolved before migration.

**Target API location:** `api/clubs/{clubId}/thumbnails/presign/` (**planned**).

**Target database query location:** N/A for presigning; confirmation would use `database/queries/clubs/thumbnails/confirm/` (**planned**).

**Portable:** Yes, after introducing a storage-provider abstraction.

**Notes:** The handler does not verify club existence or caller authority and does not create image metadata or update `clubs.fk_logo_id`.

## Events

### 🟢 ✅ GET `/events`

**Purpose:** Lists posted events with optional date and pagination filters.

**Authentication:** None.

**Legacy route:** [`gateway/routes/event_routes.go`](../infrastructure/legacy/gateway/routes/event_routes.go).

**Legacy handler:** [`lambda/api/events/get.go`](../infrastructure/legacy/lambda/api/events/get.go).

**Database query:** [Public/composite event read](../database/README.md#11-public-and-composite-event-read) — [`events/SELECT_events.sql`](../infrastructure/legacy/utils/query_client/queries/events/SELECT_events.sql).

**Authorization dependencies:** None; the handler hardcodes `posted` status.

**Target API location:** `api/events/list/` (**planned**).

**Target database query location:** `database/queries/events/read/` with a list wrapper (**planned**).

**Portable:** Yes, after extracting the HTTP and database adapters.

**Notes:** Preserve or deliberately revise strict date bounds, joined-row pagination, unordered map grouping, and the fixed placeholder thumbnail through contract tests.

### 🟢 ✅ GET `/events/{eventId}`

**Purpose:** Returns one posted event with its description, owner-club ID, and associate-club IDs.

**Authentication:** None.

**Legacy route:** [`gateway/routes/event_routes.go`](../infrastructure/legacy/gateway/routes/event_routes.go).

**Legacy handler:** [`lambda/api/events/eventId/get.go`](../infrastructure/legacy/lambda/api/events/eventId/get.go).

**Database query:** Reuses [public/composite event read](../database/README.md#11-public-and-composite-event-read) and [`events/SELECT_events.sql`](../infrastructure/legacy/utils/query_client/queries/events/SELECT_events.sql) with `status='posted'` and the ID as a `LIKE` argument.

**Authorization dependencies:** None; drafted and archived events are excluded by the hardcoded status.

**Target API location:** `api/events/{eventId}/get/` (**planned**).

**Target database query location:** Reuse `database/queries/events/read/` with a single-event wrapper (**planned**).

**Portable:** Yes, after extracting the HTTP and database adapters.

**Notes:** This composite response supersedes historical dedicated description and club subresources. The handler requires only a non-empty ID and caps joined rows at 100.

### 🟢 🟨 GET `/events/{eventId}/images`

**Purpose:** Historical public gallery read for a posted event.

**Authentication:** Historically public; no active exact-path route exists.

**Legacy implementation:** Inactive stub route in [`stubLambda.go`](../infrastructure/legacy/internal/stack/stubLambda.go) with hard-coded handler [`stub/lambda/events/eventId/images/get.go`](../infrastructure/legacy/stub/lambda/events/eventId/images/get.go). A changed club-scoped route is active but cannot load its missing query.

**Database query:** [Event-image list](../database/README.md#16-event-image-list) — required `images`/`event_images` join is missing.

**Authorization dependencies:** Public reads must first establish that the event is posted; the current club-scoped handler does not perform that check.

**Target API location:** `api/events/{eventId}/images/get/` (**planned**).

**Target database query location:** `database/queries/events/images/list/` (**planned**).

**Portable:** Partial; requires a query, a posted-visibility policy, and a storage-provider abstraction.

**Notes:** Do not treat the hard-coded stub response as an implementation contract.

### 🟢 🟨 GET `/events/{eventId}/description`

**Purpose:** Historical public description subresource for a posted event.

**Authentication:** Historically public; no active dedicated route exists.

**Legacy implementation:** Inactive stub handler [`stub/lambda/events/eventId/description/get.go`](../infrastructure/legacy/stub/lambda/events/eventId/description/get.go). The real active `GET /events/{eventId}` already returns the description.

**Database query:** Reuse [public/composite event read](../database/README.md#11-public-and-composite-event-read); a stale stub query remains at [`description.sql`](../infrastructure/legacy/stub/lambda/events/eventId/description/description.sql).

**Authorization dependencies:** Posted-event visibility.

**Target API location:** No separate endpoint is planned; use `api/events/{eventId}/get/` unless compatibility requires a subresource adapter.

**Target database query location:** Reuse `database/queries/events/read/` (**planned**).

**Portable:** Partial as a historical route; the combined behavior is portable.

**Notes:** Avoid duplicating the event read solely to recreate the old path.

### 🟢 🟨 GET `/events/{eventId}/clubs`

**Purpose:** Historical public listing of the owner and associated clubs for a posted event.

**Authentication:** Historically public; no active dedicated route exists.

**Legacy implementation:** Inactive hard-coded handler [`stub/lambda/events/eventId/clubs/get.go`](../infrastructure/legacy/stub/lambda/events/eventId/clubs/get.go). The active event-detail response already groups owner and associate club IDs.

**Database query:** Reuse [public/composite event read](../database/README.md#11-public-and-composite-event-read).

**Authorization dependencies:** Posted-event visibility.

**Target API location:** No separate endpoint is planned; use `api/events/{eventId}/get/` unless compatibility requires a subresource adapter.

**Target database query location:** Reuse `database/queries/events/read/` (**planned**).

**Portable:** Partial as a historical route; the combined behavior is portable.

**Notes:** Current composite responses include club IDs but leave club names empty and substitute a fixed thumbnail URL.

## Authorized events

The historical revamp introduced `/auth/events` because an event may be associated with several clubs; authorization should not depend on one `{clubId}` in the route. No route in this section is active today.

### 🔴 🟨 GET `/auth/events/{eventId}`

**Purpose:** Returns a draft, posted, or otherwise non-public event to an authorized club manager.

**Authentication:** Historically protected; no active route exists.

**Legacy implementation:** No composed handler. The active public [`eventId/get.go`](../infrastructure/legacy/lambda/api/events/eventId/get.go) and unused event-authorization helper are separate reusable primitives.

**Database query:** [Authorized event read](../database/README.md#24-authorized-event-read) can reuse [event read](../database/README.md#11-public-and-composite-event-read) plus [event authorization](../database/README.md#14-event-authorization).

**Authorization dependencies:** Verified `sub`; e-board/owner of any club linked to the event.

**Target API location:** `api/auth/events/{eventId}/get/` (**planned**).

**Target database query location:** Reuse `database/queries/events/read/` and `database/queries/authorization/events/can_manage/` (**planned**).

**Portable:** Partial; the read and policy primitives exist but are not composed.

**Notes:** Define which statuses an authorized caller may see and whether authors have any independent permission.

### 🔴 🟨 GET `/auth/events/{eventId}/images`

**Purpose:** Returns image metadata/URLs for an event, including non-public events, after event-role authorization.

**Authentication:** Historically protected; no active route exists.

**Legacy implementation:** A public club-scoped handler exists at [`lambda/api/clubs/events/images/get/get.go`](../infrastructure/legacy/lambda/api/clubs/events/images/get/get.go), but it references missing SQL and performs no role check.

**Database query:** [Event-image list](../database/README.md#16-event-image-list) plus [event authorization](../database/README.md#14-event-authorization).

**Authorization dependencies:** Verified `sub`; e-board/owner of a linked club.

**Target API location:** `api/auth/events/{eventId}/images/get/` (**planned**).

**Target database query location:** `database/queries/events/images/list/` plus shared authorization query (**planned**).

**Portable:** Partial; requires missing SQL, policy composition, and a storage adapter.

**Notes:** The current handler ignores `{clubId}` and should not be duplicated merely to support this route.

### 🔴 ⬜ PATCH `/auth/events/{eventId}`

**Purpose:** Updates event fields and supports state changes such as draft-to-posted publication.

**Authentication:** Historically protected; no active route exists.

**Legacy implementation:** No route, handler, update query, or transaction orchestration found.

**Database query:** [Event update and publish](../database/README.md#25-event-update-and-publish) — required but not implemented.

**Authorization dependencies:** Verified `sub`; event e-board/owner policy; field-level validation and an explicit status-transition policy.

**Target API location:** `api/auth/events/{eventId}/update/` (**planned**).

**Target database query location:** `database/queries/events/update/` (**planned**).

**Portable:** N/A until implemented.

**Notes:** Define patch semantics, allowed fields, optimistic concurrency, associate-club changes, and publication rules before writing SQL.

### 🔴 🟨 POST `/auth/events/{eventId}/thumbnails`

**Purpose:** Returns an upload URL for an event thumbnail after event-role authorization.

**Authentication:** Historically protected; no active `/auth` route exists.

**Legacy implementation:** Core S3 signing exists at the public club-scoped handler [`lambda/api/clubs/events/thumbnails/post/post.go`](../infrastructure/legacy/lambda/api/clubs/events/thumbnails/post/post.go).

**Database query:** N/A for presigning. [Thumbnail metadata assignment](../database/README.md#27-club-and-event-thumbnail-metadata-assignment) is absent.

**Authorization dependencies:** Verified `sub`; event e-board/owner policy.

**Target API location:** `api/auth/events/{eventId}/thumbnails/presign/` (**planned**).

**Target database query location:** N/A for signing; confirmation would use `database/queries/events/thumbnails/confirm/` (**planned**).

**Portable:** Partial; the signer is reusable behind a storage-provider abstraction, but authorization and confirmation are missing.

**Notes:** Decide whether the future canonical route remains club-scoped or moves to `/auth`; do not maintain two independent signers.

### 🔴 🟨 POST `/auth/events/{eventId}/images`

**Purpose:** Returns an upload URL for an event-gallery image after event-role authorization.

**Authentication:** Historically protected; no active `/auth` route exists.

**Legacy implementation:** Core S3 signing exists at the public club-scoped handler [`lambda/api/clubs/events/images/post/post.go`](../infrastructure/legacy/lambda/api/clubs/events/images/post/post.go), with a separate public confirmation handler.

**Database query:** N/A for presigning; [event-image metadata confirmation](../database/README.md#15-event-image-metadata-confirmation) applies after upload.

**Authorization dependencies:** Verified `sub`; event e-board/owner policy.

**Target API location:** `api/auth/events/{eventId}/images/presign/` and `confirm/` (**planned**).

**Target database query location:** `database/queries/events/images/confirm/` for confirmation (**planned**).

**Portable:** Partial; the signer/confirmation behavior is reusable after policy, storage, and transaction extraction.

**Notes:** The current endpoints trust client-provided metadata and do not verify the S3 object before database insertion.

### 🔴 ⬜ DELETE `/auth/events/{eventId}`

**Purpose:** Archives or deletes an event according to the product retention policy.

**Authentication:** Historically protected; no active route exists.

**Legacy implementation:** No route, handler, archive/delete query, or cleanup workflow found.

**Database query:** [Event deletion](../database/README.md#26-event-deletion) — required but not implemented.

**Authorization dependencies:** Verified `sub`; event e-board/owner policy.

**Target API location:** `api/auth/events/{eventId}/delete/` (**planned**).

**Target database query location:** `database/queries/events/delete/` (**planned**).

**Portable:** N/A until implemented.

**Notes:** The schema supports both `status='archived'` and `deleted_at`, while current public reads do not filter `deleted_at`; choose one lifecycle policy before implementation.

## Admins

The historical design required every admin route to verify the caller is already an admin. The schema contains an `admins` table, but the active application has no admin route or admin-check helper.

### 🔴 ⬜ GET `/admins`

**Purpose:** Lists administrator records and any approved related information.

**Authentication:** Historically protected with an admin check; no active route exists.

**Legacy implementation:** No executable route, handler, or list query found.

**Database query:** [Admin CRUD](../database/README.md#22-admin-crud) — list behavior is required but not implemented.

**Authorization dependencies:** Verified `sub` plus reusable admin-status check, which is also missing.

**Target API location:** `api/admins/list/` (**planned**).

**Target database query location:** `database/queries/admins/list/` and a shared `is_admin` query (**planned**).

**Portable:** N/A until implemented.

**Notes:** Define whether “their id and clubs” means managed clubs, ordinary memberships, or a separate response before implementation.

### 🔴 🟨 GET `/admins/{studentId}`

**Purpose:** Returns one administrator record if present.

**Authentication:** Historically protected with an admin check; no active route exists.

**Legacy implementation:** Only an unwired SQL stub at [`GET_admins_studentId.sql`](../infrastructure/legacy/stub/lambda/admins/studentId/GET_admins_studentId.sql); it hardcodes student ID `2` and has no handler.

**Database query:** [Admin CRUD](../database/README.md#22-admin-crud).

**Authorization dependencies:** Verified caller plus admin-status check.

**Target API location:** `api/admins/{studentId}/get/` (**planned**).

**Target database query location:** `database/queries/admins/get/` (**planned**).

**Portable:** Partial; the schema intent is clear, but the stub is not parameterized migration-ready code.

**Notes:** Current student IDs are Cognito UUID strings, not the integer used by the stale stub.

### 🔴 ⬜ POST `/admins`

**Purpose:** Promotes a student by inserting an admin record.

**Authentication:** Historically protected with an admin check; no active route exists.

**Legacy implementation:** No route, handler, insert query, or admin authorization helper found.

**Database query:** [Admin CRUD](../database/README.md#22-admin-crud) — create behavior is required but not implemented.

**Authorization dependencies:** Verified caller plus admin-status check; target student existence validation.

**Target API location:** `api/admins/create/` (**planned**).

**Target database query location:** `database/queries/admins/create/` (**planned**).

**Portable:** N/A until implemented.

**Notes:** Define bootstrap/last-admin policy so administration cannot be granted or removed without a recoverable authority path.

### 🔴 🟨 DELETE `/admins/{studentId}`

**Purpose:** Demotes a student by deleting the admin record.

**Authentication:** Historically protected with an admin check; no active route exists.

**Legacy implementation:** Only an unwired SQL stub at [`DELETE_admins_studentId.sql`](../infrastructure/legacy/stub/lambda/admins/studentId/DELETE_admins_studentId.sql); it hardcodes student ID `2` and has no handler.

**Database query:** [Admin CRUD](../database/README.md#22-admin-crud).

**Authorization dependencies:** Verified caller plus admin-status check.

**Target API location:** `api/admins/{studentId}/delete/` (**planned**).

**Target database query location:** `database/queries/admins/delete/` (**planned**).

**Portable:** Partial; only obsolete SQL-shaped evidence exists.

**Notes:** A last-admin/self-demotion policy requires manual product review.

## Verification and club administration

Public verification reads are already implemented by [`GET /clubs?verified=true`](#clubs). The historical write routes below are not.

### 🔴 ⬜ POST `/clubs/{clubId}/verification`

**Purpose:** Marks a club verified by inserting it into `verified_clubs`.

**Authentication:** Historically protected; no active route exists.

**Legacy implementation:** Schema support and read filtering exist, but no route, handler, insert query, or admin check was found.

**Database query:** [Club verification writes](../database/README.md#23-club-verification-writes) — required but not implemented.

**Authorization dependencies:** Verified caller plus an approved club-administration policy, historically admin-only.

**Target API location:** `api/clubs/{clubId}/verification/create/` (**planned**).

**Target database query location:** `database/queries/clubs/verification/create/` (**planned**).

**Portable:** N/A until implemented.

**Notes:** The unique table row makes the write naturally idempotent only if duplicate handling is defined.

### 🔴 ⬜ DELETE `/clubs/{clubId}/verification`

**Purpose:** Removes a club from the verified set.

**Authentication:** Historically protected; no active route exists.

**Legacy implementation:** No route, handler, delete query, or admin check found.

**Database query:** [Club verification writes](../database/README.md#23-club-verification-writes) — required but not implemented.

**Authorization dependencies:** Verified caller plus club-administration policy.

**Target API location:** `api/clubs/{clubId}/verification/delete/` (**planned**).

**Target database query location:** `database/queries/clubs/verification/delete/` (**planned**).

**Portable:** N/A until implemented.

**Notes:** Unverification currently affects discovery filtering only; it is not a membership or authorization revocation.

## Images

The current image flow signs short-lived S3 URLs so clients transfer bytes directly to S3. Storage signing is application behavior that can be retained behind an adapter; S3 itself should not define the API. See the [legacy image-flow documentation](../infrastructure/legacy/docs/architecture/image-uploads.md).

### 🟢 🟨 GET `/clubs/{clubId}/events/{eventId}/images`

**Purpose:** Loads event-image metadata and returns one-minute signed GET URLs.

**Authentication:** None in the active route.

**Legacy route:** [`gateway/routes/event_image_routes.go`](../infrastructure/legacy/gateway/routes/event_image_routes.go).

**Legacy handler:** [`lambda/api/clubs/events/images/get/get.go`](../infrastructure/legacy/lambda/api/clubs/events/images/get/get.go).

**Database query:** [Event-image list](../database/README.md#16-event-image-list). The referenced `images/SELECT_event_images.sql` file is absent.

**Authorization dependencies:** None currently. The handler neither checks posted visibility nor a club/event role.

**Target API location:** Prefer public `api/events/{eventId}/images/get/` for posted events and authenticated `api/auth/events/{eventId}/images/get/` for drafts (**planned**).

**Target database query location:** Shared `database/queries/events/images/list/` (**planned**).

**Portable:** Partial; the response/signing behavior is reusable after adding SQL, visibility policy, and a storage-provider abstraction.

**Notes:** The active handler ignores `{clubId}` and deterministically returns its query-error response before reaching S3 signing.

### 🟢 ✅ POST `/clubs/{clubId}/events/{eventId}/images`

**Purpose:** Creates a UUID-based object key and a one-minute signed PUT URL for an event-gallery image.

**Authentication:** None in the active route.

**Legacy route:** [`gateway/routes/event_image_routes.go`](../infrastructure/legacy/gateway/routes/event_image_routes.go).

**Legacy handler:** [`lambda/api/clubs/events/images/post/post.go`](../infrastructure/legacy/lambda/api/clubs/events/images/post/post.go).

**Database query:** N/A; metadata is written only by the separate confirmation endpoint.

**Authorization dependencies:** None currently; the handler does not verify club/event existence, association, or caller role.

**Target API location:** `api/auth/events/{eventId}/images/presign/` or a retained club-scoped adapter (**planned**).

**Target database query location:** N/A for presigning.

**Portable:** Yes, after introducing a storage-provider abstraction and authorization policy.

**Notes:** The handler ignores `{clubId}`, accepts any non-empty filename/MIME type, and signs `events/{eventId}/images/{uuid}{extension}`.

### 🟢 ✅ POST `/clubs/{clubId}/events/{eventId}/images/confirm`

**Purpose:** Records uploaded image metadata and links the image to an event.

**Authentication:** None in the active route.

**Legacy route:** [`gateway/routes/event_image_routes.go`](../infrastructure/legacy/gateway/routes/event_image_routes.go).

**Legacy handler:** [`lambda/api/clubs/events/images/confirm/post.go`](../infrastructure/legacy/lambda/api/clubs/events/images/confirm/post.go).

**Database query:** [Event-image metadata confirmation](../database/README.md#15-event-image-metadata-confirmation) — [`images/INSERT_image.sql`](../infrastructure/legacy/utils/query_client/queries/images/INSERT_image.sql) then [`images/INSERT_event_image.sql`](../infrastructure/legacy/utils/query_client/queries/images/INSERT_event_image.sql).

**Authorization dependencies:** None currently.

**Target API location:** `api/auth/events/{eventId}/images/confirm/` or a retained club-scoped adapter (**planned**).

**Target database query location:** `database/queries/events/images/confirm/` (**planned**).

**Portable:** Yes, after extracting the database/storage adapters and making confirmation transactional.

**Notes:** The handler ignores `{clubId}`, trusts client-supplied IDs/keys/MIME metadata, does not check S3 object existence, and commits its two inserts independently.

### 🟢 ✅ POST `/clubs/{clubId}/events/{eventId}/thumbnails`

**Purpose:** Returns a one-minute signed PUT URL for an event thumbnail.

**Authentication:** None in the active route.

**Legacy route:** [`gateway/routes/event_image_routes.go`](../infrastructure/legacy/gateway/routes/event_image_routes.go).

**Legacy handler:** [`lambda/api/clubs/events/thumbnails/post/post.go`](../infrastructure/legacy/lambda/api/clubs/events/thumbnails/post/post.go).

**Database query:** N/A. No metadata confirmation or `events.fk_thumbnail_id` assignment exists; see [thumbnail metadata assignment](../database/README.md#27-club-and-event-thumbnail-metadata-assignment).

**Authorization dependencies:** None currently.

**Target API location:** `api/auth/events/{eventId}/thumbnails/presign/` or a retained club-scoped adapter (**planned**).

**Target database query location:** N/A for presigning; confirmation would use `database/queries/events/thumbnails/confirm/` (**planned**).

**Portable:** Yes, after introducing storage and authorization adapters.

**Notes:** The handler ignores `{clubId}` and signs `events/{eventId}/thumbnails/{uuid}{extension}`.

### 🔴 ⬜ POST `/clubs/{clubId}/thumbnails/confirm`

**Purpose:** Historical image-flow step to persist club-thumbnail metadata and assign the uploaded image to the club.

**Authentication:** No active route; future write should require an approved club role.

**Legacy implementation:** No route, handler, query transaction, or S3 verification found.

**Database query:** [Thumbnail metadata assignment](../database/README.md#27-club-and-event-thumbnail-metadata-assignment) — required but not implemented.

**Authorization dependencies:** Verified caller plus club e-board/owner policy.

**Target API location:** `api/clubs/{clubId}/thumbnails/confirm/` (**planned**).

**Target database query location:** `database/queries/clubs/thumbnails/confirm/` (**planned**).

**Portable:** N/A until implemented.

**Notes:** Confirmation should create/reuse image metadata, update `clubs.fk_logo_id`, and define replacement cleanup atomically where possible.

### 🔴 ⬜ POST `/clubs/{clubId}/events/{eventId}/thumbnails/confirm`

**Purpose:** Historical image-flow step to persist event-thumbnail metadata and assign it to the event.

**Authentication:** No active route; future write should require event authority.

**Legacy implementation:** No route, handler, query transaction, or S3 verification found.

**Database query:** [Thumbnail metadata assignment](../database/README.md#27-club-and-event-thumbnail-metadata-assignment) — required but not implemented.

**Authorization dependencies:** Verified caller plus event e-board/owner policy.

**Target API location:** `api/auth/events/{eventId}/thumbnails/confirm/` (**planned**).

**Target database query location:** `database/queries/events/thumbnails/confirm/` (**planned**).

**Portable:** N/A until implemented.

**Notes:** The current presign response provides the required identifiers, but no code consumes them to update `events.fk_thumbnail_id`.

### 🔵 🟨 Internal image metadata write (historical `POST /images`)

**Purpose:** The historical revamp described an internal post-upload function that inserts generic image metadata.

**Invocation:** Historically blue/internal, not an active API Gateway route.

**Legacy implementation:** No generic handler. Equivalent metadata insertion exists only inside the active event-image confirmation handler.

**Database query:** Reuse [`images/INSERT_image.sql`](../infrastructure/legacy/utils/query_client/queries/images/INSERT_image.sql) from [event-image metadata confirmation](../database/README.md#15-event-image-metadata-confirmation).

**Authorization dependencies:** Depends on the calling workflow; the current event confirmation has none.

**Target API location:** Prefer a shared internal image-metadata service under `api/internal/images/confirm/`, called by route-specific confirmation handlers (**planned**).

**Target database query location:** Shared image insert within the route-appropriate confirmation transaction (**planned**).

**Portable:** Partial; the insert primitive exists, but the generic function/contract does not.

**Notes:** Do not expose a generic public write merely because the historical label resembles an HTTP path.

### 🔴 ⬜ DELETE `/images/{imageId}`

**Purpose:** Deletes image metadata and, under an explicit lifecycle policy, the corresponding stored object.

**Authentication:** Historically protected; no active route exists.

**Legacy implementation:** No route, handler, SQL cleanup, or S3 delete workflow found.

**Database query:** [Image deletion](../database/README.md#28-image-deletion) — required but not implemented.

**Authorization dependencies:** Verified caller plus ownership/club/event/admin policy based on the image's purpose and links.

**Target API location:** `api/images/{imageId}/delete/` (**planned**).

**Target database query location:** `database/queries/images/delete/` plus a storage deletion adapter (**planned**).

**Portable:** N/A until implemented.

**Notes:** Define link cleanup, thumbnail foreign-key clearing, object deletion ordering, failure recovery, and orphan retention before implementation.

## Operational routes

### 🟢 ✅ GET `/health`

**Purpose:** Returns a fixed “server running” response for basic API reachability.

**Authentication:** None.

**Legacy route:** [`gateway/routes/test_routes.go`](../infrastructure/legacy/gateway/routes/test_routes.go), registered by both active API stacks.

**Legacy handler:** [`lambda/api/test/ping/main.go`](../infrastructure/legacy/lambda/api/test/ping/main.go).

**Database query:** N/A.

**Authorization dependencies:** None.

**Target API location:** `api/operations/health/get/` (**planned**).

**Target database query location:** N/A.

**Portable:** Yes, after extracting the Lambda/API Gateway adapter.

**Notes:** Decide later whether liveness and readiness are separate; this handler currently proves only that Lambda executed, not database or storage health.

### 🟢 🟨 GET `/database/test`

**Purpose:** Dormant connectivity diagnostic that opens the configured MySQL connection and evaluates `SELECT 1 + 1`.

**Authentication:** The helper would be public if registered, but neither active API stack calls it.

**Legacy implementation:** Route helper [`gateway/routes/database_routes.go`](../infrastructure/legacy/gateway/routes/database_routes.go), integration [`gateway/integrations/test_database.go`](../infrastructure/legacy/gateway/integrations/test_database.go), and handler [`lambda/api/database/test/main.go`](../infrastructure/legacy/lambda/api/database/test/main.go).

**Database query:** Embedded diagnostic SQL, not an application query group.

**Authorization dependencies:** None in the dormant registration.

**Target API location:** No public business route is planned. Move equivalent coverage to protected operational diagnostics or integration tests.

**Target database query location:** N/A.

**Portable:** Partial as a connectivity test; it should not be migrated as a public API contract.

**Notes:** The source exists, but inactive wiring means this is not one of the 18 active routes.

## Recommended migration order

1. Define provider-neutral request, response, identity, database, and storage interfaces while snapshot-testing current response contracts.
2. Port shared identity extraction, student synchronization, validation, response mapping, and authorization policies once; remove duplicate role-helper implementations during that later migration.
3. Port the implemented read routes and their query groups before write workflows.
4. Port club membership and club creation with explicit decisions for route authorization and creator ownership.
5. Repair and transaction-test event creation before exposing it from the new module.
6. Introduce the storage-provider abstraction, then migrate presign, confirmation, and gallery-read behavior with authorization and lifecycle tests.
7. Implement historically planned routes only after their product/authorization decisions are approved; do not create empty endpoint folders solely to mirror this document.
8. Switch API Gateway/Lambda wiring only after parity tests pass. Keep `infrastructure/legacy/` as the working reference until an independently reviewed cutover.

## Manual review required

- Choose canonical protected event routes: historical `/auth/events/{eventId}` versus current club-scoped media paths.
- Decide whether club creation automatically grants ownership and whether creation requires an existing admin.
- Define owner-only versus e-board-or-owner policy for member-role changes.
- Define admin bootstrap, self-demotion, and last-admin protections.
- Correct the active leave-route authorizer mismatch without treating this documentation pass as the behavior change.
- Define event publication/update/archive semantics, including `status`, `deleted_at`, and associated-club changes.
- Define image authorization, content validation, S3 existence verification, confirmation transactions, replacement cleanup, and deletion recovery.
- Resolve current pagination/grouping and response-shape quirks through tests before deciding whether migration means strict parity or a versioned contract change.
- Review Cognito trigger ownership: the student-sync function is instantiated/wired in two stacks, and Cognito trigger slots are singleton configuration.

## Current versus planned state

- `api/README.md`: migration map only.
- `api/`: no migrated handler implementation yet.
- [`infrastructure/legacy/lambda/api/`](../infrastructure/legacy/lambda/api/): current handler source.
- [`infrastructure/legacy/gateway/`](../infrastructure/legacy/gateway/): current API Gateway route/integration source.
- [`infrastructure/legacy/utils/auth/`](../infrastructure/legacy/utils/auth/): current authentication/authorization helper source.
- [`database/README.md`](../database/README.md): linked query and schema migration map.

This pass changes documentation only. It does not move handlers, alter route/auth behavior, change response formats, or deploy anything.
