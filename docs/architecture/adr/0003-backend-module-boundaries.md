# ADR 0003: Backend Module Boundaries

## Status

Accepted for initial scaffold.

## Decision

Organize the Go backend around MVC-like boundaries while preserving Go idioms.

```text
cmd/kbase-server/              Process entrypoint
internal/config/               Runtime configuration
internal/controllers/rest/     Thin REST/JSON controllers
internal/controllers/protobuf/ Thin protobuf/connect controllers
internal/controllers/mcp/      MCP tool/resource/prompt registration
internal/contracts/proto/      Protobuf contracts and generated code boundary
internal/domain/               Domain entities, value objects, domain errors
internal/services/             Business logic and application orchestration
internal/repositories/         Repository interfaces and composition
internal/clients/              Outbound clients for online services
internal/identity/             People, teams, organizations, roles
internal/topics/               Topic hierarchy and classification
internal/knowledge/            Sources, chunks, zettels, claims, links
internal/timeline/             Events, deadlines, review windows, temporal facts
internal/ingestion/            Importers, extractors, chunkers, summarizers
internal/sync/                 Offline-first sync operations and cursors
internal/storage/turso/        Turso/Limbo-compatible SQL implementations
internal/storage/graph/        LadybugDB graph implementations
internal/storage/search/       Search/vector implementations
migrations/                    Turso/Limbo-compatible SQL migrations
```

## Rules

- Controllers decode, validate transport shape, call services, and map errors to protocol responses.
- Services own business rules, permissions, transactions, cross-store write coordination, and domain events.
- Repositories communicate with persistence layers and hide SQL/Ladybug implementation details.
- Clients communicate with online services and hide external SDK details.
- Domain and service packages must not import controller, SQL driver, graph driver, or external provider SDK packages.
- MCP, REST, and protobuf endpoints share services instead of duplicating business logic.
