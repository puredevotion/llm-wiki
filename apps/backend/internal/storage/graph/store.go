package graph

import (
	"context"
	"fmt"
	"strings"

	lb "github.com/LadybugDB/go-ladybug"
	"llm-wiki/apps/backend/internal/domain"
	"llm-wiki/apps/backend/internal/repositories"
)

type Store struct {
	db *lb.Database
}

func NewStore(path string) (*Store, error) {
	// Initialize the database (on-disk)
	db, err := lb.OpenDatabase(path, lb.DefaultSystemConfig())
	if err != nil {
		return nil, fmt.Errorf("failed to open ladybugdb: %w", err)
	}
	return &Store{db: db}, nil
}

func (s *Store) Close() error {
	s.db.Close()
	return nil
}

func (s *Store) Migrate(ctx context.Context) error {
	conn, err := lb.OpenConnection(s.db)
	if err != nil {
		return err
	}
	defer conn.Close()

	// Initial schema
	schemaQueries := []string{
		"CREATE NODE TABLE Person(id STRING, name STRING, PRIMARY KEY (id))",
		"CREATE NODE TABLE Team(id STRING, name STRING, PRIMARY KEY (id))",
		"CREATE NODE TABLE Organization(id STRING, name STRING, PRIMARY KEY (id))",
		"CREATE NODE TABLE Topic(id STRING, name STRING, PRIMARY KEY (id))",
		"CREATE NODE TABLE Project(id STRING, name STRING, PRIMARY KEY (id))",
		"CREATE NODE TABLE Source(id STRING, title STRING, kind STRING, PRIMARY KEY (id))",
		"CREATE NODE TABLE Zettel(id STRING, title STRING, lifecycle STRING, PRIMARY KEY (id))",
		"CREATE NODE TABLE Event(id STRING, title STRING, kind STRING, PRIMARY KEY (id))",
		"CREATE REL TABLE AUTHORED_BY(FROM Source TO Person)",
		"CREATE REL TABLE RELATED_TO(FROM Source TO Topic, FROM Source TO Project, FROM Zettel TO Topic)",
		"CREATE REL TABLE DERIVED_FROM(FROM Zettel TO Source)",
		"CREATE REL TABLE MEMBER_OF(FROM Person TO Team)",
		"CREATE REL TABLE PART_OF(FROM Team TO Organization)",
		"CREATE REL TABLE REPORTS_TO(FROM Person TO Person)",
		"CREATE REL TABLE HAPPENED_DURING(FROM Event TO Project)",
		"CREATE REL TABLE INVOLVES(FROM Event TO Person)",
		"CREATE REL TABLE PRECEDES(FROM Event TO Event)",
		"CREATE REL TABLE FOLLOWS(FROM Event TO Event)",
		"CREATE REL TABLE REFERS_TO(FROM Event TO Zettel)",
	}

	for _, q := range schemaQueries {
		_, err := conn.Query(q)
		if err != nil {
			// Ignore if table already exists
			if !strings.Contains(err.Error(), "already exists") {
				return fmt.Errorf("failed to execute schema query %q: %w", q, err)
			}
		}
	}
	return nil
}

func (s *Store) Graph() repositories.GraphRepository {
	return &graphRepo{s}
}

type graphRepo struct{ *Store }

func (r *graphRepo) UpsertNode(ctx context.Context, id, label string, properties map[string]any) error {
	conn, err := lb.OpenConnection(r.db)
	if err != nil {
		return err
	}
	defer conn.Close()

	props := make([]string, 0, len(properties)+1)
	props = append(props, fmt.Sprintf("id: '%s'", id))
	for k, v := range properties {
		props = append(props, fmt.Sprintf("%s: '%v'", k, v))
	}
	query := fmt.Sprintf("CREATE (:%s {%s})", label, strings.Join(props, ", "))

	_, err = conn.Query(query)
	if err != nil {
		errLower := strings.ToLower(err.Error())
		// Check for missing property error as well to improve coverage
		if strings.Contains(errLower, "cannot find property") {
			return fmt.Errorf("property not in schema: %w", err)
		}
		if strings.Contains(errLower, "primary key") && strings.Contains(errLower, "violates") {
			// Update properties if node exists
			updateProps := make([]string, 0, len(properties))
			for k, v := range properties {
				updateProps = append(updateProps, fmt.Sprintf("n.%s = '%v'", k, v))
			}
			if len(updateProps) > 0 {
				updateQuery := fmt.Sprintf("MATCH (n:%s {id: '%s'}) SET %s", label, id, strings.Join(updateProps, ", "))
				_, err = conn.Query(updateQuery)
				return err
			}
			return nil
		}
		return fmt.Errorf("failed to upsert node %s:%s: %w", label, id, err)
	}
	return nil
}

func (r *graphRepo) CreateRelationship(ctx context.Context, fromID, fromLabel, toID, toLabel, relType string) error {
	conn, err := lb.OpenConnection(r.db)
	if err != nil {
		return err
	}
	defer conn.Close()

	query := fmt.Sprintf("MATCH (a:%s {id: '%s'}), (b:%s {id: '%s'}) CREATE (a)-[:%s]->(b)", fromLabel, fromID, toLabel, toID, relType)
	_, err = conn.Query(query)
	if err != nil {
		return fmt.Errorf("failed to create relationship %s:%s -> %s:%s [%s]: %w", fromLabel, fromID, toLabel, toID, relType, err)
	}
	return nil
}

func (r *graphRepo) FetchGraph(ctx context.Context) (*domain.GraphData, error) {
	conn, err := lb.OpenConnection(r.db)
	if err != nil {
		return nil, err
	}
	defer conn.Close()

	data := &domain.GraphData{
		Nodes: []domain.GraphNode{},
		Links: []domain.GraphLink{},
	}

	// 1. Fetch Nodes
	tables := []struct {
		label string
		prop  string
	}{
		{"Person", "name"},
		{"Team", "name"},
		{"Organization", "name"},
		{"Topic", "name"},
		{"Project", "name"},
		{"Source", "title"},
		{"Zettel", "title"},
		{"Event", "title"},
	}

	for _, t := range tables {
		query := fmt.Sprintf("MATCH (n:%s) RETURN n.id, n.%s", t.label, t.prop)
		res, err := conn.Query(query)
		if err != nil {
			continue
		}
		for res.HasNext() {
			tuple, err := res.Next()
			if err != nil {
				continue
			}
			idVal, _ := tuple.GetValue(0)
			nameVal, _ := tuple.GetValue(1)
			id, _ := idVal.(string)
			name, _ := nameVal.(string)
			
			data.Nodes = append(data.Nodes, domain.GraphNode{
				ID:    id,
				Label: t.label,
				Name:  name,
			})
		}
		res.Close()
	}

	// 2. Fetch Relationships
	relQuery := "MATCH (a)-[r]->(b) RETURN a.id, b.id, label(r)"
	res, err := conn.Query(relQuery)
	if err == nil {
		for res.HasNext() {
			tuple, err := res.Next()
			if err != nil {
				continue
			}
			srcVal, _ := tuple.GetValue(0)
			dstVal, _ := tuple.GetValue(1)
			typeVal, _ := tuple.GetValue(2)
			src, _ := srcVal.(string)
			dst, _ := dstVal.(string)
			rType, _ := typeVal.(string)
			
			data.Links = append(data.Links, domain.GraphLink{
				Source: src,
				Target: dst,
				Type:   rType,
			})
		}
		res.Close()
	}

	return data, nil
}
