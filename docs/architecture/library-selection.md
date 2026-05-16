# Go Library Selection

## Selection Principles

- Prefer standard library packages unless a focused library clearly reduces risk.
- Keep business logic independent from transport, database, and provider SDKs.
- Prefer libraries with active maintenance, small API surfaces, and good context support.
- Keep generated code at the API and SQL boundaries, not in domain services.

## Selected Baseline

| Concern | Library | Status | Reasoning |
| --- | --- | --- | --- |
| HTTP server | `net/http` | Selected | Go 1.22+ routing is adequate for a small controller layer and avoids framework lock-in. |
| REST routing middleware | `github.com/go-chi/chi/v5` | Optional | Add when route groups, middleware chains, or URL params outgrow `net/http`. It remains idiomatic `http.Handler`. |
| Protobuf RPC | `connectrpc.com/connect` | Selected candidate | Simple protobuf APIs over HTTP, browser/mobile friendly, supports gRPC-compatible workflows without requiring gRPC-only infrastructure. |
| Protobuf generation | `buf.build` tooling | Selected candidate | Standardizes proto linting, breaking-change checks, and code generation. |
| MCP server | `github.com/modelcontextprotocol/go-sdk/mcp` | Selected | Official Go SDK, already scaffolded for Streamable HTTP at `/mcp`. |
| Primary SQL database | Turso/Limbo-compatible engine | Selected direction | Fits embedded/lightweight goal and sync direction while retaining SQLite-compatible schema habits. |
| Go SQL access | `database/sql` plus Turso/libSQL driver | Selected candidate | Keeps repository code idiomatic and driver-swappable. Exact driver should be validated against local embedded Limbo and remote Turso requirements. |
| SQL generation | `sqlc` | Selected candidate | Type-safe SQL without ORM abstractions; good fit for repository packages. |
| Migrations | `pressly/goose` | Selected candidate | Lightweight Go migrations with SQL files; works well for embedded/local development. |
| Graph database | LadybugDB Go binding | Selected candidate | Separate property graph database for direct graph storage and traversal; validate Go binding, concurrency, backup, and deployment before production data. |
| Logging | `log/slog` | Selected | Standard library structured logging. |
| Configuration | `os.Getenv` wrapper first | Selected | Current needs are small; add `caarlos0/env` or similar only when config complexity justifies it. |
| UUID/ULID | UUIDv7 or ULID package | Selected candidate | Needed for offline-first operation IDs and sortable event logs. Choose after mobile compatibility check. |
| Validation | Hand-written validators plus protobuf validation | Selected candidate | Keeps domain validation explicit; add `protovalidate-go` for protobuf contracts if contracts become complex. |
| Testing | Go `testing` plus `testify` optional | Selected | Start with standard library; add assertions/mocks only where they improve clarity. |

## Deliberately Avoided For Now

| Library/Class | Reason |
| --- | --- |
| Full-stack web framework | Controllers should remain thin and Go-native. |
| ORM | The data model needs explicit SQL, migrations, and sync/audit control. |
| Filesystem markdown store | Zettelkasten is conceptual only; durable content belongs in database tables. |
| Graph-only source of truth | Mobile sync, audit, and content provenance need SQL operation logs and stable records. |
| Heavy queue service | Start with database-backed jobs and background workers before adding external queue operations. |

## Open Validation Tasks

- Verify the best Go driver path for embedded Limbo versus remote Turso.
- Build a small LadybugDB proof of concept: create person/topic/zettel nodes, edges, and query traversal from Go.
- Confirm backup and restore procedures for both Turso/Limbo data and LadybugDB graph data.
- Confirm mobile local database options for Kotlin and iOS before locking sync protocol details.

## References

- LadybugDB: https://github.com/LadybugDB/ladybug
- Turso/Limbo: https://github.com/tursodatabase/limbo
- Connect for Go: https://connectrpc.com/docs/go/getting-started/
- chi: https://github.com/go-chi/chi
- sqlc: https://sqlc.dev/
- goose: https://github.com/pressly/goose
- MCP Go SDK: https://github.com/modelcontextprotocol/go-sdk
