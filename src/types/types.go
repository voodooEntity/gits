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

// - - - - - - - - - - - - - - - - - - - - - - - - - -
type Query struct {
	Method int
	Match  [][][3]string
	Map    []Query
	Mode   [][]string
}

type Result struct {
	Entities  map[int]StorageEntity
	Relations map[int]StorageRelation
}

var qry = Query{
	Method: 1,
	Match: [][][3]string{
		{
			{"type", "==", "testtype"},
			{"properties.Something", "!=", "idontwantthis"},
		},
	},
	Map: []Query{
		{
			Method: 2,
			Match: [][][3]string{
				{
					{"type", "==", "testchild"},
					{"context", "==", "findme"},
				},
			},
		},
	},
	Mode: [][]string{
		{"order", "id", "desc"},
		{"limit", "10"},
	},
}

// - - - - - - - - - - - - - - - - - - - - - - - - - -
/**
Methods:
-> READ
-> REDUCE
-> UPDATE
-> DELETE
-> COUNT

Type:
-> Entity
-> Relation

Filter:
-> Value
-> Context
-> Property
-> ID
-> Type

Compare Operators:
-> equals
-> prefix
-> suffix
-> substring
-> >=
-> <=
-> ==

AFTERPROCESSING:
-> ORDER BY % ASC/DESC

SPECIAL:
-> LIMIT
-> TRAVERSE
-> RTRAVERSE

*/
