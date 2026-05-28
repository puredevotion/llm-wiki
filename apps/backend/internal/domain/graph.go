package domain

// GraphNode represents a node in the knowledge graph.
type GraphNode struct {
	ID    string `json:"id"`
	Label string `json:"label"` // Person, Team, Topic, Project, Zettel, Event
	Name  string `json:"name"`
	Kind  string `json:"kind,omitempty"`
}

// GraphLink represents a relationship between two nodes.
type GraphLink struct {
	Source string `json:"source"`
	Target string `json:"target"`
	Type   string `json:"type"`
}

// GraphData represents the complete visualization-ready graph.
type GraphData struct {
	Nodes []GraphNode `json:"nodes"`
	Links []GraphLink `json:"links"`
}
