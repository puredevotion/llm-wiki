# ADR 0001: Persistence Strategy

## Status

Accepted for initial scaffold.

## Context

The system needs durable knowledge storage, full-text search, graph-like relationships, temporal reasoning, mobile sync, and background re-evaluation. It should remain operationally light and favor embedded or single-binary components.

Kuzu is no longer a suitable planned dependency: the upstream GitHub repository was archived by its owner on October 10, 2025. A separate graph database would also increase operational complexity before the graph workload is proven.

## Decision

Use Turso/Limbo-compatible SQL as the preferred authoritative persistence direction for the initial backend.

Keep the schema portable across Turso/Limbo and SQLite-compatible tooling where practical. Treat graph, search, vector, and object storage as adapters or projections, not independent sources of truth.

Turso Database is still marked beta by its upstream README, so production hardening must include backups, migration rehearsal, and a compatibility fallback path before real user data depends on it.

## Store Choices

| Concern | Initial Choice | Rationale |
| --- | --- | --- |
| Durable records | Turso/Limbo-compatible SQL | Embedded, SQLite-compatible, Rust implementation direction, sync roadmap |
| Graph relationships | Typed edge tables in the primary store | Keeps who/what/when relationships in the source of truth |
| Advanced graph traversal | Recursive CTEs and indexed edge queries first | Avoid abandoned or immature graph dependencies in v1 |
| Full-text content search | Adapter seam; use built-in compatible FTS when available | Search capabilities differ across engines and bindings, so isolate it |
| JSON metadata | Portable JSON stored in text columns, upgraded to JSON functions where supported | Avoid coupling core schema to optional extensions |
| Temporal model | Event log plus `valid_from`, `valid_to`, `observed_at`, `recorded_at` fields | Avoid a separate temporal DB; preserves auditability |
| Raw artifacts | Filesystem object store with DB references | Keeps large PDFs/media out of hot relational tables |
| Vector search | Adapter seam; evaluate Turso vector support and alternatives after retrieval tests | Avoid locking into a vector engine before retrieval quality is measured |

## Consequences

- The first backend can run as one Go binary plus local database/files.
- The project aligns with Turso/Limbo rather than vanilla SQLite as the preferred database path.
- SQLite-compatible fallback remains useful for development, tests, and platform gaps.
- Graph queries are modeled explicitly in the schema instead of outsourced to a graph DB.
- Search and vector behavior must be isolated behind interfaces because Turso/Limbo features are still evolving.
- Early production use needs conservative backup and restore procedures because the preferred engine is still evolving.

## References

- Turso/Limbo repository: https://github.com/tursodatabase/limbo
- Turso embedded replicas and sync direction: https://docs.turso.tech/features/embedded-replicas/introduction
- Kuzu archived repository notice: https://github.com/kuzudb/kuzu
