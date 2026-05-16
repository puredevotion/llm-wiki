# Sync Strategy

## Baseline

The server is the authority. Mobile apps keep local caches and an outbox of user operations. The backend exposes sync endpoints that accept idempotent operations and return ordered changes since a cursor.

The preferred storage direction is Turso/Limbo-compatible local databases because that aligns with embedded use, SQLite-compatible semantics, and Turso's sync direction. Where a platform binding is not ready, clients may use native SQLite locally behind the same repository interface.

## Why Not Full Multi-Master Initially

A virtual chief of staff needs correct provenance, auditability, and conflict handling more than low-latency peer-to-peer writes. A server-authoritative model keeps the first version understandable while still supporting offline mobile capture.

## Operation Log

Every write creates an operation record:

- `operation_id`: client-generated ULID/UUIDv7 for idempotency
- `actor_id`
- `device_id`
- `entity_type`
- `entity_id`
- `operation_type`
- `payload_json`
- `base_version`
- `created_at`
- `applied_at`
- `status`

## Mobile Buffering

Mobile clients should support:

- Local draft capture for text, links, photos, PDFs, and notes.
- Outbox retry with exponential backoff.
- Conflict UI for edits to the same durable zettel or relationship.
- Read cache for recent topics, people, notes, and tasks.

## Server Sync Endpoints

Initial HTTP API shape:

```text
POST /api/sync/push       Submit client operations
GET  /api/sync/pull       Pull changes since cursor
GET  /api/sync/bootstrap  Get initial compact replica snapshot
```

## Turso Sync

Turso's current documentation distinguishes legacy embedded replicas from newer Turso Sync for explicit local-first push/pull. That maps well to the product direction, but the first backend should keep explicit operation records because they are easier to audit, authorize, and debug across Go, Kotlin, Swift, and web clients.

References:

- https://docs.turso.tech/features/embedded-replicas/introduction
- https://github.com/tursodatabase/limbo
