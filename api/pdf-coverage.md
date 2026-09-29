# Planning PDF coverage

How the planning PDF, "GWC Website Documentation", maps to the code: which of its endpoints exist, at which paths, and how complete each one is. Endpoint behavior is documented in [README.md](README.md); the `api/` module's layout is in [LAYOUT.md](LAYOUT.md).

## Readiness legend

This file rates each item by how much working code exists for it. That's a different scale from the status legend in [README.md](README.md#legend), where anything without a deployed route is ⬜.

- ✅ **Built:** a working route, handler, and supporting behavior exist.
- 🟨 **Partial:** only a stub or supporting primitive exists, the route is inactive, the item is built at a different path, or a defect blocks the intended behavior.
- ⬜ **Not started:** no route, handler, or query exists.
- ❓ **Needs a decision:** the evidence conflicts, or a product decision has to come first.

Access markers follow [README.md](README.md#legend). For a route that isn't built, the marker is the design's intent, not a claim that it's deployed.

## Sources

The ratings were checked against:

- route registration in [`gateway/routes/`](../infrastructure/legacy/gateway/routes/);
- both API stacks in [`productionApi.go`](../infrastructure/legacy/internal/stack/productionApi.go) and [`developmentApi.go`](../infrastructure/legacy/internal/stack/developmentApi.go);
- Lambda handlers under [`lambda/api/`](../infrastructure/legacy/lambda/api/);
- embedded SQL under [`utils/query_client/queries/`](../infrastructure/legacy/utils/query_client/queries/);
- authentication and authorization helpers under [`utils/auth/`](../infrastructure/legacy/utils/auth/);
- image/S3 integrations and Cognito trigger wiring; and
- the inactive stub stack and stub handlers, which count only as partial evidence.

The PDF, `GWC Website Documentation.pdf`, supplies the route grouping, the design intent, and the green/red/blue access model. It isn't tracked in the repository. Its checkmarks, assignments, dates, and project-management notes aren't evidence that something is built.

In the 52-page copy of the PDF, the relevant sections are the first endpoint list (pp. 9–13), the "Endpoints Revamp" list (pp. 15–19), and the frontend image-upload flow (pp. 25–26). Page numbers vary between exports. In this one the revamp list has no heading of its own: it's the second endpoint list, the one that adds `/auth/events` and marks `GET /clubs/{clubId}` as new.

## Deployed routes

| Readiness | Count |
| --- | ---: |
| ✅ Built | 14 |
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

## Design coverage

The PDF's "Endpoints Revamp" lists 35 items. Their readiness:

| Readiness | Count |
| --- | ---: |
| ✅ Built at the designed route or intent | 8 |
| 🟨 Partial, at a different path, stubbed, or incomplete | 16 |
| ⬜ Nothing executable | 11 |
| ❓ Needs a decision | 0 |
| **Design items** | **35** |

Two of the items are internal (blue) functions, not HTTP routes: the Cognito student sync (✅) and the generic image metadata write (🟨).

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
| Club-scoped event detail, update, and delete (first list) | Nothing. The revamp designs these under `/auth/events`, and only the club-scoped media paths are built. |

## Design checklist

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
