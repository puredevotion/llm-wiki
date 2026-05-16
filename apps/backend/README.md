# Backend

Go backend for the knowledge base and virtual chief of staff.

## Responsibilities

- Own the authoritative Turso/Limbo-compatible content database and migrations.
- Own direct graph persistence through LadybugDB.
- Expose thin REST and protobuf controllers for web/mobile sync and administration.
- Expose MCP Streamable HTTP at `/mcp` for agents.
- Run ingestion, summarization, indexing, review, and re-evaluation workers.
- Maintain search/vector projections.

## Development Commands

```bash
go test ./...
go run ./cmd/kbase-server
```

The scaffold pins the official MCP Go SDK in `go.mod`.

## Package Boundaries

See [../../docs/architecture/go-backend.md](../../docs/architecture/go-backend.md) and [../../docs/architecture/adr/0003-backend-module-boundaries.md](../../docs/architecture/adr/0003-backend-module-boundaries.md).
