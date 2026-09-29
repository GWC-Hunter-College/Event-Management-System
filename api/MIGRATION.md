# API migration map

How the API is planned to move out of [`infrastructure/legacy/`](../infrastructure/legacy/) and into this provider-independent `api/` module. This material used to live in [`README.md`](README.md), which is now the endpoint reference.

No handlers have moved here yet. The integrated implementation under `infrastructure/legacy/` is still what runs, and that code, not the planning document, is the source of truth for every status. The companion [database migration map](../database/README.md) covers schema and query groups.

## Boundary

```text
Frontend -> API -> Database
```

The API remains the layer through which separately maintained frontend applications communicate with backend data. AWS Lambda, API Gateway, Cognito, S3, and RDS are current adapters or hosting choices. They must not define the future application's core endpoint, authorization, storage, or database behavior.

## Migration status legend

The migration views on this page use this four-state legend. It isn't the same as the endpoint reference's legend in [README.md](README.md#legend), where anything without a deployed route is ⬜.

- ✅ **Implemented:** a corresponding active legacy route, handler, and required supporting behavior are identifiable for migration.
- 🟨 **Partial:** only a stub or supporting primitive exists, the route is inactive, or a deterministic wiring or query defect prevents the intended behavior.
- ⬜ **Not found:** no corresponding route, handler, or query implementation was found.
- ❓ **Ambiguous:** evidence conflicts, or a product decision is required before status can be resolved.

Access markers follow [README.md](README.md#legend). For inactive historical routes, the marker records the historical intent and never implies that an absent route is deployed.

**Portability:** "portable" describes reusable application behavior, not whether the current Lambda can be copied unchanged. AWS coupling normally means an adapter must be extracted; it doesn't make the behavior unusable.

## Migration summary

The counts are kept in two views that don't add up, because the historical list overlaps the active API and several routes changed shape.

### Active legacy API

| Status | Count |
| --- | ---: |
| ✅ Implemented | 14 |
| 🟨 Partial | 4 |
| ⬜ Not found | 0 |
| ❓ Ambiguous | 0 |
| **Active HTTP routes** | **18** |

The active access configuration is 12 🟢 public routes and 6 🔴 Cognito-JWT routes. There is no API-wide default authorizer.

The four active partial routes are:

1. `POST /clubs/{clubId}/events`: wired, but its event and description `INSERT` statements are invalid.
2. `DELETE /clubs/{clubId}/members/me`: its handler requires JWT claims, but its active API Gateway route omits the authorizer.
3. `GET /clubs/{clubId}/events/{eventId}/images`: its handler references a SQL file that doesn't exist.
4. `POST /clubs/{clubId}/events/{eventId}/images/confirm`: its handler closes its package-level database client at the end of every request, so warm invocations fail. The Phase 1 docs restructure moved this route from ✅ to 🟨; the earlier map counted 15 ✅ and 3 🟨.

The active development and production APIs register the same 18 method/path contracts. Three event-image `POST` paths also register `OPTIONS` against their Lambda integration; those preflight registrations aren't counted as application endpoints. See the [detailed legacy route matrix](../infrastructure/legacy/docs/api/README.md#active-route-matrix). The `GET /database/test` helper in `gateway/routes` isn't called by either stack, so it isn't one of the 18.

### Historical "Endpoints Revamp" reconciliation

| Status | Count |
| --- | ---: |
| ✅ Implemented at the historical route/intent | 8 |
| 🟨 Partial, moved, stubbed, or incomplete | 16 |
| ⬜ No executable implementation found | 11 |
| ❓ Ambiguous | 0 |
| **Historical items** | **35** |

Two blue (internal) functions are documented: the current Cognito student synchronization (✅) and the historical generic image metadata write (🟨). They aren't active HTTP routes.

## Evidence and interpretation rules

The inventory was checked against:

- active route registration in [`gateway/routes/`](../infrastructure/legacy/gateway/routes/);
- both API stacks in [`productionApi.go`](../infrastructure/legacy/internal/stack/productionApi.go) and [`developmentApi.go`](../infrastructure/legacy/internal/stack/developmentApi.go);
- Lambda handlers under [`lambda/api/`](../infrastructure/legacy/lambda/api/);
- embedded SQL under [`utils/query_client/queries/`](../infrastructure/legacy/utils/query_client/queries/);
- authentication and authorization helpers under [`utils/auth/`](../infrastructure/legacy/utils/auth/);
- image/S3 integrations and Cognito trigger wiring; and
- the inactive stub stack and stub handlers, which count only as partial evidence.

The planning document, `GWC Website Documentation.pdf`, supplied the route grouping, intent, and the green/red/blue access model. The PDF isn't tracked in the repository. Its checkmarks, assignments, dates, and project-management notes were not used as implementation evidence.

Page numbers differ between exports of the PDF. In the 52-page copy used for the Phase 1 docs restructure, the relevant sections are the original endpoint list (pp. 9–13), the "Endpoints Revamp" list (pp. 15–19), and the frontend image-upload flow (pp. 25–26). The earlier version of this map cited pp. 14–18, 20–25, and 33–35 from a locally stashed pre-refactor copy. The 52-page copy has no "Endpoints Revamp" heading; the revamp list is the second endpoint list, the one that adds `/auth/events` and marks `GET /clubs/{clubId}` as new.

## Planned API directory structure

The following tree is a **representative planned skeleton only**. These directories don't exist yet and should be created incrementally as behavior is migrated and tested. Some leaf operations are collapsed for readability; the [per-endpoint targets](#per-endpoint-migration-targets) below are the complete mapping.

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
| Separate event description and club subresources | The active `GET /events/{eventId}` response already joins the description and associated-club IDs. The historical subresource routes are inactive stubs. |
| `/auth/events/{eventId}` revamp | No `/auth/events` route is active. Some image operations remain on older club-scoped paths and are currently public. |
| Generic `POST /images` | Event-gallery metadata is instead written by the active `POST .../images/confirm` route. No generic image route is active. |
| Club/event thumbnail confirmations | Presign handlers exist, but neither confirmation/metadata-assignment route exists. |
| Original club-scoped event detail/update/delete | The PDF later proposed `/auth/events`; neither protected form is active. Only club-scoped media paths remain. |

### Historical endpoint checklist

This compact view keeps every item in the historical "Endpoints Revamp", with statuses in the [migration legend](#migration-status-legend). The endpoint reference documents each item under its current path; for example, every 🟨 item below whose only evidence is a stub or a primitive at another path is ⬜ there.

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

## Per-endpoint migration targets

Every target below is **planned**. The SQL for query groups 1–15 has already been copied into the matching `database/queries/` paths in the database module's Phase 1; "planned" here refers to moving the callers, as explained in the [database migration map](../database/README.md#migration-summary).

### Internal

| Function | Target API location | Target query location | Portable |
| --- | --- | --- | --- |
| [🔵 Cognito student sync](endpoints/internal.md#-cognito-student-sync-trigger) | `api/internal/students/sync/` | `database/queries/students/ensure/` | Yes, after extracting Cognito event parsing and the AWS credential/connection adapter. |
| [🔵 Image metadata write (historical `POST /images`)](endpoints/internal.md#-image-metadata-write-pdf-post-images) | A shared internal image-metadata service under `api/internal/images/confirm/`, called by route-specific confirmation handlers. | The shared image insert, inside the route-appropriate confirmation transaction. | Partial; the insert primitive exists, but the generic function/contract doesn't. |

### Me

| Endpoint | Target API location | Target query location | Portable |
| --- | --- | --- | --- |
| [`GET /me`](endpoints/me.md#-get-me) | `api/me/get/` | `database/queries/me/get/` | Yes, after extracting API Gateway/JWT parsing and database connection creation. |
| [`GET /me/clubs`](endpoints/me.md#-get-meclubs) | `api/me/clubs/get/` | `database/queries/me/clubs/list/` | Yes, after extracting HTTP/JWT and database adapters. |
| [`GET /me/events`](endpoints/me.md#-get-meevents) | `api/me/events/get/` | `database/queries/me/events/list/` | Yes, after extracting HTTP/JWT and database adapters. |
| Historical `GET /me/clubs/events` | Reuse `api/me/events/get/`; retain the old path only if an explicit compatibility requirement is approved. | `database/queries/me/events/list/` | Partial; the real replacement is portable, but the historical route itself isn't active. |
| [`GET /me/clubs/eboard`](endpoints/me.md#-get-meclubseboard) | `api/me/clubs/eboard/get/` | `database/queries/me/clubs/eboard/list/`, or reuse `database/queries/me/clubs/list/` with an explicit role filter. | Partial. |

### Clubs and memberships

| Endpoint | Target API location | Target query location | Portable |
| --- | --- | --- | --- |
| [`GET /clubs`](endpoints/clubs.md#-get-clubs) | `api/clubs/list/` | `database/queries/clubs/list/` | Yes, after extracting the HTTP and database adapters. |
| [`POST /clubs`](endpoints/clubs.md#-post-clubs) | `api/clubs/create/` | `database/queries/clubs/create/` | Yes, after extracting HTTP/JWT and database adapters. |
| [`GET /clubs/{clubId}`](endpoints/clubs.md#-get-clubsclubid) | `api/clubs/{clubId}/get/` | `database/queries/clubs/get/` | Yes, after extracting the HTTP and database adapters. |
| [`POST /clubs/{clubId}/members/me`](endpoints/memberships.md#-post-clubsclubidmembersme) | `api/clubs/{clubId}/members/join/` | `database/queries/clubs/members/create/`, or a caller-specific wrapper under `join/`. | Yes, after extracting HTTP/JWT and database adapters. |
| [`DELETE /clubs/{clubId}/members/me`](endpoints/memberships.md#-delete-clubsclubidmembersme) | `api/clubs/{clubId}/members/leave/` | `database/queries/clubs/members/leave/` | Partial; the handler/query logic is reusable after correcting the transport authorization boundary. |
| [`GET /clubs/{clubId}/members`](endpoints/memberships.md#-get-clubsclubidmembers) | `api/clubs/{clubId}/members/list/` | `database/queries/clubs/members/list/` | N/A until implemented. |
| [`GET /clubs/{clubId}/eboard`](endpoints/memberships.md#-get-clubsclubideboard) | `api/clubs/{clubId}/members/list/` with a role filter, or `api/clubs/{clubId}/eboard/get/` if a distinct contract is retained. | Prefer the shared `database/queries/clubs/members/list/`. | N/A until implemented. |
| [`PUT /clubs/{clubId}/members/roles`](endpoints/memberships.md#-put-clubsclubidmembersroles) | `api/clubs/{clubId}/members/roles/update/` | `database/queries/clubs/members/update_role/` | N/A until implemented. |
| [`POST /clubs/{clubId}/verification`](endpoints/verification.md#-post-clubsclubidverification) | `api/clubs/{clubId}/verification/create/` | `database/queries/clubs/verification/create/` | N/A until implemented. |
| [`DELETE /clubs/{clubId}/verification`](endpoints/verification.md#-delete-clubsclubidverification) | `api/clubs/{clubId}/verification/delete/` | `database/queries/clubs/verification/delete/` | N/A until implemented. |

### Events

| Endpoint | Target API location | Target query location | Portable |
| --- | --- | --- | --- |
| [`GET /clubs/{clubId}/events`](endpoints/club-events.md#-get-clubsclubidevents) | `api/clubs/{clubId}/events/list/` | `database/queries/clubs/events/list/` | Yes, after extracting the HTTP and database adapters. |
| [`POST /clubs/{clubId}/events`](endpoints/club-events.md#-post-clubsclubidevents) | `api/clubs/{clubId}/events/create/` | `database/queries/events/create/`, shared across route adapters. | Partial; validation and orchestration are reusable after SQL repair, one transaction boundary, and authorization extraction. |
| [`GET /clubs/{clubId}/events/drafts`](endpoints/club-events.md#-get-clubsclubideventsdrafts) | `api/clubs/{clubId}/events/drafts/get/`, or an authenticated filter on the shared list service. | Reuse `database/queries/clubs/events/list/` with an explicit authorized status policy. | N/A as an endpoint; the underlying read query is portable. |
| [`GET /events`](endpoints/events.md#-get-events) | `api/events/list/` | `database/queries/events/read/` with a list wrapper. | Yes, after extracting the HTTP and database adapters. |
| [`GET /events/{eventId}`](endpoints/events.md#-get-eventseventid) | `api/events/{eventId}/get/` | Reuse `database/queries/events/read/` with a single-event wrapper. | Yes, after extracting the HTTP and database adapters. |
| Historical `GET /events/{eventId}/description` and `/clubs` | No separate endpoint is planned; use `api/events/{eventId}/get/` unless compatibility requires a subresource adapter. | Reuse `database/queries/events/read/`. | Partial as historical routes; the combined behavior is portable. |
| [`GET /events/{eventId}/images`](endpoints/events.md#-get-eventseventidimages) | `api/events/{eventId}/images/get/` | `database/queries/events/images/list/` | Partial; requires a query, a posted-visibility policy, and a storage-provider abstraction. |
| [`GET /auth/events/{eventId}`](endpoints/event-management.md#-get-autheventseventid) | `api/auth/events/{eventId}/get/` | Reuse `database/queries/events/read/` and `database/queries/authorization/events/can_manage/`. | Partial; the read and policy primitives exist but aren't composed. |
| [`GET /auth/events/{eventId}/images`](endpoints/event-management.md#-get-autheventseventidimages) | `api/auth/events/{eventId}/images/get/` | `database/queries/events/images/list/` plus the shared authorization query. | Partial; requires missing SQL, policy composition, and a storage adapter. |
| [`PATCH /auth/events/{eventId}`](endpoints/event-management.md#-patch-autheventseventid) | `api/auth/events/{eventId}/update/` | `database/queries/events/update/` | N/A until implemented. |
| [`POST /auth/events/{eventId}/thumbnails`](endpoints/event-management.md#-post-autheventseventidthumbnails) | `api/auth/events/{eventId}/thumbnails/presign/` | N/A for signing; confirmation would use `database/queries/events/thumbnails/confirm/`. | Partial; the signer is reusable behind a storage-provider abstraction, but authorization and confirmation are missing. |
| [`POST /auth/events/{eventId}/images`](endpoints/event-management.md#-post-autheventseventidimages) | `api/auth/events/{eventId}/images/presign/` and `confirm/` | `database/queries/events/images/confirm/` for confirmation. | Partial; the signer/confirmation behavior is reusable after policy, storage, and transaction extraction. |
| [`DELETE /auth/events/{eventId}`](endpoints/event-management.md#-delete-autheventseventid) | `api/auth/events/{eventId}/delete/` | `database/queries/events/delete/` | N/A until implemented. |

### Images

| Endpoint | Target API location | Target query location | Portable |
| --- | --- | --- | --- |
| [`POST /clubs/{clubId}/thumbnails`](endpoints/images.md#-post-clubsclubidthumbnails) | `api/clubs/{clubId}/thumbnails/presign/` | N/A for presigning; confirmation would use `database/queries/clubs/thumbnails/confirm/`. | Yes, after introducing a storage-provider abstraction. |
| [`POST /clubs/{clubId}/thumbnails/confirm`](endpoints/images.md#-post-clubsclubidthumbnailsconfirm) | `api/clubs/{clubId}/thumbnails/confirm/` | `database/queries/clubs/thumbnails/confirm/` | N/A until implemented. |
| [`POST /clubs/{clubId}/events/{eventId}/thumbnails`](endpoints/images.md#-post-clubsclubideventseventidthumbnails) | `api/auth/events/{eventId}/thumbnails/presign/`, or a retained club-scoped adapter. | N/A for presigning; confirmation would use `database/queries/events/thumbnails/confirm/`. | Yes, after introducing storage and authorization adapters. |
| [`POST /clubs/{clubId}/events/{eventId}/thumbnails/confirm`](endpoints/images.md#-post-clubsclubideventseventidthumbnailsconfirm) | `api/auth/events/{eventId}/thumbnails/confirm/` | `database/queries/events/thumbnails/confirm/` | N/A until implemented. |
| [`POST /clubs/{clubId}/events/{eventId}/images`](endpoints/images.md#-post-clubsclubideventseventidimages) | `api/auth/events/{eventId}/images/presign/`, or a retained club-scoped adapter. | N/A for presigning. | Yes, after introducing a storage-provider abstraction and authorization policy. |
| [`POST /clubs/{clubId}/events/{eventId}/images/confirm`](endpoints/images.md#-post-clubsclubideventseventidimagesconfirm) | `api/auth/events/{eventId}/images/confirm/`, or a retained club-scoped adapter. | `database/queries/events/images/confirm/` | Yes, after extracting the database/storage adapters and making confirmation transactional. |
| [`GET /clubs/{clubId}/events/{eventId}/images`](endpoints/images.md#-get-clubsclubideventseventidimages) | Prefer public `api/events/{eventId}/images/get/` for posted events and authenticated `api/auth/events/{eventId}/images/get/` for drafts. | The shared `database/queries/events/images/list/`. | Partial; the response/signing behavior is reusable after adding SQL, a visibility policy, and a storage-provider abstraction. |
| [`DELETE /images/{imageId}`](endpoints/images.md#-delete-imagesimageid) | `api/images/{imageId}/delete/` | `database/queries/images/delete/` plus a storage deletion adapter. | N/A until implemented. |

### Admins

| Endpoint | Target API location | Target query location | Portable |
| --- | --- | --- | --- |
| [`GET /admins`](endpoints/admins.md#-get-admins) | `api/admins/list/` | `database/queries/admins/list/` and a shared `is_admin` query. | N/A until implemented. |
| [`GET /admins/{studentId}`](endpoints/admins.md#-get-adminsstudentid) | `api/admins/{studentId}/get/` | `database/queries/admins/get/` | Partial; the schema intent is clear, but the stub isn't parameterized migration-ready code. |
| [`POST /admins`](endpoints/admins.md#-post-admins) | `api/admins/create/` | `database/queries/admins/create/` | N/A until implemented. |
| [`DELETE /admins/{studentId}`](endpoints/admins.md#-delete-adminsstudentid) | `api/admins/{studentId}/delete/` | `database/queries/admins/delete/` | Partial; only obsolete SQL-shaped evidence exists. |

### Operational

| Endpoint | Target API location | Target query location | Portable |
| --- | --- | --- | --- |
| [`GET /health`](endpoints/operational.md#-get-health) | `api/operations/health/get/` | N/A | Yes, after extracting the Lambda/API Gateway adapter. |
| [`GET /database/test`](endpoints/operational.md#-get-databasetest) | No public business route is planned. Move equivalent coverage to protected operational diagnostics or integration tests. | N/A | Partial as a connectivity test; it shouldn't be migrated as a public API contract. |

### Parity checks before porting

These current behaviors need contract tests before migration, so each one is either preserved or deliberately changed:

- The public event lists' strict date bounds, joined-row pagination, unordered map grouping, partial club associations, and fixed placeholder thumbnails ([event object](endpoints/events.md#event-object)).
- `GET /events/{eventId}` requires only a non-empty ID, matches it with `LIKE`, and caps joined rows at 100.
- The current event-image handlers trust client-provided metadata and don't verify the S3 object before inserting into the database.
- The club-scoped image handlers ignore `{clubId}`; don't duplicate them just to support the `/auth/events` routes.
- Don't treat hard-coded stub responses (for example, the `GET /events/{eventId}/images` stub) as implementation contracts.

## Shared authentication and authorization dependencies

Current behavior of each helper is in [endpoints/internal.md](endpoints/internal.md).

| Capability | Legacy source | Current use | Planned responsibility |
| --- | --- | --- | --- |
| Extract verified `sub` and optional email | [`extract_sub.go`](../infrastructure/legacy/utils/auth/extract_sub.go) | Club/event creation and membership handlers | HTTP/identity adapter plus application identity context |
| Ensure a student row exists | [`ensure_student.go`](../infrastructure/legacy/utils/auth/ensure_student.go) | All six JWT-attached handlers | Shared student synchronization service |
| Check club e-board/owner role | [`club_authorization.go`](../infrastructure/legacy/utils/auth/club_authorization.go) and duplicate [`lambda/internal` helper](../infrastructure/legacy/lambda/internal/auth/club_authorization/club_authorization.go) | **No active handler calls it** | Shared club authorization policy using [query group 13](../database/README.md#13-club-authorization) |
| Check event-associated club role | [`event_authorization.go`](../infrastructure/legacy/utils/auth/event_authorization.go) and duplicate [`lambda/internal` helper](../infrastructure/legacy/lambda/internal/auth/event_authorization/event_authorization.go) | **No active handler calls it** | Shared event authorization policy using [query group 14](../database/README.md#14-event-authorization) |
| Format validation errors | [`validation_error.go`](../infrastructure/legacy/utils/errors/validation_error.go) | Club and event create handlers | Shared transport-neutral validation result mapping |

A Cognito JWT establishes identity only. It doesn't establish admin, membership, e-board, owner, or event authority. The current enforcement gaps must be preserved as known migration work, not silently presented as implemented authorization.

## Recommended migration order

1. Define provider-neutral request, response, identity, database, and storage interfaces while snapshot-testing current response contracts.
2. Port shared identity extraction, student synchronization, validation, response mapping, and authorization policies once; remove duplicate role-helper implementations during that later migration.
3. Port the implemented read routes and their query groups before write workflows.
4. Port club membership and club creation with explicit decisions for route authorization and creator ownership.
5. Repair and transaction-test event creation before exposing it from the new module.
6. Introduce the storage-provider abstraction, then migrate presign, confirmation, and gallery-read behavior with authorization and lifecycle tests.
7. Implement historically planned routes only after their product/authorization decisions are approved; don't create empty endpoint folders solely to mirror this document.
8. Switch API Gateway/Lambda wiring only after parity tests pass. Keep `infrastructure/legacy/` as the working reference until an independently reviewed cutover.

The product and design decisions that block these steps are listed under [Open questions](README.md#open-questions).

## Current versus planned state

- [`api/README.md`](README.md): endpoint reference index.
- [`api/endpoints/`](endpoints/): per-resource endpoint reference.
- `api/MIGRATION.md`: this migration map.
- `api/`: no migrated handler implementation yet.
- [`infrastructure/legacy/lambda/api/`](../infrastructure/legacy/lambda/api/): current handler source.
- [`infrastructure/legacy/gateway/`](../infrastructure/legacy/gateway/): current API Gateway route/integration source.
- [`infrastructure/legacy/utils/auth/`](../infrastructure/legacy/utils/auth/): current authentication/authorization helper source.
- [`database/README.md`](../database/README.md): linked query and schema migration map.

The docs restructure changed documentation only. It doesn't move handlers, alter route or auth behavior, change response formats, or deploy anything.
