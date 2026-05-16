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
  REST controllers                                           |
  protobuf/connect controllers                               |
  MCP controllers/tools/resources/prompts                    |
  services                                                   |
  repositories + clients                                     |
  background workers                                         |
  offline-first sync engine                                  |
        |                                                    |
        +--> Turso/Limbo-compatible SQL store                |
        +--> LadybugDB graph database                        |
        +--> search/vector adapter                           |
        +--> object/file store for original artifacts        |
                                                             |
Web / Android / iOS clients <------ offline-first sync ------+
```

## Design Constraints

- Keep operations light: prefer embedded libraries and single-binary deployment over heavy distributed systems.
- Store durable content in Turso/Limbo-compatible database tables, not markdown files.
- Use LadybugDB as a separate direct graph database for relationship storage and traversal.
- Keep all agent-facing actions behind typed MCP tools with validation, authorization, and audit records.
- Treat mobile clients as offline-first replicas with local writes, durable outboxes, and sync reconciliation.
- Preserve provenance for every piece of content: source, timestamp, actor, transform chain, and confidence.

## Planes

| Plane | Purpose | Primary Model |
| --- | --- | --- |
| Who | People, teams, accounts, organizations, roles | LadybugDB graph nodes/edges plus SQL metadata |
| What | Topics, zettels, source documents, summaries, claims | Database-stored zettelkasten-inspired notes and topic records |
| When | Events, validity windows, reminders, review cadence | Append-only event log plus temporal fields |
| Why | Provenance, decisions, confidence, re-evaluation triggers | Metadata and audit trail |

## Backend Responsibilities

- Ingest content from URLs, PDFs, pasted text, conversations, and future connectors.
- Normalize content into sources, chunks, notes, claims, topics, and graph relationships.
- Maintain search indexes and graph state from service-level writes and repair jobs.
- Expose MCP tools for search, capture, linking, summarization requests, and review queues.
- Expose REST/protobuf sync APIs with conflict detection and durable operation logs.
- Run scheduled and event-driven re-evaluation jobs.

## Early Non-Goals

- Multi-node database clustering.
- Filesystem markdown knowledge store.
- Real-time peer-to-peer mobile sync.
- Letting agents write directly to database tables.
- Premature adoption of heavy queue/search infrastructure.

## More Detail

- [Go backend architecture](go-backend.md)
- [Go library selection](library-selection.md)
- [Domain model](domain-model.md)
- [Sync strategy](sync.md)
