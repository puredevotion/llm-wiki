# Turso Storage

Authoritative store for durable content, sources, chunks, zettels, citations, events, operation logs, sync cursors, and audit records.

The preferred engine direction is Turso/Limbo. Keep schema and queries SQLite-compatible where practical so development, tests, and platform fallbacks remain simple.

Do not store zettelkasten content as `.md` files. Store durable content in database tables.
