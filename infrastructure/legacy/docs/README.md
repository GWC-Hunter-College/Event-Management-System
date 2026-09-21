# Backend and infrastructure documentation

These documents describe the preserved integrated implementation under `infrastructure/legacy/`. They are not a description of the planned modular architecture at the repository root.

This directory documents the backend that stores and serves Hunter College club and event data: the AWS CDK application, Go Lambda APIs, MySQL schema, Cognito authentication, database-backed authorization, and S3 image workflows.

The public club website and the administrative/event-management application are separate frontend repositories that consume these APIs. The preserved implementation provisions their AWS delivery layer along with the backend infrastructure.

Start with the [architecture overview](architecture/overview.md) for the end-to-end request path. Use the topic guides below for implementation details.

## Documentation map

| Area | Documents |
| --- | --- |
| System design | [Architecture overview](architecture/overview.md), [authentication](architecture/authentication.md), [authorization](architecture/authorization.md), [image uploads](architecture/image-uploads.md) |
| AWS and CDK | [Serverless overview](architectures/serverless/overview.md), [infrastructure](architectures/serverless/infrastructure.md), [stack catalog](architectures/serverless/stacks.md), [deployment](architectures/serverless/deployment.md), [costs](architectures/serverless/costs.md) |
| HTTP API | [API reference](api/README.md), [student and identity routes](api/students.md), [club routes](api/clubs.md), [event routes](api/events.md), [image routes](api/images.md) |
| Data | [Database overview](database/README.md), [schema reference](database/schema.md) |
| Contributor setup | [Local development](development/local-development.md), [deployment workflow](architectures/serverless/deployment.md) |

## Implementation sources

The active CDK entrypoint, route registration, Lambda handlers, and current database initializer are authoritative:

- [`cdk-infrastructure.go`](../cdk-infrastructure.go) selects the stacks that synthesize.
- [`gateway/routes/`](../gateway/routes) defines the routes registered on the development and production HTTP APIs.
- [`lambda/api/`](../lambda/api) and [`lambda/internal/`](../lambda/internal) contain the active handlers and shared backend behavior.
- [`11_04_2025_create_core_tables_up.sql`](../lambda/internal/database/init/migrations/11_04_2025_create_core_tables_up.sql) defines the schema created by a fresh database initialization.

The repository also contains a commented-out stub stack, an unregistered database-test route helper, and earlier SQL snapshots. They are not part of the active API or fresh-install schema unless the CDK entrypoint or an active handler references them.
