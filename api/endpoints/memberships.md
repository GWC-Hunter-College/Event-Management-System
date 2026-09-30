# Memberships

Joining and leaving clubs, plus the planned routes for listing members and changing roles. Membership and role flags live in `club_members`; see [Roles](../README.md#roles).

Routes are registered in [`club_routes.go`](../../infrastructure/legacy/gateway/routes/club_routes.go). Response shapes come from the handlers' response maps; example values are made up, following the [implementation's club docs](../../infrastructure/legacy/docs/api/clubs.md).

## 🔴 POST `/clubs/{clubId}/members/me`

Joins the caller to a club as a regular member. The frontend's Club page calls it from the Join button.

**Auth:** 🔴 JWT. Caller-scoped: the student ID comes from the token's `sub`, never from the request, so nobody can add someone else. No role is needed, and the club doesn't have to be verified.

**Path params:** `clubId`, a required integer.

**Request body:** none.

**Response `200`:**

```json
{ "message": "Successfully joined club", "clubId": 7, "joined": true }
```

**Errors:**

| Status | Body | When |
| --- | --- | --- |
| `401` | From API Gateway | Token missing or invalid. |
| `400` | `{"error": "Error extracting sub from request: <reason>"}` | No claims or `sub`. |
| `400` | `{"error": "clubId must be a valid integer"}` | `clubId` isn't an integer (an empty one gets `clubId is required in path parameters`). |
| `400` | `{"error": "Unable to join club. You may already be a member or the club does not exist."}` | Already a member, or no such club. |
| `500` | `{"error": "Error ensuring student exists: <reason>"}` or `{"error": "Error joining club"}` | `RequireStudent` or the insert failed. |

**Status:** ✅ Implemented.

**Known issues:**

- Both role flags are inserted as `false`, and no endpoint can raise them yet (see [`PUT /clubs/{clubId}/members/roles`](#-put-clubsclubidmembersroles)).
- The frontend's `docs/api.md` expects `404` for a missing club; the backend returns `400`.
- Deploy risk: both API stacks declare this Lambda with the same fixed name, `PostJoinClubMemberMe`, so they can't both deploy it in one account and region. The repo doesn't show which API serves the route today.

**Proposed changes** (Need: **Now**; screen: Club page "+ JOIN CLUB"): `404` for a missing club and `409` for an existing membership, instead of `400` for both ([M3](../coverage.md#m3-400-where-the-frontend-expects-401-403-or-404)). Request and `200` response unchanged; the frontend only checks `res.ok` and then re-reads `GET /me/clubs`.

**PDF path:** the planning PDF writes this route as `POST /clubs/{clubId}/members` ("join a club as a member"). By design it's self-join only; no endpoint adds a different student ([query group 19](../../database/README.md#19-add-a-specified-club-member)).

**Queries** (2): 1. [`clubs/get/EXISTS_club.sql`](../../database/queries/clubs/get/EXISTS_club.sql) (read, `404`) → 2. [`clubs/members/create/INSERT_club_member.sql`](../../database/queries/clubs/members/create/INSERT_club_member.sql) (write, 0 rows is `409`). No transaction: the insert is guarded on its own.

**Code:** route [`club_routes.go`](../../infrastructure/legacy/gateway/routes/club_routes.go) · handler [`clubs/clubId/members/me/post/post.go`](../../infrastructure/legacy/lambda/api/clubs/clubId/members/me/post/post.go) · SQL [`clubs/INSERT_club_member.sql`](../../infrastructure/legacy/utils/query_client/queries/clubs/INSERT_club_member.sql) · [query group 8](../../database/README.md#8-join-caller-to-club)

## 🟢 DELETE `/clubs/{clubId}/members/me`

Leaves a club as the caller; an owner can't leave. The frontend's Club page calls it from the Leave button, after a confirm dialog. **It's broken:** the route has no authorizer, so the handler never gets a `sub` and returns `400` on every call.

**Auth:** 🟢 on the deployed route (no Cognito authorizer), but the handler requires JWT claims. The PDF intended 🔴, self-delete only. Sending a bearer token doesn't help, because API Gateway only passes claims on routes that have the authorizer.

**Path params:** `clubId`, a required integer.

**Request body:** none.

**Response `200`** (only when the handler receives claims, which the deployed route never provides):

```json
{ "message": "Successfully left club", "clubId": 7, "left": true }
```

**Errors:**

| Status | Body | When |
| --- | --- | --- |
| `400` | `{"error": "Error extracting sub from request: Request context has no authorizer"}` | Every call through the deployed route. |
| `400` | `{"error": "clubId must be a valid integer"}` | Bad `clubId` (only reachable with claims). |
| `400` | `{"error": "Unable to leave club. You may not be a member or you are the club owner."}` | No row deleted (only reachable with claims). |
| `500` | `{"error": "Error ensuring student exists: <reason>"}` or `{"error": "Error leaving club"}` | `RequireStudent` or the delete failed (only reachable with claims). |

**Status:** 🟨 Broken: the route registration omits the Cognito authorizer.

**Known issues:**

- Fix: add the Cognito authorizer to this route's registration in `club_routes.go`.
- The dev API's CORS config doesn't allow `DELETE` (prod's does), so browsers block this call against dev at the preflight.
- The SQL owner guard is `member_is_owner = FALSE`, which doesn't match `NULL`, so a membership with a null owner flag can't be deleted.
- The frontend expects `401`, `403` (owner), and `404` (not a member); the handler uses `400` for all of them.
- Same deploy risk as joining: both API stacks declare the fixed Lambda name `DeleteClubMemberMe`.

**Proposed changes** (Need: **Now**; screen: Club page "Leave club" in the JOINED menu, after the confirm modal):

- **Auth:** 🔴 JWT: add the Cognito authorizer to the route. Allow `DELETE` in the dev API's CORS config.
- **Response `200`:** unchanged, `{"message": "Successfully left club", "clubId": 7, "left": true}`.
- **Errors:** `401` no token (API Gateway), `403` the caller is the club's last owner, `404` the caller isn't a member or the club doesn't exist, `500` ([M3](../coverage.md#m3-400-where-the-frontend-expects-401-403-or-404)).
- **Owners:** an owner can leave while the club has another owner; the last owner can't ([decision 11](../README.md#decisions)). The deployed SQL refuses every owner.

**Queries** (3; all in one transaction): 1. [`clubs/members/update_role/SELECT_club_owners_for_update.sql`](../../database/queries/clubs/members/update_role/SELECT_club_owners_for_update.sql) (read, locks the owner rows) → 2. [`clubs/members/leave/DELETE_club_member.sql`](../../database/queries/clubs/members/leave/DELETE_club_member.sql) (write) → on 0 rows: 3. [`authorization/clubs/is_member/IS_club_member.sql`](../../database/queries/authorization/clubs/is_member/IS_club_member.sql) (read, 0 is `404`; 1 is `403`, the caller is the last owner)

**Code:** route [`club_routes.go`](../../infrastructure/legacy/gateway/routes/club_routes.go) (no `Authorizer` on this registration) · handler [`clubs/clubId/members/me/delete/delete.go`](../../infrastructure/legacy/lambda/api/clubs/clubId/members/me/delete/delete.go) · SQL [`clubs/DELETE_club_member.sql`](../../infrastructure/legacy/utils/query_client/queries/clubs/DELETE_club_member.sql) · [query group 9](../../database/README.md#9-leave-callers-club)

## 🔴 GET `/clubs/{clubId}/members`

Lists a club's members from `club_members`, paginated if needed (PDF). No frontend screen lists members yet: the Club page's Manage tab lists events only, and its Board tab is a planned photo-and-notes wall, not a member list.

**Auth:** 🔴 JWT + e-board or owner of the club (PDF), using the [club role check](internal.md#-club-role-check).

**Path params:** `clubId`. Pagination isn't defined.

**Response:** not defined. Still open: which fields to return, how to paginate, and whether ordinary members may see any of it.

**Proposed contract** (Need: **Later**; screen: a member list in the Club page's Manage tab, not built):

- **Query params:** `role` (optional: `member`, `eboard`, or `owner`), `limit`, `page`.
- **Response `200`:**

  ```json
  {
    "message": "Succesfully fetched 2 members",
    "members": [
      { "studentId": "11111111-2222-3333-4444-555555555555", "email": "student@example.edu", "firstName": "Ada", "lastName": "L.", "role": "owner" },
      { "studentId": "aaaaaaaa-bbbb-cccc-dddd-eeeeeeeeeeee", "email": null, "firstName": null, "lastName": null, "role": "member" }
    ]
  }
  ```

  `role` is the member's one role (`club_members.role` in the baseline schema, [decision 11](../README.md#decisions)). With `role=eboard`, the list has e-board members and owners. Names come from `student_info`, which nothing writes yet, so they're usually `null`. Owners and e-board members see `email`; a regular member never does ([decision 6](../README.md#decisions)). Since this route is limited to e-board and owners, its callers always see emails. If the list is ever opened to members, `email` is left out for them.
- **Errors:** `401`, `403` not e-board or owner, `404` unknown club.

**Status:** ⬜ Not built. No route or handler; there's only placeholder text in [`stub/lambda/clubs/clubId/members/`](../../infrastructure/legacy/stub/lambda/clubs/clubId/members/). The SQL exists ([query group 18](../../database/README.md#18-club-member-and-e-board-listing)).

**Queries** (3): 1. [`clubs/get/EXISTS_club.sql`](../../database/queries/clubs/get/EXISTS_club.sql) (read, `404`) → 2. [`authorization/clubs/can_manage/IS_student_authorized_club.sql`](../../database/queries/authorization/clubs/can_manage/IS_student_authorized_club.sql) (auth, `403`) → 3. [`clubs/members/list/SELECT_club_members.sql`](../../database/queries/clubs/members/list/SELECT_club_members.sql) (read, `role` filter from the query string, or `NULL`)

**Plan:** PDF "Endpoints Revamp", p. 16 · [query group 18](../../database/README.md#18-club-member-and-e-board-listing)

## 🔴 GET `/clubs/{clubId}/eboard`

Lists a club's e-board members and owners (PDF). The PDF itself asks whether this should be a separate endpoint.

**Auth:** 🔴 JWT + e-board or owner of the club (PDF).

**Path params:** `clubId`.

**Response:** not defined. A role filter on the members list would avoid a second copy of the member-list SQL.

**Proposed** (Need: **Later**; no screen lists the e-board yet): serve this as [`GET /clubs/{clubId}/members?role=eboard`](#-get-clubsclubidmembers), returning e-board members and owners, rather than a separate route. If the route is kept for the PDF's sake, it returns the same `members` shape.

**Status:** ⬜ Not built. No route or handler; only placeholder material. It reuses the member-list SQL ([query group 18](../../database/README.md#18-club-member-and-e-board-listing)).

**Queries** (3): 1. [`clubs/get/EXISTS_club.sql`](../../database/queries/clubs/get/EXISTS_club.sql) (read, `404`) → 2. [`authorization/clubs/can_manage/IS_student_authorized_club.sql`](../../database/queries/authorization/clubs/can_manage/IS_student_authorized_club.sql) (auth, `403`) → 3. [`clubs/members/list/SELECT_club_members.sql`](../../database/queries/clubs/members/list/SELECT_club_members.sql) (read, `role` `eboard`: e-board members and owners)

**Plan:** PDF "Endpoints Revamp", p. 16 · [query group 18](../../database/README.md#18-club-member-and-e-board-listing)

## 🔴 PUT `/clubs/{clubId}/members/roles`

Promotes a member to e-board or owner, or demotes them (PDF: `?is_eboard=<TRUE|FALSE>&is_owner=<TRUE|FALSE>`). In the PDF's page plan, owners do this from the club's member and e-board lists.

**Auth:** 🔴 JWT + **owner only** ([decision 6](../README.md#decisions); PDF: "ONLY OWNERS CAN PROMOTE MEMBERS TO EBOARD OR OWNERS", with one query to check ownership and another to update). The existing [club role check](internal.md#-club-role-check) accepts e-board *or* owner, so it isn't enough on its own.

**Params (PDF):** path `clubId`; query `is_eboard` and `is_owner`. Nothing in the PDF's parameters says *which* member to change.

**Proposed contract** (Need: **Later**; screen: promote and demote controls in the Manage tab, not built). Names the member in the body, and takes one role value, which the baseline schema stores in one `role` column, so a member always has exactly one role ([decision 11](../README.md#decisions)):

- **Request body:** `{ "studentId": "aaaaaaaa-bbbb-cccc-dddd-eeeeeeeeeeee", "role": "eboard" }`, where `role` is `member`, `eboard`, or `owner`.
- **Response `200`:** `{"message": "Updated role", "clubId": 7, "studentId": "aaaaaaaa-bbbb-cccc-dddd-eeeeeeeeeeee", "role": "eboard"}`.
- **Errors:** `400` bad role, `401`, `403` caller isn't an owner, `404` the student isn't a member, `409` the change would leave the club with no owner.

**Response:** not defined.

**Status:** ⬜ Not built. No route or handler. The owner-only check and the role update exist as SQL ([query group 20](../../database/README.md#20-update-member-roles)).

**Queries** (5; steps 3–5 in one transaction): 1. [`clubs/get/EXISTS_club.sql`](../../database/queries/clubs/get/EXISTS_club.sql) (read, `404`) → 2. [`authorization/clubs/is_owner/IS_club_owner.sql`](../../database/queries/authorization/clubs/is_owner/IS_club_owner.sql) (auth, the caller; `403`) → 3. [`clubs/members/update_role/SELECT_club_owners_for_update.sql`](../../database/queries/clubs/members/update_role/SELECT_club_owners_for_update.sql) (read, locks the owner rows) → 4. [`clubs/members/update_role/UPDATE_club_member_role.sql`](../../database/queries/clubs/members/update_role/UPDATE_club_member_role.sql) (write) → on 0 rows: 5. [`authorization/clubs/is_member/IS_club_member.sql`](../../database/queries/authorization/clubs/is_member/IS_club_member.sql) (read, the target: 0 is `404`; otherwise the locked owner rows tell "already in that role" (`200`) from "last owner" (`409`))

**Plan:** PDF "Endpoints Revamp", p. 16 · [query group 20](../../database/README.md#20-update-member-roles)
