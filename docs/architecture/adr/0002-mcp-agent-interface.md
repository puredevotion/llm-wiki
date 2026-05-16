# ADR 0002: MCP Agent Interface

## Status

Accepted for initial scaffold.

## Context

Claude, ChatGPT, Copilot, and future agents need controlled access to the knowledge base. Direct database access is unsafe and bypasses validation, authorization, audit logging, and provenance.

## Decision

Expose agent capabilities through the backend's MCP server at `/mcp` using Streamable HTTP.

The MCP layer will register tools, resources, and prompts that call backend services. It must not call storage adapters directly.

## Initial MCP Surface

Tools:

- `kb.search`: search notes, sources, topics, people, and events.
- `kb.capture`: submit pasted text, URL references, or conversation excerpts to the ingestion inbox.
- `kb.link`: propose or create typed relationships between entities.
- `kb.review_queue`: list items requiring human or agent review.
- `kb.summarize_thread`: request summarization of conversation/source chunks.

Resources:

- `kb://topics/{id}`
- `kb://zettels/{id}`
- `kb://people/{id}`
- `kb://sources/{id}`

Prompts:

- `chief_of_staff_briefing`
- `topic_reassessment`
- `source_to_zettels`

## Security Rules

- Every MCP request is authenticated when exposed remotely.
- Every tool input is schema-validated.
- Every write records actor, client, request id, source/provenance, and before/after metadata.
- MCP tools return structured, citation-friendly results and never raw stack traces.
- High-impact writes can be staged as proposals until trust and policy are mature.

## References

- MCP Go SDK overview: https://go.sdk.modelcontextprotocol.io/
- MCP SDK list: https://modelcontextprotocol.io/docs/sdk
