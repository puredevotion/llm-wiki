# Domain Model

## Core Concepts

### Source

A source is an imported artifact: URL article, Hacker News thread, PDF, conversation transcript, copied text, uploaded file, or agent-generated note. Sources are immutable after capture except for metadata corrections.

Key fields:

- `id`
- `kind`: `url`, `pdf`, `conversation`, `paste`, `file`, `agent_note`
- `uri`
- `title`
- `author_ref`
- `captured_at`
- `content_hash`
- `raw_object_ref`
- `metadata_json`

### Chunk

A chunk is a bounded extract from a source that can be indexed, embedded, summarized, and cited. Chunking is deterministic per source version.

### Zettel

A zettel is a durable atomic note. It may be short-lived, long-running, or evergreen. Zettels are not raw captures; they are interpreted knowledge units linked to sources and other notes.

Key fields:

- `id`
- `title`
- `body`
- `lifecycle`: `ephemeral`, `project`, `evergreen`
- `status`: `inbox`, `active`, `archived`, `superseded`
- `created_by`
- `created_at`
- `updated_at`
- `valid_from`
- `valid_to`
- `review_after`

### Topic

A topic is a nested knowledge area. Topic hierarchy should be explicit but not exclusive: a note can belong to multiple topics.

### Person / Team

People and teams represent the who-plane. They are graph-friendly entities because relationships matter: reports-to, member-of, collaborated-with, owns-topic, mentioned-by, authored.

### Event

An event captures time-sensitive context: meeting, deadline, decision, reminder, source capture, agent action, topic review, or relationship change.

## Relationship Types

Use typed edges to connect entities:

- `MENTIONS`
- `AUTHORED_BY`
- `MEMBER_OF`
- `OWNS_TOPIC`
- `RELATED_TO`
- `SUPPORTS`
- `CONTRADICTS`
- `SUPERSEDES`
- `DERIVED_FROM`
- `SCHEDULED_FOR`
- `VALID_DURING`

The initial implementation stores edges in the primary Turso/Limbo-compatible database. Relationship queries should use indexed edge tables and recursive CTEs before introducing any graph-specific engine.

## Lifecycle Buckets

| Bucket | Meaning | Review Behavior |
| --- | --- | --- |
| Ephemeral | Useful for hours/days, such as active conversation context | Auto-expire or summarize into project notes |
| Project | Useful while a workstream is active | Review on cadence or project milestones |
| Evergreen | Durable knowledge, principles, people context, decisions | Periodic validation and contradiction checks |

## Ingestion Flow

```text
Capture source
  -> store raw artifact
  -> extract text and metadata
  -> chunk deterministically
  -> classify people/topics/time references
  -> create inbox zettels and candidate edges
  -> queue review/re-evaluation jobs
```
