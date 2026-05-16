# Sync Strategy

## Baseline

Mobile clients are offline-first. They must allow capture, draft editing, recent knowledge browsing, and queued actions without network access. The backend remains the reconciliation authority, but mobile writes are first persisted locally and then synchronized.

The preferred local storage direction is Turso/Limbo-compatible local databases because that aligns with embedded use, SQLite-compatible semantics, and Turso's sync direction. Where a platform binding is not ready, clients may use native SQLite locally behind the same repository interface.

## Offline-First Behavior

Clients should support:

- Local durable outbox for every user action.
- Local read cache for recent topics, people, zettels, sources, tasks, and review items.
- Local draft capture for text, links, photos, PDFs, and notes.
- Conflict UI for edits to the same durable zettel or relationship.
- Background retry with exponential backoff.
- Clear sync state: pending, synced, conflict, failed.

## Why Not Peer-To-Peer Initially

A virtual chief of staff needs correct provenance, auditability, and conflict handling more than direct device-to-device sync. Offline-first client behavior gives users local continuity while server reconciliation keeps the system understandable.

## Operation Log

Every write creates a local operation first and then a server operation after sync:

- `operation_id`: client-generated ULID/UUIDv7 for idempotency
- `actor_id`
- `device_id`
- `entity_type`
- `entity_id`
- `operation_type`
- `payload_json`
- `base_version`
- `created_at`
- `synced_at`
- `applied_at`
- `status`

## Server Sync Endpoints

Initial API shape should be available over REST and protobuf/connect:

```text
POST /api/v1/sync/operations  Submit client operation batches
GET  /api/v1/sync/changes     Pull ordered changes with cursor and limit
GET  /api/v1/sync/bootstrap   Get initial compact replica snapshot
```

The protobuf API should model the same commands and results as REST so mobile clients can use the more efficient contract without changing backend services.

## Graph Sync

Mobile clients should not embed the full graph database in v1. They sync the subset of graph facts required for local UI and capture. The server writes full graph state to LadybugDB and can return graph-derived views through sync/read APIs.

## Turso Sync

Turso's current documentation distinguishes legacy embedded replicas from newer Turso Sync for explicit local-first push/pull. That maps well to the product direction, but the first backend should keep explicit operation records because they are easier to audit, authorize, and debug across Go, Kotlin, Swift, and web clients.

References:

- https://docs.turso.tech/features/embedded-replicas/introduction
- https://github.com/tursodatabase/limbo
