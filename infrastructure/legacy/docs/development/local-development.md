# Local development

Unless a section explicitly changes directories, run commands in this guide from `infrastructure/legacy/`.

## What runs locally

The preserved legacy implementation is a Go CDK application plus Lambda handler source; it does not contain a local HTTP server, SAM template, LocalStack configuration, Makefile, or `scripts/` directory. The normal local loop compiles/tests Go packages and synthesizes CloudFormation. API Gateway, Cognito, S3, Secrets Manager, and RDS behavior is exercised against a deployed AWS environment.

The database initializer has a Docker Compose/RIE harness, but it still calls AWS Secrets Manager and a MySQL endpoint. It is not a self-contained local database.

## Tooling

Install or provide:

- Go matching the legacy module's `go 1.23.0` language directive and `go1.24.3` toolchain declaration;
- Node.js and the AWS CDK v2 CLI;
- AWS CLI credentials for synthesis context and deployed-service checks;
- Docker for the initializer image asset and local RIE harness; and
- internet access when modules or container images are not cached.

Verify the active tools and AWS identity:

```bash
go version
cdk --version
docker version
aws sts get-caller-identity
aws configure get region
```

## Environment setup

Create the ignored local file:

```bash
cp .env.example .env
```

Populate it without committing or sharing the values:

```dotenv
GOOGLE_CLIENT_ID=<google-oauth-client-id>
GOOGLE_CLIENT_SECRET=<google-oauth-client-secret>

CALLBACK_URLS=http://localhost:5173
LOGOUT_URLS=http://localhost:5173
PRODUCTION_STATUS=false
```

`CALLBACK_URLS` and `LOGOUT_URLS` can contain comma-separated URLs. Each URL must also be allowed by the Google/Cognito application configuration. `PRODUCTION_STATUS=false` is shown only to make the local staging selection explicit; missing or any non-`true` value has the same effect in `AuthorizationStack`.

The app calls `godotenv.Load()` and falls back to process environment variables if `.env` is absent. `GOOGLE_CLIENT_SECRET` cannot be empty during synthesis. Keep `cdk.out` private as well as `.env`, because synthesis-time identity-provider values can appear in the cloud assembly.

The CDK CLI sets `CDK_DEFAULT_ACCOUNT` and `CDK_DEFAULT_REGION` for its app subprocess. A direct `go run cdk-infrastructure.go` does not infer them by itself; prefer CDK commands or set those values deliberately when debugging the entry point.

## Install and compile

The legacy module contains the CDK app, API handlers, and shared packages:

```bash
go mod download
go test ./...
```

The database initializer is a nested Go module and needs its own check:

```bash
cd lambda/internal/database/init
go mod download
go test ./...
cd ../../../..
```

There are currently no active test assertions in the root CDK test file and no package test files, so a green run confirms compilation rather than behavioral coverage.

Format and vet changed Go packages before review:

```bash
gofmt -w path/to/changed.go
go vet ./...
```

## Synthesize and inspect

`cdk.json` defines the app command as:

```text
go mod download && go run cdk-infrastructure.go
```

Use the CDK CLI for the normal feedback loop:

```bash
cdk list
cdk synth --all
cdk diff STACK_NAME
```

Expected `cdk list` output is the 11-stack inventory in [the stack catalog](../architectures/serverless/stacks.md). The stub stack is intentionally absent because its entry-point constructor is commented out.

Synthesis bundles each active `GoFunction` and builds the Docker image asset for `DatabaseInitStack`; Docker can therefore be required even when inspecting a different stack. The first run can be substantially slower while Go modules, CDK/jsii packages, Lambda binaries, and container layers are cached.

The CDK watch configuration includes legacy implementation files broadly while excluding its README, `cdk*.json`, Go module files, and tests. If using watch against a development stack, first inspect the initial diff and remember that hotswap behavior changes AWS resources:

```bash
cdk watch DevApiStack
```

## Code organization and route workflow

| Location | Development role |
| --- | --- |
| `internal/stack/` | Stack boundaries and shared-resource wiring |
| `gateway/routes/` | API paths, methods, and authorizer attachment |
| `gateway/integrations/` | Lambda constructs, entries, environment variables, integrations, and grants |
| `gateway/parameters/` | Shared RDS/S3 construct parameters |
| `lambda/api/` | API Gateway v2 handlers |
| `lambda/internal/auth/` | Cognito-trigger handler |
| `utils/query_client/` | Secret loading, MySQL connections, embedded SQL, and query helpers |
| `database/models/` | Shared data models |

Production and development call the same route helpers and handler entries. Their stack constructors provide `deploymentTarget` suffixes and different `DB_NAME` values. When reviewing a route change, trace all four layers: route method/path and authorizer, integration entry/environment, IAM/S3/RDS grants, and handler behavior.

Some source helpers are inactive: `DatabaseRoutes`, the read club-thumbnail integration, the read event-thumbnail integration, RDS Proxy code, and `StubLambdaStack` are not called by either active API stack. `gateway/stagingApi/README.md` describes an earlier directory convention, while the active development API uses the shared `gateway/routes` package.

## Local Lambda checks

Individual handler packages can be compiled directly, for example:

```bash
go test ./lambda/api/events/...
go test ./lambda/api/clubs/...
go test ./lambda/api/me/...
```

Running a handler with `go run` starts the AWS Lambda runtime client; it does not expose a local HTTP API. Handler events are API Gateway v2 payloads, and database-backed handlers also require `DB_SECRET_ARN`, `DB_HOST`, `DB_NAME`, AWS credentials, private network reachability, and the deployed security/IAM boundary. Prefer unit tests around handler/query logic or invoke a deployed development API for an end-to-end check.

## Database initializer with Docker and RIE

`lambda/internal/database/init/docker-compose.yml` builds the initializer image, mounts the AWS Lambda Runtime Interface Emulator from `~/.aws-lambda-rie`, and exposes the Lambda runtime endpoint on local port `9000`.

Before running it:

1. install the RIE binary at the mounted path;
2. supply a database secret ARN belonging to the selected account—do not rely on the tracked example ARN;
3. supply `DB_HOST`, which the Compose file does not currently define;
4. make AWS credentials and Region available to the container so the SDK can call Secrets Manager; and
5. provide network reachability to the private RDS endpoint. An isolated RDS endpoint is not reachable from an ordinary local Docker network without an authorized private-network or SSM forwarding path.

Keep environment-specific overrides outside the repository or in an ignored private file. Do not paste the secret value into Compose; the initializer needs the secret **ARN** and loads the value through AWS SDK credentials.

After those prerequisites are in place, the harness commands are:

```bash
cd lambda/internal/database/init
docker compose build lambda
docker compose up -d lambda
curl http://localhost:9000/2015-03-31/functions/function/invocations -d '{}'
docker compose down
cd ../../../..
```

The invocation creates both logical databases/tables and seeds `STAGING`. It changes live data. Although database/table creation uses `IF NOT EXISTS`, foreign-key additions and seed inserts are not fully idempotent; do not repeatedly invoke it against an initialized database. The Lambda's one-minute deployed timeout also applies operational pressure even if a local container is allowed to run longer.

## Testing a deployed development API

Read the endpoint from CloudFormation:

```bash
aws cloudformation describe-stacks \
  --stack-name DevApiStack \
  --query 'Stacks[0].Outputs[?OutputKey==`myHttpApiEndpoint`].OutputValue' \
  --output text
```

The health route is public:

```bash
curl https://API_ID.execute-api.REGION.amazonaws.com/health
```

JWT routes require a token issued for the shared Cognito user pool/client. Do not commit tokens or put them in reusable command examples. Image upload handlers return presigned S3 operations; test both the API response and the subsequent direct S3 transfer.

## Development constraints

- Both API stacks explicitly name the join and leave Lambdas without `Prod`/`Dev` suffixes. Deploying both into one account/Region creates a physical-name ownership conflict even though synthesis succeeds.
- `AuthorizationStack` and `DevApiStack` both update the same Cognito post-confirmation/post-authentication trigger slots; the last deployment controls the active upsert Lambda.
- Development CORS does not advertise `DELETE`, although the route exists, and both APIs allow all origins/headers.
- Several mutating routes do not attach the JWT authorizer, including membership deletion and image/thumbnail writes.
- VPC Lambdas have no NAT path. New runtime calls to public AWS or internet endpoints require an already-declared reachable endpoint/path; compilation success does not verify runtime egress.
- Development and production schemas share one RDS instance, secret, endpoint, backup policy, and capacity boundary.
- Fixed bucket, domain, IAM, and Lambda names limit parallel stacks and test copies.
- RDS, Cognito, and frontend stack destruction can delete data. Use `cdk diff` and the [deployment safety guidance](../architectures/serverless/deployment.md#change-and-rollback-safety) before applying changes.
- No Route 53 records, ACM certificates, API custom domains, WAF, API access logs, or alarms are created by the active code.

For the full resource and route description, see [infrastructure.md](../architectures/serverless/infrastructure.md).
