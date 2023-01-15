package types

import "sync"

// - - - - - - - - - - - - - - - - - - - - - - - - - -
// - - - - - - - - STORAGE STRUCTS - - - - - - - - - -
// - - - - - - - - - - - - - - - - - - - - - - - - - -

type Storage struct {
	Mutex struct {
		EntityType sync.RWMutex
		Entity     sync.RWMutex
		Relation   sync.RWMutex
	}
	Relations struct {
		DirectedIndex   map[int]map[int]map[int]map[int]*StorageRelation
		DirectedRIndex  map[int]map[int]map[int]map[int]*StorageRelation
		UndirectedIndex map[int]struct {
			Alpha   [2]int
			Beta    [2]int
			Pointer *StorageRelation
		}
		Data map[int]StorageRelation
	}
	Entities struct {
		IDMax map[int]int
		Index map[int]map[int]*StorageEntity
		Data  map[int]StorageEntity
	}
	Types struct {
		IDMax int
		Types map[int]string
	}
}

// - - - - - - - - - - - - - - - - - - - - - - - - - -
// StorageEntity struct
type StorageEntity struct {
	ID         int
	Type       int
	Context    string
	Value      string
	Properties map[string]string
	Version    int
}

// - - - - - - - - - - - - - - - - - - - - - - - - - -
// relation struct
type StorageRelation struct {
	SourceType int
	SourceID   int
	TargetType int
	TargetID   int
	Context    string
	Properties map[string]string
	Version    int
}

// - - - - - - - - - - - - - - - - - - - - - - - - - -
// - - - - - - PERSISTANCE STRUCTS - - - - - - - - - -
// - - - - - - - - - - - - - - - - - - - - - - - - - -

// - - - - - - - - - - - - - - - - - - - - - - - - - -
// persistance payload struct
type PersistencePayload struct {
	Type        string
	Method      string
	Entity      StorageEntity
	Relation    StorageRelation
	EntityTypes map[int]string
}

// - - - - - - - - - - - - - - - - - - - - - - - - - -
// persistance config struct
type PersistenceConfig struct {
	PersistenceChannelBufferSize int
	Active                       bool
	RotationEntriesMax           int
}
