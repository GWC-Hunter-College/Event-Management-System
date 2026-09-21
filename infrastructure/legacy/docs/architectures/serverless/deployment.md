# Deployment guide

Unless a section explicitly changes directories, run commands in this guide from `infrastructure/legacy/`.

This guide deploys the CDK application defined by `cdk-infrastructure.go`. Every CDK command starts the complete Go app, loads `.env` when present, constructs all 11 stacks, and stages Lambda assets. Even a command targeting one stack therefore needs valid identity configuration, the Go toolchain, and Docker access for the database-initializer image.

## Prerequisites

- Go with the module's declared `go1.24.3` toolchain available
- Node.js and an AWS CDK v2 CLI compatible with the pinned CDK libraries (`v2.208.0`)
- AWS CLI credentials for the intended account and Region
- Docker daemon access and enough storage to build the initializer image
- Network access on the first build for Go modules, CDK assets, and container base images
- Google OAuth client credentials configured with every callback/logout URL used below
- AWS permission to bootstrap and deploy CloudFormation, IAM, S3, CloudFront, EC2/VPC, RDS, Secrets Manager, Lambda, API Gateway, Cognito, ECR, Logs, and Systems Manager resources

## Configuration

Copy the tracked template and keep the populated file local:

```bash
cp .env.example .env
```

| Variable | Required behavior |
| --- | --- |
| `GOOGLE_CLIENT_ID` | Must contain the Google OAuth client ID. An empty identity-provider configuration prevents synthesis. |
| `GOOGLE_CLIENT_SECRET` | Must be non-empty. It is passed to the Cognito Google provider during synthesis. |
| `CALLBACK_URLS` | Comma-separated OAuth callback URLs. Whitespace and empty entries are removed. |
| `LOGOUT_URLS` | Comma-separated post-logout URLs. Whitespace and empty entries are removed. |
| `PRODUCTION_STATUS` | Used only by `AuthorizationStack`: case-insensitive `true` sets `DB_NAME=PRODUCTION`; missing/blank/other values set `DB_NAME=STAGING`. |

The development API always uses `STAGING`, and the production API always uses `PRODUCTION`, regardless of `PRODUCTION_STATUS`.
The tracked `.env.example` does not include `PRODUCTION_STATUS`; add it to the local `.env` when the identity-sync target must be explicit.

`.env` and `cdk.out` are ignored by Git. Keep both private: the Google provider secret is supplied as a synthesis-time CDK value and can appear in the generated cloud assembly and CloudFormation deployment data. Do not put credentials in shell examples, documentation, commits, issue text, or build logs.

## Select and verify the AWS environment

The CDK CLI derives `CDK_DEFAULT_ACCOUNT` and `CDK_DEFAULT_REGION` from the selected AWS credentials; the `env()` helper applies them to every stack.

```bash
aws sts get-caller-identity
aws configure get region
cdk --version
docker version
go version
```

If profiles are used, select one consistently for both AWS CLI and CDK commands, for example with `AWS_PROFILE`. Confirm the identity again when changing shells, profiles, or Regions. Fixed S3 bucket names and other explicit physical names can collide with resources already present in the target or another account.

## Bootstrap

Bootstrap each account/Region before its first deployment:

```bash
cdk bootstrap aws://ACCOUNT_ID/REGION
```

Use the account reported by `aws sts get-caller-identity` and the configured Region; do not copy an account ID from context files or examples. CDK uses the bootstrap resources to publish file and container assets.

## Build and synthesize

Run compilation checks for both Go modules:

```bash
go mod download
go test ./...
```

```bash
cd lambda/internal/database/init
go test ./...
cd ../../../..
```

The repository's test files currently contain no active assertions, so these commands primarily compile packages. Then produce and inspect the cloud assembly:

```bash
cdk list
cdk synth --all
cdk diff --all
```

`cdk list` should report exactly the 11 stacks listed in [stacks.md](stacks.md). Synthesis currently emits CDK library notices for the CloudFront S3 origin, Cognito Google client-secret property, and bastion machine-image helper; notices are not deployment success signals, so inspect the final command exit status.

## Ordered deployment

CDK understands the token-derived stack dependencies, but database readiness and Cognito trigger ownership add application ordering outside that graph.

### 1. Deploy independent roots

```bash
cdk deploy \
  FrontendStack \
  FrontendHccStack \
  NetworkStack \
  ImageStack \
  AuthenticationStack
```

The frontend stacks are independent and can be omitted when deploying only the API/data system. `DatabaseStack` cannot precede `NetworkStack`.

### 2. Deploy the database

```bash
cdk deploy DatabaseStack
```

This creates a generated database secret and the single-AZ MySQL instance in isolated subnets. Wait for CloudFormation to finish before initializing or deploying traffic that expects schemas.

### 3. Initialize schemas and staging data

```bash
cdk diff DatabaseInitStack
cdk deploy DatabaseInitStack
```

Creation of the custom resource invokes the initializer. It creates both logical databases and tables, then seeds `STAGING`. A CloudFormation success status is the readiness gate for database-backed API requests. Recreating the custom resource or invoking the image again is data-changing and can fail on duplicate foreign keys or seed rows.

### 4. Deploy administrative access when required

```bash
cdk deploy BastionStack
```

This also creates the shared VPC's SSM interface endpoints and S3 gateway endpoint. It is not a prerequisite in the synthesized graph for either API.

### 5. Deploy identity trigger and APIs deliberately

For the production user-upsert Lambda, set the selection explicitly when deploying `AuthorizationStack`:

```bash
PRODUCTION_STATUS=true cdk deploy AuthorizationStack
```

For a production API deployment:

```bash
cdk deploy ProdApiStack
```

For a development API deployment:

```bash
cdk deploy DevApiStack
```

Deploying `DevApiStack` updates the shared user pool's post-confirmation and post-authentication triggers to `PostConfirmUserUpsertDev`, which writes to `STAGING`. Deploying `AuthorizationStack` afterward updates both triggers to `PostConfirmUserUpsert`, whose database is selected by `PRODUCTION_STATUS`. The last successful custom-resource update controls the shared trigger slots.

## Deployment-sensitive constraints

Review these before using `cdk deploy --all`:

- `ProdApiStack` and `DevApiStack` both request the explicit Lambda names `PostJoinClubMemberMe` and `DeleteClubMemberMe`. A Lambda name can be owned only once per account/Region, so a clean deployment of the second API stack encounters a physical-name conflict. Stack order does not make two CloudFormation stacks share ownership.
- `AuthorizationStack` and `DevApiStack` both replace the same Cognito trigger configuration. There is no CloudFormation dependency between them, so `--all` does not encode which trigger should win.
- The APIs depend on RDS but not on `DatabaseInitStack`; CDK can deploy an API before the schema custom resource succeeds.
- Fixed S3 names are globally unique. A name held in any AWS account blocks bucket creation.
- The Cognito domain prefix and numerous Lambda/IAM names are fixed and constrain parallel copies in the same naming scope.
- `DatabaseInitStack` mutates data and does not provide a rollback migration. CloudFormation rollback of the stack does not undo SQL that already succeeded.

The generic all-stack command is:

```bash
cdk deploy --all
```

It is suitable only when the target's physical-name ownership, desired Cognito trigger, and database readiness have already been accounted for. `cdk synth --all` can succeed even when a later AWS deployment rejects duplicate physical names.

## Verify the deployment

Read outputs from CloudFormation rather than copying values from synthesized templates:

```bash
aws cloudformation describe-stacks \
  --stack-name ProdApiStack \
  --query 'Stacks[0].Outputs' \
  --output table
```

Repeat with `DevApiStack`, `AuthenticationStack`, the frontend stacks, or `BastionStack` as needed. Test the public health route using the endpoint output:

```bash
curl https://API_ID.execute-api.REGION.amazonaws.com/health
```

For the bastion, obtain `InstanceId` and start a Session Manager session:

```bash
aws ssm start-session --target INSTANCE_ID
```

The instance has no SSH ingress or key pair. Database credentials are not granted to its instance role; retrieve them only through an identity authorized for the generated secret.

For frontend releases, upload built assets to the path exposed by the appropriate `S3...Destination`/bucket output and invalidate the matching distribution. Frontend building and upload automation are not part of the legacy implementation.

## Change and rollback safety

- Run `cdk diff STACK_NAME` immediately before every deployment and review IAM, replacement, and removal-policy changes.
- Do not disable the default CDK security-change approval without an established review mechanism.
- Treat replacement or destruction of `DatabaseStack`, `AuthenticationStack`, `FrontendStack`, and `FrontendHccStack` as destructive: RDS and Cognito use `DESTROY`, and both frontend buckets use `DESTROY` with automatic object deletion.
- RDS deletion protection is disabled and backup retention is one day. Confirm an independent recovery path before a destructive operation.
- The image bucket uses the S3 construct's default removal behavior, but retained resources and fixed names can block later re-creation.
- A failed initializer can leave partially applied SQL because statements are executed sequentially. Inspect `DatabaseInitializerLogs` before retrying or replacing the custom resource.

No legacy deployment step creates Route 53 records, custom API/CloudFront domains, or ACM certificates. Post-deployment endpoints remain the generated AWS service domains.
