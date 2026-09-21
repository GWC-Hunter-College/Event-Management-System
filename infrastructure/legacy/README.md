# Legacy integrated Event Management System implementation

> This directory preserves the currently working, all-in-one Event Management System while the repository is modularized in phases. It is the active historical implementation, not a broken or abandoned copy. The first phase changes its location and only the minimum path documentation needed for that move; it does not migrate or redesign the database, API, authentication, authorization, or AWS backend resources.

Run commands in this guide from `infrastructure/legacy/`, the directory containing this implementation's `go.mod` and `cdk.json`, unless a command explicitly changes directories.

This preserved implementation is the backend, API, data, authentication/authorization, and AWS infrastructure layer for Hunter College's club event-management ecosystem. It stores and serves club, event, description, tag, membership, permission, and image metadata for Girls Who Code at Hunter and the broader club-event experience.

The public-facing club website and the administrative/event-management application live in separate repositories and consume this backend. The preserved implementation also provisions their S3 and CloudFront delivery infrastructure; it does not contain or build the frontend source.

## High-level architecture

![High-level architecture showing separate clients using CloudFront and private website buckets, Cognito, API Gateway, Lambda, an image bucket, and RDS](docs/assets/high-level-architecture.png)

Client applications load the separately built frontends through CloudFront, authenticate with Cognito when required, and call the development or production HTTP API. Go Lambda handlers read and write MySQL data and create signed S3 requests; image bytes then transfer directly between the client and the private image bucket.

The CDK application synthesizes **11 active stacks** for frontend delivery, networking, authentication, authorization, HTTP APIs, image storage, MySQL, database initialization, and private administrative access.

The application currently provides:

- public club and posted-event discovery, authenticated caller and membership reads, club and event creation, and direct image-upload primitives;
- two private S3 website origins, each served through two CloudFront distributions;
- separate production and development API Gateway HTTP APIs backed by Go Lambda functions;
- a Cognito user pool with Cognito and Google sign-in;
- a shared VPC, RDS MySQL instance, generated Secrets Manager secret, and logical `STAGING` and `PRODUCTION` databases;
- a private S3 image bucket used through presigned requests; and
- an SSM-managed bastion host in an isolated subnet.

The source of truth for stack composition is [`cdk-infrastructure.go`](cdk-infrastructure.go). The `StubLambdaStack` source remains in the repository but its constructor is commented out, so it is not one of the 11 active stacks.

## Documentation

- [Architecture overview](docs/architecture/overview.md) — detailed system relationships and request flows
- [Serverless implementation overview](docs/architectures/serverless/overview.md) — AWS boundaries, environments, and resources
- [Infrastructure reference](docs/architectures/serverless/infrastructure.md) — VPC, APIs, Lambda, Cognito, RDS, S3, CloudFront, and IAM
- [Stack catalog](docs/architectures/serverless/stacks.md) — all 11 stacks, dependencies, values, and outputs
- [API reference](docs/api/README.md) — active routes, request/response contracts, and route-level authentication
- [Database reference](docs/database/README.md) — provisioning, initialization, and the implemented 13-table schema
- [Authentication and authorization](docs/architecture/authentication.md) — Cognito identity flow and database role boundaries
- [Image uploads](docs/architecture/image-uploads.md) — signed S3 transfers and event-gallery confirmation
- [Deployment guide](docs/architectures/serverless/deployment.md) — bootstrap, synthesis, ordering, deployment, and verification
- [Local development](docs/development/local-development.md) — environment setup, tests, bundling, and local database-initializer invocation
- [Documentation index](docs/README.md) — complete topic map

## Prerequisites

- Go compatible with the legacy module's `go 1.23.0` directive and `go1.24.3` toolchain declaration
- AWS CDK v2 CLI and Node.js
- AWS CLI credentials for the target account and Region
- Docker for the database-initializer image asset
- Google OAuth client credentials and valid callback/logout URLs

The CDK CLI supplies `CDK_DEFAULT_ACCOUNT` and `CDK_DEFAULT_REGION` from the selected AWS credentials. All stacks use that same concrete environment.

## Quick start

Create a local environment file and replace every blank value. `.env` is ignored by Git.

```bash
cp .env.example .env
```

`GOOGLE_CLIENT_ID` and `GOOGLE_CLIENT_SECRET` must be non-empty for synthesis. `CALLBACK_URLS` and `LOGOUT_URLS` accept comma-separated URLs. `PRODUCTION_STATUS` is also read by `AuthorizationStack`: only the case-insensitive value `true` selects the `PRODUCTION` database; every other value selects `STAGING`. The tracked template does not include that variable, so add it to `.env` when the trigger database selection must be explicit.

Confirm the active AWS identity before doing any CDK operation that can change resources:

```bash
aws sts get-caller-identity
aws configure get region
```

Install modules, compile the repository, and synthesize the application:

```bash
go mod download
go test ./...
cdk list
cdk synth
```

Bootstrap each account/Region once, then inspect changes before deployment:

```bash
cdk bootstrap aws://ACCOUNT_ID/REGION
cdk diff --all
```

Read the [deployment guide](docs/architectures/serverless/deployment.md) before running `cdk deploy`. The active stacks share generated values and some explicitly named resources, and the database-initializer custom resource changes live data.

## Repository layout

| Path | Purpose |
| --- | --- |
| `cdk-infrastructure.go` | CDK application entry point and the 11 active stack constructors |
| `internal/stack/` | Stack definitions for frontends, network, database, images, identity, APIs, bastion, and initialization |
| `gateway/routes/` | HTTP route declarations shared by the production and development APIs |
| `gateway/integrations/` | Lambda constructs, API integrations, environment variables, and resource grants |
| `lambda/api/` | Go handlers for deployed HTTP endpoints |
| `lambda/internal/auth/` | Cognito post-confirmation/post-authentication database upsert handler |
| `lambda/internal/database/init/` | Docker-image Lambda and SQL migrations for database initialization |
| `utils/` | Shared authorization, validation, and MySQL query helpers |
| `database/models/` | API/database model types |
| `stub/` | Stub handlers and SQL retained in source; the stub stack is inactive |
| `.env.example` | Non-secret template for Cognito/Google redirect configuration |
| `cdk.json` | CDK command (`go mod download && go run cdk-infrastructure.go`) and feature flags |

There is no project `scripts/` directory, Makefile, or package script layer; development and deployment commands are run directly through Go, Docker, AWS CLI, and the CDK CLI.

## Important lifecycle behavior

Several resources intentionally have destructive removal behavior in the current definitions: both frontend buckets delete their objects on stack removal, and the Cognito user pool and RDS instance use `DESTROY`. The RDS instance also has deletion protection disabled. Treat `cdk destroy`, stack replacement, and changes that force replacement as data-impacting operations.

No Route 53 hosted zones, DNS records, or ACM certificates are defined here. CloudFront uses its generated domain and managed HTTPS certificate; the APIs use `execute-api` endpoints and Cognito uses its hosted-domain URL.

## License

See the repository [LICENSE](../../LICENSE).
