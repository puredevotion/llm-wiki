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

A zettel is a durable atomic note inspired by zettelkasten practice. It may be short-lived, long-running, or evergreen. Zettels are not raw captures and are not markdown files; they are database records containing interpreted knowledge linked to sources and graph relationships.

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

People and teams represent the who-plane. They are graph-native entities because relationships matter: reports-to, member-of, collaborated-with, owns-topic, mentioned-by, authored.

### Event

An event captures time-sensitive context: meeting, deadline, decision, reminder, source capture, agent action, topic review, or relationship change.

## Relationship Types

Use typed LadybugDB edges to connect graph nodes:

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

Graph nodes should use stable IDs that can reference SQL records for sources, chunks, zettels, topics, people, teams, and events.

## Lifecycle Buckets

| Bucket | Meaning | Review Behavior |
| --- | --- | --- |
| Ephemeral | Useful for hours/days, such as active conversation context | Auto-expire or summarize into project notes |
| Project | Useful while a workstream is active | Review on cadence or project milestones |
| Evergreen | Durable knowledge, principles, people context, decisions | Periodic validation and contradiction checks |

## Ingestion Flow

```text
Capture source
  -> store raw artifact if needed
  -> extract text and metadata into SQL tables
  -> chunk deterministically
  -> classify people/topics/time references
  -> create inbox zettels and candidate graph relationships
  -> queue review/re-evaluation jobs
```
