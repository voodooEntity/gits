package types

// - - - - - - - - - - - - - - - - - - - - - - - - - -
// - - - - - - - - STORAGE STRUCTS - - - - - - - - - -
// - - - - - - - - - - - - - - - - - - - - - - - - - -

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
