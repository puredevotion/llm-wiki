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
go test ./internal/... -cover
go run ./cmd/kbase-server
```

The scaffold pins the official MCP Go SDK in `go.mod`.

Code changes must be developed red-green with tests written before implementation. Active backend packages should maintain 90%+ coverage; process-only wiring such as `cmd/kbase-server/main.go` may be excluded from the active package coverage target until it contains testable logic.

## Package Boundaries

See [../../docs/architecture/go-backend.md](../../docs/architecture/go-backend.md) and [../../docs/architecture/adr/0003-backend-module-boundaries.md](../../docs/architecture/adr/0003-backend-module-boundaries.md).
