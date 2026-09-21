# Database

The application stores identity, club, membership, event, and image metadata in one Amazon RDS for MySQL instance. Development handlers connect to the `STAGING` database and production handlers connect to `PRODUCTION`; the databases share the instance, generated administrator credential, network boundary, and backup policy.

See the [schema reference](schema.md) for the current dbdiagram visual, maintainable ERD, and column-level constraints. Image metadata writes are described in [Image uploads](../architecture/image-uploads.md).

## Source of truth

Fresh initialization currently runs [`11_04_2025_create_core_tables_up.sql`](../../lambda/internal/database/init/migrations/11_04_2025_create_core_tables_up.sql). That migration is the authoritative schema for the legacy implementation. Earlier dated migrations, SQL under `stub/environment`, Go model structs, and diagrams do not override it.

The schema reference includes a committed dbdiagram export and links to the external [Hunter Club Event System dbdiagram](https://dbdiagram.io/d/Hunter-Club-Event-System-6861eb43f413ba35086e147c). These are visual references; they are not executed during deployment and do not override the active migration.

## Provisioned database

[`internal/stack/database.go`](../../internal/stack/database.go) creates the following database resources:

| Setting | Current value |
| --- | --- |
| Engine | MySQL `8.0.37` |
| Instance | Single-AZ `db.t3.micro` |
| Storage | 20 GiB initially, autoscaling up to 100 GiB |
| Placement | Private isolated VPC subnets |
| Credentials | Secrets Manager-generated secret for user `dbadmin` |
| Connectivity | Lambdas connect directly to the instance on TCP 3306; no RDS Proxy is created |
| Backups | One-day retention |
| Deletion behavior | Deletion protection disabled and CloudFormation removal policy `DESTROY` |

The database stack permits port 3306 from the Lambda database security group, and `BastionStack` adds a separate ingress rule from its bastion security group. Database-aware Lambdas also use the VPC security group that reaches the Secrets Manager interface endpoint so they can load the generated credential.

The API stacks select a database through `DB_NAME`:

| API stack | HTTP API | Database |
| --- | --- | --- |
| `DevApiStack` | `ClubEventApiDev` | `STAGING` |
| `ProdApiStack` | `ClubEventApiProd` | `PRODUCTION` |

This is logical data separation inside one RDS instance, not separate development and production database infrastructure.

## Initialization

[`DatabaseInitStack`](../../internal/stack/database_init.go) deploys the `InitRDS` container-image Lambda and invokes it through a CloudFormation custom resource. The function has a one-minute timeout, 256 MiB of memory, access to the RDS secret and instance, and a `DatabaseInitializerLogs` log group with seven-day retention.

On an invocation, [`lambda/internal/database/init/main.go`](../../lambda/internal/database/init/main.go) performs these operations in order:

1. Load the database username and password from `DB_SECRET_ARN` and use `DB_HOST` as the RDS endpoint.
2. Connect without a selected database and run `07_11_2025_create_databases_up.sql`, which executes `CREATE DATABASE IF NOT EXISTS` for `STAGING` and `PRODUCTION`.
3. Connect to each database and run the single active table migration, `11_04_2025_create_core_tables_up.sql`.
4. Connect to `STAGING` and run `09_14_2025_seed_tables.sql`.
5. Return `200` with `Database initialization complete`; an initialization error is returned in a `500` HTTP-shaped response body.

`PRODUCTION` receives the schema but no seed data. The staging seed creates sample students, profiles, clubs, memberships, events, descriptions, club-event associations, verified clubs, and admins.

### Initialization behavior

- Table creation uses `CREATE TABLE IF NOT EXISTS`, but the later `ALTER TABLE ... ADD FOREIGN KEY` statements have no equivalent guard. Re-running the initializer against an already initialized schema can fail while adding existing constraints.
- The staging seed uses ordinary `INSERT` statements and fixed identifiers. It is not idempotent and can fail on duplicate primary or unique values.
- The migration runner splits files on semicolons and executes each statement individually without a transaction. A failure leaves earlier statements committed, so a database can be partially initialized.
- There is no migration history table or version check. Only the migration names in `initTableMigrationFiles` execute; the earlier 07/11 and 09/08 core-table files are inactive.
- No 11/04 down migration is present or selected by the initializer.
- The environment lookup reuses one `success` variable and checks it only after reading `DB_HOST`. A missing `DB_SECRET_ARN` is not detected by that condition when `DB_HOST` is present; secret loading instead proceeds with an empty ARN.
- The initializer's MySQL connection has TLS disabled in code. API query connections created by `NewClientFromHost` request the MySQL driver's registered `true` TLS configuration.
- The custom-resource handler uses API Gateway v2 request and response types even though it is invoked by the CDK custom-resource provider. It ignores the incoming event and returns HTTP-shaped fields.
- Because the handler does not inspect the custom-resource request type, provider invocations for create, update, or delete all run the same database initialization sequence.

## Application query access

[`utils/query_client`](../../utils/query_client) is the shared SQL access layer used by database-backed Lambdas:

- SQL files under `utils/query_client/queries` are embedded into each Go binary with `go:embed`.
- Handlers construct a query with a file path and positional arguments; `sqlx` sends the SQL with `?` parameters rather than interpolating request values.
- `NewClientFromHost` loads the username, password, and port from Secrets Manager, combines them with `DB_HOST` and `DB_NAME`, enables the driver's `true` TLS configuration, and opens a direct connection.
- `Get`, `Select`, `Exec`, and the multi-query helpers map query results to Go structs or execute data changes. Transaction helpers exist, but handlers choose whether to use them.

The query layer reads each embedded SQL file into a fixed 4 KiB buffer and does not use the byte count returned by `Read`. Consequently, files longer than 4 KiB are truncated and shorter files are returned with the unread portion of the buffer still present as zero bytes. This is the behavior of the current loader.

The alternate `NewClient` constructor accepts a database name but does not assign it to the MySQL DSN. Active API and identity handlers use `NewClientFromHost`, which does select `DB_NAME`.

## Go models and DDL

The structs in [`database/models`](../../database/models) describe shapes used by selected handlers, not the complete database definition. The package currently contains representations for students, clubs, events, event descriptions, images, and event-image links. It does not model every table, database default, key, or nullable column.

Important differences between the Go layer and DDL include:

- the DDL permits `NULL` in nearly every non-primary-key column, while several model fields are non-pointer Go values and some API validators require them;
- `EventImage.EventID` is a Go `string`, while `event_images.fk_event_id` is an SQL `INT`;
- validation tags such as `required` and `uuid` are application metadata and do not create database constraints; and
- role strings computed by queries and the image `purpose` strings used by handlers are not SQL enums.

Use the active DDL when evaluating persistence constraints and the relevant handler, query, and model together when evaluating an API operation.

## Operational characteristics

- The schema declares no `ON DELETE` or `ON UPDATE` actions. MySQL therefore applies its default restrictive foreign-key behavior; deleting a referenced row fails while dependent rows remain.
- No table declares a storage engine, character set, collation, or schema-level time-zone policy, so instance and database defaults apply.
- With MySQL 8.0's default `explicit_defaults_for_timestamp=ON` (the stack supplies no custom parameter group), `created_at` and `updated_at` columns have no automatic defaults or update expressions. Inserts must supply values when the application needs them.
- The DDL has no `CHECK` constraints for date order, exactly one event-owning club, exactly one club owner, MIME types, image purposes, or boolean combinations.
- Nullable `UNIQUE` columns allow multiple `NULL` values under MySQL semantics. This affects optional usernames, names, thumbnails, verification rows, admin rows, and event descriptions.
- The image bucket is shared by development and production and its keys have no environment prefix, even though image metadata is stored in separate MySQL databases.
