# Girls Who Code Hunter Website & Event Management System

Website hosting and event-management services for Girls Who Code at Hunter College and the broader Hunter club community. Frontend application source is maintained separately.

## System Overview

This repository brings together two major systems:

- **Frontend delivery:** the infrastructure used to deploy and serve the project's websites, including the Girls Who Code Hunter website and event-management frontend.
- **Event-management backend:** the APIs, data, authentication, and image storage that support club information, events, memberships, and administrative functionality.

## Website Hosting

The hosting layer provides reusable infrastructure for publishing the project's frontend applications. Static builds are stored in Amazon S3 and distributed through CloudFront. When a custom domain is configured, Route 53 manages its DNS and ACM supplies its HTTPS certificate.

One result of this hosting system is the Girls Who Code Hunter website:

**[www.girlswhocodehunter.org](https://girlswhocodehunter.org)**

![Girls Who Code Hunter homepage](docs/assets/gwc-homepage.png)

**[View the Hosting & Deployment Architecture →](hosting/README.md)**

## Event Management Backend

The backend manages event information, club data, memberships, and images for the public website and management interface. Visitors can browse public event and club information without signing in; administrative actions are intended for authorized E-board members managing that content.

Amazon Cognito provides sign-in for the management interface, with application-level club roles supporting authorization. API Gateway exposes the backend, Go Lambda handlers implement its behavior, RDS MySQL persists application data, and S3 stores images.

![High-level architecture showing client applications, website delivery, Cognito, API Gateway, Go Lambda handlers, image storage, and RDS MySQL within AWS CDK and CloudFormation](infrastructure/legacy/docs/assets/high-level-architecture.png)

[Explore the backend documentation](infrastructure/legacy/README.md)

## Modular Backend Architecture

The backend is being separated into three primary modules so application behavior and cloud infrastructure can evolve independently.

### Database

`database/` is the persistence boundary for schema, migrations, queries, and database behavior. Its purpose is to keep data logic independent of the AWS service used to host the database.

### API

`api/` is the application boundary for routes, service behavior, and the interface through which frontends read or modify data. API Gateway and Lambda provide deployment mechanisms; they should not define the application itself.

### Infrastructure

`infrastructure/` owns cloud-specific deployment concerns, with `infrastructure/aws/` defining the modular AWS boundary. Provisioning and connecting compute, database hosting, identity, storage, and networking belong here rather than in the API or database modules.

The previous integrated implementation remains in `infrastructure/legacy/` while functionality moves into these boundaries. It combines API code, database integration, and AWS deployment definitions; the refactor separates those responsibilities while preserving the implementation during migration.

## Repository Structure

```text
hosting/                  Frontend hosting and delivery infrastructure
database/                 Database and persistence layer
api/                      Backend API/application layer
infrastructure/           Backend cloud deployment infrastructure
├── aws/                  Modular AWS deployment boundary
└── legacy/               Previous integrated implementation
```

## License

See [LICENSE](LICENSE).
