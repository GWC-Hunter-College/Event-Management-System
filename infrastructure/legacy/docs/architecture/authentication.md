# Authentication

Amazon Cognito supplies user registration, sign-in, social federation, and JWT validation for the HTTP APIs. The development and production APIs share one user pool and one public browser app client. Authentication is attached to individual API Gateway routes rather than enabled as an API-wide default.

This page describes identity proof and synchronization. Database role decisions are separate and are documented in [Authorization](authorization.md).

## Cognito resources

[`AuthenticationStack`](../../internal/stack/authentication.go) creates the following configuration:

| Concern | Current configuration |
| --- | --- |
| Self-service registration | Enabled. |
| Native sign-in alias | Email. Email is required and mutable. |
| Account recovery | Email only. |
| Password policy | Minimum 7 characters; lowercase, uppercase, digit, and symbol requirements are all disabled; temporary passwords are valid for 7 days. |
| User-pool retention | CloudFormation removal policy is `DESTROY`. |
| Hosted UI | Cognito domain with the fixed prefix `event-manager-authz`. |
| Native client auth flow | User-password authentication enabled. |
| Browser client secret | None; the app client is public. |
| OAuth flow | Authorization code grant. |
| OAuth scopes | `openid`, `email`, and `profile`. |
| Identity providers | Cognito and Google. |
| Token validity | Access token 1 hour, ID token 1 hour, refresh token 30 days. |
| Token revocation | Enabled. |
| User-existence errors | Prevented by the app-client setting. |

The Google provider requests `openid`, `email`, and `profile` and maps Google email, given name, family name, and profile picture into Cognito attributes. Cognito requires the Google provider to exist before creating the web client.

The stack outputs the user-pool ID, app-client ID, hosted-UI base URL, and configured callback and logout URL strings. It does not output the Google client secret.

## Deployment configuration inputs

Authentication synthesis reads these environment variables without committing their values:

| Variable | Use |
| --- | --- |
| `GOOGLE_CLIENT_ID` | Google identity-provider client identifier. |
| `GOOGLE_CLIENT_SECRET` | Google identity-provider credential. |
| `CALLBACK_URLS` | Comma-separated OAuth callback URLs. Values are trimmed and blank entries are discarded. |
| `LOGOUT_URLS` | Comma-separated post-logout URLs, processed the same way. |
| `PRODUCTION_STATUS` | Selects the identity-sync Lambda's database in `AuthorizationStack`: case-insensitive `true` selects `PRODUCTION`; every other value selects `STAGING`. |
| `CDK_DEFAULT_ACCOUNT`, `CDK_DEFAULT_REGION` | Select the AWS environment used by all stacks. |

Database host, schema name, and secret ARN are injected into the identity-sync and request Lambdas by CDK. Database credentials stay in Secrets Manager and are loaded at runtime.

## JWT request path

The production API receives its authorizer from [`AuthorizationStack`](../../internal/stack/authorization.go). The development API creates a corresponding authorizer inside [`developmentApi.go`](../../internal/stack/developmentApi.go). Both are `HttpUserPoolAuthorizer` instances configured with the shared user pool and shared app client.

For a protected route:

1. The client sends the Cognito token in `Authorization: Bearer <token>`.
2. API Gateway validates it against the configured user pool and app client.
3. API Gateway places verified claims in `requestContext.authorizer.jwt.claims`.
4. Lambda reads `sub` and, where needed, `email` from that map.
5. The handler uses `sub` as the primary key in `students` and as the caller identity in membership or author fields.

The Lambdas do not decode or verify a raw bearer token themselves. A route without the CDK authorizer does not gain a JWT claims context merely because the request includes an `Authorization` header.

The active handlers consume no Cognito groups, custom role claims, or OAuth scopes. The stable caller key is `sub`; `email` is treated as optional by the helper code.

## Routes with JWT validation

Exactly six active business routes attach the Cognito authorizer:

| Route | Identity use after validation |
| --- | --- |
| `GET /me` | Read the caller's student row. |
| `GET /me/clubs` | Read memberships for the caller's `sub`. |
| `GET /me/events` | Read events for clubs joined by the caller. |
| `POST /clubs` | Ensure the caller has a student row, then use only authenticated identity; no privilege check. |
| `POST /clubs/{clubId}/events` | Store the caller's `sub` as event author; no club-role check. |
| `POST /clubs/{clubId}/members/me` | Insert a regular membership for the caller. |

API Gateway rejects missing or invalid tokens before invoking these Lambdas. The exact public and protected route inventory is in the [API reference](../api/README.md).

## Cognito-to-database identity synchronization

The sync handler is [`lambda/internal/auth/postConfirm/upsert.go`](../../lambda/internal/auth/postConfirm/upsert.go). Its active SQL is [`students/UPSERT_student.sql`](../../utils/query_client/queries/students/UPSERT_student.sql):

```sql
INSERT INTO students (id, email)
VALUES (?, ?)
ON DUPLICATE KEY UPDATE
  email = VALUES(email);
```

The same handler accepts two Cognito trigger sources:

- `PostConfirmation_ConfirmSignUp`, which creates or updates the row after sign-up confirmation; and
- `PostAuthentication_Authentication`, which repeats the upsert after authentication and captures an email change.

Other trigger-source strings return the original event without a write. If `sub` is absent, the handler logs and skips the write. Database and affected-row errors are also logged and swallowed: the original Cognito event is returned without an error, so a synchronization failure does not block confirmation or authentication.

### Trigger wiring in the synthesized app

Two active stacks create a sync Lambda and call Cognito `UpdateUserPool` for the same shared user pool:

| Stack | Function name | Database selection |
| --- | --- | --- |
| `AuthorizationStack` | `PostConfirmUserUpsert` | `PRODUCTION` only when `PRODUCTION_STATUS` is case-insensitive `true`; otherwise `STAGING`. |
| `DevApiStack` | `PostConfirmUserUpsertDev` | Always `STAGING`. |

Each custom resource writes both `PostConfirmation` and `PostAuthentication` in the user pool's single `LambdaConfig`. Cognito has one configured function per trigger slot, so the stack whose `UpdateUserPool` call runs last determines which of these Lambda ARNs remains attached. The source does not establish a cross-stack ordering between those two updates. Deleting either custom resource sends an empty `LambdaConfig`, clearing the user pool's trigger configuration rather than only its own ARN.

Both sync Lambdas run in the VPC, receive permission for Cognito to invoke them from this user pool, read the RDS secret, and connect to the database selected above.

## Request-time student synchronization

All six JWT-attached route handlers also call [`RequireStudent`](../../utils/auth/ensure_student.go). This fallback checks for `students.id = sub` and creates the row when it is missing:

- when an email claim exists, it uses the same ID-and-email upsert;
- without email, it inserts the ID with a null email and preserves any existing email on a duplicate key; and
- when the row already exists, it does not update its email.

This fallback allows protected request handling to establish a missing student even when a Cognito trigger write did not occur. Depending on the handler, a missing `sub` returns `400`, while database synchronization failures return either `400` or `500` as documented per endpoint.

The database column permits a null email, but [`models.Student`](../../database/models/student.go) exposes email as a non-nullable Go string. `GET /me` can therefore fail while scanning a student row that was created by the sub-only path.

## Authentication boundary

Cognito validation proves which user supplied a protected request. It does not prove that the student is an admin, a club member, an e-board member, or a club owner. Those states live in MySQL, and the current handlers enforce them only where explicitly described in [Authorization](authorization.md).
