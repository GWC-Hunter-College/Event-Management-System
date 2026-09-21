# AWS Backend Infrastructure

This directory is the future boundary for modular AWS infrastructure used to operate the backend API and database. It is a documented placeholder in this first modularization phase; no API, Amazon EC2, Amazon RDS, or other backend infrastructure has been implemented or moved here yet.

Future infrastructure may cover:

- compute
- networking
- database hosting
- security
- secrets
- observability
- backups

The API application and MySQL implementation should remain independent of these AWS deployment choices. The current integrated AWS implementation remains under [`../legacy/`](../legacy/).
