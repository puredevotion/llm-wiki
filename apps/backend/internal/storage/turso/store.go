package turso

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	_ "github.com/mattn/go-sqlite3"
	_ "github.com/tursodatabase/libsql-client-go/libsql"
	"llm-wiki/apps/backend/internal/domain"
	"llm-wiki/apps/backend/internal/repositories"
)

type Store struct {
	db *sql.DB
}

func NewStore(dsn string) (*Store, error) {
	driver := "libsql"
	if strings.HasPrefix(dsn, "file:") {
		driver = "sqlite3"
		dsn = strings.TrimPrefix(dsn, "file:")
	}
	db, err := sql.Open(driver, dsn)
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}
	if err := db.Ping(); err != nil {
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}
	return &Store{db: db}, nil
}

func (s *Store) Close() error {
	return s.db.Close()
}

func (s *Store) Migrate(ctx context.Context, schema string) error {
	_, err := s.db.ExecContext(ctx, schema)
	return err
}

func (s *Store) Actors() repositories.ActorRepository   { return &actorRepo{s} }
func (s *Store) Sources() repositories.SourceRepository { return &sourceRepo{s} }
func (s *Store) Zettels() repositories.ZettelRepository { return &zettelRepo{s} }
func (s *Store) Topics() repositories.TopicRepository   { return &topicRepo{s} }
func (s *Store) Operations() repositories.OperationRepository { return &operationRepo{s} }

type actorRepo struct{ *Store }

func (r *actorRepo) FindByName(ctx context.Context, name string) (*domain.Actor, error) {
	row := r.db.QueryRowContext(ctx, "SELECT id, kind, display_name, metadata_json, created_at FROM actors WHERE display_name = ?", name)
	var a domain.Actor
	var metadataJSON string
	var createdAt string
	err := row.Scan(&a.ID, &a.Kind, &a.DisplayName, &metadataJSON, &createdAt)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("failed to scan actor: %w", err)
	}
	if err := json.Unmarshal([]byte(metadataJSON), &a.Metadata); err != nil {
		return nil, fmt.Errorf("failed to unmarshal actor metadata: %w", err)
	}
	a.CreatedAt, _ = time.Parse(time.RFC3339, createdAt)
	return &a, nil
}

func (r *actorRepo) Save(ctx context.Context, a *domain.Actor) error {
	metadata, err := json.Marshal(a.Metadata)
	if err != nil {
		return fmt.Errorf("failed to marshal actor metadata: %w", err)
	}
	_, err = r.db.ExecContext(ctx,
		"INSERT INTO actors (id, kind, display_name, metadata_json, created_at) VALUES (?, ?, ?, ?, ?) ON CONFLICT(id) DO UPDATE SET display_name=excluded.display_name, metadata_json=excluded.metadata_json",
		a.ID, a.Kind, a.DisplayName, string(metadata), a.CreatedAt.Format(time.RFC3339),
	)
	return err
}

type sourceRepo struct{ *Store }

func (r *sourceRepo) Save(ctx context.Context, src *domain.Source) error {
	metadata, err := json.Marshal(src.Metadata)
	if err != nil {
		return fmt.Errorf("failed to marshal source metadata: %w", err)
	}
	_, err = r.db.ExecContext(ctx,
		"INSERT INTO sources (id, kind, uri, title, content_hash, raw_object_ref, metadata_json, captured_by, captured_at) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?) ON CONFLICT(id) DO UPDATE SET title=excluded.title, metadata_json=excluded.metadata_json",
		src.ID, src.Kind, src.URI, src.Title, src.ContentHash, src.RawObjectRef, string(metadata), src.CapturedBy, src.CapturedAt.Format(time.RFC3339),
	)
	return err
}

type zettelRepo struct{ *Store }

func (r *zettelRepo) Save(ctx context.Context, z *domain.Zettel) error {
	var vFrom, vTo, rAfter *string
	if z.ValidFrom != nil {
		s := z.ValidFrom.Format(time.RFC3339)
		vFrom = &s
	}
	if z.ValidTo != nil {
		s := z.ValidTo.Format(time.RFC3339)
		vTo = &s
	}
	if z.ReviewAfter != nil {
		s := z.ReviewAfter.Format(time.RFC3339)
		rAfter = &s
	}

	_, err := r.db.ExecContext(ctx,
		"INSERT INTO zettels (id, title, body, lifecycle, status, created_by, valid_from, valid_to, review_after, created_at, updated_at) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?) ON CONFLICT(id) DO UPDATE SET title=excluded.title, body=excluded.body, status=excluded.status, updated_at=excluded.updated_at",
		z.ID, z.Title, z.Body, z.Lifecycle, z.Status, z.CreatedBy, vFrom, vTo, rAfter, z.CreatedAt.Format(time.RFC3339), z.UpdatedAt.Format(time.RFC3339),
	)
	return err
}

func (r *zettelRepo) SearchZettels(ctx context.Context, query string, limit int) ([]*domain.Zettel, error) {
	rows, err := r.db.QueryContext(ctx,
		"SELECT z.id, z.title, z.body, z.lifecycle, z.status, z.created_by, z.created_at, z.updated_at FROM zettels z JOIN zettels_fts f ON z.rowid = f.rowid WHERE zettels_fts MATCH ? LIMIT ?",
		query, limit,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to search zettels: %w", err)
	}
	defer rows.Close()

	var results []*domain.Zettel
	for rows.Next() {
		var z domain.Zettel
		var createdAt, updatedAt string
		err := rows.Scan(&z.ID, &z.Title, &z.Body, &z.Lifecycle, &z.Status, &z.CreatedBy, &createdAt, &updatedAt)
		if err != nil {
			return nil, fmt.Errorf("failed to scan search result: %w", err)
		}
		z.CreatedAt, _ = time.Parse(time.RFC3339, createdAt)
		z.UpdatedAt, _ = time.Parse(time.RFC3339, updatedAt)
		results = append(results, &z)
	}
	return results, nil
}

type topicRepo struct{ *Store }

func (r *topicRepo) FindByName(ctx context.Context, name string) (*domain.Topic, error) {
	row := r.db.QueryRowContext(ctx, "SELECT id, parent_id, name, created_at FROM topics WHERE name = ?", name)
	var t domain.Topic
	var parentID sql.NullString
	var createdAt string
	err := row.Scan(&t.ID, &parentID, &t.Name, &createdAt)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("failed to scan topic: %w", err)
	}
	if parentID.Valid {
		t.ParentID = parentID.String
	}
	t.CreatedAt, _ = time.Parse(time.RFC3339, createdAt)
	return &t, nil
}

func (r *topicRepo) Save(ctx context.Context, t *domain.Topic) error {
	var parentID sql.NullString
	if t.ParentID != "" {
		parentID.String = t.ParentID
		parentID.Valid = true
	}
	_, err := r.db.ExecContext(ctx,
		"INSERT INTO topics (id, parent_id, name, created_at, updated_at) VALUES (?, ?, ?, ?, ?) ON CONFLICT(id) DO UPDATE SET name=excluded.name, parent_id=excluded.parent_id, updated_at=excluded.updated_at",
		t.ID, parentID, t.Name, t.CreatedAt.Format(time.RFC3339), time.Now().Format(time.RFC3339),
	)
	return err
}

type operationRepo struct{ *Store }

func (r *operationRepo) Save(ctx context.Context, op *domain.Operation) error {
	var appliedAt *string
	if op.AppliedAt != nil {
		s := op.AppliedAt.Format(time.RFC3339)
		appliedAt = &s
	}
	_, err := r.db.ExecContext(ctx,
		"INSERT INTO operations (id, actor_id, device_id, entity_kind, entity_id, operation_type, payload_json, base_version, status, created_at, applied_at) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?) ON CONFLICT(id) DO UPDATE SET status=excluded.status, applied_at=excluded.applied_at",
		op.ID, op.ActorID, op.DeviceID, op.EntityKind, op.EntityID, op.OperationType, string(op.Payload), op.BaseVersion, string(op.Status), op.CreatedAt.Format(time.RFC3339), appliedAt,
	)
	return err
}

func (r *operationRepo) FindByID(ctx context.Context, id string) (*domain.Operation, error) {
	row := r.db.QueryRowContext(ctx, "SELECT id, actor_id, device_id, entity_kind, entity_id, operation_type, payload_json, base_version, status, created_at, applied_at FROM operations WHERE id = ?", id)
	var op domain.Operation
	var payload string
	var createdAt string
	var appliedAt sql.NullString
	err := row.Scan(&op.ID, &op.ActorID, &op.DeviceID, &op.EntityKind, &op.EntityID, &op.OperationType, &payload, &op.BaseVersion, &op.Status, &createdAt, &appliedAt)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	op.Payload = json.RawMessage(payload)
	op.CreatedAt, _ = time.Parse(time.RFC3339, createdAt)
	if appliedAt.Valid {
		t, _ := time.Parse(time.RFC3339, appliedAt.String)
		op.AppliedAt = &t
	}
	return &op, nil
}

func (r *operationRepo) FetchChanges(ctx context.Context, cursor string, limit int) ([]*domain.Operation, error) {
	query := "SELECT id, actor_id, device_id, entity_kind, entity_id, operation_type, payload_json, base_version, status, created_at, applied_at FROM operations WHERE status = 'applied'"
	args := []any{}
	if cursor != "" {
		query += " AND applied_at > ?"
		args = append(args, cursor)
	}
	query += " ORDER BY applied_at ASC LIMIT ?"
	args = append(args, limit)

	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var results []*domain.Operation
	for rows.Next() {
		var op domain.Operation
		var payload string
		var createdAt string
		var appliedAt sql.NullString
		err := rows.Scan(&op.ID, &op.ActorID, &op.DeviceID, &op.EntityKind, &op.EntityID, &op.OperationType, &payload, &op.BaseVersion, &op.Status, &createdAt, &appliedAt)
		if err != nil {
			return nil, err
		}
		op.Payload = json.RawMessage(payload)
		op.CreatedAt, _ = time.Parse(time.RFC3339, createdAt)
		if appliedAt.Valid {
			t, _ := time.Parse(time.RFC3339, appliedAt.String)
			op.AppliedAt = &t
		}
		results = append(results, &op)
	}
	return results, nil
}
