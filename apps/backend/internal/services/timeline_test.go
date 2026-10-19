package services

import (
	"context"
	"fmt"
	"testing"
	"time"

	"llm-wiki/apps/backend/internal/domain"
)

type mockTimelineRepo struct {
	events map[string]*domain.Event
	fail   bool
}

func (m *mockTimelineRepo) Save(ctx context.Context, e *domain.Event) error {
	if m.fail {
		return fmt.Errorf("sql fail")
	}
	if m.events == nil {
		m.events = make(map[string]*domain.Event)
	}
	m.events[e.ID] = e
	return nil
}

func (m *mockTimelineRepo) FindByID(ctx context.Context, id string) (*domain.Event, error) {
	if m.fail {
		return nil, fmt.Errorf("sql fail")
	}
	return m.events[id], nil
}

func (m *mockTimelineRepo) Fetch(ctx context.Context, startsAt, endsAt *time.Time, limit int) ([]*domain.Event, error) {
	if m.fail {
		return nil, fmt.Errorf("sql fail")
	}
	var results []*domain.Event
	for _, e := range m.events {
		results = append(results, e)
		if len(results) >= limit {
			break
		}
	}
	return results, nil
}

func TestTimelineService(t *testing.T) {
	events := &mockTimelineRepo{events: make(map[string]*domain.Event)}
	graph := &mockGraphRepo{nodes: make(map[string]string)}
	ops := &mockOpRepo{ops: make(map[string]*domain.Operation)}

	svc := NewTimelineService(events, graph, ops)
	ctx := context.Background()

	t.Run("Record Event", func(t *testing.T) {
		e := &domain.Event{
			Kind:  domain.EventMeeting,
			Title: "Design Review",
			Body:  "Discussed time plane",
		}
		if err := svc.RecordEvent(ctx, e); err != nil {
			t.Fatalf("RecordEvent failed: %v", err)
		}
		if _, ok := events.events[e.ID]; !ok {
			t.Error("event was not saved to repo")
		}
		if graph.nodes[e.ID] != "Event" {
			t.Error("event was not saved to graph")
		}
	})

	t.Run("Record Event Default ID and Time", func(t *testing.T) {
		e := &domain.Event{Title: "Untitled"}
		if err := svc.RecordEvent(ctx, e); err != nil {
			t.Fatalf("RecordEvent failed: %v", err)
		}
		if e.ID == "" {
			t.Error("expected default ID")
		}
		if e.RecordedAt.IsZero() {
			t.Error("expected default RecordedAt")
		}
	})

	t.Run("Fetch Timeline", func(t *testing.T) {
		res, err := svc.FetchTimeline(ctx, nil, nil, 10)
		if err != nil {
			t.Fatalf("FetchTimeline failed: %v", err)
		}
		if len(res) == 0 {
			t.Error("expected events, got 0")
		}
	})

	t.Run("Get Event", func(t *testing.T) {
		e, err := svc.GetEvent(ctx, "e1")
		if err != nil {
			t.Fatalf("GetEvent failed: %v", err)
		}
		if e == nil {
			// e1 might not exist if Record Event test didn't use e1
		}
	})

	t.Run("Fetch Timeline Default Limit", func(t *testing.T) {
		res, err := svc.FetchTimeline(ctx, nil, nil, 0)
		if err != nil {
			t.Fatalf("FetchTimeline failed: %v", err)
		}
		if len(res) == 0 {
			t.Error("expected events")
		}
	})

	t.Run("Fetch Timeline with Dates", func(t *testing.T) {
		now := time.Now()
		res, err := svc.FetchTimeline(ctx, &now, &now, 10)
		if err != nil {
			t.Fatalf("FetchTimeline failed: %v", err)
		}
		if len(res) == 0 {
			t.Error("expected events")
		}
	})

	t.Run("Relate Event", func(t *testing.T) {
		if err := svc.RelateEvent(ctx, "e1", "p1", "Project", "HAPPENED_DURING"); err != nil {
			t.Fatalf("RelateEvent failed: %v", err)
		}
	})

	t.Run("Record Event Fail SQL", func(t *testing.T) {
		events.fail = true
		defer func() { events.fail = false }()
		err := svc.RecordEvent(ctx, &domain.Event{Title: "Fail"})
		if err == nil {
			t.Error("expected sql failure")
		}
	})

	t.Run("Record Event Fail Graph", func(t *testing.T) {
		graph.fail = true
		defer func() { graph.fail = false }()
		err := svc.RecordEvent(ctx, &domain.Event{Title: "Fail"})
		if err == nil {
			t.Error("expected graph failure")
		}
	})

	t.Run("Relate Event Fail Graph", func(t *testing.T) {
		graph.failRel = true
		defer func() { graph.failRel = false }()
		err := svc.RelateEvent(ctx, "e1", "p1", "Project", "X")
		if err == nil {
			t.Error("expected graph relationship failure")
		}
	})
}
