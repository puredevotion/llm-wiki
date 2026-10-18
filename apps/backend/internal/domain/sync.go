package domain

import (
	"encoding/json"
	"time"
)

// OperationStatus represents the state of a synchronized write.
type OperationStatus string

const (
	OperationPending  OperationStatus = "pending"
	OperationApplied  OperationStatus = "applied"
	OperationRejected OperationStatus = "rejected"
)

// Operation represents a discrete write request from a client or service.
type Operation struct {
	ID            string          `json:"id"`             // Unique ULID/UUIDv7
	ActorID       string          `json:"actor_id"`       // Who performed the write
	DeviceID      string          `json:"device_id"`      // Originating device
	EntityKind    string          `json:"entity_kind"`    // zettel, topic, etc.
	EntityID      string          `json:"entity_id"`      // ID of the target entity
	OperationType string          `json:"operation_type"` // upsert, delete
	Payload       json.RawMessage `json:"payload"`        // Data for the operation
	BaseVersion   *int            `json:"base_version"`   // Version client last saw
	Status        OperationStatus `json:"status"`
	CreatedAt     time.Time       `json:"created_at"`
	AppliedAt     *time.Time      `json:"applied_at"`
}

// SyncBatch is a set of operations submitted by a client.
type SyncBatch struct {
	Operations []Operation `json:"operations"`
}

// SyncChange is a wrapper for a confirmed change in the feed.
type SyncChange struct {
	Operation Operation `json:"operation"`
	Cursor    string    `json:"cursor"` // Sequential identifier for polling
}
