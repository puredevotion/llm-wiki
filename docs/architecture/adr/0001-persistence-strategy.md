# ADR 0001: Persistence Strategy

## Status

Accepted for initial scaffold.

## Context

The system needs durable knowledge storage, full-text search, direct graph relationships, temporal reasoning, offline-first mobile sync, and background re-evaluation. It should remain operationally light and favor embedded or single-binary components where possible.

Kuzu is no longer a suitable planned dependency: the upstream GitHub repository was archived by its owner on October 10, 2025.

## Decision

Use two first-class persistence stores:

- Turso/Limbo-compatible SQL for durable content, operation logs, sync state, events, audit records, and database-stored zettelkasten-inspired notes.
- LadybugDB as a separate graph database for direct who/what/when relationship storage and traversal.

Content is not stored as markdown files. Zettelkasten is a modeling inspiration only.

## Store Choices

| Concern | Initial Choice | Rationale |
| --- | --- | --- |
| Durable content | Turso/Limbo-compatible SQL | Embedded/lightweight direction, SQLite-compatible habits, sync roadmap |
| Notes/zettels | SQL tables | Queryable, syncable, auditable; no `.md` file store |
| Graph relationships | LadybugDB | Direct graph database with property graph/Cypher direction |
| Cross-store consistency | Operation log plus idempotent graph writes | Keeps offline sync and graph projection recoverable |
| Full-text content search | Search adapter over SQL content | Search capabilities differ by engine; isolate behind repository/service contracts |
| JSON metadata | Portable JSON stored in text columns, upgraded to JSON functions where supported | Avoid coupling core schema to optional extensions |
| Temporal model | Event log plus `valid_from`, `valid_to`, `observed_at`, `recorded_at` fields | Avoid a separate temporal DB; preserves auditability |
| Raw artifacts | Filesystem object store with DB references | Keeps large PDFs/media out of hot relational tables while extracted content goes to SQL |
| Vector search | Adapter seam; evaluate Turso vector support and alternatives after retrieval tests | Avoid locking into a vector engine before retrieval quality is measured |

## Consistency Model

Turso/Limbo owns content and operation history. LadybugDB owns graph traversal state. Services coordinate writes through repositories:

```text
write SQL content + operation log
  -> write graph node/edge facts directly to LadybugDB or enqueue graph update
  -> record graph projection status
  -> retry idempotently on failure
```

Graph nodes and edges must use stable IDs that reference SQL content records where applicable. If LadybugDB is unavailable, writes should still preserve enough operation data to repair or replay graph state.

## Consequences

- The backend uses a direct graph database instead of SQL edge tables as the primary graph mechanism.
- Mobile sync and audit remain grounded in operation logs and SQL content records.
- The project aligns with Turso/Limbo rather than vanilla SQLite as the preferred database path.
- SQLite-compatible fallback remains useful for development, tests, and platform gaps.
- Early production use needs conservative backup and restore procedures for both stores.
- Cross-store failure handling must be explicit from the beginning.

## References

- Turso/Limbo repository: https://github.com/tursodatabase/limbo
- Turso embedded replicas and sync direction: https://docs.turso.tech/features/embedded-replicas/introduction
- LadybugDB repository: https://github.com/LadybugDB/ladybug
- Kuzu archived repository notice: https://github.com/kuzudb/kuzu
