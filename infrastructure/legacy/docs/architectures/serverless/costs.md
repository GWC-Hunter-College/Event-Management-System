# AWS cost characteristics

This page explains which resources in the current CDK application generate AWS charges. It is not a quote. Rates vary by Region, account age, usage tier, discounts, and AWS pricing changes. Use the [AWS Pricing Calculator](https://calculator.aws/) with the target account and Region before deployment.

## Baseline resources

These resources can accrue charges even when no user calls the API:

| Resource | Provisioned shape | Billing dimensions |
| --- | --- | --- |
| RDS for MySQL | One Single-AZ `db.t3.micro`; 20 GiB initially, autoscaling up to 100 GiB; one-day backup retention | DB instance time, provisioned storage, backup or snapshot storage, data transfer, and burstable CPU credits when applicable |
| Secrets Manager endpoint | One interface endpoint network interface in one isolated subnet | Endpoint-interface hours and processed data |
| Bastion endpoints | SSM, SSM Messages, and EC2 Messages interface endpoints in isolated subnets across up to two AZs | One endpoint-interface hour per service per AZ and processed data |
| Bastion host | One Linux `t3.micro` in an isolated subnet | Running EC2 time and its EBS root volume; stopping the instance stops compute charges but not provisioned EBS charges |
| RDS secret | One CDK-generated database credential | Secret-month and API calls |
| CloudFront | Four distributions across the two frontend stacks | Requests, data transfer, and invalidations beyond the applicable allowance |
| Website S3 | Two buckets | Stored bytes, requests, retrieval or transfer where applicable, and object removal operations |
| Image S3 | One retained bucket | Stored bytes, PUT and GET requests, and transfer; no lifecycle rule removes unconfirmed uploads |
| CloudWatch Logs | Lambda log groups, including a database-initializer group with one-week retention | Ingested bytes, retained bytes, and query scanning |

The VPC deliberately creates no NAT gateway, so there is no NAT gateway hourly or per-GB charge. `BastionStack` adds an S3 gateway endpoint, for which AWS documents [no additional endpoint charge](https://docs.aws.amazon.com/vpc/latest/privatelink/vpc-endpoints-s3.html).

## Interface endpoint calculation

AWS bills each interface endpoint network interface by the hour. With the repository's default two-AZ VPC layout:

- the Secrets Manager endpoint is pinned to one isolated subnet, yielding one endpoint network interface;
- each of the three bastion and SSM endpoint services selects the isolated subnets, yielding up to two interfaces per service; and
- deploying all stacks therefore yields up to seven billable endpoint network interfaces.

Using the `$0.01` per endpoint ENI-hour rate currently published by AWS, a 730-hour illustrative month is:

```text
7 endpoint ENIs × 730 hours × $0.01 = $51.10/month
```

Data processing is additional. This formula is an example, not an application-wide monthly total; a different Region, selected AZ count, or price changes it. Check [AWS PrivateLink pricing](https://aws.amazon.com/privatelink/pricing/) before estimating a deployment.

## Usage-based services

| Service | What increases cost in this repository |
| --- | --- |
| API Gateway HTTP API | API requests and outbound data transfer. |
| Lambda | Invocations and execution duration, weighted by memory. The database initializer uses 256 MiB; route functions otherwise use their construct defaults. |
| Cognito | Monthly active users and the user-pool feature tier; Google sign-in is configured as a social identity provider. |
| Secrets Manager | Lambda credential retrievals in addition to secret storage. |
| S3 | Website and image storage, signed PUTs and GETs, listing, deletion, and transfer. |
| CloudFront | Viewer requests and bytes delivered by four distributions. |
| CloudWatch | Lambda log ingestion, retention, and log-query scanning. |

AWS maintains the current rates: [RDS for MySQL](https://aws.amazon.com/rds/mysql/pricing/), [API Gateway](https://aws.amazon.com/api-gateway/pricing/), [Lambda](https://aws.amazon.com/lambda/pricing/), [Cognito](https://aws.amazon.com/cognito/pricing/), [Secrets Manager](https://aws.amazon.com/secrets-manager/pricing/), [S3](https://aws.amazon.com/s3/pricing/), [CloudFront](https://aws.amazon.com/cloudfront/pricing/), [EC2](https://aws.amazon.com/ec2/pricing/on-demand/), [EBS](https://aws.amazon.com/ebs/pricing/), and [CloudWatch](https://aws.amazon.com/cloudwatch/pricing/).

## Cost centers by stack

| Stack group | Cost centers |
| --- | --- |
| `NetworkStack` | One Secrets Manager interface endpoint. |
| `DatabaseStack` | RDS instance, storage and backups, and one generated secret. |
| `BastionStack` | One EC2 instance, EBS root storage, and up to six interface endpoint network interfaces across two AZs; the S3 gateway endpoint has no added endpoint fee. |
| `ProdApiStack`, `DevApiStack`, `AuthorizationStack` | API requests, Lambda requests and duration, Secrets Manager reads, and logs. Deploying both APIs duplicates most route Lambdas. |
| `ImageStack` | S3 image storage and requests. |
| `FrontendStack`, `FrontendHccStack` | Two website buckets and four CloudFront distributions in total. |
| `DatabaseInitStack` | Container image asset storage, custom-resource provider and Lambda execution, and one log group configured for seven-day retention. |

## Historical estimate context

The project planning PDF recorded a point-in-time backend estimate of `$22.97` per month excluding its database-access host. That number is not a forecast for the current source. The PDF assumed 8 GB of database storage, while CDK provisions 20 GiB, and its interface-endpoint subtotal does not reflect the current three SSM services spread across up to two AZs. It also included a Route 53 domain cost, while the repository provisions no Route 53 resources.

For an estimate that matches a deployment, inventory the synthesized templates, confirm the target AZ count, and enter measured request, execution, storage, and transfer volumes in the AWS Pricing Calculator.
