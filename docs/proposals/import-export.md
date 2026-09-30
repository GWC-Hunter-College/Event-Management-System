# Import and export

**Proposed, 2026-09-30.** Nothing here is built. Every route is ⬜ and **Proposed**, in the [api/ reference](../../api/README.md#legend) style.

**Depends on:** the core API's role and admin checks; the read queries each sheet reuses. Optional: [notifications](notifications.md) ("your export is ready") and [edit history](edit-history.md) (the record of who exported what).
**Needed by:** nothing.

## Problem

**Export.** E-boards turn over every year, and a club's record (its roster, its past events, its photos) should leave with the club, not with whoever held the password. Officers also get asked for rosters and event lists by Student Activities, and clubs sometimes move to another tool. Admins want a copy of everything the platform holds.

**Import.** A club that already keeps its roster and events in a spreadsheet wants to start with that data rather than type it in. And an admin needs a way to put a club back from a backup: after a purge, after a mistake beyond the 30-day restore window, or to move a club from `STAGING` to `PRODUCTION`.

**Who uses it and how:**

- **A club's e-board or owner** exports their own club from Club → Manage → "Export club data": club info, members, events with descriptions, announcements, and images. A minute later they download a `.zip`.
- **An admin** exports any club, or every club.
- **An admin** (first) imports a `.zip` or `.xlsx`: uploads it, reads a dry-run preview of what would be created and what's wrong, fixes the file or confirms, and gets a new club.

## The round-trip format

One format for export and import, so an export is a valid import:

```text
girls-who-code-at-hunter-2026-09-30.zip
├── manifest.json     format version, source, options, counts, file checksums
├── club.xlsx         every table, one sheet each
└── images/
    ├── 1f0c8a4e-….png   the club logo
    ├── 7d2b19c0-….jpg   a flyer
    └── …                one file per stored image, named by its image id
```

**Why a `.zip` of an `.xlsx` and images.** Images don't fit in cells, and officers edit rosters in Excel or Google Sheets, not JSON. One workbook with several sheets keeps related tables together, where CSV would need a file per table and still trips over encodings, line breaks inside descriptions, and Excel's type guessing. `manifest.json` is for the importer, not for people. An import may also be a bare `club.xlsx`, with no images.

**`manifest.json`:**

```json
{
  "format": "hunter-clubs-export",
  "version": 1,
  "exportedAt": "2026-09-30T18:04:11Z",
  "source": { "environment": "PRODUCTION", "clubId": 2 },
  "scope": "club",
  "options": { "includeImages": true, "includeEmails": false },
  "counts": { "members": 58, "events": 41, "eventPhotos": 312, "announcements": 23, "images": 355 },
  "files": [ { "path": "images/1f0c8a4e-….png", "bytes": 48213, "sha256": "…" } ]
}
```

### Sheets

The first sheet, **README**, explains the file to a person and is ignored on import. Then:

| Sheet | One row per | Columns |
| --- | --- | --- |
| **Club** | the club (one row) | `name`, `description`, `website_url`, `topics` (up to three keys, `;`-separated), `logo` (a path in `images/`), `logo_alt_text`, and export-only `verified` |
| **Members** | membership | `email` (only with `includeEmails`), `first_name`, `last_name`, `username`, `role` (`member`, `eboard`, `owner`), `student_id` (the Cognito `sub`, for restores) |
| **Events** | event the club hosts or co-hosts | `ref` (the source event id), `title`, `status` (`draft`, `posted`, `cancelled`), `start`, `end`, `timezone`, `location`, `rsvp_link`, `description`, `flyer`, `flyer_alt_text`, `hosted_as` (`host` or `co-host`), `co_hosts` (club names), export-only `created_at`, `updated_at` |
| **Event photos** | gallery image | `event_ref`, `image`, `alt_text`, `added_at` (gallery order) |
| **Announcements** | announcement the club owns or shares | `ref`, `title`, `body`, `status`, `posted_at`, `posted_as` (`owner` or `shared`), `also_posted_to` (club names), export-only `created_at` |
| **Images** | stored file | `file`, `image_id`, `filename`, `mimetype`, `alt_text`, `purpose`, `uploaded_at`, `bytes`, `sha256` |
| **Board pins** (later) | approved pin | `kind`, `image`, `note_text`, `event_ref`, `caption`, `style`, `x`, `y`, `width`, `height`, `rotation`, `stack_order`, `approved_at` |

The export reads through the same views and queries the API uses (`club_details`, `event_details`, `announcement_details`, the gallery query), so it holds exactly what the app shows its e-board, and never soft-deleted rows.

### Conventions

- **References, not database ids.** `ref` identifies a row within the file; `event_ref` points at it. Import creates new ids for everything. The export writes source ids as refs, so a round trip is stable.
- **Event times** are written the way people read them: `start` and `end` as local wall time in the row's `timezone` (`2026-10-14 18:00`, `America/New_York`), stored as text cells. The importer converts them to UTC ([decision 2](../../database/docs/schema-review.md#decisions)). It also accepts a date-typed Excel cell (read as wall time in that zone) and full ISO 8601 with an offset. A local time that doesn't exist or happens twice at a daylight-saving change is an error asking for an offset.
- **Other timestamps** (`posted_at`, `created_at`) are ISO 8601 UTC with `Z`, as in the API.
- **Lists in a cell** are `;`-separated. Club names in `co_hosts` and `also_posted_to` are matched to existing clubs by name, under the same case- and accent-insensitive collation as `clubs.name`.
- **Images** are referenced by their path in `images/`. A file used twice (a flyer that's also in the gallery) is stored once and referenced twice, as in the database ([decision 1](../../database/docs/schema-review.md#decisions)).
- **Text safety.** Every value is written as a string cell, never a formula, so a title like `=HYPERLINK(…)` can't run when an officer opens the file. The importer refuses formula cells.
- **Cell limit.** Excel holds at most 32,767 characters in a cell, but `event_descriptions.description` and `club_info.description` are `TEXT` (about 64 KB). Proposed: cap event descriptions at 10,000 characters and club descriptions at 5,000 in the API's validation, like the announcement body's 5,000 ([announcements.md](announcements.md#composer)). Otherwise a long description can't round-trip.

## Export

### How large exports run

| Export | Size at Hunter scale |
| --- | --- |
| One club, data only | Under 1 MB: 60 members, 40 events, 25 announcements. |
| One club, with images | Typically 0.3–2 GB after a few years of galleries (a few hundred photos at ~2.5 MB); the largest clubs may reach 5 GB. |
| Every club, data only | A few MB. |
| Every club, with images | Tens of GB. Not offered: see below. |

**Recommendation: every export is an async job that writes a `.zip` to S3 and returns a presigned download link.** A direct response is ruled out for anything with images: API Gateway HTTP APIs time out after 30 seconds, and a Lambda's synchronous response is capped at 6 MB. A data-only workbook would fit, but one code path for every export is simpler to authorize, audit, and test, and a data-only job finishes in a few seconds anyway.

```text
Browser ── POST /clubs/{clubId}/exports ──> API Lambda (in the VPC)
                                              ├─ INSERT data_jobs (queued)
                                              └─ PUT s3://…/jobs/{jobId}.json   (S3 gateway endpoint)
S3 event ──> Data-jobs worker Lambda (in the VPC)
              ├─ reads RDS through the same queries the API uses
              ├─ streams club.xlsx and images/ into a zip
              └─ multipart upload ──> s3://…/exports/{clubId}/{jobId}.zip
Browser ── GET /exports/{jobId} (polls every 2 s) ──> status, then a 15-minute download URL
```

- **The trigger is an S3 event,** not a queue or a direct Lambda invoke. The legacy network, which the new modules are expected to keep, has no NAT gateway: Lambdas reach Secrets Manager through an interface endpoint and S3 through a free gateway endpoint, and nothing else. Invoking a Lambda or sending to SQS from inside the VPC would need another interface endpoint (about $7.30 a month). Writing a small request object to S3 costs nothing, and S3 invokes the worker directly. The upload of an import `.zip` is the same kind of event.
- **Streaming.** The worker writes the zip straight into an S3 multipart upload: Go's `archive/zip` over an `io.Pipe` into the SDK's upload manager, with images copied from `GetObject` bodies as they're read. Nothing is held in memory or on `/tmp`. Images are stored uncompressed in the zip (JPEG and PNG don't compress further); the workbook is deflated. Go writes Zip64 on its own past 4 GB or 65,535 files.
- **Limits.** A Lambda runs at most 15 minutes. At a conservative 50 MB/s from S3 in-region that's over 40 GB, so any one club fits. Memory: 1,024–2,048 MB. A job still `running` after 20 minutes is marked `failed` the next time anyone reads it.
- **Download.** `GET /exports/{jobId}` signs a fresh 15-minute URL on every call, with `Content-Disposition: attachment` and a readable filename, so a pasted link dies quickly. Presigned URLs are never logged.
- **Clean-up.** An S3 lifecycle rule deletes `exports/` objects after 7 days and `jobs/` objects after a day; the job row becomes `expired`.
- **Admin, every club:** data only, one workbook with a `club` column on each sheet. Images for every club would be tens of GB in one file; an admin who needs them exports clubs one at a time. **Exports aren't backups:** RDS automated backups and S3 versioning are. The legacy stack keeps RDS backups for one day ([`database.go`](../../infrastructure/legacy/internal/stack/database.go)); seven is the usual minimum, and backup storage up to the database's size costs nothing extra.

## Import

### Modes

| Mode | For | Who (proposed) | What it keeps |
| --- | --- | --- | --- |
| `restore` | Putting a club back from one of our exports. | Admins | Statuses, `posted_at`, `created_at`, gallery order, memberships of students who still exist. |
| `create` | A new club from a spreadsheet made from our template. | Admins first; any signed-in student later, if club creation stays open to everyone ([open question](../../api/README.md#open-questions)). | Statuses as given (default `draft`); new timestamps. No memberships except the importer's. |

### Flow

```text
1. POST /imports {mode}                  → importId and a presigned POST form (size-limited)
2. Browser uploads the .zip straight to s3://…/imports/{importId}/upload.zip
3. S3 event → worker: validate everything, then a dry run → report on the job row (validated or invalid)
4. GET /imports/{importId}               → the preview: counts, errors, warnings, sample rows
5. POST /imports/{importId}/commit       → worker: re-check, store images, one transaction → the new club
```

**Upload.** A presigned **POST**, not PUT, because only a POST policy can cap the size (`content-length-range`, up to 2 GB) and pin the key and content type. The browser uploads directly, as it does for images.

**Validation** checks every row against the same rules as the API, and reports errors by sheet, row, and column:

- **The zip itself:** only `manifest.json`, one `.xlsx`, and files under `images/`; no `..`, absolute paths, or links (zip slip); at most 5,000 entries; at most 4 GB uncompressed in total, checked while reading so a zip bomb stops early; the manifest's checksums match.
- **Club:** a name that's free (`clubs.name` is unique; `409` in the report, fixable with a new name at commit); up to three known topics; an `http(s)` website.
- **Events:** title 1–255 characters; a status of `draft`, `posted`, or `cancelled`; a posted or cancelled event needs a location, as posting does ([S3](../../database/docs/schema-review.md#s3-allowed-transitions)); `end` not before `start`; a valid IANA zone; description within the cap.
- **Announcements:** a posted one needs a body; title and body within their limits.
- **Images:** every referenced file is present; type checked from its first bytes (JPEG, PNG, WebP; no SVG); at most 15 MB each; alt text within 1,000 characters.
- **Warnings, not errors:** co-host names that match no club (the link is dropped); rows with `hosted_as = co-host` or `posted_as = shared` (they belong to another club and are skipped); members who aren't on the platform yet.

**The dry run** then runs every insert in one transaction and **rolls it back**, so the database's own rules (unique name, topic keys, the `CHECK`s on dates and `posted_at`) are tested for real without keeping anything. It writes no files. The rolled-back inserts leave gaps in auto-increment ids, which is harmless.

**Commit** re-runs validation (a club may have taken the name since), checks the upload's SHA-256 against the one it validated, then:

1. **Files first.** Each image gets a new image id and is written to `imported/{importId}/{imageId}.{ext}`, outside the prefixes that lifecycle rules expire. Object keys are opaque once stored ([I3](../../database/docs/schema-review.md#i3-required-image-columns)); only the upload confirms check a prefix.
2. **Then one short transaction:** the club, `club_info`, topics, the owner membership, events, descriptions, club links, image rows (`fk_club_id` the new club, `fk_uploaded_by` the importer, original `purpose` and alt text), flyers, gallery links, announcements and their links, and an `imported` [history](edit-history.md) entry. Holding the transaction open during minutes of uploads would lock the tables it touches, so files come first.
3. **If the transaction fails,** the worker deletes the files it wrote. Any it misses are orphans nothing references, which I4 already treats as harmless.

### Members on import

People can't be enrolled in a club without doing anything, and most of a roster may never have signed in, so they have no `students` row to link to.

- **`create`:** the importer becomes the owner, as with `POST /clubs` ([M4](../../api/coverage.md#m4-the-creator-doesnt-become-the-clubs-owner)). The **Members** sheet creates no memberships. The report says "58 people listed; share the club link so they can join." A later phase can add invitations: a `club_invitations (fk_club_id, email, role, fk_invited_by, expires_at)` table, consumed when that student next signs in and accepts.
- **`restore`:** rows whose `student_id`, or else `email`, matches an existing student become memberships with their role. At least one owner must result; if none matches, the commit body names one (`ownerStudentId`).
- `verified` is never imported: verifying a club is an admin action.

## Data model

```sql
-- Import and export jobs. The API creates the row, the data-jobs worker does the
-- work and records the result. object_key is the zip in S3 (under exports/ or
-- imports/, which lifecycle rules delete after 7 days). report holds the counts and,
-- for an import, the dry run's errors, warnings and preview.
-- One export per club (or one all-clubs export), and one import per requester, may be
-- queued or running at a time. Only those rows get a key in the functional UNIQUE index.
CREATE TABLE `data_jobs` (
  `id` CHAR(36) NOT NULL,
  `kind` ENUM('export', 'import') NOT NULL,
  `scope` ENUM('club', 'all') NOT NULL DEFAULT 'club',
  `import_mode` ENUM('create', 'restore') NULL,
  `fk_club_id` INT NULL,
  `fk_requested_by` CHAR(36) NULL,
  `options` JSON NOT NULL,
  `status` ENUM('queued', 'running', 'validated', 'invalid', 'succeeded', 'failed', 'cancelled', 'expired') NOT NULL DEFAULT 'queued',
  `object_key` VARCHAR(255) NULL,
  `object_sha256` CHAR(64) NULL,
  `report` JSON NULL,
  `error` VARCHAR(1000) NULL,
  `created_at` TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
  `started_at` TIMESTAMP NULL,
  `finished_at` TIMESTAMP NULL,
  `expires_at` TIMESTAMP NULL,
  PRIMARY KEY (`id`),
  UNIQUE KEY `uq_data_jobs_one_active` ((IF(`status` IN ('queued', 'running'),
    CONCAT(`kind`, ':', IF(`kind` = 'export', CONCAT(`scope`, ':', COALESCE(`fk_club_id`, 0)), COALESCE(`fk_requested_by`, ''))),
    NULL))),
  KEY `idx_data_jobs_club` (`fk_club_id`, `created_at`),
  KEY `idx_data_jobs_requested_by` (`fk_requested_by`, `created_at`),
  CONSTRAINT `fk_data_jobs_club` FOREIGN KEY (`fk_club_id`) REFERENCES `clubs` (`id`),
  CONSTRAINT `fk_data_jobs_requested_by` FOREIGN KEY (`fk_requested_by`) REFERENCES `students` (`id`) ON DELETE SET NULL
);
```

- `fk_club_id` is the exported club, or the club an import created (set at commit). `RESTRICT` until club deletion is designed ([F1](../../database/docs/schema-review.md#f1-on-delete-behavior)).
- Job rows are small and kept; the files expire. The notification types `export.ready` and `import.finished` point here through `notifications.fk_data_job_id` ([notifications.md](notifications.md#ddl-sketch)).
- No existing table changes.

## Endpoints

All Proposed, all ⬜. Responses use `200`, as every create does today ([Responses](../../api/README.md#responses)).

| Access | Method | Path | What it does | Role | Phase |
| --- | --- | --- | --- | --- | --- |
| 🔴 | POST | `/clubs/{clubId}/exports` | Starts an export of one club. | E-board or owner of `clubId`, or an admin | 1 |
| 🔴 | GET | `/clubs/{clubId}/exports` | The club's exports from the last 7 days: who, when, options, status. No links. | E-board or owner of `clubId`, or an admin | 1 |
| 🔴 | GET | `/exports/{exportId}` | Status; when done, a 15-minute download URL. | The student who asked, or an admin | 1 |
| 🔴 | POST | `/admins/exports` | Starts a data-only export of every club. | Admin | 4 |
| 🔴 | POST | `/imports` | Starts an import and returns an upload form. | Admin (phase 3); any signed-in student for `create`, if decided (phase 4) | 3 |
| 🔴 | GET | `/imports/{importId}` | Status and the dry-run report. | The student who started it, or an admin | 3 |
| 🔴 | POST | `/imports/{importId}/commit` | Imports what the dry run validated. | The student who started it | 3 |
| 🔴 | DELETE | `/imports/{importId}` | Discards an import and its upload. | The student who started it, or an admin | 3 |
| 🟢 | GET | `/imports/template` | The blank template (a README sheet and empty sheets). Could equally be a static file on the frontend. | Public | 4 |
| 🔵 | Lambda | Data-jobs worker | Runs exports, dry runs, and commits, from S3 events. | — | 1 |

### 🔴 POST `/clubs/{clubId}/exports`

**Auth:** 🔴 JWT + e-board or owner of `clubId`, or an admin. `403` otherwise; `404` for an unknown club.

**Request body (Proposed):**

```json
{ "includeImages": true, "includeEmails": false, "confirmEmailUse": false }
```

| Field | Rule |
| --- | --- |
| `includeImages` | Default `true`. Phase 1 accepts only `false`. |
| `includeEmails` | Default `false`. When `true`, `confirmEmailUse` must be `true`: the UI shows "I'll use member emails only for club business" and sends the checkbox. |

**Response `200` (Proposed):** `{ "message": "Export started", "exportId": "…", "status": "queued" }`.

**Errors:** `400`, `401`, `403`, `404`, `409` an export of this club is already running, `500`.

**Writes:** the `data_jobs` row and an `exported` [history](edit-history.md) entry naming the options (phase 2 of edit history), in one transaction; then the S3 request object.

### 🔴 GET `/exports/{exportId}`

**Auth:** 🔴 JWT + the requester, or an admin. Another e-board member gets `404`; they can start their own export.

**Response `200` (Proposed):**

```json
{
  "message": "Export ready",
  "export": {
    "id": "…", "clubId": 2, "status": "succeeded", "createdAt": "…Z", "finishedAt": "…Z", "expiresAt": "…Z",
    "options": { "includeImages": true, "includeEmails": false },
    "counts": { "members": 58, "events": 41, "announcements": 23, "images": 355 }, "bytes": 912345678,
    "downloadUrl": "<15-minute-signed-url>"
  }
}
```

`downloadUrl` is present only when `status` is `succeeded` and the file hasn't expired.

### 🔴 POST `/imports` and `/imports/{importId}/commit`

- **`POST /imports`** body `{ "mode": "restore" | "create" }`; response `{ "importId", "upload": { "url", "fields" }, "maxBytes": 2147483648 }`: the browser posts the file with those form fields.
- **`GET /imports/{importId}`** returns `status` (`queued` until the upload lands, `running`, `validated`, `invalid`, then `succeeded` or `failed`) and `report`: `{ "counts": {…}, "errors": [{ "sheet", "row", "column", "message" }], "warnings": […], "preview": { "Events": [ first 20 rows as they'll be created ], … } }`.
- **`POST /imports/{importId}/commit`** body `{ "clubName": "optional new name", "ownerStudentId": "restore only, if no owner matched" }`; `409` unless `validated`; response `{ "message": "Import started", "status": "running" }`. When it finishes, `GET` returns `"clubId"`.

## Authorization

| Who | Export | Import |
| --- | --- | --- |
| E-board or owner | Their own club. Emails only with the confirmation. Download only their own exports. | — (phase 4: `create`, if decided) |
| Admin | Any club, or every club (data only). Download any export. | Both modes. |
| Member, public | — | — |

An export holds only what its requester can already see in the app, for that club: drafts yes, other clubs' drafts or members no, soft-deleted rows no.

## Privacy

Exports contain student emails. [Decision 6](../../api/README.md#decisions) lets owners and e-board see member emails in the app, so an e-board export may include them. A file is different from a screen, though: once downloaded it's outside the platform's access control, it can be forwarded, it outlives the officer's term, and it can't be revoked. So:

1. **Emails are opt-in.** `includeEmails` defaults to `false`, and `true` needs the confirmation checkbox. Without it, the Members sheet has names and roles only.
2. **Every export is recorded** in the club's [history](edit-history.md) with who ran it and whether it held emails, visible to the whole e-board and to admins.
3. **Short-lived access.** Only the requester (and admins) can download; each click signs a new 15-minute URL; the file is deleted after 7 days; the bucket blocks public access and encrypts at rest. The "export ready" notification links to the app, never to the file.
4. **Minimum content.** The Members sheet has email, names, role, and the opaque student id: nothing about a student's other clubs. Not exported: soft-deleted items, pending or declined board pins, notifications, edit history, and co-hosting clubs' data beyond their names.
5. **Imports don't collect third-party emails.** In `create` mode, roster emails are counted in the report and not stored. `restore` only matches students already on the platform.
6. **Policy check needed.** Whether student officers may take rosters with emails off the platform, and whether students with a directory-information hold must be left out, is a question for Hunter's Office of Student Activities; these docs can't settle it ([open questions](#open-questions)).

## Images and storage

- **Exports** read every image the club uses (logo, flyers, galleries; later board uploads and announcement images) by `images.fk_club_id` and by reference, and copy the originals into the zip. Nothing changes in the image tables.
- **Imports** create new `images` rows and new S3 objects under `imported/{importId}/`, owned by the new club. Files used twice in the source are stored once, as in the database.
- **Temporary storage** (`exports/`, `imports/`, `jobs/`) is expired by lifecycle rules, never by the purge.

## Rough AWS cost

Using the [shared assumptions](README.md#scale-and-price-assumptions):

| Item | Cost |
| --- | --- |
| A club export with 750 MB of images | Worker: 2 GB × ~60 s = 120 GB-s, ~$0.002. S3 requests: under $0.01. Storage for 7 days: ~$0.004. **Download: 0.75 GB × $0.09 = ~$0.07**, or nothing inside the 100 GB monthly free allowance. |
| A data-only export | Well under $0.01. |
| A 1 GB import | Upload is free (data in). Worker ~$0.01. Temporary storage: cents. New images: ~$0.02 a month. |
| Standing cost | None: no queue, no endpoint, no NAT. Lifecycle rules and S3 event notifications are free. |
| **50 full exports a semester** | **Under $5 a semester**, almost all of it download bandwidth. |

## Changes to existing tables and endpoints

- **Tables:** none altered. `data_jobs` is new.
- **Validation:** event descriptions capped at 10,000 characters and club descriptions at 5,000, on `POST /clubs/{clubId}/events`, `PATCH /auth/events/{eventId}`, `POST /clubs`, and `PATCH /clubs/{clubId}`, so every value fits in a cell.
- **Infrastructure (in the new modules, not `infrastructure/legacy`):** the worker Lambda; S3 event notifications on `jobs/` and `imports/`; lifecycle rules on `exports/`, `imports/` and `jobs/`; an S3 gateway endpoint on the new VPC (the legacy one is in its bastion stack); and, separately, RDS backup retention raised from one day.
- **Reused:** the read views and queries for the sheets, `POST /clubs`'s creator-ownership insert, and the image insert.
- **Other proposals:** `exported` and `imported` history actions; `export.ready` and `import.finished` notification types.

## Open questions

1. **Emails:** may an e-board export them at all, or only admins? Opt-in (proposed), or default on since they can see them anyway? And the policy check with Student Activities on rosters and directory holds.
2. **Who may import in `create` mode:** admins only (proposed first), or anyone who can create a club?
3. **Invitations:** build `club_invitations` so an imported roster can be invited, or leave joining to students?
4. **Retention:** 7 days for export files and a 15-minute link: right?
5. **Admin all-clubs with images:** needed, or are RDS and S3 backups enough?
6. **History in exports:** should an admin's export include the club's edit history?
7. **Co-hosted items:** export them as rows marked `co-host` (proposed), or leave them out?

## Phased plan

| Phase | Scope | Size |
| --- | --- | --- |
| **1. First version** | Club export, **data only**: `data_jobs`, the S3-triggered worker, lifecycle rules, the three export routes, the zip with `manifest.json` and `club.xlsx` (no `images/`), emails opt-in, and the description caps. | Medium: 1 table, 3 routes, 1 Lambda. |
| 2 | Images in the export: the streaming zip. | Small. |
| 3 | Import in `restore` mode, for admins: presigned POST upload, validation, dry run with rollback, preview, commit. | Medium to large: most of the validation work. |
| 4 | `create` mode with the template, co-host matching, invitations if accepted, the admin all-clubs export, and board pins in the format. | Medium. |
