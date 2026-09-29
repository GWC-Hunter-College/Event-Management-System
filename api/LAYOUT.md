# API module layout

The layout of the provider-independent `api/` module: its directory tree, the module path for each endpoint, and the shared authentication and authorization pieces those endpoints depend on. The module has no code yet. Every endpoint runs from the Go Lambda handlers in [`infrastructure/legacy/`](../infrastructure/legacy/), which are the reference for behavior. Endpoint behavior is documented in [README.md](README.md), the planning PDF is compared with the code in [pdf-coverage.md](pdf-coverage.md), and the schema and query groups are in [database/README.md](../database/README.md).

## Boundary

```text
Frontend -> API -> Database
```

The API is the layer through which separately maintained frontends reach backend data. AWS Lambda, API Gateway, Cognito, S3, and RDS are the current adapters and hosting choices. In this layout they sit behind `adapters/`, so the module's endpoint, authorization, storage, and database behavior doesn't depend on them.

**Reusable**, the last column in the tables below, describes application behavior that can carry over from the legacy handler, not whether a Lambda can be copied unchanged. AWS coupling usually means extracting an adapter; it doesn't make the behavior unusable.

## Directory layout

This tree is a **representative skeleton**. None of these directories exist: each one is created with the first behavior it holds, once that behavior is built and tested. Some leaf operations are collapsed for readability; [Endpoint locations](#endpoint-locations) is the complete mapping.

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
│       ├── update/
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
│           ├── restore/
│           ├── images/
│           └── thumbnails/
├── admins/
│   ├── list/
│   ├── create/
│   ├── purge/
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

Route-shaped folders are navigation aids, not a reason to duplicate shared authorization, query, response, or storage logic. The matching query tree is in the [database reference](../database/README.md#directory-layout).

## Endpoint locations

None of the module paths below exist yet. The SQL for query groups 1–15 is already at its path in `database/queries/`, but no module code calls it (see the [database summary](../database/README.md#query-group-summary)).

### Internal

| Function | Module path | Queries | Reusable |
| --- | --- | --- | --- |
| [🔵 Cognito student sync](endpoints/internal.md#-cognito-student-sync-trigger) | `api/internal/students/sync/` | `database/queries/students/ensure/` | Yes, after extracting Cognito event parsing and the AWS credential/connection adapter. |
| [🔵 Image metadata write (PDF `POST /images`)](endpoints/internal.md#-image-metadata-write-pdf-post-images) | A shared internal image-metadata service under `api/internal/images/confirm/`, called by each route-specific confirm handler. | The shared image insert, inside the route-appropriate confirm transaction. | Partial; the insert primitive exists, but the generic function and its contract don't. |

### Me

| Endpoint | Module path | Queries | Reusable |
| --- | --- | --- | --- |
| [`GET /me`](endpoints/me.md#-get-me) | `api/me/get/` | `database/queries/me/get/` | Yes, after extracting API Gateway/JWT parsing and database connection creation. |
| [`GET /me/clubs`](endpoints/me.md#-get-meclubs) | `api/me/clubs/get/` | `database/queries/me/clubs/list/` | Yes, after extracting the HTTP/JWT and database adapters. |
| [`GET /me/events`](endpoints/me.md#-get-meevents) | `api/me/events/get/` | `database/queries/me/events/list/` | Yes, after extracting the HTTP/JWT and database adapters. |
| PDF path `GET /me/clubs/events` | Served by `api/me/events/get/`; the PDF path is added as an alias only if a client needs it. | `database/queries/me/events/list/` | Partial; `GET /me/events` is reusable, and the PDF path itself isn't built. |
| [`GET /me/clubs/eboard`](endpoints/me.md#-get-meclubseboard) | `api/me/clubs/eboard/get/` | `database/queries/me/clubs/eboard/list/`, or `database/queries/me/clubs/list/` with an explicit role filter. | Partial. |

### Clubs and memberships

| Endpoint | Module path | Queries | Reusable |
| --- | --- | --- | --- |
| [`GET /clubs`](endpoints/clubs.md#-get-clubs) | `api/clubs/list/` | `database/queries/clubs/list/` | Yes, after extracting the HTTP and database adapters. |
| [`POST /clubs`](endpoints/clubs.md#-post-clubs) | `api/clubs/create/` | `database/queries/clubs/create/` | Yes, after extracting the HTTP/JWT and database adapters. |
| [`GET /clubs/{clubId}`](endpoints/clubs.md#-get-clubsclubid) | `api/clubs/{clubId}/get/` | `database/queries/clubs/get/` | Yes, after extracting the HTTP and database adapters. |
| [`PATCH /clubs/{clubId}`](endpoints/clubs.md#-patch-clubsclubid) | `api/clubs/{clubId}/update/` | `database/queries/clubs/update/` and `database/queries/clubs/tags/replace/` ([query group 30](../database/README.md#30-club-update)). | N/A until built. |
| [`POST /clubs/{clubId}/members/me`](endpoints/memberships.md#-post-clubsclubidmembersme) | `api/clubs/{clubId}/members/join/` | `database/queries/clubs/members/create/`, or a caller-specific wrapper under `join/`. | Yes, after extracting the HTTP/JWT and database adapters. |
| [`DELETE /clubs/{clubId}/members/me`](endpoints/memberships.md#-delete-clubsclubidmembersme) | `api/clubs/{clubId}/members/leave/` | `database/queries/clubs/members/leave/` | Partial; the handler and query logic are reusable once the route's authorizer is fixed. |
| [`GET /clubs/{clubId}/members`](endpoints/memberships.md#-get-clubsclubidmembers) | `api/clubs/{clubId}/members/list/` | `database/queries/clubs/members/list/` | N/A until built. |
| [`GET /clubs/{clubId}/eboard`](endpoints/memberships.md#-get-clubsclubideboard) | `api/clubs/{clubId}/members/list/` with a role filter, or `api/clubs/{clubId}/eboard/get/` if it keeps a distinct contract. | Prefer the shared `database/queries/clubs/members/list/`. | N/A until built. |
| [`PUT /clubs/{clubId}/members/roles`](endpoints/memberships.md#-put-clubsclubidmembersroles) | `api/clubs/{clubId}/members/roles/update/` | `database/queries/clubs/members/update_role/` | N/A until built. |
| [`POST /clubs/{clubId}/verification`](endpoints/verification.md#-post-clubsclubidverification) | `api/clubs/{clubId}/verification/create/` | `database/queries/clubs/verification/create/` | N/A until built. |
| [`DELETE /clubs/{clubId}/verification`](endpoints/verification.md#-delete-clubsclubidverification) | `api/clubs/{clubId}/verification/delete/` | `database/queries/clubs/verification/delete/` | N/A until built. |

### Events

| Endpoint | Module path | Queries | Reusable |
| --- | --- | --- | --- |
| [`GET /clubs/{clubId}/events`](endpoints/club-events.md#-get-clubsclubidevents) | `api/clubs/{clubId}/events/list/` | `database/queries/clubs/events/list/` | Yes, after extracting the HTTP and database adapters. |
| [`POST /clubs/{clubId}/events`](endpoints/club-events.md#-post-clubsclubidevents) | `api/clubs/{clubId}/events/create/` | `database/queries/events/create/`, shared across route adapters. | Partial; validation and orchestration are reusable after the SQL repair, a single transaction, and extracting authorization. |
| [`GET /clubs/{clubId}/events/drafts`](endpoints/club-events.md#-get-clubsclubideventsdrafts) | `api/clubs/{clubId}/events/drafts/get/`, or an authenticated filter on the shared list service. | `database/queries/clubs/events/drafts/`, which fixes the status to `draft` ([query group 21](../database/README.md#21-club-draft-events)). | N/A as an endpoint; the underlying read query is reusable. |
| [`GET /events`](endpoints/events.md#-get-events) | `api/events/list/` | `database/queries/events/read/` with a list wrapper. | Yes, after extracting the HTTP and database adapters. |
| [`GET /events/{eventId}`](endpoints/events.md#-get-eventseventid) | `api/events/{eventId}/get/` | `database/queries/events/read/` with a single-event wrapper. | Yes, after extracting the HTTP and database adapters. |
| PDF paths `GET /events/{eventId}/description` and `/clubs` | No separate endpoint; served by `api/events/{eventId}/get/`, with a subresource adapter only if a client needs one. | `database/queries/events/read/` | Partial as separate routes; the combined behavior is reusable. |
| [`GET /events/{eventId}/images`](endpoints/events.md#-get-eventseventidimages) | `api/events/{eventId}/images/get/` | `database/queries/events/images/list/` | Partial; needs a query, a posted-visibility policy, and a storage-provider abstraction. |
| [`GET /auth/events/{eventId}`](endpoints/event-management.md#-get-autheventseventid) | `api/auth/events/{eventId}/get/` | `database/queries/events/read/` and `database/queries/authorization/events/can_manage/`. | Partial; the read and policy primitives exist but aren't combined. |
| [`GET /auth/events/{eventId}/images`](endpoints/event-management.md#-get-autheventseventidimages) | `api/auth/events/{eventId}/images/get/` | `database/queries/events/images/list/` plus the shared authorization query. | Partial; needs the missing SQL, the policy, and a storage adapter. |
| [`PATCH /auth/events/{eventId}`](endpoints/event-management.md#-patch-autheventseventid) | `api/auth/events/{eventId}/update/` | `database/queries/events/update/` | N/A until built. |
| [`POST /auth/events/{eventId}/thumbnails`](endpoints/event-management.md#-post-autheventseventidthumbnails) | `api/auth/events/{eventId}/thumbnails/presign/` | N/A for signing; the confirm step would use `database/queries/events/thumbnails/confirm/`. | Partial; the signer is reusable behind a storage-provider abstraction, but authorization and confirmation are missing. |
| [`POST /auth/events/{eventId}/images`](endpoints/event-management.md#-post-autheventseventidimages) | `api/auth/events/{eventId}/images/presign/` and `confirm/` | `database/queries/events/images/confirm/` for the confirm step. | Partial; the signer and confirm behavior are reusable after adding the policy, storage adapter, and transaction. |
| [`DELETE /auth/events/{eventId}`](endpoints/event-management.md#-delete-autheventseventid) | `api/auth/events/{eventId}/delete/` | `database/queries/events/delete/` | N/A until built. |
| [`POST /auth/events/{eventId}/restore`](endpoints/event-management.md#-post-autheventseventidrestore) | `api/auth/events/{eventId}/restore/` | `database/queries/events/restore/` and `database/queries/authorization/events/manages_owner_club/`. | N/A until built. |

### Images

| Endpoint | Module path | Queries | Reusable |
| --- | --- | --- | --- |
| [`POST /clubs/{clubId}/thumbnails`](endpoints/images.md#-post-clubsclubidthumbnails) | `api/clubs/{clubId}/thumbnails/presign/` | N/A for presigning; the confirm step would use `database/queries/clubs/thumbnails/confirm/`. | Yes, after introducing a storage-provider abstraction. |
| [`POST /clubs/{clubId}/thumbnails/confirm`](endpoints/images.md#-post-clubsclubidthumbnailsconfirm) | `api/clubs/{clubId}/thumbnails/confirm/` | `database/queries/clubs/thumbnails/confirm/` | N/A until built. |
| [`POST /clubs/{clubId}/events/{eventId}/thumbnails`](endpoints/images.md#-post-clubsclubideventseventidthumbnails) | `api/auth/events/{eventId}/thumbnails/presign/`, or a club-scoped adapter if that path stays. | N/A for presigning; the confirm step would use `database/queries/events/thumbnails/confirm/`. | Yes, after introducing storage and authorization adapters. |
| [`POST /clubs/{clubId}/events/{eventId}/thumbnails/confirm`](endpoints/images.md#-post-clubsclubideventseventidthumbnailsconfirm) | `api/auth/events/{eventId}/thumbnails/confirm/` | `database/queries/events/thumbnails/confirm/` | N/A until built. |
| [`POST /clubs/{clubId}/events/{eventId}/images`](endpoints/images.md#-post-clubsclubideventseventidimages) | `api/auth/events/{eventId}/images/presign/`, or a club-scoped adapter if that path stays. | N/A for presigning. | Yes, after introducing a storage-provider abstraction and an authorization policy. |
| [`POST /clubs/{clubId}/events/{eventId}/images/confirm`](endpoints/images.md#-post-clubsclubideventseventidimagesconfirm) | `api/auth/events/{eventId}/images/confirm/`, or a club-scoped adapter if that path stays. | `database/queries/events/images/confirm/` | Yes, after extracting the database and storage adapters and making the confirm step one transaction. |
| [`GET /clubs/{clubId}/events/{eventId}/images`](endpoints/images.md#-get-clubsclubideventseventidimages) | Public `api/events/{eventId}/images/get/` for posted events, and authenticated `api/auth/events/{eventId}/images/get/` for drafts. | The shared `database/queries/events/images/list/`. | Partial; the response and signing behavior are reusable after adding the SQL, a visibility policy, and a storage-provider abstraction. |
| [`DELETE /images/{imageId}`](endpoints/images.md#-delete-imagesimageid) | `api/images/{imageId}/delete/` | `database/queries/images/delete/` plus a storage deletion adapter. | N/A until built. |

### Admins

| Endpoint | Module path | Queries | Reusable |
| --- | --- | --- | --- |
| [`GET /admins`](endpoints/admins.md#-get-admins) | `api/admins/list/` | `database/queries/admins/list/` and a shared `is_admin` query. | N/A until built. |
| [`GET /admins/{studentId}`](endpoints/admins.md#-get-adminsstudentid) | `api/admins/{studentId}/get/` | `database/queries/admins/get/` | Partial; the schema intent is clear, but the stub isn't parameterized, working SQL. |
| [`POST /admins`](endpoints/admins.md#-post-admins) | `api/admins/create/` | `database/queries/admins/create/` | N/A until built. |
| [`DELETE /admins/{studentId}`](endpoints/admins.md#-delete-adminsstudentid) | `api/admins/{studentId}/delete/` | `database/queries/admins/delete/` | Partial; only obsolete SQL-shaped notes exist. |
| [`POST /admins/purge`](endpoints/admins.md#-post-adminspurge) | `api/admins/purge/` | `database/queries/events/purge/` and `database/queries/images/purge/`, plus the storage adapter for S3 deletes. | N/A until built. |

### Operational

| Endpoint | Module path | Queries | Reusable |
| --- | --- | --- | --- |
| [`GET /health`](endpoints/operational.md#-get-health) | `api/operations/health/get/` | N/A | Yes, after extracting the Lambda/API Gateway adapter. |
| [`GET /database/test`](endpoints/operational.md#-get-databasetest) | Not a public route in the design; equivalent coverage belongs in protected diagnostics or integration tests. | N/A | Partial as a connectivity test; it isn't part of the public API contract. |

### Behaviors without contract tests

These legacy behaviors have no contract tests. Each needs one before its `api/` version replaces the legacy handler, so that the behavior is kept or changed on purpose:

- The public event lists' strict date bounds, joined-row pagination, unordered map grouping, partial club associations, and fixed placeholder thumbnails ([event object](endpoints/events.md#event-object)).
- `GET /events/{eventId}` requires only a non-empty ID, matches it with `LIKE`, and caps joined rows at 100.
- The current event-image handlers trust client-provided metadata and don't verify the S3 object before inserting into the database.
- The club-scoped image handlers ignore `{clubId}`; that behavior isn't carried into the `/auth/events` routes.
- Hard-coded stub responses (for example, the `GET /events/{eventId}/images` stub) aren't contracts.

## Shared authentication and authorization dependencies

Current behavior of each helper is in [endpoints/internal.md](endpoints/internal.md).

| Capability | Current source | Current use | Module home |
| --- | --- | --- | --- |
| Extract verified `sub` and optional email | [`extract_sub.go`](../infrastructure/legacy/utils/auth/extract_sub.go) | Club and event creation, and the membership handlers | HTTP/identity adapter plus an application identity context |
| Ensure a student row exists | [`ensure_student.go`](../infrastructure/legacy/utils/auth/ensure_student.go) | All six JWT-attached handlers | Shared student sync service |
| Check club e-board/owner role | [`club_authorization.go`](../infrastructure/legacy/utils/auth/club_authorization.go) and a duplicate [`lambda/internal` helper](../infrastructure/legacy/lambda/internal/auth/club_authorization/club_authorization.go) | **No deployed handler calls it** | Shared club authorization policy using [query group 13](../database/README.md#13-club-authorization) |
| Check event-associated club role | [`event_authorization.go`](../infrastructure/legacy/utils/auth/event_authorization.go) and a duplicate [`lambda/internal` helper](../infrastructure/legacy/lambda/internal/auth/event_authorization/event_authorization.go) | **No deployed handler calls it** | Shared event authorization policy using [query group 14](../database/README.md#14-event-authorization) |
| Format validation errors | [`validation_error.go`](../infrastructure/legacy/utils/errors/validation_error.go) | Club and event create handlers | Shared, transport-neutral validation result mapping |

A Cognito JWT establishes identity only. It doesn't establish admin, membership, e-board, owner, or event authority. No handler calls a role check, so each endpoint lists its missing role check as a known issue.

## Build dependencies

The module's pieces depend on each other in this order. Each step needs the ones before it. None is started.

1. Provider-neutral request, response, identity, database, and storage interfaces, with snapshot tests of the current response contracts.
2. Shared identity extraction, student sync, validation, response mapping, and authorization policies, built once. The duplicate role helpers are dropped at this step.
3. The implemented read routes and their query groups, before any write workflow.
4. Club membership and club creation, which need decisions on route authorization and creator ownership.
5. Event creation, which needs its SQL repaired and a transaction test before the module exposes it.
6. The storage-provider abstraction, then presigning, confirmation, and gallery reads with authorization and lifecycle tests.
7. The planned (⬜) routes, each after its product and authorization decisions are approved. Endpoint folders aren't created just to mirror the docs.
8. Switching the API Gateway/Lambda wiring to the module, after the contract tests pass. `infrastructure/legacy/` stays the working reference until that switch is reviewed.

The product and design decisions that block these steps are listed under [Open questions](README.md#open-questions).

## Files

- [`api/README.md`](README.md): the endpoint index.
- [`api/endpoints/`](endpoints/): the per-resource endpoint reference.
- `api/LAYOUT.md`: this module layout.
- [`api/pdf-coverage.md`](pdf-coverage.md): the planning PDF compared with the code.
- [`api/coverage.md`](coverage.md): the frontend's screens mapped to endpoints.
- `api/`: no module code yet.
- [`infrastructure/legacy/lambda/api/`](../infrastructure/legacy/lambda/api/): the current handler source.
- [`infrastructure/legacy/gateway/`](../infrastructure/legacy/gateway/): the current API Gateway route and integration source.
- [`infrastructure/legacy/utils/auth/`](../infrastructure/legacy/utils/auth/): the current authentication and authorization helpers.
- [`database/README.md`](../database/README.md): schema and query groups.
