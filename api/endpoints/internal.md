# Internal functions

🔵 Code that runs outside API Gateway: the Cognito trigger, plus the shared checks that handlers call or are meant to call. None of these has a URL.

A Cognito JWT only establishes identity. It doesn't establish admin, membership, e-board, owner, or event authority; those need the checks below, and today no deployed handler runs any of the role checks.

## 🔵 Cognito student sync trigger

Creates or updates the user's `students` row from Cognito's `sub` and `email`, so the database knows everyone who signs up or signs in. This is the PDF's "Register Lambda".

**Invoked by:** Cognito's `PostConfirmation` trigger (source `PostConfirmation_ConfirmSignUp`) and `PostAuthentication` trigger (source `PostAuthentication_Authentication`), not API Gateway. Other trigger sources are ignored.

**Input:** the Cognito event's `userAttributes.sub` and `userAttributes.email`. The email can be empty.

**Does:** runs [`students/UPSERT_student.sql`](../../infrastructure/legacy/utils/query_client/queries/students/UPSERT_student.sql), which inserts `(sub, email)` or updates the email if the row already exists.

**On failure:** logs the error and returns the Cognito event unchanged, so a database problem never blocks sign-up or sign-in. A missing `sub` skips the write.

**Status:** ✅ Implemented.

**Known issues:**

- Two copies are wired to the same user pool. `PostConfirmUserUpsert` in `AuthorizationStack` writes to `PRODUCTION` only when `PRODUCTION_STATUS` is `true`, otherwise `STAGING`. `PostConfirmUserUpsertDev` in `DevApiStack` always writes to `STAGING`. A user pool holds one function per trigger, so whichever stack updated it last wins, and deleting either stack's custom resource clears both triggers.

**Code:** handler [`lambda/internal/auth/postConfirm/upsert.go`](../../infrastructure/legacy/lambda/internal/auth/postConfirm/upsert.go) · wired by [`authorization.go`](../../infrastructure/legacy/internal/stack/authorization.go) and [`developmentApi.go`](../../infrastructure/legacy/internal/stack/developmentApi.go) · [query group 1](../../database/README.md#1-student-existence-and-upsert)

## 🔵 RequireStudent (request-time student sync)

Makes sure the caller has a `students` row before a protected handler does anything else. It's the fallback for when the Cognito trigger didn't write the row.

**Called by:** all six handlers behind the Cognito authorizer, plus the leave-club handler, which never gets this far (see [`DELETE /clubs/{clubId}/members/me`](memberships.md#-delete-clubsclubidmembersme)).

**Does:**

1. Returns `ErrNoSub` if `sub` is empty.
2. Checks for the row with [`EXISTS_student_by_sub.sql`](../../infrastructure/legacy/utils/query_client/queries/students/EXISTS_student_by_sub.sql).
3. If it's missing, inserts `(sub, email)` with [`UPSERT_student.sql`](../../infrastructure/legacy/utils/query_client/queries/students/UPSERT_student.sql) when there's an email claim, or `(sub, NULL)` with [`UPSERT_student_sub_only.sql`](../../infrastructure/legacy/utils/query_client/queries/students/UPSERT_student_sub_only.sql) when there isn't.

It never updates the email of an existing row.

**Status:** ✅ Implemented.

**Known issues:**

- Handlers map its errors differently: `POST /clubs` returns `400`, the other handlers return `500`, and the `/me` handlers return `400` only for a missing `sub`.
- The frontend sends access tokens, which have no `email` claim, so rows created here get a `NULL` email, and a `NULL` email breaks [`GET /me`](me.md#-get-me).

**Code:** [`utils/auth/ensure_student.go`](../../infrastructure/legacy/utils/auth/ensure_student.go) · [query group 1](../../database/README.md#1-student-existence-and-upsert)

## 🔵 Club role check

`AuthorizeStudentClub(qc, sub, clubId)` returns true if the student is an e-board member **or** owner of the club.

**SQL:** [`authorization/IS_student_authorized_club.sql`](../../infrastructure/legacy/utils/query_client/queries/authorization/IS_student_authorized_club.sql) matches `member_is_eboard = 1 OR member_is_owner = 1`. The database module's version reads the baseline's single role column: `role IN ('eboard', 'owner')` ([query group 13](../../database/README.md#13-club-authorization)).

**Errors:** `ErrNoSub` for an empty `sub`; `ErrBadClub` when `clubId <= 0`.

**Status:** 🟨 Built but **not called by any handler**, so no club write is role-checked today.

**Known issues:**

- It isn't enough for owner-only rules such as [`PUT /clubs/{clubId}/members/roles`](memberships.md#-put-clubsclubidmembersroles).
- A duplicate copy lives in [`lambda/internal/auth/club_authorization/`](../../infrastructure/legacy/lambda/internal/auth/club_authorization/club_authorization.go). The two are identical in purpose and neither is called.

**Code:** [`utils/auth/club_authorization.go`](../../infrastructure/legacy/utils/auth/club_authorization.go) · [query group 13](../../database/README.md#13-club-authorization)

## 🔵 Event role check

`AuthorizeStudentEvent(ctx, qc, sub, eventId)` returns true if the student is an e-board member or owner of **any** club linked to the event, so authorization doesn't depend on a single `clubId`.

**SQL:** [`authorization/IS_student_authorized_event.sql`](../../infrastructure/legacy/utils/query_client/queries/authorization/IS_student_authorized_event.sql).

**Errors:** `ErrNoSub` for an empty `sub`; `ErrBadEvent` when `eventId <= 0`.

**Status:** 🟨 Built but not called by any handler. It's meant for the planned [`/auth/events`](event-management.md) routes.

**Known issues:**

- A duplicate copy lives in [`lambda/internal/auth/event_authorization/`](../../infrastructure/legacy/lambda/internal/auth/event_authorization/event_authorization.go).

**Code:** [`utils/auth/event_authorization.go`](../../infrastructure/legacy/utils/auth/event_authorization.go) · [query group 14](../../database/README.md#14-event-authorization)

## 🔵 Admin check

Answers "is the caller an admin?", meaning "does the caller have a row in `admins`?". Every [admin route](admins.md) needs it (PDF), and so do the [verification writes](verification.md) if they're admin-only.

**Status:** ⬜ Not built. No query or helper reads `admins`.

**Plan:** a shared `is_admin` query in [query group 22](../../database/README.md#22-admin-crud).

## 🔵 Image metadata write (PDF `POST /images`)

The PDF's internal function for storing an image's metadata after it's uploaded to S3. Its planned input is `{ eventId: number | null, clubId: number | null, purpose: string, objectKey: string }`. Per the PDF, `purpose` and `objectKey` are "given to you" and must not be modified.

**Status:** ⬜ Not built as a shared function. The same insert ([`images/INSERT_image.sql`](../../infrastructure/legacy/utils/query_client/queries/images/INSERT_image.sql)) runs inside the event-image [confirm route](images.md#-post-clubsclubideventseventidimagesconfirm), which has no auth.

**Notes:** build it as a shared service that each route-specific confirm handler calls. Don't expose a public generic write just because the PDF's label looks like an HTTP path; it's blue (internal) in the PDF.

**Plan:** PDF "Endpoints Revamp", p. 19 · [query group 15](../../database/README.md#15-event-image-metadata-confirmation)

## Shared helpers

| Helper | What it does | Used by |
| --- | --- | --- |
| [`ExtractSubFromRequest`](../../infrastructure/legacy/utils/auth/extract_sub.go) | Reads `sub` (required) and `email` (optional) from `requestContext.authorizer.jwt.claims`. Fails when there's no authorizer context or no `sub`. | Club creation, event creation, join, and leave. The `/me` handlers read the claims map directly. |
| [`FormatValidationError`](../../infrastructure/legacy/utils/errors/validation_error.go) | Turns validator errors into `{"errors": [{"field", "message"}]}`. | Club creation and event creation. |
