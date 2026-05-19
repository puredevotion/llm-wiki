package domain

import "time"

// Actor represents a person, agent, or service.
type Actor struct {
	ID          string
	Kind        string // person, agent, service
	DisplayName string
	Metadata    map[string]any
	CreatedAt   time.Time
}

// Source represents an imported artifact.
type Source struct {
	ID           string
	Kind         string // url, pdf, conversation, paste, file, agent_note
	URI          string
	Title        string
	ContentHash  string
	RawObjectRef string
	Metadata     map[string]any
	CapturedBy   string // Actor ID
	CapturedAt   time.Time
}

// Zettel represents an atomic note.
type Zettel struct {
	ID         string
	Title      string
	Body       string
	Lifecycle  string // ephemeral, project, evergreen
	Status     string // inbox, active, archived, superseded
	CreatedBy  string // Actor ID
	ValidFrom  *time.Time
	ValidTo    *time.Time
	ReviewAfter *time.Time
	CreatedAt  time.Time
	UpdatedAt  time.Time
}

// Topic represents a knowledge area.
type Topic struct {
	ID        string
	Name      string
	ParentID  string
	CreatedAt time.Time
}

// SummaryPayload represents the input for ingesting a conversation summary.
type SummaryPayload struct {
	Project      string    `json:"project"`
	Participants []string  `json:"participants"`
	Topics       []string  `json:"topics"`
	Timestamp    time.Time `json:"timestamp"`
	Summary      string    `json:"summary"`
	Conclusions  []string  `json:"conclusions"`
}
