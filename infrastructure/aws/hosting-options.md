# Hosting options

**Proposed, 2026-09-30.** Nothing here is chosen or built. Three ways to host the backend (the [`api/`](../../api/) Go module and the [`database/`](../../database/) MySQL schema) on AWS, replacing [`infrastructure/legacy/`](../legacy/), which isn't being reused.

**Prices** are us-east-1 list prices **at full price, with no free tier**. The one allowance counted is Cognito's 10,000 monthly active users, which AWS says doesn't expire and isn't part of the 12-month free tier. The figures were checked against search results on 2026-09-30, but AWS's own pricing pages couldn't be opened from that session, so confirm them in the [AWS Pricing Calculator](https://calculator.aws/) before budgeting. Usage follows the [shared assumptions](../../docs/proposals/README.md#scale-and-price-assumptions): about 1,000 students a day, which averages well under one request a second.

## What the backend needs

- **The Go API over HTTPS,** for the frontend (served from CloudFront by [`hosting/`](../../hosting/)) and the GWC website.
- **MySQL 8.0.17 or later.** The baseline uses `CHECK` constraints, functional indexes, and JSON; the [edit history proposal](../../docs/proposals/edit-history.md) adds a multi-valued index.
- **S3 for images.** Browsers upload and download directly through presigned URLs, the same in every option.
- **Cognito for sign-in.** Nothing is deployed and there are no users yet, so starting a fresh user pool costs nothing.
- **Background work,** from the [feature proposals](../../docs/proposals/README.md): sending email, export and import jobs, and image thumbnails.
- **Backups.** The club export is for handing data to officers, not for disaster recovery ([why](../../docs/proposals/import-export.md#export)).

## Why not the legacy stack

- It deploys about 25 separate Lambdas and two API stacks, which declare some of the same fixed function names ([api/README.md](../../api/README.md#open-questions)).
- Its Lambdas sit in private subnets with no NAT gateway, so reaching SES, SQS, or any other AWS service needs a paid endpoint per service (about $7.30 a month each) or a NAT gateway (about $33).
- Database access goes through a bastion host and three SSM endpoints.
- It keeps one day of database backups.

## Option 1: one server on Lightsail

```text
Browser ── CloudFront (the frontend, hosting/)
Browser ── HTTPS ──> Lightsail instance, 1 GB
                      ├─ Caddy: the TLS certificate for api.<domain>, proxying to the API
                      ├─ Go API: every route, plus goroutines for email, jobs, and thumbnails
                      └─ MySQL 8.0, on the same instance
                   ──> S3 (images) · SES (email) · Cognito (its public token keys)
```

| Item | Monthly |
| --- | --- |
| Lightsail $7 bundle: 1 GB RAM, 40 GB SSD, 2 TB of transfer, a public IPv4 address | $7.00 |
| Automatic daily snapshots, 7 kept (at most 40 GB × $0.05; incremental, so usually less) | up to $2.00 |
| Nightly `mysqldump` to S3, 30 kept (a few MB each) | under $0.05 |
| **Total** | **~$9** |

- **HTTPS:** Caddy gets and renews a Let's Encrypt certificate on its own. It needs a domain: a subdomain of `girlswhocodehunter.org`, whose Route 53 zone already exists, or a domain for the Hunter clubs site (about $13 a year).
- **AWS credentials:** Lightsail instances can't take an IAM role, so the API uses an IAM user's access keys, limited to the image bucket, SES, and Cognito, stored on the instance, and rotated each year. This is the option's main security trade-off; EC2 avoids it.
- **Token checks:** in the legacy stack, API Gateway checked Cognito tokens. Here the Go app checks them against the user pool's public keys (JWKS), with a small library.
- **MySQL in 1 GB:** a small buffer pool (about 256 MB) and `performance_schema` off. If memory runs short, move to the $12 bundle (2 GB).
- **Deploys:** GitHub Actions builds the binary, copies it over SSH, and restarts a systemd service, about a second of downtime. A reboot or patch is a few minutes.
- **Patching:** unattended security updates on Ubuntu or Amazon Linux.
- **Traffic:** the 2 TB included covers the API's JSON many times over. Images come straight from S3 and are billed there.
- **Growing:** snapshot to a bigger bundle, or move MySQL to RDS (option 2) with a dump and restore.

## Option 2: EC2 and RDS

```text
Browser ── HTTPS ──> EC2 t4g.micro, public subnet
                      ├─ Caddy (TLS) → Go API: every route, plus goroutines for email, jobs, and thumbnails
                      └──> RDS MySQL db.t4g.micro, private subnet (reached only from the instance)
                   ──> S3 · SES · Cognito
```

| Item | Monthly |
| --- | --- |
| EC2 t4g.micro (1 GB) | $6.13 |
| EBS disk, 20 GB gp3 | $1.60 |
| Public IPv4 address | $3.65 |
| RDS db.t4g.micro, MySQL, single-AZ | $11.68 |
| RDS storage, 20 GB | $2.30 |
| RDS automated backups, 7 days (backup storage up to the database's size is included) | $0 |
| API traffic out (~12 GB of JSON at $0.09) | ~$1.10 |
| **Total** | **~$26** |

- **Managed database:** automated backups with point-in-time restore, and minor-version patching, handled by AWS. That's what the extra ~$17 over option 1 buys.
- **No stored keys:** the instance takes an IAM role. SSM Session Manager replaces SSH, so no port is open for logins.
- **No NAT:** the database needs no outbound access, and the instance is in a public subnet.
- **Variant:** MySQL on the instance instead of RDS is about $13 a month, but then it's option 1 with more parts.

## Option 3: serverless, with a sleeping database

```text
Public reads      Browser ──> CloudFront ──> S3: JSON files for clubs, events, announcements, boards
Signed-in reads   Browser ──> API Gateway ──> Lambda (in no VPC) ──> DynamoDB: my clubs, unread counts
Writes            Browser ──> API Gateway ──> Lambda ── Data API (HTTPS) ──> Aurora Serverless v2
                                                 └─> rebuild the affected JSON in S3, update DynamoDB
Waking            an edit page opens ──> POST /wake ──> Data API "SELECT 1"
```

**Aurora Serverless v1 is retired** (March 2025). **Aurora Serverless v2** had a floor of 0.5 capacity units until November 2024, which is $43.80 a month with nobody using it. Since then it can **scale to zero**:

- With a minimum of 0, it pauses after an idle period you choose, from 5 minutes to a day.
- While paused you pay nothing for compute, only storage.
- Waking takes up to about 15 seconds.
- It needs Aurora MySQL 3.08 or later (MySQL 8.0-compatible).

The Lambdas reach Aurora through the Data API over HTTPS, so they sit in no VPC and need no NAT or endpoints. Aurora itself lives in a VPC but needs no outbound access.

| Item | Monthly |
| --- | --- |
| Aurora storage and I/O (~1 GB) | ~$0.30 |
| Aurora automated backups (up to the database's size included) | $0 |
| API Gateway and Lambda (~1M requests) | ~$1.50 |
| Data API ($0.35 per million requests) | ~$0.05 |
| The read cache: CloudFront, S3 writes, DynamoDB | ~$1.50–2.50 |
| **Base** | **~$4** |
| Plus Aurora compute | **$0.06 per hour awake** at the smallest size (0.5 units), more if it scales up under load |

Every wake costs at least the 5-minute idle timeout, so a one-minute edit bills about six minutes.

| Database awake | Total |
| --- | --- |
| ~1 hour a day (summer, quiet weeks) | ~$6 |
| ~3 hours a day (a busy semester) | ~$9 |
| ~8 hours a day | ~$18 |
| Never sleeps | ~$48 |

**For it to stay asleep:**

- **Page views never read MySQL.** Public data comes from the JSON files in S3; signed-in data (my clubs, drafts, the unread badge) from DynamoDB.
- **Nothing touches the database on a timer.** No pollers, health checks, or connection pools. Email is sent right after the write that caused it.
- **Wake early.** An edit page calls `/wake` when it opens, and again every few minutes while it stays open, so the database is up before the person submits. The route needs a login and a rate limit, or anyone could keep the database awake.
- **Set a budget alert** at about $10 in AWS Budgets, so a database that stops sleeping is noticed within days.

**What changes in the API:** public list routes become static files, so paging and date filters (`?when=upcoming&page=2`) move into the browser, which is fine at a club's size (tens of KB per list). Every write rebuilds the files it affects. Signed-in reads move to DynamoDB. A write that arrives while the database is asleep waits up to about 15 seconds.

## Shared costs

The same in every option, and not in the totals above:

| Item | Monthly |
| --- | --- |
| Image storage in S3 (~50 GB, growing each semester) | ~$1.15 |
| Image downloads, with the 640-pixel copies the [board proposal](../../docs/proposals/bulletin-board.md#images-and-storage) calls for (~40 GB at $0.09) | ~$3.60 |
| Email through SES (~13,000, once email notifications ship) | ~$1.30 |
| Cognito, up to 10,000 monthly active users | $0 |
| Route 53 (the zone already exists for the GWC site) | $0 extra |
| **Shared** | **~$6** |

Without the smaller image copies, downloads cost ten times as much or more, which would outweigh the hosting itself.

## Side by side

| | 1. Lightsail | 2. EC2 + RDS | 3. Serverless, sleeping Aurora |
| --- | --- | --- | --- |
| Hosting, per month | ~$9, flat | ~$26, flat | ~$4 + $0.06 per awake hour: ~$6–9 typical, ~$48 worst |
| With the shared costs | ~$15 | ~$32 | ~$12–15 typical |
| Predictable bill | Yes | Yes | No: it follows awake hours |
| Backups | Snapshots and a nightly dump, which you set up | Managed, point-in-time | Managed, point-in-time |
| After a quiet spell | Instant | Instant | A write can wait ~15 seconds, hidden by waking early |
| What you maintain | The OS, MySQL, TLS, deploys | The OS, TLS, deploys | Cache rebuilds, the wake flow, many small functions |
| AWS credentials | Stored access keys | IAM role | IAM roles |
| API contract changes | None | None | Public lists become static files |
| Work to build | Least | A little more | Most |
| How it grows | A bigger bundle, then option 2 | Bigger instances | On its own |

## Recommendation

**Start with option 1.** It's the cheapest predictable bill, the least to build, and it fits not knowing how many students will use the site. The `api/` and `database/` modules don't depend on where they run, so moving later is a deploy change plus a database dump and restore.

- **Choose option 2** if managed backups and point-in-time restore are worth about $17 a month more, for example once the site holds a few years of club history.
- **Choose option 3** if a bill that falls to about $5 in quiet months matters more than the extra design work and the API changes.

## What each option means for the feature proposals

| Piece | Options 1 and 2 (a server) | Option 3 (serverless) |
| --- | --- | --- |
| [Email sender](../../docs/proposals/notifications.md#email-delivery-phase-3) | A loop in the API sends pending emails every minute, straight to SES. | Sent right after the write commits; a timer would keep the database awake. |
| Unread-count polling | Reads MySQL. | Reads DynamoDB, or it keeps Aurora awake. |
| [Export and import jobs](../../docs/proposals/import-export.md#export) | A goroutine takes queued jobs from `data_jobs`, with no time limit. | A Lambda started by an S3 event, with a 15-minute limit. |
| Image thumbnails | Made by the API when an upload is confirmed. | A Lambda triggered by the upload. |
| Cognito token checks | In the Go app. | API Gateway's JWT authorizer. |
| The admin purge | Unchanged. | Unchanged; it wakes the database once. |

## Open questions

1. **Which option?**
2. **The API's domain:** a subdomain of `girlswhocodehunter.org`, or a domain for the Hunter clubs site? The email sending domain is the same question.
3. **Region:** us-east-1? The legacy stack took it from the deploying machine's environment.
4. **Ownership:** who holds the AWS account and its credentials from one e-board year to the next?
5. **Budget alert:** at what amount?

## Sources

- [Aurora Serverless v2 supports scaling to zero capacity](https://aws.amazon.com/about-aws/whats-new/2024/11/amazon-aurora-serverless-v2-scaling-zero-capacity/) and [the auto-pause documentation](https://docs.aws.amazon.com/AmazonRDS/latest/AuroraUserGuide/aurora-serverless-v2-auto-pause.html) (AWS)
- [Aurora Serverless v2 pricing guide](https://www.usage.ai/blogs/aws/rds/aurora-serverless-v2/) (Usage.ai) and [Data API pricing](https://dev.to/aws-builders/data-api-for-amazon-aurora-serverless-v2-with-aws-sdk-for-java-part-12-data-api-quotas-11h2) (DEV Community)
- [Amazon Cognito pricing](https://aws.amazon.com/cognito/pricing/) (AWS)
- [Lightsail pricing](https://www.cloudzero.com/blog/amazon-lightsail-pricing/) (CloudZero) and [Lightsail and IAM roles](https://repost.aws/questions/QUMbXYXqiKTQ6L2Q2Z_FNLGQ/i-want-to-use-lightsail-to-but-it-doesn-t-seem-to-allow-uploads-to-s3-buckets) (AWS re:Post)
- [db.t4g.micro pricing](https://instances.vantage.sh/aws/rds/db.t4g.micro) (Vantage)
