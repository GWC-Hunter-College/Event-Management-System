# Notifications

**Proposed, 2026-09-30.** Nothing here is built, and nothing in the baseline changes until the proposal is accepted. Every route is ⬜ and **Proposed**, in the [api/ reference](../../api/README.md#legend) style.

**Depends on:** the core API from the "Needed now" list ([coverage.md](../../api/coverage.md#needed-now)): `RequireStudent`, the club role check, and `GET /me/clubs`.
**Needed by:** [announcements](announcements.md) ("a new post from your club"), the [bulletin board](bulletin-board.md) ("your pin is up", "your pin was declined", "a pin is waiting for review"), and [import and export](import-export.md) ("your export is ready"). Later: event reminders (the club page's "Event reminders" menu item, SOON today).

## Problem

Two features that are already designed can't be finished without a way to tell a student that something happened while they weren't looking:

- **Bulletin board.** A member sends a pin for review and leaves. The Figma panel promises "You'll get a notification when it's up", and a declined pin comes back with "a short note" from the e-board. The e-board also needs to know that submissions are waiting.
- **Announcements.** A member of a club learns about a new post without checking every club page. The API's open questions already defer this to "a notification system, which the club bulletin board also needs" ([api/README.md](../../api/README.md#open-questions)).

**Who uses it.** Every signed-in student receives notifications; the e-board's actions (posting, approving, declining) create them. Nothing here is public, and the GWC website never reads it.

**How.** A bell in the app shell shows the number of unread notifications. Opening it lists the latest ones, newest first: "Girls Who Code posted: HunterHacks registration is open", "Your pin is up on the Girls Who Code board", "Your pin for Girls Who Code wasn't added: please keep photos on topic". Clicking one opens what it's about (the announcement, the board with the pin, the export) and marks it read. "Mark all as read" clears the badge. Later, a settings page turns email on or off per kind of notification.

## In-app vs email

| | In-app | Email |
| --- | --- | --- |
| What it needs | One table and four routes. MySQL and the API only. | Everything in-app needs, plus SES production access, a sending domain with DKIM, SPF and DMARC, unsubscribe handling, and bounce handling. |
| Reaches | Students who open the site. | Students who don't, as long as the Cognito trigger stored their email. `students.email` is nullable, and the access token carries no email ([M5](../../api/coverage.md#m5-the-access-token-has-no-email-claim)). |
| Standing cost | About nothing. | About nothing beyond SES's $0.10 per 1,000 emails, on any of the [hosting options](../../infrastructure/aws/hosting-options.md). Only a private network with no NAT, like the legacy stack's, would need a ~$7.30-a-month SES endpoint. |
| Risk | Low. | A spam complaint or bounce rate can suspend the SES account; emails to CUNY addresses may be filtered. |

**Recommendation: in-app first, email second.** Ship in-app notifications as version 1, with every type. Add email in phase 3, per type and per student, behind preferences. No web push or SMS: there's no native app, and web push needs a service worker and per-browser subscriptions for little gain at this scale.

**Real-time:** none. The frontend polls [`GET /me/notifications/unread-count`](#-get-menotificationsunread-count) when a page loads, when the window regains focus, and every 60 seconds while it's visible. An API Gateway WebSocket API would add connection state and cost for updates nobody needs within the second.

## What gets stored

**One row per recipient** ("fan-out on write"). Posting an announcement to a 60-member club inserts 60 rows with one `INSERT … SELECT` from `club_members`, in the same transaction as the post.

The alternative, fan-out on read, stores one row per happening and computes each student's list by joining their memberships. It writes less, but it still needs per-student state for read/unread and for email delivery, a student who joins a club later would suddenly "receive" its old posts, and every list read becomes a multi-club join. At Hunter's scale the largest fan-out is one club's roster, about 500 rows, so writing them is cheap.

**Rows point at what they're about instead of copying it.** A notification stores its recipient, its type, its club, who caused it, and a foreign key to its subject (an announcement, a pin, an export job). The API builds the text ("Girls Who Code posted: …") when it reads the list, from the subject's current title. So an edited title shows up correctly, a soft-deleted subject hides its notification, and purging the subject deletes its notifications through the foreign key. The one piece of free text a notification shows, a declined pin's note, lives on the pin.

**Read/unread** is `read_at`: `NULL` is unread. There's no separate "seen" state; opening the bell doesn't mark anything read, clicking an item does, and "Mark all as read" does for all.

### DDL sketch

Version 1. Written in the baseline's conventions (named foreign keys, `TIMESTAMP` for UTC times, no semicolons inside comments), so it could go straight into the baseline while there's no live database, as the announcement tables did.

```sql
-- Notifications: one row per recipient. The API builds the text from the subject
-- when it reads the list, so nothing here copies a title or body.
-- type is a key from the API's notification registry, for example
-- announcement.posted or board.pin_declined. Not constrained, like images.purpose:
-- a new type brings new code anyway.
-- Exactly one subject column is set, chosen by type. Later features add theirs.
CREATE TABLE `notifications` (
  `id` BIGINT NOT NULL AUTO_INCREMENT,
  `fk_recipient_id` CHAR(36) NOT NULL,
  `type` VARCHAR(40) NOT NULL,
  `fk_club_id` INT NULL,
  `fk_actor_id` CHAR(36) NULL,
  `fk_announcement_id` INT NULL,
  `created_at` TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
  `read_at` TIMESTAMP NULL,
  PRIMARY KEY (`id`),
  KEY `idx_notifications_recipient` (`fk_recipient_id`, `id`),
  KEY `idx_notifications_recipient_unread` (`fk_recipient_id`, `read_at`),
  KEY `idx_notifications_club` (`fk_club_id`),
  KEY `idx_notifications_actor` (`fk_actor_id`),
  KEY `idx_notifications_announcement` (`fk_announcement_id`),
  CONSTRAINT `fk_notifications_recipient` FOREIGN KEY (`fk_recipient_id`) REFERENCES `students` (`id`) ON DELETE CASCADE,
  CONSTRAINT `fk_notifications_club` FOREIGN KEY (`fk_club_id`) REFERENCES `clubs` (`id`),
  CONSTRAINT `fk_notifications_actor` FOREIGN KEY (`fk_actor_id`) REFERENCES `students` (`id`) ON DELETE SET NULL,
  CONSTRAINT `fk_notifications_announcement` FOREIGN KEY (`fk_announcement_id`) REFERENCES `announcements` (`id`) ON DELETE CASCADE
);
```

Why each foreign key behaves as it does, following [F1](../../database/docs/schema-review.md#f1-on-delete-behavior):

| Column | `ON DELETE` | Why |
| --- | --- | --- |
| `fk_recipient_id` | `CASCADE` | A notification is part of its recipient's inbox. |
| `fk_actor_id` | `SET NULL` | Attribution, like `events.fk_author_id`. |
| `fk_club_id` | `RESTRICT` | Club deletion isn't designed yet (F1). |
| `fk_announcement_id` (and later subjects) | `CASCADE` | The purge hard-deletes announcements; their notifications go with them. |

**Indexes.** The list reads `(fk_recipient_id, id)` newest first with an id cursor. The unread count is `COUNT(*) … WHERE fk_recipient_id = ? AND read_at IS NULL`, served by `(fk_recipient_id, read_at)` without touching the rows.

**Later features add columns** in their own migrations, each a nullable foreign key with `ON DELETE CASCADE`:

```sql
ALTER TABLE `notifications`
  ADD COLUMN `fk_board_pin_id` INT NULL,       -- bulletin-board.md
  ADD COLUMN `fk_data_job_id` CHAR(36) NULL;   -- import-export.md
```

**Phase 2 adds grouping**, so 30 submissions to one board don't become 30 unread rows for each e-board member. An unread row with the same `group_key` is bumped instead of duplicated, using the baseline's functional-index trick (M3): only unread rows get a key, so read ones never collide.

```sql
ALTER TABLE `notifications`
  ADD COLUMN `group_key` VARCHAR(80) NULL,
  ADD COLUMN `group_count` INT NOT NULL DEFAULT 1,
  ADD UNIQUE KEY `uq_notifications_unread_group` (`fk_recipient_id`, (IF(`read_at` IS NULL, `group_key`, NULL)));
-- The writer uses INSERT ... ON DUPLICATE KEY UPDATE
--   group_count = group_count + 1, created_at = CURRENT_TIMESTAMP
```

**Phase 3 adds email** delivery state and preferences:

```sql
ALTER TABLE `notifications`
  ADD COLUMN `email_status` ENUM('none', 'pending', 'sent', 'failed', 'skipped') NOT NULL DEFAULT 'none',
  ADD COLUMN `emailed_at` TIMESTAMP NULL,
  ADD KEY `idx_notifications_email` (`email_status`, `id`);

-- Only a student's overrides are stored. The defaults live in the API's type registry.
CREATE TABLE `notification_preferences` (
  `fk_student_id` CHAR(36) NOT NULL,
  `type` VARCHAR(40) NOT NULL,
  `email` BOOL NOT NULL,
  `updated_at` TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  PRIMARY KEY (`fk_student_id`, `type`),
  CONSTRAINT `fk_notification_preferences_student` FOREIGN KEY (`fk_student_id`) REFERENCES `students` (`id`) ON DELETE CASCADE
);
```

`skipped` covers a student with no stored email, or email turned off for that type when the row was written.

### Types

The registry of types lives in the API, as one Go table: key, recipients, subject column, how to build the text and link, and the email default. The types these proposals need:

| Type | Recipients | Subject | Text (built on read) | Email default (phase 3) | From |
| --- | --- | --- | --- | --- | --- |
| `announcement.posted` | Members of the owner club, and of associate clubs on whose e-board the author sits, deduplicated, except the author ([why](announcements.md#notifications)) | announcement | "{club} posted: {title}" | On | [announcements.md](announcements.md) |
| `announcement.shared` | E-board and owners of the other associate clubs | announcement | "{club} shared an announcement with {your club}" | Off | [announcements.md](announcements.md) |
| `board.pin_submitted` | E-board and owners of the club, grouped per club | pin | "{n} pins waiting for review on the {club} board" | Off | [bulletin-board.md](bulletin-board.md) |
| `board.pin_approved` | The member who submitted it | pin | "Your pin is up on the {club} board" | Off | [bulletin-board.md](bulletin-board.md) |
| `board.pin_declined` | The member who submitted it | pin | "Your pin for {club} wasn't added: {note}" | Off | [bulletin-board.md](bulletin-board.md) |
| `export.ready`, `import.finished` | The student who asked | job | "Your export of {club} is ready" | On | [import-export.md](import-export.md) |

The reviewer is stored as the actor but never shown to the member ("the e-board", not a name). Student names aren't stored anyway: nothing writes `student_info` yet.

### Visibility on read

The list query hides a row whose subject the recipient can't see any more: a soft-deleted announcement (restoring it brings the notification back, without a new one), a removed pin, a job whose export expired. It doesn't hide a row because the recipient has since left the club; announcements are public anyway.

### Retention

Rows are small, but they shouldn't live forever. Proposed: the admin's [purge](../../api/endpoints/admins.md#-post-adminspurge) (decision 8) also deletes read notifications older than 180 days and every notification older than a year. That's one more statement in its first transaction. The numbers are an [open question](#open-questions).

## Endpoints

All Proposed, all ⬜. Every 🔴 route first runs [`RequireStudent`](../../api/endpoints/internal.md#-requirestudent-request-time-student-sync).

| Access | Method | Path | What it does | Role | Phase |
| --- | --- | --- | --- | --- | --- |
| 🔴 | GET | `/me/notifications` | The caller's notifications, newest first. | Any signed-in student; own rows only | 1 |
| 🔴 | GET | `/me/notifications/unread-count` | The badge count, for polling. | Any signed-in student | 1 |
| 🔴 | PATCH | `/me/notifications/{notificationId}` | Marks one read or unread. | The recipient | 1 |
| 🔴 | POST | `/me/notifications/read-all` | Marks everything up to a cursor read. | Any signed-in student | 1 |
| 🔴 | GET | `/me/notification-settings` | Email on or off, per type. | Any signed-in student | 3 |
| 🔴 | PUT | `/me/notification-settings` | Changes those settings. | Any signed-in student | 3 |
| 🟢 | POST | `/notifications/unsubscribe` | One-click unsubscribe from an email link. | Anyone holding a valid signed token | 3 |
| 🔵 | function | `Notify` | Inserts notifications inside the caller's transaction. | Called by handlers | 1 |
| 🔵 | worker | Email sender | Sends pending emails through SES. | — | 3 |

**Authorization.** Every `/me` route is scoped to the caller: each query has `fk_recipient_id = <sub>` in its `WHERE` clause. Someone else's notification id is `404`, not `403`, so ids don't reveal anything. There's no public or admin route that creates a notification; only the 🔵 `Notify` function does, from inside a handler that already passed its own role check. An admin broadcast to every student is an [open question](#open-questions).

### 🔴 GET `/me/notifications`

**Auth:** 🔴 JWT. Returns only the caller's rows.

**Query params:** `limit` (default 20, `1`–`50`), `before` (a notification id: returns older ones, for "load more"), `unread` (`true` returns unread only).

**Response `200` (Proposed):**

```json
{
  "message": "Successfully fetched 2 notifications",
  "unreadCount": 1,
  "notifications": [
    {
      "id": 812,
      "type": "announcement.posted",
      "createdAt": "2026-09-28T14:00:00Z",
      "readAt": null,
      "club": { "id": 2, "name": "Girls Who Code @ Hunter", "thumbnailUrl": "<readable-logo-url>" },
      "text": "Girls Who Code @ Hunter posted: HunterHacks registration is open",
      "link": { "kind": "announcement", "id": 7 }
    },
    {
      "id": 790,
      "type": "board.pin_declined",
      "createdAt": "2026-09-27T20:11:00Z",
      "readAt": "2026-09-27T21:00:00Z",
      "club": { "id": 2, "name": "Girls Who Code @ Hunter", "thumbnailUrl": "<readable-logo-url>" },
      "text": "Your pin for Girls Who Code @ Hunter wasn't added: please keep photos on topic",
      "link": { "kind": "boardPin", "clubId": 2, "id": 55 }
    }
  ]
}
```

`link` is data, not a URL, so the frontend owns its routes (`/announcement/7`, `/club/2?tab=board&pin=55`). Times are ISO 8601 UTC with `Z`, as everywhere else.

**Errors:** `400` bad `limit` or `before`, `401`, `500`.

**Queries (proposed, not written):** `notifications/list/SELECT_my_notifications.sql` (read; left-joins each subject table to build the text and apply [visibility](#visibility-on-read)) → `notifications/count/SELECT_my_unread_count.sql` (read).

### 🔴 GET `/me/notifications/unread-count`

**Auth:** 🔴 JWT. The cheap call the frontend polls.

**Response `200` (Proposed):** `{ "message": "Successfully counted unread notifications", "unreadCount": 3 }`.

**Queries:** `notifications/count/SELECT_my_unread_count.sql` (read, index-only).

### 🔴 PATCH `/me/notifications/{notificationId}`

**Auth:** 🔴 JWT + the caller is the recipient; otherwise `404`.

**Request body (Proposed):** `{ "read": true }` or `{ "read": false }`. Marking read sets `read_at` only if it's `NULL`, so the first read time is kept.

**Response `200` (Proposed):** `{ "message": "Successfully updated notification 812", "unreadCount": 2 }`.

**Errors:** `400`, `401`, `404`, `500`.

**Queries:** `notifications/update/UPDATE_notification_read.sql` (write, scoped to the caller; 0 rows is `404`) → `SELECT_my_unread_count.sql` (read).

### 🔴 POST `/me/notifications/read-all`

**Auth:** 🔴 JWT.

**Request body (Proposed):** `{ "upTo": 812 }`, the newest id the caller has on screen, so a notification that arrives while the request is in flight stays unread. Omitted means everything.

**Response `200` (Proposed):** `{ "message": "Marked 3 notifications read", "unreadCount": 0 }`.

**Queries:** `notifications/update/UPDATE_my_notifications_read.sql` (write).

### 🔵 `Notify`

A Go function in the API module, not a route. Called by a handler **inside its own transaction**, after its writes and before commit:

```go
notify.Send(ctx, tx, notify.AnnouncementPosted{AnnouncementID: 7, ClubIDs: []int{3, 2}, ActorID: sub})
```

Each type runs one `INSERT … SELECT` that picks its recipients in SQL (for example, `SELECT DISTINCT fk_student_id FROM club_members WHERE fk_club_id IN (…) AND fk_student_id <> ?`). Because it shares the transaction, a post that rolls back never notifies anyone, and a post that commits always has its notifications. With email on (phase 3), it also sets `email_status` to `pending` or `skipped` from the recipient's preference and stored email.

### Phase 3 routes, in brief

- 🔴 **GET/PUT `/me/notification-settings`:** `{ "email": { "announcement.posted": true, "board.pin_declined": false, … } }`, every type with its effective value. `PUT` stores only the values that differ from the default. In-app notifications can't be turned off.
- 🟢 **POST `/notifications/unsubscribe`:** the target of the `List-Unsubscribe` header ([RFC 8058](https://www.rfc-editor.org/rfc/rfc8058) one-click) and of the footer link. The body carries a token: an HMAC over the student id, the type, and an expiry, signed with a key in Secrets Manager or SSM. It's 🟢 because mail clients call it without a login; the token is the authorization. Turns email off for that type.
- 🔵 **Email sender:** see below.

## Email delivery (phase 3)

```text
API handler
  └─ same transaction ─> notifications row, email_status = 'pending'

Email sender
  ├─ SELECT … WHERE email_status = 'pending' ORDER BY id LIMIT 100 FOR UPDATE SKIP LOCKED
  ├─ render, send ─> SES ─> the student's inbox
  └─ UPDATE email_status = 'sent' or 'failed', emailed_at
```

The table is the outbox: delivery state stays in the database, and a failed send stays `pending` for the next pass. **What runs the sender depends on the [hosting option](../../infrastructure/aws/hosting-options.md):**

- **On a server** (options 1 and 2): a goroutine in the API process makes a pass every minute and calls SES directly. It's a loop inside the API, not a scheduled job, and it only sends what's already committed.
- **Serverless with a sleeping database** (option 3): a timer would keep Aurora awake, so the handler makes one pass right after its transaction commits. A send that fails stays `pending` until the next write wakes the database.

Neither needs a NAT gateway or a VPC endpoint. Those were only needed because the legacy stack's Lambdas sat in private subnets with no way out.

**SES setup.** Request production access (the sandbox only sends to verified addresses, 200 a day). Verify a domain with Easy DKIM, set a custom MAIL FROM for SPF, and publish DMARC. **Which domain** is open: `girlswhocodehunter.org` is live in Route 53 ([hosting/README.md](../../hosting/README.md)), so a subdomain of it could send today, but this platform serves every Hunter club. Leave SES's account-level suppression list on, so bounces and complaints stop further sends without code; a bounce handler (SNS to Lambda, marking the student) can come later.

**What an email contains.** The same text as the in-app row, a button back to the app, and an unsubscribe link. Never an export download link (see [import-export.md](import-export.md#privacy)). One email per notification; a daily digest is an open question.

## Impact on images and storage

None. Notifications store no files. Emails don't embed images; the club logo, if shown, is linked through the app, never attached. Row storage is covered under [Cost](#rough-aws-cost).

## Rough AWS cost

Using the [shared assumptions](README.md#scale-and-price-assumptions): 50 active clubs a week, 60 members each, one announcement per active club per week, 1,000 daily active students.

| Item | Phase 1 (in-app) | Phase 3 (email) |
| --- | --- | --- |
| Rows | ~13,000 announcement notifications a month, plus board and job rows. ~400 bytes each with indexes: ~6 MB a month, a rounding error on any database's storage. | Same. |
| Polling | 1,000 students × ~20 checks a day × 30 = ~600,000 requests. Nothing extra on a server; ~$0.75 of API Gateway and Lambda serverless, where the counts come from DynamoDB so Aurora can sleep. | Same. |
| Email | — | ~13,000 emails at $0.10 per 1,000: ~$1.30. |
| **Total** | **Under $1 a month** | **About $2 a month** |

## Changes to existing tables and endpoints

- **Tables:** none in version 1. `notifications` references `students`, `clubs` and `announcements` without changing them.
- **Announcements:** [`POST /clubs/{clubId}/announcements`](../../api/endpoints/announcements.md#-post-clubsclubidannouncements) (when created as `posted`) and [`PATCH /auth/announcements/{announcementId}`](../../api/endpoints/announcements.md#-patch-authannouncementsannouncementid) (on `draft → posted`) add one `Notify` insert to their transactions. Details in [announcements.md](announcements.md#notifications).
- **Purge:** [`POST /admins/purge`](../../api/endpoints/admins.md#-post-adminspurge) removes notifications of purged announcements through the cascade, and (proposed) old notifications under the retention rule.
- **Email (phase 3):** depends on `students.email` being filled by the Cognito trigger, never from the access token ([M5](../../api/coverage.md#m5-the-access-token-has-no-email-claim)).
- **Frontend:** a bell and panel in the app shell, and a settings page in phase 3. [coverage.md](../../api/coverage.md#every-page-app-shell) gets rows for them when they're built.
- **Docs, once accepted:** the api/README open question on notifying members is answered, and `notifications` joins the schema, `schema.dbml`, and the query groups.

## Open questions

1. **Email defaults.** Recommended: announcements on, export-ready on, board outcomes off. Is emailing every announcement by default acceptable, given that joining a club is the student's only consent?
2. **Sending domain** for SES: a subdomain of `girlswhocodehunter.org`, a new domain for the Hunter clubs site, or a Hunter-owned one?
3. **The sender's loop.** On a server, the sender makes a pass every minute inside the API. Decision 8 ruled out a scheduled job for the purge; is a sending loop acceptable?
4. **Retention:** 180 days for read notifications and a year for all?
5. **Per-club mute** ("don't notify me about this club") in phase 1, or later? It would be a column on `club_members` or a small table.
6. **Admin broadcasts** to every student (for example, "club fair on Thursday")? Nothing here supports that.
7. **Digests:** one email per notification, or a daily summary per student?
8. Should `board.pin_submitted` notify every e-board member, or only owners?

## Phased plan

| Phase | Scope | Size |
| --- | --- | --- |
| **1. First version** | The `notifications` table as sketched, `Notify`, the four `/me/notifications` routes, the bell and panel, and one type: `announcement.posted`, shipped with the [Announcements tab](announcements.md). In-app only. | Small: 1 table, 4 routes, ~5 queries. |
| 2 | Board types with grouping (`group_key`), and job types, as those features ship. | Small, per feature. |
| 3 | Email: SES setup and domain, the sender, preferences, the settings page, one-click unsubscribe, retention in the purge. | Medium. |
| Later | Event reminders (needs a "remind me" or RSVP record, and a schedule per event), digests, per-club mute, web push. | — |
