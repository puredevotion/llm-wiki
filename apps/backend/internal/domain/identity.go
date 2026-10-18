package domain

import "time"

// Team represents a group of actors.
type Team struct {
	ID           string
	OrgID        string
	Name         string
	Metadata     map[string]any
	CreatedAt    time.Time
	UpdatedAt    time.Time
}

// Organization represents a top-level organizational entity.
type Organization struct {
	ID           string
	Name         string
	Metadata     map[string]any
	CreatedAt    time.Time
	UpdatedAt    time.Time
}

// TeamMember represents a membership relationship in SQL.
type TeamMember struct {
	TeamID    string
	ActorID   string
	Role      string
	CreatedAt time.Time
}
