# Admins

All planned. The PDF says **every** admin route must first check that the caller is an admin. The schema has an `admins (fk_student_id)` table, but there's no admin route, handler, or [admin check](internal.md#-admin-check). According to the stub note in [`stub/lambda/admins/admins.txt`](../../infrastructure/legacy/stub/lambda/admins/admins.txt), admins are the people who verify clubs, so random users can't flood the directory with fake ones.

The PDF shows ✅ next to two of these routes. Those marks tracked task progress; no admin route exists in the code.

Student IDs are Cognito `sub` UUID strings (`students.id` is `CHAR(36)`).

## 🔴 GET `/admins`

Lists admins, "their id and clubs" (PDF).

**Auth:** 🔴 JWT + admin.

**Response:** not defined. Open: whether "clubs" means clubs they manage, clubs they've joined, or a separate response.

**Status:** ⬜ Not built. Only planning text and the schema exist.

**Plan:** PDF "Endpoints Revamp", p. 18 · [query group 22](../../database/README.md#22-admin-crud)

## 🔴 GET `/admins/{studentId}`

Returns one admin's record, if there is one.

**Auth:** 🔴 JWT + admin.

**Path params:** `studentId`.

**Response:** not defined.

**Status:** ⬜ Not built. There's only an unwired SQL stub, [`GET_admins_studentId.sql`](../../infrastructure/legacy/stub/lambda/admins/studentId/GET_admins_studentId.sql), which hard-codes the obsolete numeric student ID `2`. Treat it as design notes, not working SQL.

**Plan:** PDF "Endpoints Revamp", p. 19 · [query group 22](../../database/README.md#22-admin-crud)

## 🔴 POST `/admins`

Promotes a student to admin by inserting their ID into `admins`.

**Auth:** 🔴 JWT + admin. It must also check that the target student exists.

**Request body:** not defined.

**Response:** not defined.

**Status:** ⬜ Not built. No route, handler, insert query, or admin check.

**Open:** how the first admin gets created, and how to make sure the last admin can't be removed.

**Plan:** PDF "Endpoints Revamp", p. 19 · [query group 22](../../database/README.md#22-admin-crud)

## 🔴 DELETE `/admins/{studentId}`

Demotes an admin by deleting their `admins` row.

**Auth:** 🔴 JWT + admin.

**Path params:** `studentId`.

**Response:** not defined.

**Status:** ⬜ Not built. There's only an unwired SQL stub, [`DELETE_admins_studentId.sql`](../../infrastructure/legacy/stub/lambda/admins/studentId/DELETE_admins_studentId.sql), with the same hard-coded ID `2`.

**Open:** whether admins may demote themselves, and what protects the last admin.

**Plan:** PDF "Endpoints Revamp", p. 19 · [query group 22](../../database/README.md#22-admin-crud)
