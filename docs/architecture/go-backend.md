# Go Backend Architecture

## Goal

Use Go idioms while keeping an MVC-like separation:

```text
controllers -> services -> repositories -> persistence
                       -> clients -> external services
```

Controllers are thin transport adapters. Services own business rules. Repositories own persistence communication. Clients own communication with online services and third-party APIs.

## Package Layout

```text
cmd/kbase-server/                 main package and process wiring
internal/config/                  environment and runtime configuration
internal/controllers/rest/        thin REST/JSON controllers
internal/controllers/protobuf/    thin protobuf/connect controllers
internal/controllers/mcp/         MCP agent endpoint and tool registration
internal/contracts/proto/         protobuf contracts and generated API boundary
internal/domain/                  domain entities, value objects, domain errors
internal/services/                application/business services
internal/repositories/            repository interfaces and composition
internal/storage/turso/           Turso/Limbo SQL implementations
internal/storage/graph/           LadybugDB graph implementations
internal/storage/search/          search/vector projection implementations
internal/clients/                 outbound clients for fetchers, LLMs, embeddings, auth, notifications
internal/ingestion/               import/extract/chunk/classify workflows
internal/sync/                    offline-first sync service and operation reconciliation
migrations/                       Turso/Limbo-compatible schema migrations
```

## MVC-Like Flow

```text
REST/protobuf/MCP controller
  -> validate transport request
  -> map request DTO to service command/query
  -> call service
  -> map service result to transport response
```

```text
service
  -> enforce business rules and permissions
  -> coordinate repositories and clients
  -> create domain events and operation-log entries
  -> return domain result or typed domain error
```

```text
repository
  -> execute persistence-specific reads/writes
  -> translate storage errors to repository errors
  -> never call controllers or external clients
```

## Controller Rules

- Keep controllers thin: no business rules, no SQL, no graph queries.
- Accept `context.Context` and pass it through every service/repository/client call.
- Decode and validate request shape at the edge.
- Use explicit response structs, not raw maps, except for temporary health checks.
- Convert domain errors to protocol-specific status codes in one place.
- REST and protobuf controllers should share service methods, not duplicate business logic.
- Version REST application endpoints under `/api/v1`; keep operational probes such as `/healthz` unversioned.

## Service Rules

- Services are the application boundary and should be easy to unit test with fake repositories/clients.
- Services own transactions and cross-store consistency policy.
- Services write durable content to Turso/Limbo and graph facts to LadybugDB through repositories.
- Services create operation records for offline sync and audit before or within the same unit of work as state changes.
- Do not put HTTP, protobuf, MCP, SQL driver, or Ladybug client types in service method signatures.

## Repository Rules

- Define small interfaces near the service that consumes them when possible.
- Implement repositories in storage-specific packages.
- Prefer typed methods over generic `Save(any)` APIs.
- Repositories should return domain types, not database row structs, unless the row is explicitly the domain model.
- Use generated SQL where it improves safety; hand-written SQL is acceptable for simple queries.

## Client Rules

Clients are outbound adapters for online services:

- URL fetchers and feed readers.
- PDF/extraction services when not local.
- LLM providers and embedding providers.
- Auth, notification, email, calendar, and future connector APIs.

Client packages should expose narrow interfaces and isolate provider SDKs from services.

## Content Storage

Zettelkasten is an inspiration for the knowledge model, not a filesystem format. Do not store notes as `.md` files.

All durable knowledge content belongs in the Turso/Limbo-compatible database:

- sources
- chunks
- zettels
- citations
- summaries
- claims
- lifecycle and review metadata

Raw uploaded artifacts such as PDFs may live in object/file storage, but extracted text and durable interpreted content must be persisted in database tables.

## Cross-Store Write Pattern

Because the architecture uses a separate graph database, services should use an explicit write pattern:

```text
1. Write content/operation/audit records to Turso/Limbo.
2. Write or enqueue graph facts for LadybugDB.
3. Record projection status and retry failures through a background worker.
4. Reads that require graph traversal query LadybugDB first and degrade to limited SQL metadata when unavailable.
```

Graph writes must be idempotent. Every graph node/edge should reference stable IDs from the content database.
