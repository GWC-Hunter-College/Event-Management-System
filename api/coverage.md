# Frontend coverage

Maps every route in the Hunter College Clubs frontend to the endpoints behind it: what a user can see or do on each screen, which endpoint serves it, and that endpoint's status. It also records where the frontend and backend disagree, and which side should change.

**Sources:** [Hunter-College-Clubs-Frontend](https://github.com/GWC-Hunter-College/Hunter-College-Clubs-Frontend) on `staging` at `3b77955` ("Frontend revamp"): the routes in `src/App.tsx`, `src/pages/`, `src/components/`, `src/hooks/`, the normalizers in `src/types/`, the design-mode mocks in `src/mocks/handlers.ts` and `src/mocks/fixtures.ts`, and `docs/api.md` and `docs/pages.md`. Backend behavior comes from the [endpoint reference](README.md), which is checked against `infrastructure/legacy/`.

**About the mocks:** design mode answers every call from MSW handlers. They return the frontend's own canonical shapes (`start`, `end`, `flyer`, `owner`), not the backend's (`startDate`, `thumbnailUrl`, `owners`). The normalizers in `src/types/` accept both, so the UI works against either. Where this page quotes a mock contract, it says so.

## How to read this

- **Status** is the endpoint's status on the [README legend](README.md#legend): ✅ implemented, 🟨 partial or broken, ⬜ not built. "—" means no endpoint is involved (the UI works client-side or shows static content).
- **Need** says when the frontend needs the endpoint or change: **Now** if the UI for it already exists, **Later** if it doesn't (including UI that only shows a "SOON" pill or "This feature isn't available yet"). Only endpoints with work to do have a Need.
- **M1–M14** refer to [Contract mismatches](#contract-mismatches).
- **Decisions** taken on 2026-09-29 are recorded in [README.md](README.md#decisions); the contracts linked here already follow them. The frontend changes they imply are collected under [Frontend changes](#frontend-changes).

## Needed now

The endpoint work the existing UI is waiting on, in the order the screens depend on it. Every item links to its **Proposed** contract.

| Work | Endpoints | Screens | Mismatches |
| --- | --- | --- | --- |
| Usable image URLs, club names, ISO dates, `cancelled` in public lists, event-based paging | [Proposed event object](endpoints/events.md#proposed-event-object) on [`GET /events`](endpoints/events.md#-get-events), [`GET /events/{eventId}`](endpoints/events.md#-get-eventseventid), [`GET /clubs/{clubId}/events`](endpoints/club-events.md#-get-clubsclubidevents), [`GET /me/events`](endpoints/me.md#-get-meevents) | Home, Events, Club, Event, My Clubs | M6, M8, M9, M10, M14 |
| Club description, tags, member count, logo URL | [Proposed club object](endpoints/clubs.md#proposed-club-object) on [`GET /clubs`](endpoints/clubs.md#-get-clubs), [`GET /clubs/{clubId}`](endpoints/clubs.md#-get-clubsclubid), [`GET /me/clubs`](endpoints/me.md#-get-meclubs) | Clubs, Club, My Clubs, Event | M6, M13 |
| Creating events that work | [`POST /clubs/{clubId}/events`](endpoints/club-events.md#-post-clubsclubidevents) | New event form | M2, M11 |
| Drafts for managers | [`GET /clubs/{clubId}/events/drafts`](endpoints/club-events.md#-get-clubsclubideventsdrafts), [`GET /auth/events/{eventId}`](endpoints/event-management.md#-get-autheventseventid) | New event step 1, New event form, Club → Manage | M1, M12 |
| Updating an event in place: saving or posting a resumed draft, and cancelling | [`PATCH /auth/events/{eventId}`](endpoints/event-management.md#-patch-autheventseventid) | New event form (resumed draft), Event manager bar "CANCEL EVENT". The "EDIT" buttons use it too, but they're Later. | M12, M10 |
| Flyer upload | [`POST /auth/events/{eventId}/thumbnails`](endpoints/event-management.md#-post-autheventseventidthumbnails), [`POST /clubs/{clubId}/events/{eventId}/thumbnails/confirm`](endpoints/images.md#-post-clubsclubideventseventidthumbnailsconfirm) | New event form | M6, M11 |
| Club creation and logo upload | [`POST /clubs`](endpoints/clubs.md#-post-clubs), [`POST /clubs/{clubId}/thumbnails`](endpoints/images.md#-post-clubsclubidthumbnails), [`POST /clubs/{clubId}/thumbnails/confirm`](endpoints/images.md#-post-clubsclubidthumbnailsconfirm) | New club | M4, M7, M13 |
| Leaving a club, and status codes | [`DELETE /clubs/{clubId}/members/me`](endpoints/memberships.md#-delete-clubsclubidmembersme), [`POST /clubs/{clubId}/members/me`](endpoints/memberships.md#-post-clubsclubidmembersme) | Club | M3 |

## Screens

### `/` Home

`src/pages/Home.tsx`

| What the user sees or does | Endpoint | Status | Need | Notes |
| --- | --- | --- | --- | --- |
| Hero strip of the next few upcoming events' flyers | [`GET /events`](endpoints/events.md#-get-events) | ✅ | Now | Event objects carry no flyer URL, so every tile shows fallback art (M6). The call sends no parameters, so it gets 10 joined rows starting in 1970 (M9). |
| "Upcoming Events" grid (16 tiles on desktop, 8 on mobile), sorted by start | [`GET /events`](endpoints/events.md#-get-events) | ✅ | Now | Client filters out drafts, cancelled, and past events. Needs the [proposed event object](endpoints/events.md#proposed-event-object). |
| Host club name and logo on each tile | [`GET /clubs?verified=true`](endpoints/clubs.md#-get-clubs) | ✅ | Now | Looked up by ID because event objects have empty club names (M8). An unverified host club isn't found. |
| "See as a list" / "See all N events" | — | — | | Links to `/events`. |

### `/events` Events

`src/pages/Events.tsx`

| What the user sees or does | Endpoint | Status | Need | Notes |
| --- | --- | --- | --- | --- |
| Agenda of upcoming events, grouped by month and day | [`GET /events`](endpoints/events.md#-get-events) | ✅ | Now | Same paging problem as Home (M9). |
| Host club name and logo on each row | [`GET /clubs?verified=true`](endpoints/clubs.md#-get-clubs) | ✅ | Now | Workaround for M8. |
| "Cancelled" pill on cancelled events | [`GET /events`](endpoints/events.md#-get-events) | ✅ | Now | Never shown: the backend returns only `posted` events and has no `cancelled` status (M10). |
| Search by title or club | — | — | | Client-side. |
| Add to calendar | — | — | | Client-side `.ics` download (`src/lib/ics.ts`). |

### `/clubs` Clubs

`src/pages/Clubs.tsx`, `src/components/ui/ClubCard.tsx`

| What the user sees or does | Endpoint | Status | Need | Notes |
| --- | --- | --- | --- | --- |
| Club cards: logo, name, description, topic tags, "N MEMBERS" | [`GET /clubs?verified=true`](endpoints/clubs.md#-get-clubs) | ✅ | Now | The list returns only `id`, `name`, and a logo S3 key. Description, tags, and member count are missing (M13), and the logo isn't a URL (M6). Member count is hidden when absent. |
| "N clubs at Hunter CS" | [`GET /clubs?verified=true`](endpoints/clubs.md#-get-clubs) | ✅ | | Array length. |
| Search by name or topic | — | — | | Client-side over `name` and `tags`; topic search finds nothing until tags are returned. |

### `/club/:clubId` Club

`src/pages/Club.tsx`, `src/components/Club/*`

| What the user sees or does | Endpoint | Status | Need | Notes |
| --- | --- | --- | --- | --- |
| Header: logo, name, bio, topic tags, member count | [`GET /clubs/{clubId}`](endpoints/clubs.md#-get-clubsclubid) | ✅ | Now | Tags and member count are missing (M13); logo is an S3 key (M6). |
| "We couldn't find that club" | [`GET /clubs/{clubId}`](endpoints/clubs.md#-get-clubsclubid) | ✅ | Now | Backend sends `400`, not `404` (M3). The page treats any failure as not found, so it still works. |
| Upcoming and past counts; Events tab with upcoming rows, past events grouped by semester, "load earlier", photo count on past rows | [`GET /clubs/{clubId}/events`](endpoints/club-events.md#-get-clubsclubidevents) | ✅ | Now | 10 joined rows by default (M9), so older semesters never load. Photo counts need `imageCount` in the [event object](endpoints/events.md#proposed-event-object); lists don't carry the gallery ([decision 3](README.md#decisions)). |
| Role label, "JOINED" menu, "+ NEW EVENT", Manage tab visibility | [`GET /me/clubs`](endpoints/me.md#-get-meclubs) | ✅ | | |
| "+ JOIN CLUB" | [`POST /clubs/{clubId}/members/me`](endpoints/memberships.md#-post-clubsclubidmembersme) | ✅ | Now | `400` for a missing club instead of `404` (M3). |
| "Leave club" (with confirm modal; hidden for owners) | [`DELETE /clubs/{clubId}/members/me`](endpoints/memberships.md#-delete-clubsclubidmembersme) | 🟨 | Now | The route has no authorizer, so every call fails with `400`. |
| "Share club" | — | — | | Client-side. |
| "Event reminders" menu item (SOON) | none | — | | [Not covered by the API yet](#not-covered-by-the-api-yet). |
| Board tab (SOON, "The club board is coming soon") | none | — | | [Not covered by the API yet](#not-covered-by-the-api-yet). |
| Announcements tab (SOON, "This feature isn't available yet") | none | — | | [Not covered by the API yet](#not-covered-by-the-api-yet). |
| Manage tab: drafts and upcoming events with a status pill | [`GET /clubs/{clubId}/events/drafts`](endpoints/club-events.md#-get-clubsclubideventsdrafts) and [`GET /clubs/{clubId}/events`](endpoints/club-events.md#-get-clubsclubidevents) | ⬜ / ✅ | Now | Today the page looks for drafts in the public list, which never has them (M1). |
| Manage tab: "EDIT" per event (disabled, SOON) | [`GET /auth/events/{eventId}`](endpoints/event-management.md#-get-autheventseventid), [`PATCH /auth/events/{eventId}`](endpoints/event-management.md#-patch-autheventseventid) | ⬜ | Later | The button is disabled with a SOON pill. `PATCH` itself is needed now for resumed drafts (M12). |
| Manage tab: member list, e-board list, promote and demote | [`GET /clubs/{clubId}/members`](endpoints/memberships.md#-get-clubsclubidmembers), [`GET /clubs/{clubId}/eboard`](endpoints/memberships.md#-get-clubsclubideboard), [`PUT /clubs/{clubId}/members/roles`](endpoints/memberships.md#-put-clubsclubidmembersroles) | ⬜ | Later | No UI yet: the Manage tab lists events only. |
| Edit club | [`PATCH /clubs/{clubId}`](endpoints/clubs.md#-patch-clubsclubid) | ⬜ | Later | No UI yet. |

### `/event/:eventId` Event

`src/pages/Event.tsx`, `src/components/Event/*`

| What the user sees or does | Endpoint | Status | Need | Notes |
| --- | --- | --- | --- | --- |
| Title, when and where (in the event's timezone), about, RSVP link | [`GET /events/{eventId}`](endpoints/events.md#-get-eventseventid) | ✅ | Now | Times shift by the viewer's UTC offset (M14). RSVP shows only for absolute `http(s)` links. |
| "We couldn't find that event" | [`GET /events/{eventId}`](endpoints/events.md#-get-eventseventid) | ✅ | | `404` works as expected. Drafts also return `404`. |
| Gallery: flyer plus extra photos; "VIEW ALL N PHOTOS"; "PAST EVENT · N PHOTOS" | [`GET /events/{eventId}`](endpoints/events.md#-get-eventseventid) | ✅ | Now | Reads `images` from the event object, which the backend doesn't return (see the [proposed event object](endpoints/events.md#proposed-event-object)); only this single-event read will carry the full gallery. There's no flyer URL either (M6). |
| "Cancelled" pill; share-only actions for cancelled events | [`GET /events/{eventId}`](endpoints/events.md#-get-eventseventid) | ✅ | Now | Needs the `cancelled` status (M10). |
| "Hosted by" logo and name, linking to the club | [`GET /clubs/{clubId}`](endpoints/clubs.md#-get-clubsclubid) | ✅ | | Uses `owners.owner.id` from the event. |
| Add to calendar, share | — | — | | Client-side. |
| Manager bar ("You own …" / "You're on the e-board of …") | [`GET /me/clubs`](endpoints/me.md#-get-meclubs) | ✅ | | Shown to the host club's e-board and owners only. |
| Manager bar: "EDIT EVENT" (shows "This feature isn't available yet") | [`GET /auth/events/{eventId}`](endpoints/event-management.md#-get-autheventseventid), [`PATCH /auth/events/{eventId}`](endpoints/event-management.md#-patch-autheventseventid) | ⬜ | Later | Editing a posted event waits for an edit form. Both endpoints are needed now for other reasons: `?draft=` prefill and resumed drafts (M1, M12). |
| Manager bar: "CANCEL EVENT" on upcoming events (shows "This feature isn't available yet") | [`PATCH /auth/events/{eventId}`](endpoints/event-management.md#-patch-autheventseventid) with `status: "cancelled"` | ⬜ | Now | Cancelling keeps the event listed; it isn't archiving or deleting (M10). |
| Delete event | [`DELETE /auth/events/{eventId}`](endpoints/event-management.md#-delete-autheventseventid) | ⬜ | Later | The manager bar has no delete button. |
| Gallery: "SOON" add-photo tile for managers | [`POST /auth/events/{eventId}/images`](endpoints/event-management.md#-post-autheventseventidimages) and its confirm step | ⬜ | Later | Placeholder only. |

### `/my-clubs` My Clubs

`src/pages/MyClubs.tsx`

| What the user sees or does | Endpoint | Status | Need | Notes |
| --- | --- | --- | --- | --- |
| Signed out: blurred preview and sign-in gate | — | — | | Static placeholder content. |
| Stat cards: clubs you're in, upcoming events from your clubs, clubs you help run | [`GET /me/clubs`](endpoints/me.md#-get-meclubs), [`GET /me/events`](endpoints/me.md#-get-meevents) | ✅ | Now | Upcoming count is limited by M9. |
| "Your clubs" cards: logo, name, role, next event | [`GET /me/clubs`](endpoints/me.md#-get-meclubs), [`GET /me/events`](endpoints/me.md#-get-meevents) | ✅ | Now | Logo is an S3 key (M6). Next event is computed client-side. |
| "From your clubs" agenda, searchable | [`GET /me/events`](endpoints/me.md#-get-meevents) | ✅ | Now | Club names come from `GET /me/clubs`. Cancelled events never arrive (M10). |
| "You haven't joined any clubs yet" | [`GET /me/clubs`](endpoints/me.md#-get-meclubs) | ✅ | | Empty `clubs`. |

### `/create` Create hub

`src/pages/Create.tsx`

| What the user sees or does | Endpoint | Status | Need | Notes |
| --- | --- | --- | --- | --- |
| "Start a club" option | — | — | | Links to `/club/create`. |
| "Post an event" option, disabled with a caption if you manage no clubs | [`GET /me/clubs`](endpoints/me.md#-get-meclubs) | ✅ | | Filters to `eboard` and `owner` client-side; [`GET /me/clubs/eboard`](endpoints/me.md#-get-meclubseboard) isn't needed. |

### `/event/create` New event, step 1

`src/pages/EventCreateStep1.tsx`, `src/hooks/useMyDrafts.ts`

| What the user sees or does | Endpoint | Status | Need | Notes |
| --- | --- | --- | --- | --- |
| Pick a club you own or are e-board of | [`GET /me/clubs`](endpoints/me.md#-get-meclubs) | ✅ | | |
| "Pick up a draft" panel across your managed clubs, with "no flyer yet" | [`GET /clubs/{clubId}/events/drafts`](endpoints/club-events.md#-get-clubsclubideventsdrafts), once per managed club | ⬜ | Now | Today it calls the public `GET /clubs/{clubId}/events` per club and filters to drafts, so the panel is always empty (M1). |

### `/club/:clubId/event/new` New event form

`src/pages/EventForm.tsx`

| What the user sees or does | Endpoint | Status | Need | Notes |
| --- | --- | --- | --- | --- |
| Hosting club chip | [`GET /clubs/{clubId}`](endpoints/clubs.md#-get-clubsclubid) | ✅ | | |
| "You don't manage this club" panel for non-managers | [`GET /me/clubs`](endpoints/me.md#-get-meclubs) | ✅ | | Client-side check only; the backend doesn't enforce it (M2). |
| `?draft=:eventId` prefills the form | [`GET /auth/events/{eventId}`](endpoints/event-management.md#-get-autheventseventid) | ⬜ | Now | Today it searches the public club list, which has no drafts (M1). |
| Title, dates and times, location, RSVP link, description | [`POST /clubs/{clubId}/events`](endpoints/club-events.md#-post-clubsclubidevents) | 🟨 | Now | Broken SQL; the body also fails validation (M2). The form has no timezone field and reads typed times as `America/New_York`. |
| "SAVE DRAFT" and "POST EVENT" | [`POST /clubs/{clubId}/events`](endpoints/club-events.md#-post-clubsclubidevents) with `status`; [`PATCH /auth/events/{eventId}`](endpoints/event-management.md#-patch-autheventseventid) for a resumed draft | 🟨 / ⬜ | Now | Resuming a draft creates a second event instead of updating it (M12). |
| Flyer dropzone and alt text | [`POST /auth/events/{eventId}/thumbnails`](endpoints/event-management.md#-post-autheventseventidthumbnails) (today [`POST /clubs/{clubId}/events/{eventId}/thumbnails`](endpoints/images.md#-post-clubsclubideventseventidthumbnails)), then [`POST /clubs/{clubId}/events/{eventId}/thumbnails/confirm`](endpoints/images.md#-post-clubsclubideventseventidthumbnailsconfirm) | ⬜ / ✅ / ⬜ | Now | The form sends a `blob:` URL as `flyer` instead (M11). |
| "More photos for this event" (SOON, "This feature isn't available yet") | [`POST /auth/events/{eventId}/images`](endpoints/event-management.md#-post-autheventseventidimages) and its confirm step | ⬜ | Later | |
| "Co-hosting clubs" (SOON, "This feature isn't available yet") | `associates` on [`POST /clubs/{clubId}/events`](endpoints/club-events.md#-post-clubsclubidevents) and [`PATCH /auth/events/{eventId}`](endpoints/event-management.md#-patch-autheventseventid) | 🟨 / ⬜ | Later | No screen shows co-hosts yet, although event objects include `owners.associates`. |
| Live preview | — | — | | Client-side. |

### `/club/create` New club

`src/pages/ClubForm.tsx`

| What the user sees or does | Endpoint | Status | Need | Notes |
| --- | --- | --- | --- | --- |
| Name and description; redirect to the new club page | [`POST /clubs`](endpoints/clubs.md#-post-clubs) | ✅ | Now | The creator doesn't become the owner, so the new club page has no Manage tab or "+ NEW EVENT" (M4). |
| Topics (up to three from a fixed list) | [`POST /clubs`](endpoints/clubs.md#-post-clubs) `tags` | ✅ | Now | Sent but ignored; there's no club tags table (M13). The form sends uppercase labels; the API takes lowercase keys ([decision 5](README.md#decisions)). |
| Logo dropzone | [`POST /clubs/{clubId}/thumbnails`](endpoints/images.md#-post-clubsclubidthumbnails), then [`POST /clubs/{clubId}/thumbnails/confirm`](endpoints/images.md#-post-clubsclubidthumbnailsconfirm) | ✅ / ⬜ | Now | The form sends a `blob:` URL as `logo` instead (M7, M13). |
| "How it looks on Discover" preview | — | — | | Client-side. |

### `/auth` Auth debug

`src/pages/Auth.tsx`. Dev-only token inspector; no endpoint.

### Every page (app shell)

| What the user sees or does | Endpoint | Status | Need | Notes |
| --- | --- | --- | --- | --- |
| Signed-in email in the user menu | — | — | | Read from the OIDC ID-token profile. [`GET /me`](endpoints/me.md#-get-me) isn't called. |
| Sign in and sign out | — | — | | Cognito Hosted UI through `react-oidc-context`. |

## Contract mismatches

Where the frontend and backend disagree. M1–M7 were recorded in the Phase 1 reference; each was re-checked against the frontend code listed. M8–M14 are new. "Change" says which side should move and why; nothing here has been changed in either repo.

### M1. Drafts in public lists

- **Frontend:** expects `GET /events` and `GET /clubs/{clubId}/events` to include drafts, and filters them out itself (`useEvents.ts`, `Club.tsx`). The step-1 drafts panel (`useMyDrafts.ts`), the Manage tab, and `?draft=` prefill (`EventForm.tsx`) all look for drafts in the public club list. `docs/api.md` documents this.
- **Backend:** every public list hard-codes `status = 'posted'`.
- **Change: frontend.** Public routes shouldn't expose unpublished events; client-side filtering still sends them to every visitor. The frontend should read drafts from [`GET /clubs/{clubId}/events/drafts`](endpoints/club-events.md#-get-clubsclubideventsdrafts) and [`GET /auth/events/{eventId}`](endpoints/event-management.md#-get-autheventseventid), which the backend needs to build.

### M2. Create-event body: `timezone`, `associates`, `description`

- **Frontend:** `EventForm.tsx` sends `event.{title, location, rsvpLink?, startDate, endDate}`, `description?`, `status`, `flyer?`, and `altText?`. It has no timezone field and no co-host picker (co-hosting is SOON).
- **Backend:** requires `event.timezone`, `description`, and `associates`, so the body fails validation with `400` even with working SQL. No role check runs.
- **Change: backend.** Make `timezone` optional (default `America/New_York`, which the form already assumes), `associates` optional (default `[]`), and `description` optional, and add the e-board-or-owner check. Neither missing field can be collected by the current UI. See the [proposed request](endpoints/club-events.md#-post-clubsclubidevents).

### M3. `400` where the frontend expects `401`, `403`, or `404`

- **Frontend:** `docs/api.md` expects `404` for a missing club or event, `403` when an owner tries to leave, and `404` when leaving a club you're not in. In code, `useClub` and `useEvent` treat any failure as "not found", and join and leave only check `res.ok`, so the visible behavior is the same today.
- **Backend:** returns `400` for a missing club (`GET /clubs/{clubId}`, join), for an owner leaving, and for not being a member. `GET /events/{eventId}` already returns `404`. Leave also returns `400` for a missing token, because its route has no authorizer.
- **Change: backend.** The codes are part of the contract for any client, including the GWC site, and `400` hides real client errors. Add the authorizer to the leave route so a missing token gets `401`.

### M4. The creator doesn't become the club's owner

- **Frontend:** `docs/api.md` says the creator becomes the owner, the mock does so, and `ClubForm.tsx` redirects to the new club page, where managing it requires `owner` or `eboard`.
- **Backend:** `POST /clubs` inserts `clubs` and `club_info` only; nobody can manage the new club.
- **Change: backend.** Insert the owner membership in the same transaction ([query group 29](../database/README.md#29-club-creator-ownership)). Whether club creation should require an admin is still an [open question](README.md#open-questions).

### M5. The access token has no `email` claim

- **Frontend:** sends the access token (`getAccessToken()` in `src/types/auth.ts`) and reads the user's email from the ID-token profile for its own display.
- **Backend:** handlers read `email` from the token. Access tokens don't carry it, so `RequireStudent` creates rows with a `NULL` email, and `GET /me` then fails with `500` on the non-nullable Go model.
- **Change: backend.** The access token is the right token for API calls. Take the email from the Cognito trigger (or the user pool), never from the access token, and make `email` nullable in the model and in `GET /me`.

### M6. No usable image URLs

- **Frontend:** puts `thumbnailUrl` straight into `<img src>`: `fromJsonClub` maps a club's `thumbnailUrl` to `logo`, `fromJsonEvent` maps an event's top-level `thumbnailUrl` to `flyer`, and `owners.*.thumbnailUrl` to the host's logo.
- **Backend:** club reads return the logo's S3 object key as `thumbnailUrl`. Event objects have no top-level `thumbnailUrl` at all, only `thumbnailId` (from `models.Event`), so the frontend never gets a flyer; `owners.*.thumbnailUrl` is a fixed external placeholder image hard-coded in the handlers.
- **Change: backend.** The bucket blocks public access, so only the API can produce a readable URL (a signed GET URL or a CDN URL). Return one as `thumbnailUrl` on clubs, events, and `owners.*`, the field names the frontend normalizers already read. See the [proposed club object](endpoints/clubs.md#proposed-club-object) and [event object](endpoints/events.md#proposed-event-object).

### M7. The club-logo signer returns `key`, not `imageId` and `objectKey`

- **Frontend:** calls no upload route yet; `ClubForm.tsx` sends the logo as a `blob:` URL in the `POST /clubs` body.
- **Backend:** `POST /clubs/{clubId}/thumbnails` returns `{uploadUrl, key}`; the event signers and the PDF's flow return `{uploadUrl, imageId, objectKey}`.
- **Change: backend.** Match the other signers before the frontend writes its upload code, so one upload helper serves all three purposes. See the [proposed response](endpoints/images.md#-post-clubsclubidthumbnails).

### M8. Event objects have empty club names

- **Frontend:** looks up every host club by ID in `GET /clubs?verified=true` (`useClubsById.ts`, used on Home, Events, and the Event page's host). An unverified host club isn't in that list, so its name falls back to "Hunter CS".
- **Backend:** `owners.*.name` is always `""`.
- **Change: backend.** The name is already joined in SQL; copy it into the response. That removes an extra request and the unverified-club gap, and the GWC site gets names without a second call.

### M9. List defaults don't match how the frontend calls lists

- **Frontend:** calls `GET /events`, `GET /clubs/{clubId}/events`, and `GET /me/events` with no parameters and expects every upcoming (and, on the Club page, past) event. Home shows up to 16 tiles.
- **Backend:** defaults to `startDate=1970-01-01` and `limit=10`, counted in joined event-to-club rows, not events. The handler then groups rows in a Go map, so the output isn't sorted even though the SQL orders by start date.
- **Change: both.** Per [decision 3](README.md#decisions), the backend pages by event with `limit` defaulting to 50 (maximum 100), orders upcoming events by start time ascending and past events descending, and puts the flyer and `imageCount` in lists (the gallery only on the single-event read). The frontend asks for `when=upcoming` or `when=past` ([proposed list parameters](endpoints/events.md#proposed-list-parameters)) and pages for older events on the Club page.

### M10. Event status vocabulary, and `cancelled`

- **Frontend:** uses `draft`, `posted`, and `cancelled`. A cancelled event stays listed with a "Cancelled" pill (Events, Club, Event page) and loses its RSVP and calendar actions. `fromJsonEvent` maps any other non-empty status to `draft`, so an `archived` event would show as a draft.
- **Backend:** the `events.status` enum is `drafted`, `posted`, `archived`, and public reads return `posted` only.
- **Change: backend first, then frontend.** Expose `draft`, `posted`, and `cancelled` in the API, mapping the database's `drafted` to `draft`. Add `cancelled` to the enum in a new dated migration. Public reads return `posted` and `cancelled`; `archived` (and anything with `deleted_at` set) never appears publicly. Cancelling is a [`PATCH`](endpoints/event-management.md#-patch-autheventseventid) to `cancelled`; archiving is [`DELETE`](endpoints/event-management.md#-delete-autheventseventid). Allowed transitions follow [decision 2](README.md#decisions): `draft → posted`, `posted → cancelled`, `cancelled → posted`, and any status → `archived`; nothing goes back to `draft`. The frontend should stop mapping unknown statuses to `draft`. See [Event status](endpoints/events.md#event-status).

### M11. Create-event extras: `status`, `flyer`, `altText`, `tags`

- **Frontend:** `EventForm.tsx` sends `status` (`draft` or `posted`), `flyer` (a `blob:` URL that only exists in the browser), and `altText`. `docs/api.md` lists `tags`, but the form has no tag input and doesn't send them, and the frontend `Event` type has no tags field.
- **Backend:** ignores all four and always stores `drafted`. The schema has no column for flyer alt text; `event_tags` exists but nothing reads or writes it.
- **Change: both.** The backend accepts `status`, and an omitted `status` means `draft` ([decision 1](README.md#decisions)). Alt text is stored on the flyer's `images` row ([decision 4](README.md#decisions)), so the frontend sends `altText` with the [flyer confirm](endpoints/images.md#-post-clubsclubideventseventidthumbnailsconfirm), not with the create body. The frontend also stops sending `flyer` and uploads the file through the [flyer flow](endpoints/images.md#how-image-uploads-work) after the event exists. Event tags wait until the UI has them.

### M12. Resuming a draft creates a second event

- **Frontend:** with `?draft=:eventId`, the form prefills from the draft but still submits `POST /clubs/{clubId}/events`, which creates a new event and leaves the draft behind.
- **Backend:** has no update route.
- **Change: frontend, after the backend builds [`PATCH /auth/events/{eventId}`](endpoints/event-management.md#-patch-autheventseventid).** A resumed draft is saved (fields only, no `status`) or posted (`"status": "posted"`) with `PATCH`. That makes `PATCH` **needed now**, even though the Edit buttons that will also use it are Later.

### M13. Club fields: description, tags, member count, logo

- **Frontend:** club cards and the club header show description, topic tags, and member count; `ClubForm.tsx` sends `tags` (up to three from a fixed list) and a `logo` `blob:` URL. Search on the Clubs page matches tags.
- **Backend:** `GET /clubs` returns only `id`, `name`, and `thumbnailUrl`. `GET /clubs/{clubId}` adds `description` and `website_url` but no tags or count. `POST /clubs` ignores `tags` and `logo`, and the schema has no table for club tags.
- **Change: backend** for the fields (and a new `club_tags` table); **frontend** for the logo, which moves to the [logo upload flow](endpoints/images.md#how-image-uploads-work). Topics are a fixed list of lowercase keys, at most 3 per club, uppercased by the UI ([decision 5](README.md#decisions)); the frontend sends keys instead of its uppercase labels. See the [proposed club object](endpoints/clubs.md#proposed-club-object).

### M14. Date and time format

- **Frontend:** sends UTC ISO 8601 strings (`wallTimeToUtcIso` in `src/lib/timezone.ts`, for example `2026-09-15T21:00:00.000Z`). It reads the backend's `YYYY-MM-DD HH:MM:SS` by replacing the space with `T` (`toIso`), which JavaScript parses as the browser's local time.
- **Backend:** returns MySQL `DATETIME` text with no offset. How the handler stores a `Z`-suffixed string hasn't been exercised, because event creation is broken.
- **Change: backend.** Store UTC and return ISO 8601 with an offset (`2026-09-15T21:00:00Z`) in every event read. Otherwise times shift by each viewer's UTC offset. The frontend's `toIso` already passes ISO strings through unchanged.

## Frontend changes

Changes the frontend needs to match the proposed contracts and the [decisions](README.md#decisions). Nothing here has been changed in the frontend repo.

| Change | Where | Why |
| --- | --- | --- |
| The design-mode mock creates a draft when `status` is omitted, instead of `posted`. | `src/mocks/handlers.ts` (`POST /clubs/:clubId/events`) | Decision 1. |
| The design-mode fixtures use the fixed topic list (lowercase keys, at most 3 per club). | `src/mocks/fixtures.ts`, `src/mocks/data.ts` | Decision 5; the fixtures use title-case tags outside the list, such as `Academic` and `Food`. |
| The New club form sends topic keys (lowercase) and displays them uppercased. | `src/pages/ClubForm.tsx` | Decision 5. |
| Past-event rows read the photo count from `imageCount`, not `images.length`. | `src/components/Club/EventsTab.tsx` (`PastEventRow`), `src/types/events.ts` | Decision 3: lists don't carry `images`. |
| Lists ask for `when=upcoming` or `when=past` and page with `limit`/`page`. | `src/hooks/useEvents.ts`, `useClubEvents.ts`, `useMyEvents.ts` | M9, decision 3. |
| Drafts come from the protected drafts list and `GET /auth/events/{eventId}`. | `src/hooks/useMyDrafts.ts`, `src/pages/Club.tsx`, `src/pages/EventForm.tsx` | M1. |
| A resumed draft is saved or posted with `PATCH`, never with `status: "draft"` on a posted event. | `src/pages/EventForm.tsx` | M12, decision 2. |
| The flyer and logo go through the upload flows; `altText` is sent with the flyer confirm. | `src/pages/EventForm.tsx`, `src/pages/ClubForm.tsx` | M7, M11, M13, decision 4. |
| Unknown statuses aren't mapped to `draft`. | `src/types/events.ts` (`fromJsonEvent`) | M10. |

## GWC website

[GWC-Website](https://github.com/GWC-Hunter-College/GWC-Website) makes no API calls. Its Events page is a "Work in Progress" placeholder; the preserved design (`src/pages/events/eventsData.ts`) shows one featured event's title, schedule, and RSVP link. It needs public reads for one club:

| What it shows | Endpoint | Status | Need | Notes |
| --- | --- | --- | --- | --- |
| Club info | [`GET /clubs/{clubId}`](endpoints/clubs.md#-get-clubsclubid) | ✅ | Later | 🟢 public. Logo needs M6. |
| Upcoming events and the featured event (title, schedule, RSVP) | [`GET /clubs/{clubId}/events?when=upcoming`](endpoints/club-events.md#-get-clubsclubidevents) | ✅ | Later | 🟢 public. Needs M8, M9, and M14 to be reliable; `when` is a [proposed list parameter](endpoints/events.md#proposed-list-parameters). |

No new endpoint is needed. CORS already allows all origins. The site finds its club through a configured club ID ([decision 7](README.md#decisions)), so no lookup by name is needed.

## Admin

The planning PDF has an admin page for verifying clubs and managing admins. The frontend has no admin route or admin UI, and doesn't call `GET /me`. So every [admin route](endpoints/admins.md) and [verification write](endpoints/verification.md) is **Later**. When an admin page is built, the frontend needs to know whether the viewer is an admin; the proposal is an `isAdmin` field on [`GET /me`](endpoints/me.md#-get-me) rather than a separate endpoint.

## Schema needs

The proposals above need these schema changes. Each goes in a new dated migration under `database/migrations/schema/`; none is written yet.

| Change | For | Need |
| --- | --- | --- |
| Add `cancelled` to `events.status` | M10 | Now |
| An alt-text column on `images` ([decision 4](README.md#decisions)) | M11 | Now |
| A `club_tags (fk_club_id, tag)` table holding lowercase topic keys, at most 3 per club (enforced by the API) ([decision 5](README.md#decisions)) | M13 | Now |

New queries: member count per club, club tags read and write, and club update. The event update ([group 25](../database/README.md#25-event-update-and-publish)), drafts list ([group 21](../database/README.md#21-club-draft-events)), thumbnail confirm ([group 27](../database/README.md#27-club-and-event-thumbnail-metadata-assignment)), and creator ownership ([group 29](../database/README.md#29-club-creator-ownership)) groups already exist.

## Not covered by the API yet

- Board tab (club photo and note wall)
- Announcements tab
- Event reminders
- Import and export
- Edit history
- Caching
