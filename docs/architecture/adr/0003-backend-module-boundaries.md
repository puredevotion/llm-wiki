# ADR 0003: Backend Module Boundaries

## Status

Accepted for initial scaffold.

## Decision

Organize the Go backend around product capabilities and isolate infrastructure.

```text
cmd/kbase-server/       Process entrypoint
internal/config/        Runtime configuration
internal/httpapi/       REST/JSON API and middleware
internal/mcp/           MCP tool/resource/prompt registration
internal/identity/      People, teams, organizations, roles
internal/topics/        Topic hierarchy and classification
internal/knowledge/     Sources, chunks, zettels, claims, links
internal/timeline/      Events, deadlines, review windows, temporal facts
internal/ingestion/     Importers, extractors, chunkers, summarizers
internal/sync/          Mobile sync operations and cursors
internal/storage/       Turso, graph, and search adapters
migrations/             Turso/Limbo-compatible SQL migrations
```

## Rules

- Domain packages define service interfaces and business rules.
- `internal/storage/*` implements persistence details.
- `internal/mcp` and `internal/httpapi` call services, not databases.
- Background workers enqueue domain jobs and persist through services.
- Agent/client-specific behavior belongs at the edge, not inside storage adapters.
