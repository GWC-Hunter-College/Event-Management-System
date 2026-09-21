# Infrastructure

This area contains provider-specific infrastructure for operating the backend API and database. It is separate from the provider-independent application boundaries in [`api/`](../api/) and [`database/`](../database/), and from frontend website infrastructure in [`hosting/`](../hosting/).

## Areas

- [`aws/`](aws/) is a documented placeholder for the future modular AWS infrastructure used to host the API and database. No new backend infrastructure is implemented there yet.
- [`legacy/`](legacy/) preserves the currently working, integrated Event Management System implementation while the repository is modularized. Its application code, database integration, AWS CDK stacks, and detailed documentation remain together in that area.
