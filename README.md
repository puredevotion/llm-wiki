# LLM Wiki

A persistent knowledge base and virtual chief of staff for people, teams, topics, time, and long-running context.

The first implementation focus is a Go backend that owns durable knowledge storage and exposes a Model Context Protocol (MCP) interface for agents such as Claude, ChatGPT, and Copilot. Web, Android, and iOS clients are planned as sync-capable frontends rather than independent sources of truth.

## Architecture Direction

- Backend: Go, single server binary, HTTP API plus MCP Streamable HTTP endpoint.
- Persistence: Turso/Limbo-compatible SQL as the preferred embedded store, with portable SQLite semantics where needed.
- Graph: typed edge tables in the primary store first; no separate graph database dependency in v1.
- MCP: Agents connect to the backend through `/mcp`; tools call service-layer APIs rather than storage directly.
- Sync: Mobile clients keep a local outbox/cache and reconcile with the server via append-only operations.
- Knowledge model: Zettelkasten-inspired notes, sources, topics, persons, teams, events, and relationships.

## Repository Layout

```text
apps/backend/        Go backend, MCP server, migrations, storage adapters
apps/web/            Future Vue/Nuxt application
apps/android/        Future Kotlin Android application
apps/ios/            Future iOS application
agent-clients/       Agent-facing integration notes and MCP client configs
docs/architecture/   Architecture, ADRs, domain model, sync strategy
tools/scripts/       Local operational scripts once needed
```

Start with [docs/architecture/README.md](docs/architecture/README.md).
