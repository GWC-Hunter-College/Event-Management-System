# Hunter College Clubs API

The HTTP API behind the Hunter College clubs and events site ([Hunter-College-Clubs-Frontend](https://github.com/GWC-Hunter-College/Hunter-College-Clubs-Frontend)). The public Girls Who Code site ([GWC-Website](https://github.com/GWC-Hunter-College/GWC-Website)) is a second, read-only client of the public routes; it makes no API calls yet.

```text
Frontend -> API -> Database
```

- **What this is:** the reference for every endpoint: how the deployed ones behave today, and the design for the ones that aren't built, which are marked ⬜.
- **Implementation:** Go Lambda handlers behind two API Gateway HTTP APIs, in [`infrastructure/legacy/`](../infrastructure/legacy/). The provider-independent `api/` module has no code yet; its layout is in [LAYOUT.md](LAYOUT.md).
- **Base URL:** whatever the frontend's `VITE_API_BASE_URL` points to. That's the API Gateway endpoint each stack prints as its `myHttpApiEndpoint` output: `ClubEventApiDev` (MySQL schema `STAGING`) or `ClubEventApiProd` (schema `PRODUCTION`). There's no custom domain, path prefix, or version prefix; append the paths below as they are.
- **Format:** request and response bodies are JSON. A few image-route errors are plain text; those endpoints say so.
- **Status:** comes from the code, never from a checkmark in the planning PDF. The PDF supplies the design intent for endpoints that aren't built yet.

```bash
curl "$VITE_API_BASE_URL/events?startDate=2026-09-01"
curl -H "Authorization: Bearer $TOKEN" "$VITE_API_BASE_URL/me/clubs"
```

## Legend

**Access**

- 🟢 **Public:** no Cognito authorizer on the route.
- 🔴 **Protected:** requires a Cognito JWT (`Authorization: Bearer <token>`). A JWT proves identity only. Member, e-board, owner, and admin checks are separate authorization and are listed per endpoint.
- 🔵 **Internal:** a Lambda or function that isn't exposed through API Gateway, such as the Cognito student trigger or an authorization check.

**Status**

- ✅ **Implemented:** deployed and working. Check the endpoint's known issues.
- 🟨 **Partial or broken:** the code exists but is broken, misconfigured, or not wired up.
- ⬜ **Planned, not built:** no route or function yet. Stubs, SQL, or helpers may exist; the endpoint's section says which.

For a planned endpoint, the access marker is what the planning PDF intended.

**Frontend need** (from [coverage.md](coverage.md))

- **Now:** the Hunter College Clubs frontend already has UI that needs this endpoint, or a change to it. The endpoint's section has a **Proposed** contract.
- **Later:** only needed by UI that doesn't exist yet, or that shows "SOON".
- **—:** no frontend change depends on it.

**Proposed** marks a request or response shape that isn't built. It follows the conventions of the built routes and is labelled wherever it appears.

## Auth

### Sending a token

Protected (🔴) routes need `Authorization: Bearer <token>`. API Gateway validates the token against the shared Cognito user pool and web app client, then passes its claims to the Lambda in `requestContext.authorizer.jwt.claims`; handlers never parse the token themselves. A missing, expired, or invalid token gets `401` from API Gateway, and the Lambda doesn't run.

On 🟢 routes the header is ignored: API Gateway only passes claims on routes that have the authorizer.

The frontend sends the Cognito **access token** (`auth.user.access_token` from `react-oidc-context`, in its `src/types/auth.ts`).

### What the token gives you

| Claim | Meaning |
| --- | --- |
| `sub` | The Cognito user ID. Always present, and used as `students.id`: it's the caller's identity everywhere. |
| `email` | Optional. Cognito ID tokens include it, but access tokens don't, so frontend calls normally reach the handlers without an email. The student's email is stored by the [Cognito trigger](endpoints/internal.md#-cognito-student-sync-trigger). |

No handler uses Cognito groups, OAuth scopes, or custom claims. Every protected handler first calls [`RequireStudent`](endpoints/internal.md#-requirestudent-request-time-student-sync), which creates the caller's `students` row if it's missing.

### Roles

A JWT tells you *who* the caller is. Roles live in MySQL, and a handler only enforces one if it queries for it.

| Role | Stored as | Set by | Checked by |
| --- | --- | --- | --- |
| `member` | A `club_members (fk_student_id, fk_club_id)` row | [`POST /clubs/{clubId}/members/me`](endpoints/memberships.md#-post-clubsclubidmembersme) | Only by scoping queries to the caller, as `/me/events` does. |
| `eboard` | `club_members.member_is_eboard = TRUE` | Nothing yet; [role changes are planned](endpoints/memberships.md#-put-clubsclubidmembersroles). | The [club](endpoints/internal.md#-club-role-check) and [event](endpoints/internal.md#-event-role-check) role checks, which no handler calls yet. |
| `owner` | `club_members.member_is_owner = TRUE` | Nothing yet; creating a club doesn't make you its owner. | The same unused role checks. The leave-club SQL refuses to delete an owner's membership. |
| `admin` | An `admins (fk_student_id)` row | Nothing yet; [admin routes are planned](endpoints/admins.md). | Nothing; there's no [admin check](endpoints/internal.md#-admin-check) yet. |

[`GET /me/clubs`](endpoints/me.md#-get-meclubs) reports one role per club, `owner` over `eboard` over `member`, because the deployed schema doesn't make the flags mutually exclusive.

The table above is the deployed (legacy) schema. In the [baseline schema](../database/migrations/schema/2026_09_29_baseline_up.sql), `club_members` has one `role` column, `member`, `eboard`, or `owner`, so a member has exactly one role ([decision 11](#decisions)). The e-board list is `role IN ('eboard', 'owner')`. Admins are still rows in `admins`. A verified club (a row in `verified_clubs`) is a directory filter, not a role, and `events_to_clubs.club_is_event_owner` marks an event's host club, not a user role.

**Today, 🔴 routes check identity only. No deployed route enforces an e-board, owner, or admin role.**

## Responses

- Most handlers share helpers: success is `200` with `{"message": "...", ...data}`, and failures are `400`, `404`, or `500` with `{"error": "...", ...}`. Creates and joins return `200`, never `201` or `204`.
- `400` covers bad input, missing claims, duplicate membership, and several not-found cases. Only `GET /events/{eventId}` returns `404`.
- Some `500` bodies include raw database error text.
- The image and health handlers build their responses by hand. Image errors use `message` instead of `error`, and some are plain text rather than JSON. Health returns `{"greeting": "Server running"}`.
- If a Lambda can't start (for example, it can't reach the database on a cold start), API Gateway returns a generic `500`.
- CORS allows all origins and headers. Prod allows `GET`, `POST`, `OPTIONS`, `PATCH`, and `DELETE`; dev leaves out `DELETE`. No route uses `PATCH` yet.
- Examples in these docs come from the handler code and Go structs, not from calls to a live API. The embedded-SQL loader reads each query into a fixed 4 KiB buffer, so exercise a deployed integration before treating an example as a runtime guarantee.

## Endpoints

Every route registered in [`gateway/routes`](../infrastructure/legacy/gateway/routes/), every internal function, every endpoint the planning PDF's "Endpoints Revamp" and image-upload flow call for, and one endpoint proposed from the [frontend coverage](coverage.md) check (`PATCH /clubs/{clubId}`). Which screen uses each endpoint is in [coverage.md](coverage.md#screens).

| Access | Method | Path | What it does | Status | Frontend need | Docs |
| --- | --- | --- | --- | --- | --- | --- |
| **Me** | | | | | | |
| 🔴 | GET | `/me` | Returns the caller's student record. | ✅ | Later | [me.md](endpoints/me.md#-get-me) |
| 🔴 | GET | `/me/clubs` | Lists the caller's clubs with their role in each. | ✅ | Now | [me.md](endpoints/me.md#-get-meclubs) |
| 🔴 | GET | `/me/events` | Lists posted events from the caller's clubs. | ✅ | Now | [me.md](endpoints/me.md#-get-meevents) |
| 🔴 | GET | `/me/clubs/eboard` | Lists clubs where the caller is e-board or owner. | ⬜ | — | [me.md](endpoints/me.md#-get-meclubseboard) |
| **Clubs** | | | | | | |
| 🟢 | GET | `/clubs` | Lists clubs; `?verified=true` returns only verified ones. | ✅ | Now | [clubs.md](endpoints/clubs.md#-get-clubs) |
| 🔴 | POST | `/clubs` | Creates a club (with no owner). | ✅ | Now | [clubs.md](endpoints/clubs.md#-post-clubs) |
| 🟢 | GET | `/clubs/{clubId}` | Returns one club, including unverified ones. | ✅ | Now | [clubs.md](endpoints/clubs.md#-get-clubsclubid) |
| 🔴 | PATCH | `/clubs/{clubId}` | Updates a club's info (owners and e-board). Proposed; not in the PDF. | ⬜ | Later | [clubs.md](endpoints/clubs.md#-patch-clubsclubid) |
| **Memberships** | | | | | | |
| 🔴 | POST | `/clubs/{clubId}/members/me` | Joins the caller to a club. | ✅ | Now | [memberships.md](endpoints/memberships.md#-post-clubsclubidmembersme) |
| 🟢 | DELETE | `/clubs/{clubId}/members/me` | Leaves a club. Broken: the route has no authorizer. | 🟨 | Now | [memberships.md](endpoints/memberships.md#-delete-clubsclubidmembersme) |
| 🔴 | GET | `/clubs/{clubId}/members` | Lists a club's members, for its e-board and owners. | ⬜ | Later | [memberships.md](endpoints/memberships.md#-get-clubsclubidmembers) |
| 🔴 | GET | `/clubs/{clubId}/eboard` | Lists a club's e-board and owners. | ⬜ | Later | [memberships.md](endpoints/memberships.md#-get-clubsclubideboard) |
| 🔴 | PUT | `/clubs/{clubId}/members/roles` | Promotes or demotes a member (owners only). | ⬜ | Later | [memberships.md](endpoints/memberships.md#-put-clubsclubidmembersroles) |
| **Club events** | | | | | | |
| 🟢 | GET | `/clubs/{clubId}/events` | Lists a club's posted events. | ✅ | Now | [club-events.md](endpoints/club-events.md#-get-clubsclubidevents) |
| 🔴 | POST | `/clubs/{clubId}/events` | Creates a draft event. Broken: invalid SQL. | 🟨 | Now | [club-events.md](endpoints/club-events.md#-post-clubsclubidevents) |
| 🔴 | GET | `/clubs/{clubId}/events/drafts` | Lists a club's draft events, for its e-board and owners. | ⬜ | Now | [club-events.md](endpoints/club-events.md#-get-clubsclubideventsdrafts) |
| **Events** | | | | | | |
| 🟢 | GET | `/events` | Lists posted events in a date window. | ✅ | Now | [events.md](endpoints/events.md#-get-events) |
| 🟢 | GET | `/events/{eventId}` | Returns one posted event with its description and clubs. | ✅ | Now | [events.md](endpoints/events.md#-get-eventseventid) |
| 🟢 | GET | `/events/{eventId}/images` | Public gallery for a posted event. | ⬜ | Later | [events.md](endpoints/events.md#-get-eventseventidimages) |
| **Event management (`/auth/events`)** | | | | | | |
| 🔴 | GET | `/auth/events/{eventId}` | Returns an event in any status, for its managers. | ⬜ | Now | [event-management.md](endpoints/event-management.md#-get-autheventseventid) |
| 🔴 | GET | `/auth/events/{eventId}/images` | Returns an event's images, drafts included, for its managers. | ⬜ | Later | [event-management.md](endpoints/event-management.md#-get-autheventseventidimages) |
| 🔴 | PATCH | `/auth/events/{eventId}` | Updates an event in place: saves or posts a resumed draft (needed now, M12), cancels, or edits. | ⬜ | Now | [event-management.md](endpoints/event-management.md#-patch-autheventseventid) |
| 🔴 | POST | `/auth/events/{eventId}/thumbnails` | Presigned upload for an event thumbnail. | ⬜ | Now | [event-management.md](endpoints/event-management.md#-post-autheventseventidthumbnails) |
| 🔴 | POST | `/auth/events/{eventId}/images` | Presigned upload for an event gallery image. | ⬜ | Later | [event-management.md](endpoints/event-management.md#-post-autheventseventidimages) |
| 🔴 | DELETE | `/auth/events/{eventId}` | Deletes an event (soft delete). | ⬜ | Later | [event-management.md](endpoints/event-management.md#-delete-autheventseventid) |
| 🔴 | POST | `/auth/events/{eventId}/restore` | Restores an event deleted less than 30 days ago (owning club's e-board or owners, or an admin). Proposed; not in the PDF. | ⬜ | Later | [event-management.md](endpoints/event-management.md#-post-autheventseventidrestore) |
| **Announcements** (Proposed; not in the PDF) | | | | | | |
| 🟢 | GET | `/clubs/{clubId}/announcements` | Lists a club's posted announcements, most recently posted first. The GWC website reads it too. | ⬜ | Later | [announcements.md](endpoints/announcements.md#-get-clubsclubidannouncements) |
| 🟢 | GET | `/announcements/{announcementId}` | Returns one posted announcement. | ⬜ | Later | [announcements.md](endpoints/announcements.md#-get-announcementsannouncementid) |
| 🔴 | GET | `/clubs/{clubId}/announcements/drafts` | Lists a club's draft announcements, for its e-board and owners. | ⬜ | Later | [announcements.md](endpoints/announcements.md#-get-clubsclubidannouncementsdrafts) |
| 🔴 | GET | `/auth/announcements/{announcementId}` | Returns an announcement in either status, for its managers. | ⬜ | Later | [announcements.md](endpoints/announcements.md#-get-authannouncementsannouncementid) |
| 🔴 | POST | `/clubs/{clubId}/announcements` | Creates an announcement, optionally shared with other clubs. | ⬜ | Later | [announcements.md](endpoints/announcements.md#-post-clubsclubidannouncements) |
| 🔴 | PATCH | `/auth/announcements/{announcementId}` | Edits an announcement, or posts a draft (`draft → posted` only). | ⬜ | Later | [announcements.md](endpoints/announcements.md#-patch-authannouncementsannouncementid) |
| 🔴 | DELETE | `/auth/announcements/{announcementId}` | Deletes an announcement (soft delete). | ⬜ | Later | [announcements.md](endpoints/announcements.md#-delete-authannouncementsannouncementid) |
| 🔴 | POST | `/auth/announcements/{announcementId}/restore` | Restores an announcement deleted less than 30 days ago. | ⬜ | Later | [announcements.md](endpoints/announcements.md#-post-authannouncementsannouncementidrestore) |
| **Images** | | | | | | |
| 🟢 | POST | `/clubs/{clubId}/thumbnails` | Presigned S3 upload URL for a club logo. | ✅ | Now | [images.md](endpoints/images.md#-post-clubsclubidthumbnails) |
| 🔴 | POST | `/clubs/{clubId}/thumbnails/confirm` | Attaches an uploaded logo to its club. | ⬜ | Now | [images.md](endpoints/images.md#-post-clubsclubidthumbnailsconfirm) |
| 🟢 | POST | `/clubs/{clubId}/events/{eventId}/thumbnails` | Presigned S3 upload URL for an event thumbnail. | ✅ | Now | [images.md](endpoints/images.md#-post-clubsclubideventseventidthumbnails) |
| 🔴 | POST | `/clubs/{clubId}/events/{eventId}/thumbnails/confirm` | Attaches an uploaded thumbnail to its event. | ⬜ | Now | [images.md](endpoints/images.md#-post-clubsclubideventseventidthumbnailsconfirm) |
| 🟢 | POST | `/clubs/{clubId}/events/{eventId}/images` | Presigned S3 upload URL for an event gallery image. | ✅ | Later | [images.md](endpoints/images.md#-post-clubsclubideventseventidimages) |
| 🟢 | POST | `/clubs/{clubId}/events/{eventId}/images/confirm` | Records an uploaded gallery image. Fails on warm invocations. | 🟨 | Later | [images.md](endpoints/images.md#-post-clubsclubideventseventidimagesconfirm) |
| 🟢 | GET | `/clubs/{clubId}/events/{eventId}/images` | Lists gallery images with signed URLs. Broken: missing SQL file. | 🟨 | Later | [images.md](endpoints/images.md#-get-clubsclubideventseventidimages) |
| 🔴 | DELETE | `/images/{imageId}` | Deletes an image. | ⬜ | Later | [images.md](endpoints/images.md#-delete-imagesimageid) |
| **Admins** | | | | | | |
| 🔴 | GET | `/admins` | Lists admins. | ⬜ | Later | [admins.md](endpoints/admins.md#-get-admins) |
| 🔴 | GET | `/admins/{studentId}` | Returns one admin. | ⬜ | Later | [admins.md](endpoints/admins.md#-get-adminsstudentid) |
| 🔴 | POST | `/admins` | Promotes a student to admin. | ⬜ | Later | [admins.md](endpoints/admins.md#-post-admins) |
| 🔴 | DELETE | `/admins/{studentId}` | Demotes an admin; never the last one. | ⬜ | Later | [admins.md](endpoints/admins.md#-delete-adminsstudentid) |
| 🔴 | POST | `/admins/purge` | Permanently removes events and images deleted more than 30 days ago (admins only). Proposed; not in the PDF. | ⬜ | Later | [admins.md](endpoints/admins.md#-post-adminspurge) |
| **Verification** | | | | | | |
| 🔴 | POST | `/clubs/{clubId}/verification` | Marks a club verified. | ⬜ | Later | [verification.md](endpoints/verification.md#-post-clubsclubidverification) |
| 🔴 | DELETE | `/clubs/{clubId}/verification` | Removes a club's verification. | ⬜ | Later | [verification.md](endpoints/verification.md#-delete-clubsclubidverification) |
| **Internal** | | | | | | |
| 🔵 | trigger | Cognito sign-up and sign-in | Upserts the student row from `sub` and `email`. | ✅ | — | [internal.md](endpoints/internal.md#-cognito-student-sync-trigger) |
| 🔵 | function | `RequireStudent` | Creates the caller's student row on first request if it's missing. | ✅ | — | [internal.md](endpoints/internal.md#-requirestudent-request-time-student-sync) |
| 🔵 | function | Club role check | Is the caller e-board or owner of a club? No handler calls it yet. | 🟨 | Now | [internal.md](endpoints/internal.md#-club-role-check) |
| 🔵 | function | Event role check | Is the caller e-board or owner of a club linked to an event? No handler calls it yet. | 🟨 | Now | [internal.md](endpoints/internal.md#-event-role-check) |
| 🔵 | function | Admin check | Is the caller an admin? | ⬜ | Later | [internal.md](endpoints/internal.md#-admin-check) |
| 🔵 | function | Image metadata write | Stores image metadata after an upload. | ⬜ | Now | [internal.md](endpoints/internal.md#-image-metadata-write-pdf-post-images) |
| **Operational** | | | | | | |
| 🟢 | GET | `/health` | Liveness check; returns a fixed greeting. | ✅ | — | [operational.md](endpoints/operational.md#-get-health) |
| 🟢 | GET | `/database/test` | Database connectivity check. Dormant: neither API registers it. | 🟨 | — | [operational.md](endpoints/operational.md#-get-databasetest) |


**Totals:** 51 HTTP endpoints and 6 internal functions. `gateway/routes` registers 19 of the endpoints: 18 are deployed on both the dev and prod APIs (12 🟢 and 6 🔴; 14 ✅ and 4 🟨), and `GET /database/test` is dormant. The other 32 endpoints are planned: 21 from the PDF, 1 proposed from the frontend, 2 from the [decisions](#decisions) (restore and purge), and 8 for [announcements](endpoints/announcements.md), for the club page's Announcements tab. By frontend need: 22 endpoints and functions are **Now**, 30 are **Later**, and 5 are **—**.

**Preflight routes:** `POST /clubs/{clubId}/events/{eventId}/thumbnails`, `.../images`, and `.../images/confirm` also register `OPTIONS` against the same Lambda for CORS preflight. Those aren't separate endpoints.

### Names in the planning PDF

The planning PDF names some endpoints differently. These docs use the paths in the code; if you're reading the PDF, look here:

| In the PDF | Where to look |
| --- | --- |
| Register Lambda | [Cognito student sync trigger](endpoints/internal.md#-cognito-student-sync-trigger) |
| `GET /me/clubs/events` | [`GET /me/events`](endpoints/me.md#-get-meevents) |
| `POST /clubs/{clubId}/members` | [`POST /clubs/{clubId}/members/me`](endpoints/memberships.md#-post-clubsclubidmembersme) (self-join only) |
| `POST /clubs/thumbnails?filetype=&filename=` | [`POST /clubs/{clubId}/thumbnails`](endpoints/images.md#-post-clubsclubidthumbnails) (JSON body) |
| `GET /clubs/verified=true` | [`GET /clubs?verified=true`](endpoints/clubs.md#-get-clubs) |
| `GET /events/{eventId}/description`, `GET /events/{eventId}/clubs` | Covered by [`GET /events/{eventId}`](endpoints/events.md#-get-eventseventid) |
| `GET`, `PATCH`, `DELETE /clubs/{clubId}/events/{eventId}` and its `/images` and `/thumbnails` (the PDF's first endpoint list) | Designed as [`/auth/events/{eventId}`](endpoints/event-management.md); the built media routes are [club-scoped](endpoints/images.md) |
| Internal `POST /images` | [Image metadata write](endpoints/internal.md#-image-metadata-write-pdf-post-images) (internal function, not a route) |

An item-by-item comparison of the PDF's 35-item "Endpoints Revamp" with the code is in [pdf-coverage.md](pdf-coverage.md#design-doc-vs-code).

## Decisions

Product decisions taken on 2026-09-29. The proposed contracts they affect are updated to match; items that need a frontend change are listed in [coverage.md](coverage.md#frontend-changes).

1. **Default event status.** When the create-event body omits `status`, the event is created as a draft. The design-mode mock, which defaults to `posted`, changes to match.
2. **Status transitions.** Statuses are `draft`, `posted`, and `cancelled`: the life of an event that still exists. Allowed: `draft → posted`, `posted → cancelled`, and `cancelled → posted`. Nothing goes back to `draft`. There's no `archived` status. [`DELETE /auth/events/{eventId}`](endpoints/event-management.md#-delete-autheventseventid) soft-deletes an event in any status: it sets `deleted_at`, and every read skips the event from then on. It can be restored for 30 days (decision 9), and after that an admin's purge removes it for good (decision 8). Updated 2026-09-29 by the [schema decisions](../database/docs/schema-review.md#decisions). See [Event status](endpoints/events.md#event-status).
3. **Event lists.** `limit` defaults to 50, with a maximum of 100. Upcoming events are ordered by start time ascending, past events descending. Lists carry the flyer URL and a photo count (`imageCount`); only the single-event read, [`GET /events/{eventId}`](endpoints/events.md#-get-eventseventid), returns the full gallery (`images`). See the [proposed event object](endpoints/events.md#proposed-event-object).
4. **Flyer alt text** is stored on `images`, with the flyer's image row.
5. **Club topics** come from a fixed list, stored as lowercase keys, at most 3 per club; the UI uppercases them. The design-mode fixtures change to use the same list. See the [proposed club object](endpoints/clubs.md#proposed-club-object).
6. **Club roles.** Owners and e-board members can edit club info. Only owners can change roles or delete the club. Owners and e-board members can see member emails; regular members can't.
7. **GWC website.** The site finds its club through a configured club ID. It reads that club's posted announcements through the public [`GET /clubs/{clubId}/announcements`](endpoints/announcements.md#-get-clubsclubidannouncements) ([coverage](coverage.md#gwc-website)).
8. **Purging is done by admins, on demand.** There's no scheduled job. [`POST /admins/purge`](endpoints/admins.md#-post-adminspurge) (🔴, admin) permanently removes events, images, and [announcements](endpoints/announcements.md) deleted **more than 30 days ago**, and the S3 files nothing uses any more.
9. **Deleted events can be restored for 30 days.** [`POST /auth/events/{eventId}/restore`](endpoints/event-management.md#-post-autheventseventidrestore) (🔴, e-board or owner of the event's **owning** club, or an admin) clears `deleted_at` only when the event was deleted less than 30 days ago. The purge uses the same 30-day cutoff, so nothing restorable is ever purged.
10. **Date-only list bounds cover whole days in `America/New_York`.** `startDate=2026-10-01` starts at 00:00 on October 1; `endDate=2026-10-31` includes all of October 31, so the range ends before 00:00 on November 1. An event matches if it overlaps the range, so a multi-day event that starts before the range still shows up. Full timestamps with an offset are accepted too. Event start and end times themselves stay full UTC timestamps. See [proposed list parameters](endpoints/events.md#proposed-list-parameters).
11. **One role per member.** `club_members.role` is `member`, `eboard`, or `owner`. The e-board list includes owners. A club's last owner can't be demoted and can't leave; an owner can leave while the club has another owner.
12. **Admin rules.** Admins can add and remove other admins, but the last admin can't be removed, including by themselves. A developer creates the first admin with a one-off SQL insert ([how](../database/README.md#the-first-admin)).

Still open: whether an event's author, or an admin, gets management rights beyond the linked clubs' e-board and owners (see [Open questions](#open-questions)). Decisions 8–12 were taken on 2026-09-29, after the first seven.

## Open questions

Product and design decisions the repos don't settle. Endpoint sections repeat the questions that apply to them.

**Auth and roles**

- Should creating a club make the creator its owner, and should creating one require an admin? The frontend already assumes the creator becomes owner ([query group 29](../database/README.md#29-club-creator-ownership)).
- How does `PUT /clubs/{clubId}/members/roles` say which member to change? (Who may change roles is [decision 6](#decisions).) The proposal puts `studentId` in the body.
- Admins: what does "their id and clubs" in `GET /admins` mean? (Bootstrapping and the last admin are [decision 12](#decisions).)
- Does an event's author, or an admin, get management rights beyond the linked clubs' e-board and owners? Today only an admin's [restore](endpoints/event-management.md#-post-autheventseventidrestore) and [takedown](endpoints/images.md#-delete-imagesimageid) go beyond them.
- Who may verify and unverify clubs? The PDF only marks those routes protected; the admins stub note says admins do it.
- Should unverified clubs stay publicly readable through `GET /clubs/{clubId}` (current behavior), or be admin-only with a verified badge (a PDF p. 50 suggestion)?
- Which token should clients send? The handlers read `email` from the token, but the access token the frontend sends has none.

**Events**

- Should the protected event routes be `/auth/events/{eventId}/...` (the revamp) or the current club-scoped media paths? Either way, keep one upload signer.
- What can `PATCH /auth/events/{eventId}` change (fields, associate clubs), and what are the publish rules? (Concurrent edits are last-write-wins for now, and deleting is a soft delete: [schema decisions](../database/docs/schema-review.md#decisions) 4 and 6.)
- Where do managers get drafts: from `GET /clubs/{clubId}/events` (what the frontend expects) or a protected `GET /clubs/{clubId}/events/drafts` (the PDF)? [coverage.md](coverage.md#m1-drafts-in-public-lists) recommends the protected route. Which statuses may `GET /auth/events/{eventId}` return, and does an event's author get rights of their own?
- Create-event contract: the frontend sends `status` and omits `timezone` and `associates`; the handler requires both, requires `description`, and always stores `drafted`. [coverage.md](coverage.md#m2-create-event-body-timezone-associates-description) recommends the backend make them optional. (The default status is [decision 1](#decisions).)
- Member lists: which fields and pagination does `GET /clubs/{clubId}/members` need, may ordinary members see the list at all (without emails, per [decision 6](#decisions)), and is `/eboard` a separate endpoint or a filter on it?

**Clubs**

- Topic keys: the fixed list is assumed to be the New club form's six topics, lowercased (`technology`, `arts`, `community`, `women in stem`, `career`, `sports`). Is that the list, and are keys with spaces acceptable, or should they be slugs such as `women-in-stem`?
- Deleting a club is owner-only ([decision 6](#decisions)), but no endpoint deletes a club, and none is proposed. Is one needed, and what happens to the club's events and members?

**Images**

- Image rules: who may upload, confirm, and delete; content checks; verifying the S3 object exists; one transaction per confirm; replacing old images; and recovering from failed deletes. Should the club-logo signer return `imageId` and `objectKey` like the others (and the PDF)?
- Should reads return usable image URLs instead of S3 object keys in `thumbnailUrl`?

**Contracts**

- Should the API keep today's quirks (`400` for not-found, joined-row paging, unordered lists, empty club names, placeholder thumbnails), or fix them in a versioned contract? The frontend already expects `404` and `403` in several places.
- `students.email` is nullable, but the Go model isn't. Should `GET /me` allow a `null` email?
- Should `/health` be split into liveness and readiness checks?
- Should the PDF's paths, such as `/me/clubs/events`, exist as aliases? Only if a client needs them.

**Future work**

- Notifications: should posting an [announcement](endpoints/announcements.md) notify members (email or push)? Not now. It depends on a notification system, which the club bulletin board (the Board tab) also needs, so both wait for it.

**Infrastructure**

- Cognito trigger ownership: two stacks wire the student-sync Lambda to the same user pool, which has one slot per trigger. Which stack should own it?
- Both API stacks declare the membership Lambdas with the same fixed names (`PostJoinClubMemberMe` and `DeleteClubMemberMe`). Which API actually serves those routes?

## More docs

- [coverage.md](coverage.md): every frontend screen mapped to the endpoints behind it, the frontend/backend contract mismatches and which side should change, and what the API doesn't cover yet.
- [LAYOUT.md](LAYOUT.md): the `api/` module's directory layout, the module path for each endpoint, and the order its pieces depend on each other.
- [pdf-coverage.md](pdf-coverage.md): the planning PDF's endpoints compared with the code, item by item.
- [database/README.md](../database/README.md): the database module, baseline schema, and the 31 query groups the endpoints link to.
- [Implementation reference](../infrastructure/legacy/docs/api/README.md) and [architecture docs](../infrastructure/legacy/docs/architecture/): the deployed stacks, authentication, and image uploads in more depth.
- The frontend's [`docs/api.md`](https://github.com/GWC-Hunter-College/Hunter-College-Clubs-Frontend/blob/staging/docs/api.md): what the frontend expects from each endpoint. It differs from the backend in places; [coverage.md](coverage.md#contract-mismatches) lists every difference.
