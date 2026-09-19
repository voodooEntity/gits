package persistence

import "github.com/voodooEntity/gits/src/types"

// EventType identifies the storage mutation represented by a PersistenceEvent.
type EventType string

const (
	EventTypeCreateEntityType EventType = "create_entity_type"
	EventTypeCreateEntity     EventType = "create_entity"
	EventTypeUpdateEntity     EventType = "update_entity"
	EventTypeDeleteEntity     EventType = "delete_entity"
	EventTypeCreateRelation   EventType = "create_relation"
	EventTypeUpdateRelation   EventType = "update_relation"
	EventTypeDeleteRelation   EventType = "delete_relation"
)

// PersistenceEvent is a single ordered mutation to be applied by an Adapter.
type PersistenceEvent struct {
	Sequence       uint64
	Type           EventType
	EntityTypeID   int
	EntityTypeName string
	Entity         types.StorageEntity
	Relation       types.StorageRelation
}

// Snapshot is the full persisted storage state returned during bootstrap.
type Snapshot struct {
	EntityTypes     map[int]string
	EntityRTypes    map[string]int
	EntityTypeIDMax int
	EntityIDMax     map[int]int
	Entities        []types.StorageEntity
	Relations       []types.StorageRelation
}

// NewSnapshot returns a Snapshot with non-nil maps for safe mutation.
func NewSnapshot() *Snapshot {
	return &Snapshot{
		EntityTypes:  make(map[int]string),
		EntityRTypes: make(map[string]int),
		EntityIDMax:  make(map[int]int),
	}
}
