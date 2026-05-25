package domain

import "time"

// EventKind represents the type of a temporal event.
type EventKind string

const (
	EventMeeting   EventKind = "meeting"
	EventMilestone EventKind = "milestone"
	EventTask      EventKind = "task"
	EventDecision  EventKind = "decision"
	EventLog       EventKind = "log"
)

// Event represents a discrete point or interval in time.
type Event struct {
	ID         string         `json:"id"`
	Kind       EventKind      `json:"kind"`
	Title      string         `json:"title"`
	Body       string         `json:"body"`
	OccurredAt *time.Time     `json:"occurred_at,omitempty"`
	StartsAt   *time.Time     `json:"starts_at,omitempty"`
	EndsAt     *time.Time     `json:"ends_at,omitempty"`
	RecordedAt time.Time      `json:"recorded_at"`
	CreatedBy  string         `json:"created_by"` // Actor ID
	Metadata   map[string]any `json:"metadata"`
}
