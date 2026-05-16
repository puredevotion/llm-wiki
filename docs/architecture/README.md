# Architecture Overview

## Goal

Build an operationally light knowledge base that can act as a virtual chief of staff: remember context, track people and teams, organize topics, reason over time, and expose safe tools to AI agents through MCP.

## System Shape

```text
Agents (Claude/ChatGPT/Copilot)
        |
        | MCP Streamable HTTP
        v
Go backend --------------------------------------------------+
  HTTP API                                                   |
  MCP tools/resources/prompts                                |
  background workers                                         |
  sync engine                                                |
        |                                                    |
        +--> Turso/Limbo-compatible SQL store                |
        +--> graph projection via typed edge tables           |
        +--> search/vector index adapter                     |
        +--> object/file store for original artifacts        |
                                                             |
Web / Android / iOS clients <------ sync API ----------------+
```

## Design Constraints

- Keep operations light: prefer embedded libraries and single-binary deployment over separate services.
- Make the Turso/Limbo-compatible store authoritative first; indexes can be rebuilt from primary tables and event logs.
- Keep graph semantics in the data model, but do not introduce a separate graph database until edge-table traversal is proven insufficient.
- Keep all agent-facing actions behind typed MCP tools with validation, authorization, and audit records.
- Treat mobile clients as occasionally connected replicas with local buffers, not as distributed masters.
- Preserve provenance for every piece of content: source, timestamp, actor, transform chain, and confidence.

## Planes

| Plane | Purpose | Primary Model |
| --- | --- | --- |
| Who | People, teams, accounts, organizations, roles | Identity records plus typed relationships |
| What | Topics, zettels, source documents, summaries, claims | Zettelkasten notes and topic hierarchy |
| When | Events, validity windows, reminders, review cadence | Append-only event log plus temporal fields |
| Why | Provenance, decisions, confidence, re-evaluation triggers | Metadata and audit trail |

## Backend Responsibilities

- Ingest content from URLs, PDFs, pasted text, conversations, and future connectors.
- Normalize content into sources, chunks, notes, claims, topics, and relationships.
- Maintain search indexes and graph projections from the authoritative store.
- Expose MCP tools for search, capture, linking, summarization requests, and review queues.
- Expose mobile/web sync APIs with conflict detection and durable operation logs.
- Run scheduled and event-driven re-evaluation jobs.

## Early Non-Goals

- Multi-node database clustering.
- Separate graph database operation.
- Real-time peer-to-peer mobile sync.
- Letting agents write directly to database tables.
- Premature adoption of a heavy graph/search server.
