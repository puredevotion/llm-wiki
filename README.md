# LLM Wiki

A persistent knowledge base and virtual chief of staff for people, teams, topics, time, and long-running context.

The first implementation focus is a Go backend that owns durable knowledge storage and exposes a Model Context Protocol (MCP) interface for agents such as Claude, ChatGPT, and Copilot. Web, Android, and iOS clients are planned as offline-first frontends with local buffering and sync.

## Architecture Direction

- Backend: Go, single server binary, thin REST/protobuf controllers, service layer, repositories, outbound clients, and MCP Streamable HTTP endpoint.
- Content persistence: Turso/Limbo-compatible SQL as the preferred embedded store, with portable SQLite semantics where needed.
- Graph persistence: separate direct LadybugDB graph database for people, teams, topics, events, and relationships.
- MCP: Agents connect to the backend through `/mcp`; tools call service-layer APIs rather than storage directly.
- Sync: Mobile clients operate offline-first with local outboxes/caches and reconcile with the server via append-only operations.
- Knowledge model: Zettelkasten-inspired notes, sources, topics, persons, teams, events, and relationships stored in databases, not markdown files.

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
