# Verification

Verified clubs are the ones the public club directory shows. The `verified_clubs` table exists, per its schema comment, "for making sure clubs are allowed to be shown to account for bad actors creating random new clubs". Reading verification works; changing it doesn't exist yet.

**Reading:** [`GET /clubs?verified=true`](clubs.md#-get-clubs) (✅) returns only verified clubs. It's a data filter, not an access rule: unverified clubs are still readable through `GET /clubs` and [`GET /clubs/{clubId}`](clubs.md#-get-clubsclubid).

**Who may verify:** the PDF's verification section only marks these routes protected (🔴). The admin requirement comes from the stub note in [`stub/lambda/admins/admins.txt`](../../infrastructure/legacy/stub/lambda/admins/admins.txt), which says admins verify clubs. This is an [open question](../README.md#open-questions).

**Need: Later.** The frontend has no admin page or admin UI, so nothing calls these routes yet; see [coverage.md](../coverage.md#admin). When an admin page exists, it can check [`isAdmin` on `GET /me`](me.md#-get-me) (Proposed) before showing itself.

## 🔴 POST `/clubs/{clubId}/verification`

Marks a club verified by inserting it into `verified_clubs` (PDF).

**Auth:** 🔴 JWT, plus an approved club-administration policy: admin-only per the stub note above, which would use the [admin check](internal.md#-admin-check).

**Path params:** `clubId`.

**Response:** not defined.

**Status:** ⬜ Not built. No route or handler. The schema, the read filter, and the SQL for the insert and the admin check exist ([query group 23](../../database/README.md#23-club-verification-writes)).

**Notes:** verifying a verified club changes nothing: `INSERT_verified_club.sql` affects 0 rows, and `EXISTS_club.sql` tells that apart from an unknown club.

**Queries** (3): 1. [`authorization/admins/is_admin/IS_admin.sql`](../../database/queries/authorization/admins/is_admin/IS_admin.sql) (auth, `403`, if verifying is admin-only ([open question](../README.md#open-questions))) → 2. [`clubs/verification/create/INSERT_verified_club.sql`](../../database/queries/clubs/verification/create/INSERT_verified_club.sql) (write) → on 0 rows: 3. [`clubs/get/EXISTS_club.sql`](../../database/queries/clubs/get/EXISTS_club.sql) (read, 0 is `404`; 1 means already verified). No transaction needed.

**Plan:** PDF "Endpoints Revamp", p. 19 · [query group 23](../../database/README.md#23-club-verification-writes)

## 🔴 DELETE `/clubs/{clubId}/verification`

Removes a club from `verified_clubs` (PDF).

**Auth:** 🔴 JWT, plus the same club-administration policy.

**Path params:** `clubId`.

**Response:** not defined.

**Status:** ⬜ Not built. No route or handler; the delete and the admin check exist as SQL ([query group 23](../../database/README.md#23-club-verification-writes)).

**Notes:** unverifying only changes directory filtering. It doesn't revoke memberships or roles.

**Queries** (3): 1. [`authorization/admins/is_admin/IS_admin.sql`](../../database/queries/authorization/admins/is_admin/IS_admin.sql) (auth, `403`, if unverifying is admin-only) → 2. [`clubs/verification/delete/DELETE_verified_club.sql`](../../database/queries/clubs/verification/delete/DELETE_verified_club.sql) (write) → on 0 rows: 3. [`clubs/get/EXISTS_club.sql`](../../database/queries/clubs/get/EXISTS_club.sql) (read, 0 is `404`; 1 means the club wasn't verified). No transaction needed.

**Plan:** PDF "Endpoints Revamp", p. 19 · [query group 23](../../database/README.md#23-club-verification-writes)
