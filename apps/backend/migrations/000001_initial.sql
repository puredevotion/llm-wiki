-- Initial Turso/Limbo-compatible SQL schema sketch. It is intentionally conservative:
-- the primary database is the source of truth, and graph/search/vector stores are
-- rebuildable projections.

PRAGMA foreign_keys = ON;

CREATE TABLE IF NOT EXISTS actors (
  id TEXT PRIMARY KEY,
  kind TEXT NOT NULL CHECK (kind IN ('person', 'agent', 'service')),
  display_name TEXT NOT NULL,
  metadata_json TEXT NOT NULL DEFAULT '{}',
  created_at TEXT NOT NULL
);

CREATE TABLE IF NOT EXISTS teams (
  id TEXT PRIMARY KEY,
  name TEXT NOT NULL,
  metadata_json TEXT NOT NULL DEFAULT '{}',
  created_at TEXT NOT NULL,
  updated_at TEXT NOT NULL
);

CREATE TABLE IF NOT EXISTS topics (
  id TEXT PRIMARY KEY,
  parent_id TEXT REFERENCES topics(id),
  name TEXT NOT NULL,
  description TEXT NOT NULL DEFAULT '',
  created_at TEXT NOT NULL,
  updated_at TEXT NOT NULL
);

CREATE TABLE IF NOT EXISTS sources (
  id TEXT PRIMARY KEY,
  kind TEXT NOT NULL CHECK (kind IN ('url', 'pdf', 'conversation', 'paste', 'file', 'agent_note')),
  uri TEXT,
  title TEXT NOT NULL DEFAULT '',
  content_hash TEXT NOT NULL,
  raw_object_ref TEXT,
  metadata_json TEXT NOT NULL DEFAULT '{}',
  captured_by TEXT REFERENCES actors(id),
  captured_at TEXT NOT NULL
);

CREATE TABLE IF NOT EXISTS chunks (
  id TEXT PRIMARY KEY,
  source_id TEXT NOT NULL REFERENCES sources(id),
  ordinal INTEGER NOT NULL,
  body TEXT NOT NULL,
  token_count INTEGER NOT NULL DEFAULT 0,
  metadata_json TEXT NOT NULL DEFAULT '{}',
  created_at TEXT NOT NULL,
  UNIQUE (source_id, ordinal)
);

CREATE TABLE IF NOT EXISTS zettels (
  id TEXT PRIMARY KEY,
  title TEXT NOT NULL,
  body TEXT NOT NULL,
  lifecycle TEXT NOT NULL CHECK (lifecycle IN ('ephemeral', 'project', 'evergreen')),
  status TEXT NOT NULL CHECK (status IN ('inbox', 'active', 'archived', 'superseded')),
  created_by TEXT REFERENCES actors(id),
  valid_from TEXT,
  valid_to TEXT,
  review_after TEXT,
  created_at TEXT NOT NULL,
  updated_at TEXT NOT NULL
);

CREATE TABLE IF NOT EXISTS citations (
  id TEXT PRIMARY KEY,
  zettel_id TEXT NOT NULL REFERENCES zettels(id),
  source_id TEXT NOT NULL REFERENCES sources(id),
  chunk_id TEXT REFERENCES chunks(id),
  quote TEXT NOT NULL DEFAULT '',
  created_at TEXT NOT NULL
);

CREATE TABLE IF NOT EXISTS events (
  id TEXT PRIMARY KEY,
  kind TEXT NOT NULL,
  title TEXT NOT NULL,
  body TEXT NOT NULL DEFAULT '',
  occurred_at TEXT,
  starts_at TEXT,
  ends_at TEXT,
  recorded_at TEXT NOT NULL,
  created_by TEXT REFERENCES actors(id),
  metadata_json TEXT NOT NULL DEFAULT '{}'
);

CREATE TABLE IF NOT EXISTS operations (
  id TEXT PRIMARY KEY,
  actor_id TEXT REFERENCES actors(id),
  device_id TEXT,
  entity_kind TEXT NOT NULL,
  entity_id TEXT NOT NULL,
  operation_type TEXT NOT NULL,
  payload_json TEXT NOT NULL,
  base_version INTEGER,
  status TEXT NOT NULL CHECK (status IN ('pending', 'applied', 'rejected')),
  created_at TEXT NOT NULL,
  applied_at TEXT
);

CREATE TABLE IF NOT EXISTS graph_events (
  id TEXT PRIMARY KEY,
  operation_id TEXT REFERENCES operations(id),
  event_type TEXT NOT NULL CHECK (event_type IN ('upsert_node', 'delete_node', 'upsert_edge', 'delete_edge')),
  graph_entity_id TEXT NOT NULL,
  payload_json TEXT NOT NULL,
  status TEXT NOT NULL CHECK (status IN ('pending', 'applied', 'failed')),
  attempts INTEGER NOT NULL DEFAULT 0,
  last_error TEXT,
  created_at TEXT NOT NULL,
  applied_at TEXT
);

CREATE VIRTUAL TABLE IF NOT EXISTS zettels_fts USING fts5(
  title,
  body,
  content='zettels',
  content_rowid='rowid'
);

-- Triggers to keep zettels_fts in sync
CREATE TRIGGER IF NOT EXISTS zettels_ai AFTER INSERT ON zettels BEGIN
  INSERT INTO zettels_fts(rowid, title, body) VALUES (new.rowid, new.title, new.body);
END;
CREATE TRIGGER IF NOT EXISTS zettels_ad AFTER DELETE ON zettels BEGIN
  INSERT INTO zettels_fts(zettels_fts, rowid, title, body) VALUES('delete', old.rowid, old.title, old.body);
END;
CREATE TRIGGER IF NOT EXISTS zettels_au AFTER UPDATE ON zettels BEGIN
  INSERT INTO zettels_fts(zettels_fts, rowid, title, body) VALUES('delete', old.rowid, old.title, old.body);
  INSERT INTO zettels_fts(rowid, title, body) VALUES (new.rowid, new.title, new.body);
END;

CREATE INDEX IF NOT EXISTS idx_events_time ON events(occurred_at, starts_at, ends_at);
CREATE INDEX IF NOT EXISTS idx_operations_actor_device ON operations(actor_id, device_id, created_at);
CREATE INDEX IF NOT EXISTS idx_graph_events_status ON graph_events(status, created_at);
