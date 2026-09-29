# API build plan

How the provider-independent `api/` module gets built. [README.md](README.md) is the blueprint: what each endpoint does and how the built ones behave. This file covers where each endpoint's code will live, what the current implementation in [`infrastructure/legacy/`](../infrastructure/legacy/) already provides to build from, and the order to build it in. That implementation stays the working reference for behavior until the module replaces it. Schema and query groups are in [database/README.md](../database/README.md).

## Boundary

```text
Frontend -> API -> Database
```

The API is the layer through which separately maintained frontends reach backend data. AWS Lambda, API Gateway, Cognito, S3, and RDS are current adapters and hosting choices. They must not define the module's endpoint, authorization, storage, or database behavior.

## Readiness legend

This file rates how ready each piece is to build from. That's a different scale from the blueprint's status legend in [README.md](README.md#legend), where anything without a deployed route is ⬜.

- ✅ **Ready:** a working route, handler, and supporting behavior exist to build from.
- 🟨 **Partial:** only a stub or supporting primitive exists, the route is inactive, or a defect blocks the intended behavior.
- ⬜ **Not started:** no route, handler, or query exists.
- ❓ **Needs a decision:** the evidence conflicts, or a product decision has to come first.

Access markers follow [README.md](README.md#legend). For a route that isn't built, the marker is the design's intent, not a claim that it's deployed.

**Reusable**, the last column in the tables below, describes application behavior that can carry over, not whether a Lambda can be copied unchanged. AWS coupling usually means extracting an adapter; it doesn't make the behavior unusable.

## Starting point

### Deployed routes

| Readiness | Count |
| --- | ---: |
| ✅ Ready | 14 |
| 🟨 Partial | 4 |
| ⬜ Not started | 0 |
| ❓ Needs a decision | 0 |
| **Deployed HTTP routes** | **18** |

That's 12 🟢 public routes and 6 🔴 Cognito-JWT routes. There's no API-wide default authorizer.

The four partial routes:

1. `POST /clubs/{clubId}/events`: wired, but its event and description `INSERT` statements are invalid.
2. `DELETE /clubs/{clubId}/members/me`: the handler requires JWT claims, but the route omits the authorizer.
3. `GET /clubs/{clubId}/events/{eventId}/images`: the handler references a SQL file that doesn't exist.
4. `POST /clubs/{clubId}/events/{eventId}/images/confirm`: the handler closes its package-level database client at the end of every request, so warm invocations fail.

The dev and prod APIs register the same 18 method/path contracts. Three event-image `POST` paths also register `OPTIONS` against their Lambda integration for CORS preflight; those aren't counted as endpoints. See the [detailed route matrix](../infrastructure/legacy/docs/api/README.md#active-route-matrix). The `GET /database/test` helper in `gateway/routes` isn't called by either stack, so it isn't one of the 18.

### Design coverage

The planning PDF's "Endpoints Revamp" lists 35 items. Their readiness:

| Readiness | Count |
| --- | ---: |
| ✅ Built at the designed route or intent | 8 |
| 🟨 Partial, at a different path, stubbed, or incomplete | 16 |
| ⬜ Nothing executable yet | 11 |
| ❓ Needs a decision | 0 |
| **Design items** | **35** |

Two of the items are internal (blue) functions, not HTTP routes: the Cognito student sync (✅) and the generic image metadata write (🟨).

## Sources

The readiness ratings were checked against:

- route registration in [`gateway/routes/`](../infrastructure/legacy/gateway/routes/);
- both API stacks in [`productionApi.go`](../infrastructure/legacy/internal/stack/productionApi.go) and [`developmentApi.go`](../infrastructure/legacy/internal/stack/developmentApi.go);
- Lambda handlers under [`lambda/api/`](../infrastructure/legacy/lambda/api/);
- embedded SQL under [`utils/query_client/queries/`](../infrastructure/legacy/utils/query_client/queries/);
- authentication and authorization helpers under [`utils/auth/`](../infrastructure/legacy/utils/auth/);
- image/S3 integrations and Cognito trigger wiring; and
- the inactive stub stack and stub handlers, which count only as partial evidence.

The planning document, `GWC Website Documentation.pdf`, supplies the route grouping, the design intent, and the green/red/blue access model. It isn't tracked in the repository. Its checkmarks, assignments, dates, and project-management notes aren't evidence that something is built.

In the 52-page copy of the PDF, the relevant sections are the first endpoint list (pp. 9–13), the "Endpoints Revamp" list (pp. 15–19), and the frontend image-upload flow (pp. 25–26). Page numbers vary between exports. In this one the revamp list has no heading of its own: it's the second endpoint list, the one that adds `/auth/events` and marks `GET /clubs/{clubId}` as new.

## Planned API directory structure

This tree is a **representative skeleton**. None of these directories exist yet; create each one when its first behavior is built and tested. Some leaf operations are collapsed for readability; [Where each endpoint will live](#where-each-endpoint-will-live) is the complete mapping.

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

Route-shaped folders are navigation aids, not a reason to duplicate shared authorization, query, response, or storage logic. The matching planned query tree is in the [database map](../database/README.md#planned-database-directory-structure).

## Design doc vs. code

| The planning PDF says | The code has |
| --- | --- |
| `GET /me/clubs/events` | `GET /me/events`. The PDF path exists only in the inactive stub stack. |
| `POST /clubs/{clubId}/members` | Self-join as `POST /clubs/{clubId}/members/me`. No endpoint adds a different student. |
| `POST /clubs/thumbnails` | A path that requires the club ID: `POST /clubs/{clubId}/thumbnails`. |
| `GET /clubs/verified=true` | A query parameter: `GET /clubs?verified=true`. |
| Separate event description and club subresources | `GET /events/{eventId}`, which already joins the description and associated-club IDs. The subresource paths are inactive stubs. |
| The `/auth/events/{eventId}` routes | No `/auth/events` route. Some of the image operations exist on club-scoped paths, and those are public. |
| A generic internal `POST /images` | Event-gallery metadata written by `POST .../images/confirm`. There's no generic image route. |
| Club and event thumbnail confirmations | Presign handlers, but neither confirmation (metadata-assignment) route. |
| Club-scoped event detail, update, and delete (first list) | Nothing yet. The revamp designs these under `/auth/events`, and only the club-scoped media paths are built. |

### Design checklist

Every item in the PDF's "Endpoints Revamp", rated on the [readiness legend](#readiness-legend). [README.md](README.md) documents each item under the path the code uses; there, an item whose only evidence is a stub or a primitive at another path is ⬜.

| Design route or function | Readiness | What exists |
| --- | --- | --- |
| Register Lambda | ✅ | The Cognito student upsert handler and query run as an internal trigger. |
| `GET /me/clubs` | ✅ | Built at this path (protected). |
| `GET /me/clubs/events` | 🟨 | The full behavior is built as `GET /me/events`; this exact path is an inactive stub. |
| `GET /me/clubs/eboard` | 🟨 | An inactive hard-coded stub and stale SQL only. |
| `POST /clubs` | ✅ | Built at this path (protected). |
| `POST /clubs/thumbnails` | 🟨 | Built as public `POST /clubs/{clubId}/thumbnails` with a JSON body. |
| `GET /clubs/{clubId}` | ✅ | Built at this path (public). |
| `GET /clubs/{clubId}/members` | ⬜ | Placeholder text only; no route, handler, or query. |
| `GET /clubs/{clubId}/eboard` | ⬜ | No route, handler, or list query. |
| `POST /clubs/{clubId}/members` | 🟨 | Built as caller-only `POST .../members/me`. |
| `PUT /clubs/{clubId}/members/roles` | ⬜ | No route, handler, owner-only check, or update query. |
| `DELETE /clubs/{clubId}/members/me` | 🟨 | The handler and query exist, but the route omits the JWT authorizer. |
| `GET /clubs/{clubId}/events` | ✅ | Built at this path (public). |
| `GET /clubs/{clubId}/events/drafts` | ⬜ | No route or handler; only reusable status-filtered SQL. |
| `POST /clubs/{clubId}/events` | 🟨 | The route exists, but its SQL is invalid and the club-role check is missing. |
| `GET /events` | ✅ | Built at this path (public, posted events). |
| `GET /events/{eventId}` | ✅ | Built at this path (public, posted events). |
| `GET /events/{eventId}/images` | 🟨 | This path is an inactive stub; the club-scoped read that does exist lacks its SQL. |
| `GET /events/{eventId}/description` | 🟨 | No dedicated route; the event detail includes the description. |
| `GET /events/{eventId}/clubs` | 🟨 | No dedicated route; the event detail includes the owner and associate IDs. |
| `GET /auth/events/{eventId}` | 🟨 | A public read and an unused authorization helper exist separately; no protected route combines them. |
| `GET /auth/events/{eventId}/images` | 🟨 | A public club-scoped handler exists but lacks its SQL and authorization. |
| `PATCH /auth/events/{eventId}` | ⬜ | No update handler or query. |
| `POST /auth/events/{eventId}/thumbnails` | 🟨 | A public club-scoped S3 signer exists; authorization and confirmation don't. |
| `POST /auth/events/{eventId}/images` | 🟨 | A public club-scoped S3 signer and confirmation exist; authorization doesn't. |
| `DELETE /auth/events/{eventId}` | ⬜ | No archive or delete handler or query. |
| `GET /admins` | ⬜ | Planning text and schema only. |
| `GET /admins/{studentId}` | 🟨 | An unwired SQL stub that hard-codes an obsolete numeric student ID. |
| `POST /admins` | ⬜ | No route, handler, insert query, or admin check. |
| `DELETE /admins/{studentId}` | 🟨 | An unwired SQL stub that hard-codes an obsolete numeric student ID. |
| `GET /clubs?verified=true` | ✅ | Built as the `verified` query parameter on `GET /clubs`. |
| `POST /clubs/{clubId}/verification` | ⬜ | Schema and read support only; no write. |
| `DELETE /clubs/{clubId}/verification` | ⬜ | Schema and read support only; no write. |
| Internal `POST /images` | 🟨 | No generic function; the metadata insert exists inside event-image confirmation. |
| `DELETE /images/{imageId}` | ⬜ | No database or storage deletion workflow. |

## Where each endpoint will live

Every location below is planned. The SQL for query groups 1–15 already sits at its target path in `database/queries/`; what still needs building is the code that calls it (see the [database map](../database/README.md#migration-summary)).

### Internal

| Function | Code | Queries | Reusable |
| --- | --- | --- | --- |
| [🔵 Cognito student sync](endpoints/internal.md#-cognito-student-sync-trigger) | `api/internal/students/sync/` | `database/queries/students/ensure/` | Yes, after extracting Cognito event parsing and the AWS credential/connection adapter. |
| [🔵 Image metadata write (PDF `POST /images`)](endpoints/internal.md#-image-metadata-write-pdf-post-images) | A shared internal image-metadata service under `api/internal/images/confirm/`, called by each route-specific confirm handler. | The shared image insert, inside the route-appropriate confirm transaction. | Partial; the insert primitive exists, but the generic function and its contract don't. |

### Me

| Endpoint | Code | Queries | Reusable |
| --- | --- | --- | --- |
| [`GET /me`](endpoints/me.md#-get-me) | `api/me/get/` | `database/queries/me/get/` | Yes, after extracting API Gateway/JWT parsing and database connection creation. |
| [`GET /me/clubs`](endpoints/me.md#-get-meclubs) | `api/me/clubs/get/` | `database/queries/me/clubs/list/` | Yes, after extracting the HTTP/JWT and database adapters. |
| [`GET /me/events`](endpoints/me.md#-get-meevents) | `api/me/events/get/` | `database/queries/me/events/list/` | Yes, after extracting the HTTP/JWT and database adapters. |
| PDF path `GET /me/clubs/events` | Served by `api/me/events/get/`; add the PDF path as an alias only if a client needs it. | `database/queries/me/events/list/` | Partial; `GET /me/events` is reusable, and the PDF path itself isn't built. |
| [`GET /me/clubs/eboard`](endpoints/me.md#-get-meclubseboard) | `api/me/clubs/eboard/get/` | `database/queries/me/clubs/eboard/list/`, or `database/queries/me/clubs/list/` with an explicit role filter. | Partial. |

### Clubs and memberships

| Endpoint | Code | Queries | Reusable |
| --- | --- | --- | --- |
| [`GET /clubs`](endpoints/clubs.md#-get-clubs) | `api/clubs/list/` | `database/queries/clubs/list/` | Yes, after extracting the HTTP and database adapters. |
| [`POST /clubs`](endpoints/clubs.md#-post-clubs) | `api/clubs/create/` | `database/queries/clubs/create/` | Yes, after extracting the HTTP/JWT and database adapters. |
| [`GET /clubs/{clubId}`](endpoints/clubs.md#-get-clubsclubid) | `api/clubs/{clubId}/get/` | `database/queries/clubs/get/` | Yes, after extracting the HTTP and database adapters. |
| [`POST /clubs/{clubId}/members/me`](endpoints/memberships.md#-post-clubsclubidmembersme) | `api/clubs/{clubId}/members/join/` | `database/queries/clubs/members/create/`, or a caller-specific wrapper under `join/`. | Yes, after extracting the HTTP/JWT and database adapters. |
| [`DELETE /clubs/{clubId}/members/me`](endpoints/memberships.md#-delete-clubsclubidmembersme) | `api/clubs/{clubId}/members/leave/` | `database/queries/clubs/members/leave/` | Partial; the handler and query logic are reusable once the route's authorizer is fixed. |
| [`GET /clubs/{clubId}/members`](endpoints/memberships.md#-get-clubsclubidmembers) | `api/clubs/{clubId}/members/list/` | `database/queries/clubs/members/list/` | N/A until built. |
| [`GET /clubs/{clubId}/eboard`](endpoints/memberships.md#-get-clubsclubideboard) | `api/clubs/{clubId}/members/list/` with a role filter, or `api/clubs/{clubId}/eboard/get/` if it keeps a distinct contract. | Prefer the shared `database/queries/clubs/members/list/`. | N/A until built. |
| [`PUT /clubs/{clubId}/members/roles`](endpoints/memberships.md#-put-clubsclubidmembersroles) | `api/clubs/{clubId}/members/roles/update/` | `database/queries/clubs/members/update_role/` | N/A until built. |
| [`POST /clubs/{clubId}/verification`](endpoints/verification.md#-post-clubsclubidverification) | `api/clubs/{clubId}/verification/create/` | `database/queries/clubs/verification/create/` | N/A until built. |
| [`DELETE /clubs/{clubId}/verification`](endpoints/verification.md#-delete-clubsclubidverification) | `api/clubs/{clubId}/verification/delete/` | `database/queries/clubs/verification/delete/` | N/A until built. |

### Events

| Endpoint | Code | Queries | Reusable |
| --- | --- | --- | --- |
| [`GET /clubs/{clubId}/events`](endpoints/club-events.md#-get-clubsclubidevents) | `api/clubs/{clubId}/events/list/` | `database/queries/clubs/events/list/` | Yes, after extracting the HTTP and database adapters. |
| [`POST /clubs/{clubId}/events`](endpoints/club-events.md#-post-clubsclubidevents) | `api/clubs/{clubId}/events/create/` | `database/queries/events/create/`, shared across route adapters. | Partial; validation and orchestration are reusable after the SQL repair, a single transaction, and extracting authorization. |
| [`GET /clubs/{clubId}/events/drafts`](endpoints/club-events.md#-get-clubsclubideventsdrafts) | `api/clubs/{clubId}/events/drafts/get/`, or an authenticated filter on the shared list service. | `database/queries/clubs/events/list/` with an explicit, authorized status policy. | N/A as an endpoint; the underlying read query is reusable. |
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

### Images

| Endpoint | Code | Queries | Reusable |
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

| Endpoint | Code | Queries | Reusable |
| --- | --- | --- | --- |
| [`GET /admins`](endpoints/admins.md#-get-admins) | `api/admins/list/` | `database/queries/admins/list/` and a shared `is_admin` query. | N/A until built. |
| [`GET /admins/{studentId}`](endpoints/admins.md#-get-adminsstudentid) | `api/admins/{studentId}/get/` | `database/queries/admins/get/` | Partial; the schema intent is clear, but the stub isn't parameterized, working SQL. |
| [`POST /admins`](endpoints/admins.md#-post-admins) | `api/admins/create/` | `database/queries/admins/create/` | N/A until built. |
| [`DELETE /admins/{studentId}`](endpoints/admins.md#-delete-adminsstudentid) | `api/admins/{studentId}/delete/` | `database/queries/admins/delete/` | Partial; only obsolete SQL-shaped notes exist. |

### Operational

| Endpoint | Code | Queries | Reusable |
| --- | --- | --- | --- |
| [`GET /health`](endpoints/operational.md#-get-health) | `api/operations/health/get/` | N/A | Yes, after extracting the Lambda/API Gateway adapter. |
| [`GET /database/test`](endpoints/operational.md#-get-databasetest) | Not a public route in the design; equivalent coverage belongs in protected diagnostics or integration tests. | N/A | Partial as a connectivity test; it isn't part of the public API contract. |

### Behaviors to pin down with tests first

These current behaviors need contract tests before their `api/` versions are built, so each one is either kept or changed on purpose:

- The public event lists' strict date bounds, joined-row pagination, unordered map grouping, partial club associations, and fixed placeholder thumbnails ([event object](endpoints/events.md#event-object)).
- `GET /events/{eventId}` requires only a non-empty ID, matches it with `LIKE`, and caps joined rows at 100.
- The current event-image handlers trust client-provided metadata and don't verify the S3 object before inserting into the database.
- The club-scoped image handlers ignore `{clubId}`; don't copy them just to support the `/auth/events` routes.
- Hard-coded stub responses (for example, the `GET /events/{eventId}/images` stub) aren't contracts.

## Shared authentication and authorization dependencies

Current behavior of each helper is in [endpoints/internal.md](endpoints/internal.md).

| Capability | Current source | Current use | Planned home |
| --- | --- | --- | --- |
| Extract verified `sub` and optional email | [`extract_sub.go`](../infrastructure/legacy/utils/auth/extract_sub.go) | Club and event creation, and the membership handlers | HTTP/identity adapter plus an application identity context |
| Ensure a student row exists | [`ensure_student.go`](../infrastructure/legacy/utils/auth/ensure_student.go) | All six JWT-attached handlers | Shared student sync service |
| Check club e-board/owner role | [`club_authorization.go`](../infrastructure/legacy/utils/auth/club_authorization.go) and a duplicate [`lambda/internal` helper](../infrastructure/legacy/lambda/internal/auth/club_authorization/club_authorization.go) | **No deployed handler calls it** | Shared club authorization policy using [query group 13](../database/README.md#13-club-authorization) |
| Check event-associated club role | [`event_authorization.go`](../infrastructure/legacy/utils/auth/event_authorization.go) and a duplicate [`lambda/internal` helper](../infrastructure/legacy/lambda/internal/auth/event_authorization/event_authorization.go) | **No deployed handler calls it** | Shared event authorization policy using [query group 14](../database/README.md#14-event-authorization) |
| Format validation errors | [`validation_error.go`](../infrastructure/legacy/utils/errors/validation_error.go) | Club and event create handlers | Shared, transport-neutral validation result mapping |

A Cognito JWT establishes identity only. It doesn't establish admin, membership, e-board, owner, or event authority. Until the role checks are built, each endpoint lists its gap as a known issue; none of those gaps is enforced authorization.

## Build order

1. Define provider-neutral request, response, identity, database, and storage interfaces, with snapshot tests of the current response contracts.
2. Build shared identity extraction, student sync, validation, response mapping, and authorization policies once, dropping the duplicate role helpers at that point.
3. Build the implemented read routes and their query groups before the write workflows.
4. Build club membership and club creation, with explicit decisions on route authorization and creator ownership.
5. Repair and transaction-test event creation before exposing it from the new module.
6. Add the storage-provider abstraction, then build presigning, confirmation, and gallery reads with authorization and lifecycle tests.
7. Build the planned routes only after their product and authorization decisions are approved. Don't create empty endpoint folders just to mirror the docs.
8. Switch the API Gateway/Lambda wiring to the new module only after the contract tests pass. `infrastructure/legacy/` stays the working reference until that switch is reviewed.

The product and design decisions that block these steps are listed under [Open questions](README.md#open-questions).

## What exists today

- [`api/README.md`](README.md): the blueprint index.
- [`api/endpoints/`](endpoints/): the per-resource endpoint reference.
- `api/MIGRATION.md`: this build plan.
- `api/`: no module code yet.
- [`infrastructure/legacy/lambda/api/`](../infrastructure/legacy/lambda/api/): the current handler source.
- [`infrastructure/legacy/gateway/`](../infrastructure/legacy/gateway/): the current API Gateway route and integration source.
- [`infrastructure/legacy/utils/auth/`](../infrastructure/legacy/utils/auth/): the current authentication and authorization helpers.
- [`database/README.md`](../database/README.md): schema and query groups.
