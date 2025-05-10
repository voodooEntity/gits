package storage

// handle all the imports
import (
	"errors"
	"regexp"
	"strconv"
	"strings"
	"sync"

	// Added import for cond package
	"github.com/voodooEntity/gits/src/query/cond"
	"github.com/voodooEntity/gits/src/transport"
	"github.com/voodooEntity/gits/src/types"
)

var persistenceFlag = false

type Storage struct {
	EntityStorage        map[int]map[int]types.StorageEntity
	EntityStorageMutex   *sync.RWMutex
	EntityIDMax          map[int]int
	EntityIDMaxMutex     *sync.RWMutex
	EntityTypes          map[int]string
	EntityRTypes         map[string]int
	EntityTypeIDMax      int
	EntityTypeMutex      *sync.RWMutex
	RelationStorage      map[int]map[int]map[int]map[int]types.StorageRelation
	RelationRStorage     map[int]map[int]map[int]map[int]bool
	RelationStorageMutex *sync.RWMutex
}

const (
	MAP_FORCE_CREATE  = -1
	MAP_IF_NOT_EXISTS = 0
)

// direction constants
const (
	DIRECTION_NONE   = -1
	DIRECTION_PARENT = 0
	DIRECTION_CHILD  = 1
)

func NewStorage() *Storage {
	return &Storage{
		EntityStorage: make(map[int]map[int]types.StorageEntity),

		// entity storage master mutex
		EntityStorageMutex: &sync.RWMutex{},

		// - - - - - - - - - - - - - - - - - - - - - - - - - -
		// entity storage id max         [Type]
		EntityIDMax: make(map[int]int),

		// master mutexd for EntityIdMax
		EntityIDMaxMutex: &sync.RWMutex{},

		// - - - - - - - - - - - - - - - - - - - - - - - - - -
		// maps to translate Types to their INT and reverse
		EntityTypes:  make(map[int]string),
		EntityRTypes: make(map[string]int),

		// and a fitting max ID
		EntityTypeIDMax: 0,

		// entity Type mutex (for adding and deleting Type types)
		EntityTypeMutex: &sync.RWMutex{},

		// - - - - - - - - - - - - - - - - - - - - - - - - - -
		// s prefix = source
		// t prefix = target
		// relation storage map             [sType][sId]   [tType][tId]
		RelationStorage: make(map[int]map[int]map[int]map[int]types.StorageRelation),

		// and relation reverse storage map
		// (for faster queries)              [tType][Tid]   [sType][sId]
		RelationRStorage: make(map[int]map[int]map[int]map[int]bool),

		// relation storage master mutex
		RelationStorageMutex: &sync.RWMutex{},
	}
}

// - - - - - - - - - - - - - - - - - - - - - - - - - -
// + + + + + + + + + +  PUBLIC  + + + + + + + + + + +
// - - - - - - - - - - - - - - - - - - - - - - - - - -

// - - - - - - - - - - - - - - - - - - - - - - - - - -
// Create an entity Type
func (s *Storage) CreateEntityType(name string) (int, error) {
	// first of allw e lock
	s.EntityTypeMutex.Lock()

	// lets check if the Type allready exists
	// if it does we just return the ID
	if id, ok := s.EntityRTypes[name]; ok {
		// dont forget to unlock
		s.EntityTypeMutex.Unlock()
		return id, nil
	}

	// ok entity doesnt exist yet, lets
	// upcount our ID Max and copy it
	// into another variable so we can be sure
	// between unlock of the ressource and return
	// it doesnt get upcounted
	s.EntityTypeIDMax++
	var newID = s.EntityTypeIDMax

	// finally create the new Type in our
	// EntityTypes index and reverse index
	s.EntityTypes[newID] = name
	s.EntityRTypes[name] = newID

	// and create mutex for EntityStorage Type+type
	s.EntityStorageMutex.Lock()

	// now we prepare the submaps in the entity
	// storage itseöf....
	s.EntityStorage[newID] = make(map[int]types.StorageEntity)

	// set the maxID for the new
	// Type type
	s.EntityIDMax[newID] = 0
	s.EntityStorageMutex.Unlock()

	// create the base maps in relation storage
	s.RelationStorageMutex.Lock()
	s.RelationStorage[newID] = make(map[int]map[int]map[int]types.StorageRelation)
	s.RelationRStorage[newID] = make(map[int]map[int]map[int]bool)
	s.RelationStorageMutex.Unlock()

	// and create the basic submaps for
	// the relation storage
	// now we unlock the mutex
	// and return the new id
	// - - - - - - - - - - - - - - - - -
	// persistence.go handling
	if true == persistenceFlag {
		//persistence.PersistenceChan <- types.PersistencePayload{
		//	Type:        "EntityType",
		//	EntityTypes: s.EntityTypes,
		//}
	}
	// - - - - - - - - - - - - - - - - -
	s.EntityTypeMutex.Unlock()
	return newID, nil
}

func (s *Storage) CreateEntityTypeUnsafe(name string) (int, error) {
	// lets check if the Type allready exists
	// if it does we just return the ID
	if id, ok := s.EntityRTypes[name]; ok {
		// dont forget to unlock
		return id, nil
	}

	// ok entity doesnt exist yet, lets
	// upcount our ID Max and copy it
	// into another variable so we can be sure
	// between unlock of the ressource and return
	// it doesnt get upcounted
	s.EntityTypeIDMax++
	var newID = s.EntityTypeIDMax

	// finally create the new Type in our
	// EntityTypes index and reverse index
	s.EntityTypes[newID] = name
	s.EntityRTypes[name] = newID

	// now we prepare the submaps in the entity
	// storage itseöf....
	s.EntityStorage[newID] = make(map[int]types.StorageEntity)

	// set the maxID for the new
	// Type type
	s.EntityIDMax[newID] = 0

	// create the base maps in relation storage
	s.RelationStorage[newID] = make(map[int]map[int]map[int]types.StorageRelation)
	s.RelationRStorage[newID] = make(map[int]map[int]map[int]bool)

	// and create the basic submaps for
	// the relation storage
	// now we unlock the mutex
	// and return the new id
	// - - - - - - - - - - - - - - - - -
	// persistence.go handling
	if true == persistenceFlag {
		//persistence.PersistenceChan <- types.PersistencePayload{
		//	Type:        "EntityType",
		//	EntityTypes: s.EntityTypes,
		//}
	}
	// - - - - - - - - - - - - - - - - -
	return newID, nil
}

func (s *Storage) CreateEntity(entity types.StorageEntity) (int, error) {
	//types.PrintMemUsage()
	// first we lock the entity Type mutex
	// to make sure while we check for the
	// existence it doesnt get deletet, this
	// may sound like a very rare upcoming case,
	//but better be safe than sorry
	s.EntityTypeMutex.RLock()

	// now
	if _, ok := s.EntityTypes[entity.Type]; !ok {
		// the Type doest exist, lets unlock
		// the Type mutex and return -1 for fail0r
		s.EntityTypeMutex.RUnlock()
		return -1, errors.New("CreateEntity.Entity Type not existing")
	}
	// the Type seems to exist, now lets lock the
	// storage mutex before Unlocking the Entity
	// Type mutex to prevent the Type beeing
	// deleted before we start locking (small
	// timing still possible )
	s.EntityTypeMutex.RUnlock()

	// upcount our ID Max and copy it
	// into another variable so we can be sure
	// between unlock of the ressource and return
	// it doesnt get upcounted
	// and set the IDMaxMutex on write Lock
	// lets upcount the entity id max fitting to
	//         [Type]
	s.EntityStorageMutex.Lock()
	s.EntityIDMax[entity.Type]++
	var newID = s.EntityIDMax[entity.Type]

	//EntityIDMaxMasterMutex.Lock()
	// and tell the entity its own id
	entity.ID = newID

	// and set the version to 1
	entity.Version = 1

	// - - - - - - - - - - - - - - - - -
	// persistence.go handling
	if true == persistenceFlag {
		//persistence.PersistenceChan <- types.PersistencePayload{
		//	Type:   "Entity",
		//	Method: "Create",
		//	Entity: entity,
		//}
	}
	// - - - - - - - - - - - - - - - - -

	// now we store the entity element
	// in the EntityStorage
	s.EntityStorage[entity.Type][newID] = entity

	//printMutexActions("CreateEntity.EntityStorageMutex.Unlock");
	s.EntityStorageMutex.Unlock()

	// create the mutex for our ressource on
	// relation. we have to create the sub maps too
	// golang things....
	s.RelationStorageMutex.Lock()
	s.RelationStorage[entity.Type][newID] = make(map[int]map[int]types.StorageRelation)
	s.RelationRStorage[entity.Type][newID] = make(map[int]map[int]bool)
	s.RelationStorageMutex.Unlock()

	// since we now stored the entity and created all
	// needed ressources we can unlock
	// the storage ressource and return the ID (or err)
	return newID, nil
}

func (s *Storage) CreateEntityUnsafe(entity types.StorageEntity) (int, error) {
	//types.PrintMemUsage()
	// first we lock the entity Type mutex
	// to make sure while we check for the
	// existence it doesnt get deletet, this
	// may sound like a very rare upcoming case,
	//but better be safe than sorry

	// now
	if _, ok := s.EntityTypes[entity.Type]; !ok {
		// the Type doest exist, lets unlock
		// the Type mutex and return -1 for fail0r
		return -1, errors.New("CreateEntity.Entity Type not existing")
	}
	// the Type seems to exist, now lets lock the
	// storage mutex before Unlocking the Entity
	// Type mutex to prevent the Type beeing
	// deleted before we start locking (small
	// timing still possible )

	// upcount our ID Max and copy it
	// into another variable so we can be sure
	// between unlock of the ressource and return
	// it doesnt get upcounted
	// and set the IDMaxMutex on write Lock
	// lets upcount the entity id max fitting to
	//         [Type]
	s.EntityIDMax[entity.Type]++
	var newID = s.EntityIDMax[entity.Type]

	//EntityIDMaxMasterMutex.Lock()
	// and tell the entity its own id
	entity.ID = newID

	// and set the version to 1
	entity.Version = 1

	// - - - - - - - - - - - - - - - - -
	// persistence.go handling
	if true == persistenceFlag {
		//persistence.PersistenceChan <- types.PersistencePayload{
		//	Type:   "Entity",
		//	Method: "Create",
		//	Entity: entity,
		//}
	}
	// - - - - - - - - - - - - - - - - -

	// now we store the entity element
	// in the EntityStorage
	s.EntityStorage[entity.Type][newID] = entity

	// create the mutex for our ressource on
	// relation. we have to create the sub maps too
	// golang things....
	s.RelationStorage[entity.Type][newID] = make(map[int]map[int]types.StorageRelation)
	s.RelationRStorage[entity.Type][newID] = make(map[int]map[int]bool)

	// since we now stored the entity and created all
	// needed ressources we can unlock
	// the storage ressource and return the ID (or err)
	return newID, nil
}

// bool return = has a new dataset been created
func (s *Storage) CreateEntityUniqueValue(entity types.StorageEntity) (int, bool, error) {
	//types.PrintMemUsage()
	// first we lock the entity Type mutex
	// to make sure while we check for the
	// existence it doesnt get deletet, this
	// may sound like a very rare upcoming case,
	//but better be safe than sorry
	s.EntityTypeMutex.RLock()

	// now we will cache the stype string due to the
	// special hack implementation of createEntityUniqueValue
	// for what we created the unsafe retrieval version  getEntitiesByTypeAndValueUnsafe()
	// that expects a string instread of the usualy on create neccesary id.
	var stype string
	if val, ok := s.EntityTypes[entity.Type]; !ok {
		// the Type doest exist, lets unlock
		// the Type mutex and return -1 for fail0r
		s.EntityTypeMutex.RUnlock()
		return -1, false, errors.New("CreateEntityUniqueValue.Entity Type not existing")
	} else {
		stype = val
	}
	// the Type seems to exist, now lets lock the
	// storage mutex before Unlocking the Entity
	// Type mutex to prevent the Type beeing
	// deleted before we start locking (small
	// timing still possible )
	s.EntityTypeMutex.RUnlock()

	// since this is the UniqueValue variant
	// we have to lock and make sure the type:value combination
	// doesnt exist. thatfor we call getEntitiesByTypeAndValueUnsafe()
	// which doesnt have any locking implemented and thatfor will be able
	// to see if we can retrieve any entity fitting
	s.EntityStorageMutex.Lock()
	entities, err := s.GetEntitiesByTypeAndValueUnsafe(stype, entity.Value, "match", entity.Context)
	if nil != err {
		s.EntityStorageMutex.Unlock()
		return -1, false, err
	}
	// ### think about update logic since collection properties might change
	if 0 < len(entities) {
		s.EntityStorageMutex.Unlock()
		//return -1,errors.New("CreateEntityUniqueValue.Entity Entity with given value already exists")
		return entities[0].ID, false, nil
	}
	// upcount our ID Max and copy it
	// into another variable so we can be sure
	// between unlock of the ressource and return
	// it doesnt get upcounted
	// and set the IDMaxMutex on write Lock
	// lets upcount the entity id max fitting to
	//         [Type]
	s.EntityIDMax[entity.Type]++
	var newID = s.EntityIDMax[entity.Type]

	//EntityIDMaxMasterMutex.Lock()
	// and tell the entity its own id
	entity.ID = newID

	// and set the version to 1
	entity.Version = 1

	// - - - - - - - - - - - - - - - - -
	// persistance handling
	if true == persistenceFlag {
		//persistence.PersistenceChan <- types.PersistencePayload{
		//	Type:   "Entity",
		//	Method: "Create",
		//	Entity: entity,
		//}
	}
	// - - - - - - - - - - - - - - - - -

	// now we store the entity element
	// in the EntityStorage
	s.EntityStorage[entity.Type][newID] = entity

	//printMutexActions("CreateEntity.EntityStorageMutex.Unlock");
	s.EntityStorageMutex.Unlock()

	// create the mutex for our ressource on
	// relation. we have to create the sub maps too
	// golang things....
	s.RelationStorageMutex.Lock()
	s.RelationStorage[entity.Type][newID] = make(map[int]map[int]types.StorageRelation)
	s.RelationRStorage[entity.Type][newID] = make(map[int]map[int]bool)
	s.RelationStorageMutex.Unlock()

	// since we now stored the entity and created all
	// needed ressources we can unlock
	// the storage ressource and return the ID (or err)
	return newID, true, nil
}

// bool return = has a new dataset been created
func (s *Storage) CreateEntityUniqueValueUnsafe(entity types.StorageEntity) (int, bool, error) {
	//types.PrintMemUsage()
	// first we lock the entity Type mutex
	// to make sure while we check for the
	// existence it doesnt get deletet, this
	// may sound like a very rare upcoming case,
	//but better be safe than sorry

	// now we will cache the stype string due to the
	// special hack implementation of createEntityUniqueValue
	// for what we created the unsafe retrieval version  getEntitiesByTypeAndValueUnsafe()
	// that expects a string instread of the usualy on create neccesary id.
	var stype string
	if val, ok := s.EntityTypes[entity.Type]; !ok {
		// the Type doest exist, lets unlock
		// the Type mutex and return -1 for fail0r
		return -1, false, errors.New("CreateEntityUniqueValue.Entity Type not existing")
	} else {
		stype = val
	}
	// the Type seems to exist, now lets lock the
	// storage mutex before Unlocking the Entity
	// Type mutex to prevent the Type beeing
	// deleted before we start locking (small
	// timing still possible )

	// since this is the UniqueValue variant
	// we have to lock and make sure the type:value combination
	// doesnt exist. thatfor we call getEntitiesByTypeAndValueUnsafe()
	// which doesnt have any locking implemented and thatfor will be able
	// to see if we can retrieve any entity fitting
	entities, err := s.GetEntitiesByTypeAndValueUnsafe(stype, entity.Value, "match", entity.Context)
	if nil != err {
		return -1, false, err
	}
	// ### think about update logic since collection properties might change
	if 0 < len(entities) {
		//return -1,errors.New("CreateEntityUniqueValue.Entity Entity with given value already exists")
		return entities[0].ID, false, nil
	}
	// upcount our ID Max and copy it
	// into another variable so we can be sure
	// between unlock of the ressource and return
	// it doesnt get upcounted
	// and set the IDMaxMutex on write Lock
	// lets upcount the entity id max fitting to
	//         [Type]
	s.EntityIDMax[entity.Type]++
	var newID = s.EntityIDMax[entity.Type]

	//EntityIDMaxMasterMutex.Lock()
	// and tell the entity its own id
	entity.ID = newID

	// and set the version to 1
	entity.Version = 1

	// - - - - - - - - - - - - - - - - -
	// persistance handling
	if true == persistenceFlag {
		//persistence.PersistenceChan <- types.PersistencePayload{
		//	Type:   "Entity",
		//	Method: "Create",
		//	Entity: entity,
		//}
	}
	// - - - - - - - - - - - - - - - - -

	// now we store the entity element
	// in the EntityStorage
	s.EntityStorage[entity.Type][newID] = entity

	//printMutexActions("CreateEntity.EntityStorageMutex.Unlock");

	// create the mutex for our ressource on
	// relation. we have to create the sub maps too
	// golang things....
	s.RelationStorage[entity.Type][newID] = make(map[int]map[int]types.StorageRelation)
	s.RelationRStorage[entity.Type][newID] = make(map[int]map[int]bool)

	// since we now stored the entity and created all
	// needed ressources we can unlock
	// the storage ressource and return the ID (or err)
	return newID, true, nil
}

func (s *Storage) GetEntityByPath(Type int, id int, context string) (types.StorageEntity, error) {
	// lets check if entity witrh the given path exists
	s.EntityStorageMutex.Lock()
	if entity, ok := s.EntityStorage[Type][id]; ok {
		// if yes we return the entity
		// and nil for error
		if "" == context || entity.Context == context {
			ret := s.deepCopyEntity(entity)
			s.EntityStorageMutex.Unlock()
			return ret, nil
		}
	}

	s.EntityStorageMutex.Unlock()

	// the path seems to transport empty , so
	// we throw an error
	return types.StorageEntity{}, errors.New("Entity on given path does not exist.")
}

func (s *Storage) GetEntityByPathUnsafe(Type int, id int, context string) (types.StorageEntity, error) {
	// lets check if entity with the given path exists
	if entity, ok := s.EntityStorage[Type][id]; ok {
		// if yes we return the entity
		// and nil for error
		if "" == context || entity.Context == context {
			return s.deepCopyEntity(entity), nil
		}
	}

	// the path seems to transport empty , so
	// we throw an error
	return types.StorageEntity{}, errors.New("Entity on given path does not exist.")
}

func (s *Storage) GetEntitiesByType(Type string, context string) (map[int]types.StorageEntity, error) {
	// retrieve the fitting id
	entityTypeID, _ := s.GetTypeIdByString(Type)

	// lock retrieve und unlock the storage
	mapRet := make(map[int]types.StorageEntity)
	i := 0
	s.EntityStorageMutex.RLock()
	for _, entity := range s.EntityStorage[entityTypeID] {
		// preset add with true
		add := true

		// check if context is set , if yes and it doesnt
		// fit we dont add
		if context != "" && entity.Context != context {
			add = false
		}

		// finally if everything is fine we add the dataset
		if add {
			mapRet[i] = s.deepCopyEntity(entity)
			i++
		}
	}

	// unlock the storage again
	s.EntityStorageMutex.RUnlock()

	// return the entity map
	return mapRet, nil
}

func (s *Storage) GetEntitiesByTypeUnsafe(Type string, context string) (map[int]types.StorageEntity, error) {
	// retrieve the fitting id
	entityTypeID, _ := s.GetTypeIdByString(Type)

	// lock retrieve und unlock the storage
	mapRet := make(map[int]types.StorageEntity)
	i := 0
	for _, entity := range s.EntityStorage[entityTypeID] {
		// preset add with true
		add := true

		// check if context is set , if yes and it doesnt
		// fit we dont add
		if context != "" && entity.Context != context {
			add = false
		}

		// finally if everything is fine we add the dataset
		if add {
			mapRet[i] = s.deepCopyEntity(entity)
			i++
		}
	}

	// return the entity map
	return mapRet, nil
}

func (s *Storage) GetEntitiesByValue(value string, mode string, context string) (map[int]types.StorageEntity, error) {
	// lets prepare the return map, counter and regex r
	entities := make(map[int]types.StorageEntity)
	i := 0
	var r *regexp.Regexp
	var err error = nil

	// first we lock the storage
	s.EntityStorageMutex.RLock()

	// if we got mode regex we prepare the regex
	// by precompiling it to have faster lookups
	if "regex" == mode {
		r, err = regexp.Compile(value)

		// check if regex could be compiled successfull,
		// else return error
		if nil != err {
			return map[int]types.StorageEntity{}, err
		}
	}

	// than we iterate through all entity storage to find a fitting value
	if 0 < len(s.EntityStorage) {
		for typeID := range s.EntityStorage {
			if 0 < len(s.EntityStorage[typeID]) {
				for _, entity := range s.EntityStorage[typeID] {
					// preset add with true
					add := true

					// check if context is set , if yes and it doesnt
					// fit we dont add
					if context != "" && entity.Context != context {
						add = false
					}

					// finally if everything is fine we add the dataset
					if add {
						switch mode {
						case "match":
							// exact match
							if entity.Value == value {
								entities[i] = s.deepCopyEntity(entity)
								i++
							}
						case "prefix":
							// starts with
							if strings.HasPrefix(entity.Value, value) {
								entities[i] = s.deepCopyEntity(entity)
								i++
							}
						case "suffix":
							// ends with
							if strings.HasSuffix(entity.Value, value) {
								entities[i] = s.deepCopyEntity(entity)
								i++
							}
						case "contain":
							// string contains string
							if strings.Contains(entity.Value, value) {
								entities[i] = s.deepCopyEntity(entity)
								i++
							}
						case "regex":
							// string matches regex
							if r.MatchString(entity.Value) {
								entities[i] = s.deepCopyEntity(entity)
								i++
							}
						}
					}
				}
			}
		}
	}

	// unlock storage again and return
	s.EntityStorageMutex.RUnlock()
	return entities, nil
}

func (s *Storage) GetEntitiesByValueUnsafe(value string, mode string, context string) (map[int]types.StorageEntity, error) {
	// lets prepare the return map, counter and regex r
	entities := make(map[int]types.StorageEntity)
	i := 0
	var r *regexp.Regexp
	var err error = nil

	// first we lock the storage

	// if we got mode regex we prepare the regex
	// by precompiling it to have faster lookups
	if "regex" == mode {
		r, err = regexp.Compile(value)

		// check if regex could be compiled successfull,
		// else return error
		if nil != err {
			return map[int]types.StorageEntity{}, err
		}
	}

	// than we iterate through all entity storage to find a fitting value
	if 0 < len(s.EntityStorage) {
		for typeID := range s.EntityStorage {
			if 0 < len(s.EntityStorage[typeID]) {
				for _, entity := range s.EntityStorage[typeID] {
					// preset add with true
					add := true

					// check if context is set , if yes and it doesnt
					// fit we dont add
					if context != "" && entity.Context != context {
						add = false
					}

					// finally if everything is fine we add the dataset
					if add {
						switch mode {
						case "match":
							// exact match
							if entity.Value == value {
								entities[i] = s.deepCopyEntity(entity)
								i++
							}
						case "prefix":
							// starts with
							if strings.HasPrefix(entity.Value, value) {
								entities[i] = s.deepCopyEntity(entity)
								i++
							}
						case "suffix":
							// ends with
							if strings.HasSuffix(entity.Value, value) {
								entities[i] = s.deepCopyEntity(entity)
								i++
							}
						case "contain":
							// string contains string
							if strings.Contains(entity.Value, value) {
								entities[i] = s.deepCopyEntity(entity)
								i++
							}
						case "regex":
							// string matches regex
							if r.MatchString(entity.Value) {
								entities[i] = s.deepCopyEntity(entity)
								i++
							}
						}
					}
				}
			}
		}
	}

	return entities, nil
}

func (s *Storage) GetEntitiesByTypeAndValue(Type string, value string, mode string, context string) (map[int]types.StorageEntity, error) {
	// lets prepare the return map, counter and regex r
	entities := make(map[int]types.StorageEntity)
	i := 0
	var r *regexp.Regexp
	var err error = nil

	// first we lock the storage
	s.EntityStorageMutex.RLock()

	// retrieve the fitting id
	entityTypeID, _ := s.GetTypeIdByString(Type)

	// if we got mode regex we prepare the regex
	// by precompiling it to have faster lookups
	if "regex" == mode {
		r, err = regexp.Compile(value)

		// check if regex could be compiled successfull,
		// else return error
		if nil != err {
			return map[int]types.StorageEntity{}, err
		}
	}

	// than we iterate through all entity storage to find a fitting value
	if 0 < len(s.EntityStorage) {
		if 0 < len(s.EntityStorage[entityTypeID]) {
			for _, entity := range s.EntityStorage[entityTypeID] {
				// preset add with true
				add := true

				// check if context is set , if yes and it doesnt
				// fit we dont add
				if context != "" && entity.Context != context {
					add = false
				}

				// finally if everything is fine we add the dataset
				if add {
					switch mode {
					case "match":
						// exact match
						if entity.Value == value {
							entities[i] = s.deepCopyEntity(entity)
							i++
						}
					case "prefix":
						// starts with
						if strings.HasPrefix(entity.Value, value) {
							entities[i] = s.deepCopyEntity(entity)
							i++
						}
					case "suffix":
						// ends with
						if strings.HasSuffix(entity.Value, value) {
							entities[i] = s.deepCopyEntity(entity)
							i++
						}
					case "contain":
						// contains
						if strings.Contains(entity.Value, value) {
							entities[i] = s.deepCopyEntity(entity)
							i++
						}
					case "regex":
						// matches regex
						if r.MatchString(entity.Value) {
							entities[i] = s.deepCopyEntity(entity)
							i++
						}
					}
				}
			}
		}
	}

	// unlock storage again and return
	s.EntityStorageMutex.RUnlock()
	return entities, nil
}

func (s *Storage) GetEntitiesByTypeAndValueUnsafe(Type string, value string, mode string, context string) (map[int]types.StorageEntity, error) {
	// lets prepare the return map, counter and regex r
	entities := make(map[int]types.StorageEntity)
	i := 0
	var r *regexp.Regexp
	var err error = nil

	// retrieve the fitting id
	entityTypeID, _ := s.GetTypeIdByStringUnsafe(Type)

	// if we got mode regex we prepare the regex
	// by precompiling it to have faster lookups
	if "regex" == mode {
		r, err = regexp.Compile(value)

		// check if regex could be compiled successfull,
		// else return error
		if nil != err {
			return map[int]types.StorageEntity{}, err
		}
	}

	// than we iterate through all entity storage to find a fitting value
	if 0 < len(s.EntityStorage) {
		if 0 < len(s.EntityStorage[entityTypeID]) {
			for _, entity := range s.EntityStorage[entityTypeID] {
				// preset add with true
				add := true

				// check if context is set , if yes and it doesnt
				// fit we dont add
				if context != "" && entity.Context != context {
					add = false
				}

				// finally if everything is fine we add the dataset
				if add {
					switch mode {
					case "match":
						// exact match
						if entity.Value == value {
							entities[i] = s.deepCopyEntity(entity)
							i++
						}
					case "prefix":
						// starts with
						if strings.HasPrefix(entity.Value, value) {
							entities[i] = s.deepCopyEntity(entity)
							i++
						}
					case "suffix":
						// ends with
						if strings.HasSuffix(entity.Value, value) {
							entities[i] = s.deepCopyEntity(entity)
							i++
						}
					case "contain":
						// contains
						if strings.Contains(entity.Value, value) {
							entities[i] = s.deepCopyEntity(entity)
							i++
						}
					case "regex":
						// matches regex
						if r.MatchString(entity.Value) {
							entities[i] = s.deepCopyEntity(entity)
							i++
						}
					}
				}
			}
		}
	}

	return entities, nil
}

func (s *Storage) UpdateEntity(entity types.StorageEntity) error {
	// - - - - - - - - - - - - - - - - -
	// lock the storage for concurrency
	s.EntityStorageMutex.Lock()
	if check, ok := s.EntityStorage[entity.Type][entity.ID]; ok {
		// - - - - - - - - - - - - - - - - -
		// lets check if the version is up to date
		if entity.Version != check.Version {
			s.EntityStorageMutex.Unlock()
			return errors.New("Mismatch of version.")
		}
		entity.Version++

		// - - - - - - - - - - - - - - - - -
		// persistence.go handling
		if true == persistenceFlag {
			//persistence.PersistenceChan <- types.PersistencePayload{
			//	Type:   "Entity",
			//	Method: "Update",
			//	Entity: entity,
			//}
		}
		// - - - - - - - - - - - - - - - - -
		s.EntityStorage[entity.Type][entity.ID] = entity
		s.EntityStorageMutex.Unlock()
		return nil
	}

	// unlock the storage and return an error in case we get here
	s.EntityStorageMutex.Unlock()
	return errors.New("Cant update non existing entity")
}

func (s *Storage) UpdateEntityUnsafe(entity types.StorageEntity) error {
	// - - - - - - - - - - - - - - - - -
	// lock the storage for concurrency
	if check, ok := s.EntityStorage[entity.Type][entity.ID]; ok {
		// - - - - - - - - - - - - - - - - -
		// lets check if the version is up to date
		if entity.Version != check.Version {
			return errors.New("Mismatch of version.")
		}
		entity.Version++

		// - - - - - - - - - - - - - - - - -
		// persistence.go handling
		if true == persistenceFlag {
			//persistence.PersistenceChan <- types.PersistencePayload{
			//	Type:   "Entity",
			//	Method: "Update",
			//	Entity: entity,
			//}
		}
		// - - - - - - - - - - - - - - - - -
		s.EntityStorage[entity.Type][entity.ID] = entity
		return nil
	}

	// unlock the storage and return an error in case we get here
	return errors.New("Cant update non existing entity")
}

func (s *Storage) DeleteEntity(Type int, id int) {
	// we gonne lock the mutex and
	// delete the element
	s.EntityStorageMutex.Lock()
	// - - - - - - - - - - - - - - - - -
	// persistence.go handling
	if true == persistenceFlag {
		//persistence.PersistenceChan <- types.PersistencePayload{
		//	Type:   "Entity",
		//	Method: "Delete",
		//	Entity: types.StorageEntity{
		//		ID:   id,
		//		Type: Type,
		//	},
		//}
	}
	// - - - - - - - - - - - - - - - - -
	delete(s.EntityStorage[Type], id)
	s.EntityStorageMutex.Unlock()
	// now we delete the relations from and to this entity
	// first child
	s.DeleteChildRelations(Type, id)
	// than parent
	s.DeleteParentRelations(Type, id)
}

func (s *Storage) DeleteEntityUnsafe(Type int, id int) {
	// we gonne lock the mutex and
	// delete the element
	// - - - - - - - - - - - - - - - - -
	// persistence.go handling
	if true == persistenceFlag {
		//persistence.PersistenceChan <- types.PersistencePayload{
		//	Type:   "Entity",
		//	Method: "Delete",
		//	Entity: types.StorageEntity{
		//		ID:   id,
		//		Type: Type,
		//	},
		//}
	}
	// - - - - - - - - - - - - - - - - -
	delete(s.EntityStorage[Type], id)
	// now we delete the relations from and to this entity
	// first child
	s.DeleteChildRelationsUnsafe(Type, id)
	// than parent
	s.DeleteParentRelationsUnsafe(Type, id)
}

func (s *Storage) GetRelation(srcType int, srcID int, targetType int, targetID int) (types.StorageRelation, error) {
	// first we lock the relation storage
	s.RelationStorageMutex.RLock()
	if _, firstOk := s.RelationStorage[srcType]; firstOk {
		if _, secondOk := s.RelationStorage[srcType][srcID]; secondOk {
			if _, thirdOk := s.RelationStorage[srcType][srcID][targetType]; thirdOk {
				if relation, fourthOk := s.RelationStorage[srcType][srcID][targetType][targetID]; fourthOk {
					s.RelationStorageMutex.RUnlock()
					return s.deepCopyRelation(relation), nil
				}
			}
		}
	}
	s.RelationStorageMutex.RUnlock()
	return types.StorageRelation{}, errors.New("Non existing relation requested")
}

func (s *Storage) GetRelationUnsafe(srcType int, srcID int, targetType int, targetID int) (types.StorageRelation, error) {
	// first we lock the relation storage
	if _, firstOk := s.RelationStorage[srcType]; firstOk {
		if _, secondOk := s.RelationStorage[srcType][srcID]; secondOk {
			if _, thirdOk := s.RelationStorage[srcType][srcID][targetType]; thirdOk {
				if relation, fourthOk := s.RelationStorage[srcType][srcID][targetType][targetID]; fourthOk {
					return s.deepCopyRelation(relation), nil
				}
			}
		}
	}
	return types.StorageRelation{}, errors.New("Non existing relation requested")
}

// maybe deprecated, check later
func (s *Storage) RelationExists(srcType int, srcID int, targetType int, targetID int) bool {
	// first we lock the relation storage
	s.RelationStorageMutex.RLock()
	if srcTypeMap, firstOk := s.RelationStorage[srcType]; firstOk {
		if srcIDMap, secondOk := srcTypeMap[srcID]; secondOk {
			if targetTypeMap, thirdOk := srcIDMap[targetType]; thirdOk {
				if _, fourthOk := targetTypeMap[targetID]; fourthOk {
					s.RelationStorageMutex.RUnlock()
					return true
				}
			}
		}
	}
	s.RelationStorageMutex.RUnlock()
	return false
}

func (s *Storage) RelationExistsUnsafe(srcType int, srcID int, targetType int, targetID int) bool {
	// first we lock the relation storage
	if srcTypeMap, firstOk := s.RelationStorage[srcType]; firstOk {
		if srcIDMap, secondOk := srcTypeMap[srcID]; secondOk {
			if targetTypeMap, thirdOk := srcIDMap[targetType]; thirdOk {
				if _, fourthOk := targetTypeMap[targetID]; fourthOk {
					return true
				}
			}
		}
	}
	return false
}

func (s *Storage) DeleteRelationList(relationList map[int]types.StorageRelation) {
	// lets walk through the iterations and delete all
	// corrosponding Relation & RRelation index entries
	if 0 < len(relationList) {
		for _, relation := range relationList {
			s.DeleteRelation(relation.SourceType, relation.SourceID, relation.TargetType, relation.TargetID)
		}
	}
	return
}

func (s *Storage) DeleteRelationListUnsafe(relationList map[int]types.StorageRelation) {
	// lets walk through the iterations and delete all
	// corrosponding Relation & RRelation index entries
	if 0 < len(relationList) {
		for _, relation := range relationList {
			s.DeleteRelationUnsafe(relation.SourceType, relation.SourceID, relation.TargetType, relation.TargetID)
		}
	}
	return
}

func (s *Storage) DeleteRelation(sourceType int, sourceID int, targetType int, targetID int) {
	s.RelationStorageMutex.Lock()
	// - - - - - - - - - - - - - - - - -
	// persistence.go handling
	if true == persistenceFlag {
		//persistence.PersistenceChan <- types.PersistencePayload{
		//	Type:   "Relation",
		//	Method: "Delete",
		//	Relation: types.StorageRelation{
		//		SourceID:   sourceID,
		//		SourceType: sourceType,
		//		TargetID:   targetID,
		//		TargetType: targetType,
		//	},
		//}
	}
	// - - - - - - - - - - - - - - - - -
	delete(s.RelationStorage[sourceType][sourceID][targetType], targetID)
	delete(s.RelationRStorage[targetType][targetID][sourceType], sourceID)
	s.RelationStorageMutex.Unlock()
}

func (s *Storage) DeleteRelationUnsafe(sourceType int, sourceID int, targetType int, targetID int) {
	// - - - - - - - - - - - - - - - - -
	// persistence.go handling
	if true == persistenceFlag {
		//persistence.PersistenceChan <- types.PersistencePayload{
		//	Type:   "Relation",
		//	Method: "Delete",
		//	Relation: types.StorageRelation{
		//		SourceID:   sourceID,
		//		SourceType: sourceType,
		//		TargetID:   targetID,
		//		TargetType: targetType,
		//	},
		//}
	}
	// - - - - - - - - - - - - - - - - -
	delete(s.RelationStorage[sourceType][sourceID][targetType], targetID)
	delete(s.RelationRStorage[targetType][targetID][sourceType], sourceID)
}

func (s *Storage) DeleteChildRelations(Type int, id int) error {
	childRelations, err := s.GetChildRelationsBySourceTypeAndSourceId(Type, id, "")
	if nil != err {
		return err
	}
	s.DeleteRelationList(childRelations)
	return nil
}

func (s *Storage) DeleteChildRelationsUnsafe(Type int, id int) error {
	childRelations, err := s.GetChildRelationsBySourceTypeAndSourceIdUnsafe(Type, id, "")
	if nil != err {
		return err
	}
	s.DeleteRelationListUnsafe(childRelations)
	return nil
}

func (s *Storage) DeleteParentRelations(Type int, id int) error {
	parentRelations, err := s.GetParentRelationsByTargetTypeAndTargetId(Type, id, "")
	if nil != err {
		return err
	}
	s.DeleteRelationList(parentRelations)
	return nil
}

func (s *Storage) DeleteParentRelationsUnsafe(Type int, id int) error {
	parentRelations, err := s.GetParentRelationsByTargetTypeAndTargetIdUnsafe(Type, id, "")
	if nil != err {
		return err
	}
	s.DeleteRelationListUnsafe(parentRelations)
	return nil
}

func (s *Storage) CreateRelation(srcType int, srcID int, targetType int, targetID int, relation types.StorageRelation) (bool, error) {
	// first we Readlock the EntityTypeMutex
	//printMutexActions("CreateRelation.EntityTypeMutex.RLock");
	s.EntityTypeMutex.RLock()
	// lets make sure the source Type exist
	if _, ok := s.EntityTypes[srcType]; !ok {
		//printMutexActions("CreateRelation.EntityTypeMutex.RUnlock");
		s.EntityTypeMutex.RUnlock()
		return false, errors.New("Source Type not existing")
	}
	// and the target Type exists too
	if _, ok := s.EntityTypes[targetType]; !ok {
		//printMutexActions("CreateRelation.EntityTypeMutex.RUnlock");
		s.EntityTypeMutex.RUnlock()
		return false, errors.New("Target Type not existing")
	}
	// finally unlock the TypeMutex again if both checks were successfull
	s.EntityTypeMutex.RUnlock()
	//// - - - - - - - - - - - - - - - - -
	// now we lock the relation mutex
	//printMutexActions("CreateRelation.RelationStorageMutex.Lock");
	s.RelationStorageMutex.Lock()
	// lets check if their exists a map for our
	// source entity to the target Type if not
	// create it.... golang things...
	if _, ok := s.RelationStorage[srcType][srcID][targetType]; !ok {
		s.RelationStorage[srcType][srcID][targetType] = make(map[int]types.StorageRelation)
		// if the map doesnt exist in this direction
		// it wont exist in the other as in reverse
		// map either so we should create it too
		// but we will store a pointer to the other
		// maps Relation instead of the complete
		// relation twice - defunct, refactor later (may create more problems then help)
		//RelationStorage[targetType][targetID][srcType] = make(map[int]Relation)
	}
	// now we prepare the reverse storage if necessary
	if _, ok := s.RelationRStorage[targetType][targetID][srcType]; !ok {
		s.RelationRStorage[targetType][targetID][srcType] = make(map[int]bool)
	}
	// set version to 1
	relation.Version = 1
	// now we store the relation
	s.RelationStorage[srcType][srcID][targetType][targetID] = relation
	// - - - - - - - - - - - - - - - - -
	// persistence.go handling
	if true == persistenceFlag {
		//persistence.PersistenceChan <- types.PersistencePayload{
		//	Type:     "Relation",
		//	Method:   "Create",
		//	Relation: relation,
		//}
	}
	// - - - - - - - - - - - - - - - - -
	// and an entry into the reverse index, its existence
	// allows us to use the coords in the normal index to revtrieve
	// the Relation. We dont create a pointer because golang doesnt
	// allow pointer on submaps in nested maps
	s.RelationRStorage[targetType][targetID][srcType][srcID] = true
	// we are done now we can unlock the entity Types
	//// - - - - - - - - - - - - - - - -
	//and finally unlock the relation Type and return
	//printMutexActions("CreateRelation.RelationStorageMutex.Unlock");
	s.RelationStorageMutex.Unlock()
	return true, nil
}

func (s *Storage) CreateRelationUnsafe(srcType int, srcID int, targetType int, targetID int, relation types.StorageRelation) (bool, error) {
	// first we Readlock the EntityTypeMutex
	//printMutexActions("CreateRelation.EntityTypeMutex.RLock");
	// lets make sure the source Type exist
	if _, ok := s.EntityTypes[srcType]; !ok {
		//printMutexActions("CreateRelation.EntityTypeMutex.RUnlock");
		return false, errors.New("Source Type not existing")
	}
	// and the target Type exists too
	if _, ok := s.EntityTypes[targetType]; !ok {
		//printMutexActions("CreateRelation.EntityTypeMutex.RUnlock");
		return false, errors.New("Target Type not existing")
	}
	// finally unlock the TypeMutex again if both checks were successfull
	//// - - - - - - - - - - - - - - - - -
	// now we lock the relation mutex
	//printMutexActions("CreateRelation.RelationStorageMutex.Lock");
	// lets check if their exists a map for our
	// source entity to the target Type if not
	// create it.... golang things...
	if _, ok := s.RelationStorage[srcType][srcID][targetType]; !ok {
		s.RelationStorage[srcType][srcID][targetType] = make(map[int]types.StorageRelation)
		// if the map doesnt exist in this direction
		// it wont exist in the other as in reverse
		// map either so we should create it too
		// but we will store a pointer to the other
		// maps Relation instead of the complete
		// relation twice - defunct, refactor later (may create more problems then help)
		//RelationStorage[targetType][targetID][srcType] = make(map[int]Relation)
	}
	// now we prepare the reverse storage if necessary
	if _, ok := s.RelationRStorage[targetType][targetID][srcType]; !ok {
		s.RelationRStorage[targetType][targetID][srcType] = make(map[int]bool)
	}
	// set version to 1
	relation.Version = 1
	// now we store the relation
	s.RelationStorage[srcType][srcID][targetType][targetID] = relation
	// - - - - - - - - - - - - - - - - -
	// persistence.go handling
	if true == persistenceFlag {
		//persistence.PersistenceChan <- types.PersistencePayload{
		//	Type:     "Relation",
		//	Method:   "Create",
		//	Relation: relation,
		//}
	}
	// - - - - - - - - - - - - - - - - -
	// and an entry into the reverse index, its existence
	// allows us to use the coords in the normal index to revtrieve
	// the Relation. We dont create a pointer because golang doesnt
	// allow pointer on submaps in nested maps
	s.RelationRStorage[targetType][targetID][srcType][srcID] = true
	// we are done now we can unlock the entity Types
	//// - - - - - - - - - - - - - - - -
	//and finally unlock the relation Type and return
	//printMutexActions("CreateRelation.RelationStorageMutex.Unlock");
	return true, nil
}

func (s *Storage) UpdateRelation(srcType int, srcID int, targetType int, targetID int, relation types.StorageRelation) (types.StorageRelation, error) {
	// first we lock the relation storage
	s.RelationStorageMutex.Lock()
	if _, firstOk := s.RelationStorage[srcType]; firstOk {
		if _, secondOk := s.RelationStorage[srcType][srcID]; secondOk {
			if _, thirdOk := s.RelationStorage[srcType][srcID][targetType]; thirdOk {
				if rel, fourthOk := s.RelationStorage[srcType][srcID][targetType][targetID]; fourthOk {
					// check if the version is fine
					if rel.Version != relation.Version {
						s.RelationStorageMutex.Unlock()
						return types.StorageRelation{}, errors.New("Mismatch of version.")
					}
					rel.Version++

					// - - - - - - - - - - - - - - - - -
					// persistence.go handling
					if true == persistenceFlag {
						//persistence.PersistenceChan <- types.PersistencePayload{
						//	Type:     "Relation",
						//	Method:   "Create",
						//	Relation: rel,
						//}
					}

					// - - - - - - - - - - - - - - - - -
					// update the data itself
					rel.Context = relation.Context
					rel.Properties = relation.Properties
					s.RelationStorage[srcType][srcID][targetType][targetID] = rel
					s.RelationStorageMutex.Unlock()
					return relation, nil
				}
			}
		}
	}
	s.RelationStorageMutex.Unlock()
	return types.StorageRelation{}, errors.New("Cant update non existing relation")
}

func (s *Storage) UpdateRelationUnsafe(srcType int, srcID int, targetType int, targetID int, relation types.StorageRelation) (types.StorageRelation, error) {
	// first we lock the relation storage
	if _, firstOk := s.RelationStorage[srcType]; firstOk {
		if _, secondOk := s.RelationStorage[srcType][srcID]; secondOk {
			if _, thirdOk := s.RelationStorage[srcType][srcID][targetType]; thirdOk {
				if rel, fourthOk := s.RelationStorage[srcType][srcID][targetType][targetID]; fourthOk {
					// check if the version is fine
					if rel.Version != relation.Version {
						return types.StorageRelation{}, errors.New("Mismatch of version.")
					}
					rel.Version++

					// - - - - - - - - - - - - - - - - -
					// persistence.go handling
					if true == persistenceFlag {
						//persistence.PersistenceChan <- types.PersistencePayload{
						//	Type:     "Relation",
						//	Method:   "Create",
						//	Relation: rel,
						//}
					}

					// - - - - - - - - - - - - - - - - -
					// update the data itself
					rel.Context = relation.Context
					rel.Properties = relation.Properties
					s.RelationStorage[srcType][srcID][targetType][targetID] = rel
					return relation, nil
				}
			}
		}
	}
	return types.StorageRelation{}, errors.New("Cant update non existing relation")
}

func (s *Storage) GetChildRelationsBySourceTypeAndSourceId(Type int, id int, context string) (map[int]types.StorageRelation, error) {
	// initialice the return map
	var mapRet = make(map[int]types.StorageRelation)
	// set counter for the loop
	var cnt = 0
	// copy the pool we have to search in
	// to prevent crashes on RW concurrency
	// we lock the RelationStorage mutex with
	// fitting Type. this allows us to proceed
	// faster since we just block to copy instead
	// of blocking for the whole process
	s.RelationStorageMutex.Lock()
	var pool = s.RelationStorage[Type][id]
	// for each possible targtType
	for _, targetTypeMap := range pool {
		// for each possible targetId per targetType
		for _, relation := range targetTypeMap {
			// context handling , default add
			add := true
			if "" != context && context != relation.Context {
				add = false
			}
			// if context is fine too (in case it got requested)
			if true == add {
				// copy the relation into the return map
				// and upcount the int
				mapRet[cnt] = s.deepCopyRelation(relation)
				cnt++
			}
		}
	}
	s.RelationStorageMutex.Unlock()
	return mapRet, nil
}

func (s *Storage) GetChildRelationsBySourceTypeAndSourceIdUnsafe(Type int, id int, context string) (map[int]types.StorageRelation, error) {
	// initialice the return map
	var mapRet = make(map[int]types.StorageRelation)
	// set counter for the loop
	var cnt = 0
	// copy the pool we have to search in
	// to prevent crashes on RW concurrency
	// we lock the RelationStorage mutex with
	// fitting Type. this allows us to proceed
	// faster since we just block to copy instead
	// of blocking for the whole process
	var pool = s.RelationStorage[Type][id]
	// for each possible targtType
	for _, targetTypeMap := range pool {
		// for each possible targetId per targetType
		for _, relation := range targetTypeMap {
			// context handling , default add
			add := true
			if "" != context && context != relation.Context {
				add = false
			}
			// if context is fine too (in case it got requested)
			if true == add {
				// copy the relation into the return map
				// and upcount the int
				mapRet[cnt] = s.deepCopyRelation(relation)
				cnt++
			}
		}
	}
	return mapRet, nil
}

func (s *Storage) GetParentEntitiesByTargetTypeAndTargetIdAndSourceType(targetType int, targetID int, sourceType int, context string) map[int]types.StorageEntity {
	// initialice the return map
	var mapRet = make(map[int]types.StorageEntity)

	// set counter for the loop
	var cnt = 0

	// we lock the RelationStorage and EntityStorage
	// mutex with.  this allows us to proceed
	// faster since we just block to copy instead
	// of blocking for the whole process
	s.EntityStorageMutex.RLock()
	s.RelationStorageMutex.RLock()
	for sourceID, _ := range s.RelationRStorage[targetType][targetID][sourceType] {
		entity := s.EntityStorage[sourceType][sourceID]
		add := true
		if "" != context && context != entity.Context {
			add = false
		}
		// copy the relation into the return map
		// and upcount the int
		if true == add {
			mapRet[cnt] = s.deepCopyEntity(s.EntityStorage[sourceType][sourceID])
			cnt++
		}

	}
	s.RelationStorageMutex.RUnlock()
	s.EntityStorageMutex.RUnlock()

	return mapRet
}

func (s *Storage) GetParentEntitiesByTargetTypeAndTargetIdAndSourceTypeUnsafe(targetType int, targetID int, sourceType int, context string) map[int]types.StorageEntity {
	// initialice the return map
	var mapRet = make(map[int]types.StorageEntity)

	// set counter for the loop
	var cnt = 0

	// we lock the RelationStorage and EntityStorage
	// mutex with.  this allows us to proceed
	// faster since we just block to copy instead
	// of blocking for the whole process
	for sourceID, _ := range s.RelationRStorage[targetType][targetID][sourceType] {
		entity := s.EntityStorage[sourceType][sourceID]
		add := true
		if "" != context && context != entity.Context {
			add = false
		}
		// copy the relation into the return map
		// and upcount the int
		if true == add {
			mapRet[cnt] = s.deepCopyEntity(s.EntityStorage[sourceType][sourceID])
			cnt++
		}

	}

	return mapRet
}

func (s *Storage) GetParentRelationsByTargetTypeAndTargetId(targetType int, targetID int, context string) (map[int]types.StorageRelation, error) {
	// initialice the return map
	var mapRet = make(map[int]types.StorageRelation)

	// set counter for the loop
	var cnt = 0

	// copy the pool we have to search in
	// to prevent crashes on RW concurrency
	// we lock the RelationStorage mutex with
	// fitting Type. this allows us to proceed
	// faster since we just block to copy instead
	// of blocking for the whole process
	s.RelationStorageMutex.RLock()
	var pool = s.RelationRStorage[targetType][targetID]
	// for each possible targtType
	for sourceTypeID, targetTypeMap := range pool {
		// for each possible targetId per targetType
		for sourceRelationID, _ := range targetTypeMap {
			// context handling, default is adding
			add := true
			if "" != context && context != s.RelationStorage[sourceTypeID][sourceRelationID][targetType][targetID].Context {
				add = false
			}
			// copy the relation into the return map
			// and upcount the int
			if true == add {
				mapRet[cnt] = s.deepCopyRelation(s.RelationStorage[sourceTypeID][sourceRelationID][targetType][targetID])
				cnt++
			}
		}
	}
	s.RelationStorageMutex.RUnlock()

	return mapRet, nil
}

func (s *Storage) GetParentRelationsByTargetTypeAndTargetIdUnsafe(targetType int, targetID int, context string) (map[int]types.StorageRelation, error) {
	// initialice the return map
	var mapRet = make(map[int]types.StorageRelation)

	// set counter for the loop
	var cnt = 0

	// copy the pool we have to search in
	// to prevent crashes on RW concurrency
	// we lock the RelationStorage mutex with
	// fitting Type. this allows us to proceed
	// faster since we just block to copy instead
	// of blocking for the whole process
	var pool = s.RelationRStorage[targetType][targetID]
	// for each possible targtType
	for sourceTypeID, targetTypeMap := range pool {
		// for each possible targetId per targetType
		for sourceRelationID, _ := range targetTypeMap {
			// context handling, default is adding
			add := true
			if "" != context && context != s.RelationStorage[sourceTypeID][sourceRelationID][targetType][targetID].Context {
				add = false
			}
			// copy the relation into the return map
			// and upcount the int
			if true == add {
				mapRet[cnt] = s.deepCopyRelation(s.RelationStorage[sourceTypeID][sourceRelationID][targetType][targetID])
				cnt++
			}
		}
	}
	return mapRet, nil
}

func (s *Storage) GetEntityTypes() map[int]string {
	// prepare the return array
	types := make(map[int]string)

	// now we lock the storage
	s.EntityTypeMutex.RLock()
	for id, Type := range s.EntityTypes {
		types[id] = Type
	}

	// unlock the mutex and return
	s.EntityTypeMutex.RUnlock()
	return types
}

func (s *Storage) GetEntityTypesUnsafe() map[int]string {
	// prepare the return array
	types := make(map[int]string)

	// now we lock the storage
	for id, Type := range s.EntityTypes {
		types[id] = Type
	}

	// unlock the mutex and return
	return types
}

func (s *Storage) GetEntityRTypes() map[string]int {
	// prepare the return array
	types := make(map[string]int)

	// now we lock the storage
	s.EntityTypeMutex.RLock()
	for Type, id := range s.EntityRTypes {
		types[Type] = id
	}

	// unlock the mutex and return
	s.EntityTypeMutex.RUnlock()
	return types
}

func (s *Storage) GetEntityRTypesUnsafe() map[string]int {
	// prepare the return array
	types := make(map[string]int)

	for Type, id := range s.EntityRTypes {
		types[Type] = id
	}

	return types
}

func (s *Storage) TypeExists(strType string) bool {
	s.EntityTypeMutex.RLock()
	// lets check if this Type exists
	if _, ok := s.EntityRTypes[strType]; ok {
		// it does lets return it
		s.EntityTypeMutex.RUnlock()
		return true
	}

	s.EntityTypeMutex.RUnlock()
	return false
}

func (s *Storage) TypeExistsUnsafe(strType string) bool {
	// lets check if this Type exists
	if _, ok := s.EntityRTypes[strType]; ok {
		// it does lets return it
		return true
	}

	return false
}

func (s *Storage) EntityExists(Type int, id int) bool {
	s.EntityStorageMutex.RLock()
	// lets check if this Type exists
	if _, ok := s.EntityStorage[Type][id]; ok {
		// it does lets return it
		s.EntityStorageMutex.RUnlock()
		return true
	}

	s.EntityStorageMutex.RUnlock()
	return false
}

func (s *Storage) EntityExistsUnsafe(Type int, id int) bool {
	// lets check if this Type exists
	if _, ok := s.EntityStorage[Type][id]; ok {
		// it does lets return it
		return true
	}

	return false
}

func (s *Storage) TypeIdExists(id int) bool {
	s.EntityTypeMutex.RLock()
	// lets check if this Type exists
	if _, ok := s.EntityTypes[id]; ok {
		// it does lets return it
		s.EntityTypeMutex.RUnlock()
		return true
	}

	s.EntityTypeMutex.RUnlock()
	return false
}

func (s *Storage) TypeIdExistsUnsafe(id int) bool {
	// lets check if this Type exists
	if _, ok := s.EntityTypes[id]; ok {
		// it does lets return it
		return true
	}

	return false
}

func (s *Storage) GetTypeIdByString(strType string) (int, error) {
	s.EntityTypeMutex.RLock()
	// lets check if this Type exists
	if id, ok := s.EntityRTypes[strType]; ok {
		// it does lets return it
		s.EntityTypeMutex.RUnlock()
		return id, nil
	}

	s.EntityTypeMutex.RUnlock()
	return -1, errors.New("Entity Type string does not exist")
}

func (s *Storage) GetTypeIdByStringUnsafe(strType string) (int, error) {
	// lets check if this Type exists
	if id, ok := s.EntityRTypes[strType]; ok {
		// it does lets return it
		return id, nil
	}

	return -1, errors.New("Entity Type string does not exist")
}

func (s *Storage) GetTypeStringById(intType int) (string, error) {
	s.EntityTypeMutex.RLock()
	// lets check if this Type exists
	if strType, ok := s.EntityTypes[intType]; ok {
		// it does lets return it
		s.EntityTypeMutex.RUnlock()
		return strType, nil
	}

	s.EntityTypeMutex.RUnlock()
	return "", errors.New("Entity Type ID does not exist")
}

func (s *Storage) GetTypeStringByIdUnsafe(intType int) (string, error) {
	// lets check if this Type exists
	if strType, ok := s.EntityTypes[intType]; ok {
		// it does lets return it
		return strType, nil
	}

	return "", errors.New("Entity Type ID does not exist")
}

func (s *Storage) GetEntityAmount() int {
	amount := 0
	s.EntityStorageMutex.RLock()
	for key, _ := range s.EntityStorage {
		amount += len(s.EntityStorage[key])
	}
	s.EntityStorageMutex.RUnlock()
	return amount
}

func (s *Storage) GetEntityAmountByType(intType int) (int, error) {
	s.EntityStorageMutex.RLock()
	// lets check if this Type exists
	if _, ok := s.EntityStorage[intType]; ok {
		// it does lets return
		amount := len(s.EntityStorage[intType])
		s.EntityStorageMutex.RUnlock()
		return amount, nil
	}

	s.EntityStorageMutex.RUnlock()
	return -1, errors.New("Entity Type does not exist")
}

func (s *Storage) MapTransportData(data transport.TransportEntity) transport.TransportEntity {
	// first we lock all the storages
	s.EntityTypeMutex.Lock()
	s.EntityStorageMutex.Lock()
	s.RelationStorageMutex.Lock()

	// lets start recursive mapping of the data
	newID := s.mapRecursive(data, -1, -1, DIRECTION_NONE)

	// now we unlock all the mutexes again
	s.EntityTypeMutex.Unlock()
	s.EntityStorageMutex.Unlock()
	s.RelationStorageMutex.Unlock()

	// we got it done lets wrap our data in an transport entity object
	ret := transport.TransportEntity{
		ID:         newID,
		Type:       data.Type,
		Value:      data.Value,
		Properties: data.Properties,
		Context:    data.Context,
		Version:    1,
	}

	return ret
}

func (s *Storage) mapRecursive(entity transport.TransportEntity, relatedType int, relatedID int, direction int) int {
	// first we get the right TypeID
	var TypeID int
	TypeID, err := s.GetTypeIdByStringUnsafe(entity.Type)
	if nil != err {
		TypeID, _ = s.CreateEntityTypeUnsafe(entity.Type)
	}

	var mapID int
	// lets see if theres an entity ID given, if its -1 the entity doesnt exist and we create it else we assume its a to map entity
	if -1 == entity.ID {
		// now we create the fitting entity
		tmpEntity := types.StorageEntity{
			ID:         -1,
			Type:       TypeID,
			Value:      entity.Value,
			Context:    entity.Context,
			Version:    1,
			Properties: entity.Properties,
		}
		// now we create the entity
		mapID, _ = s.CreateEntityUnsafe(tmpEntity)
	} else if 0 == entity.ID {
		// if entity.ID == 0 its an upsert by Value and Context(if given)
		entities, err := s.GetEntitiesByTypeAndValueUnsafe(entity.Type, entity.Value, "match", entity.Context)

		// if we got an error or no entities returned we gonne create the entity like in ID==0 case | ###todo refactor if this could somehow be done nicer / merged with the case before
		if nil != err || 0 == len(entities) {
			tmpEntity := types.StorageEntity{
				ID:         -1,
				Type:       TypeID,
				Value:      entity.Value,
				Context:    entity.Context,
				Version:    1,
				Properties: entity.Properties,
			}
			// now we create the entity
			mapID, _ = s.CreateEntityUnsafe(tmpEntity)
		} else {
			// we gonne take ID of the first entry of the return entities map
			mapID = entities[0].ID
		}

	} else {
		// it seems we got an already existing entity given so we use this id to map
		mapID = entity.ID
	}

	// lets map the child elements
	if len(entity.ChildRelations) != 0 {
		// there are children lets iteater over
		// the map
		for _, childRelation := range entity.ChildRelations {
			// pas the child entity and the parent coords to
			// create the relation after inserting the entity
			s.mapRecursive(childRelation.Target, TypeID, mapID, DIRECTION_CHILD)
		}
	}
	// than map the parent elements
	if len(entity.ParentRelations) != 0 {
		// there are children lets iteater over
		// the map
		for _, parentRelation := range entity.ParentRelations {
			// pas the child entity and the parent coords to
			// create the relation after inserting the entity
			s.mapRecursive(parentRelation.Target, TypeID, mapID, DIRECTION_PARENT)
		}
	}
	// now lets check if ourparent Type and id
	// are not -1 , if so we need to create
	// a relation
	if relatedType != -1 && relatedID != -1 {
		// lets create the relation to our parent
		if DIRECTION_CHILD == direction {
			// first we make sure the relation doesnt already exist (because we allow mapped existing data inside a to map json)
			if !s.RelationExistsUnsafe(relatedType, relatedID, TypeID, mapID) {
				tmpRelation := types.StorageRelation{
					SourceType: relatedType,
					SourceID:   relatedID,
					TargetType: TypeID,
					TargetID:   mapID,
					Version:    1,
				}
				s.CreateRelationUnsafe(relatedType, relatedID, TypeID, mapID, tmpRelation)
			}
		} else if DIRECTION_PARENT == direction {
			// first we make sure the relation doesnt already exist (because we allow mapped existing data inside a to map json)
			if !s.RelationExistsUnsafe(TypeID, mapID, relatedType, relatedID) {
				// or relation towards the child
				tmpRelation := types.StorageRelation{
					SourceType: TypeID,
					SourceID:   mapID,
					TargetType: relatedType,
					TargetID:   relatedID,
					Version:    1,
				}
				s.CreateRelationUnsafe(TypeID, mapID, relatedType, relatedID, tmpRelation)
			}
		}
	}
	// only the first return is interesting since it
	// returns the most parent id
	return mapID
}

func (s *Storage) GetEntitiesByQueryFilter(
	typePool []string,
	rootFilter cond.Condition, // New parameter
	conditions [][][3]string, // Legacy conditions (used if rootFilter is nil)
	idFilter [][]int, // Legacy (used if rootFilter is nil)
	valueFilter [][]int, // Legacy (used if rootFilter is nil)
	contextFilter [][]int, // Legacy (used if rootFilter is nil)
	propertyList []map[string][]int, // Legacy (used if rootFilter is nil)
	returnDataFlag bool,
) (
	[]transport.TransportEntity,
	[][2]int,
	int,
) {

	// check the pools given
	typeList := []int{}
	for _, eType := range typePool {
		if val, ok := s.EntityRTypes[eType]; ok {
			typeList = append(typeList, val)
		}
	}

	// do we have any types in pool left?
	if 0 == len(typeList) {
		return []transport.TransportEntity{}, nil, 0
	}

	// prepare results
	var resultEntities []transport.TransportEntity
	var resultAddresses [][2]int

	// if we get here we got some valid types in our typelist,
	// so lets walk through the pools and apply our condition groups
	for _, typeID := range typeList {
		// lets walk through this pools entities
		for entityID, entity := range s.EntityStorage[typeID] {
			add := false
			if rootFilter != nil {
				// New complex filter logic
				if s.evaluateCondition(entity, rootFilter) {
					add = true
				}
			} else {
				// Legacy filter logic
				// if there are matchgroups
				if 0 < len(conditions) { // conditions is the legacy [][][3]string
					for conditionGroupKey, conditionGroup := range conditions {
						// first we check if there is an ID filter
						// ### could have a special case for == on
						// id since this can be resolved very fast
						if 0 < len(idFilter[conditionGroupKey]) && !s.matchGroup(idFilter[conditionGroupKey], conditionGroup, strconv.Itoa(entityID)) {
							continue
						}
						// now we value
						if 0 < len(valueFilter[conditionGroupKey]) && !s.matchGroup(valueFilter[conditionGroupKey], conditionGroup, entity.Value) {
							continue
						}
						// than context
						if 0 < len(contextFilter[conditionGroupKey]) && !s.matchGroup(contextFilter[conditionGroupKey], conditionGroup, entity.Context) {
							continue
						}
						// and now the properties
						contGroupLoop := false
						for propertyKey, propertyConditions := range propertyList[conditionGroupKey] {
							if _, ok := entity.Properties[propertyKey]; ok {
								if !s.matchGroup(propertyConditions, conditionGroup, entity.Properties[propertyKey]) {
									contGroupLoop = true // ### refactor this i dont like it a bit but dont see a better way right now
									break
								}
							} else {
								// property does not exist
								contGroupLoop = true
								break
							}
						}
						// ### we broke out of the inner loop means we have to continue the condition loop
						if contGroupLoop {
							continue
						}
						// if we are still in here all the applied filters worked
						add = true
						// if we got here we can break out since the entity has been added
						break
					}
				} else {
					// we got no conditions so basicly just hit on every of this type
					add = true
				}
			}

			// do we need to add this dataset?
			if true == add {
				// and we can add the entity to our resultList
				if returnDataFlag {
					// first we copy the properties
					props := make(map[string]string)
					for key, value := range entity.Properties {
						props[key] = value
					}
					// than we add the ResultEntity itself
					resultEntities = append(resultEntities, transport.TransportEntity{
						Type:            s.EntityTypes[entity.Type],
						ID:              entity.ID,
						Value:           entity.Value,
						Context:         entity.Context,
						Version:         entity.Version,
						Properties:      props,
						ParentRelations: []transport.TransportRelation{},
						ChildRelations:  []transport.TransportRelation{},
					})
				}
				resultAddresses = append(resultAddresses, [2]int{entity.Type, entityID})
			}
		}
	}
	return resultEntities, resultAddresses, len(resultAddresses)
}

func (s *Storage) GetEntitiesByQueryFilterAndSourceAddress(
	typePool []string,
	rootFilter cond.Condition, // New parameter
	conditions [][][3]string, // Legacy conditions
	idFilter [][]int, // Legacy
	valueFilter [][]int, // Legacy
	contextFilter [][]int, // Legacy
	propertyList []map[string][]int, // Legacy
	sourceAddress [2]int,
	direction int,
	returnDataFlag bool,
) (
	[]transport.TransportRelation,
	[][2]int,
	int,
) {

	// check the pools given
	typeList := []int{}
	for _, eType := range typePool {
		if val, ok := s.EntityRTypes[eType]; ok {
			typeList = append(typeList, val)
		}
	}

	// do we have any types in pool left?
	if 0 == len(typeList) {
		return nil, nil, 0
	}

	// prepare results
	var resultEntities []transport.TransportRelation
	var resultAddresses [][2]int

	// based on the possible relations
	relPool := make(map[int][]int)
	for _, typeID := range typeList {
		// 1 -> towards children
		if 1 == direction {
			relPool[typeID] = s.getRelationTargetIDsBySourceAddressAndTargetType(sourceAddress[0], sourceAddress[1], typeID)
		} else {
			// else is -1 -> towards parents
			relPool[typeID] = s.getRRelationTargetIDsBySourceAddressAndTargetType(sourceAddress[0], sourceAddress[1], typeID)
		}
	}

	// now we know which IDs we have to check, so lets iterate through them
	for targetType, targetIDlist := range relPool {
		for _, targetID := range targetIDlist {
			add := false
			entity := s.EntityStorage[targetType][targetID]

			if rootFilter != nil {
				// New complex filter logic
				if s.evaluateCondition(entity, rootFilter) {
					add = true
				}
			} else {
				// Legacy filter logic
				if 0 < len(conditions) { // conditions is the legacy [][][3]string
					for conditionGroupKey, conditionGroup := range conditions {
						// first we check if there is an ID filter
						// ### could have a special case for == on
						// id since this can be resolved very fast
						if !s.matchGroup(idFilter[conditionGroupKey], conditionGroup, strconv.Itoa(targetID)) {
							continue
						}
						// now we value
						if !s.matchGroup(valueFilter[conditionGroupKey], conditionGroup, entity.Value) {
							continue
						}
						// than context
						if !s.matchGroup(contextFilter[conditionGroupKey], conditionGroup, entity.Context) {
							continue
						}
						// and now the properties
						contGroupLoop := false
						for propertyKey, propertyConditions := range propertyList[conditionGroupKey] {
							if _, ok := entity.Properties[propertyKey]; ok {
								if !s.matchGroup(propertyConditions, conditionGroup, entity.Properties[propertyKey]) {
									contGroupLoop = true // ### refactor this i dont like it a bit but dont see a better way right now
									break
								}
							} else {
								// property does not exist
								contGroupLoop = true
								break
							}
						}
						// ### we broke out of the inner loop means we have to continue the condition loop
						if contGroupLoop {
							continue
						}
						// if we are still in here all the applied filters worked
						// and we can add the entity to our resultList
						add = true
						// if we got here we can break out since the entity has been added
						break
					}
				} else {
					// we have no conditions so we add it anyway
					add = true
				}
			}

			// if we add the data
			if true == add {
				if returnDataFlag {
					// first we copy the properties
					props := make(map[string]string)
					for key, value := range entity.Properties {
						props[key] = value
					}
					// than we add the ResultEntity itself
					resultEntities = append(resultEntities, transport.TransportRelation{
						Context:    s.getRelationContextByAddressAndDirection(sourceAddress[0], sourceAddress[1], targetType, targetID, direction),
						Properties: s.getRelationPropertiesByAddressAndDirection(sourceAddress[0], sourceAddress[1], targetType, targetID, direction),
						Target: transport.TransportEntity{
							Type:            s.EntityTypes[entity.Type],
							ID:              entity.ID,
							Value:           entity.Value,
							Context:         entity.Context,
							Version:         entity.Version,
							Properties:      props,
							ParentRelations: []transport.TransportRelation{},
							ChildRelations:  []transport.TransportRelation{},
						},
					})
				}
				resultAddresses = append(resultAddresses, [2]int{entity.Type, targetID})
			}
		}
	}

	return resultEntities, resultAddresses, len(resultAddresses)
}

func (s *Storage) BatchUpdateAddressList(addressList [][2]int, values map[string]string) {
	for _, address := range addressList {
		entity, _ := s.GetEntityByPathUnsafe(address[0], address[1], "")
		for key, value := range values {
			switch key {
			case "Value":
				entity.Value = value
			case "Context":
				entity.Context = value
			default:
				if -1 != strings.Index(key, "Properties") {
					// ### we nmeed to prepare the map here if it doesnt exist
					entity.Properties[key[11:]] = value
				}
			}
		}
		// ### handle errors
		s.UpdateEntityUnsafe(entity)
	}
}

func (s *Storage) BatchDeleteAddressList(addressList [][2]int) {
	for _, address := range addressList {
		s.DeleteEntityUnsafe(address[0], address[1])
	}
}

func (s *Storage) LinkAddressLists(from [][2]int, to [][2]int, relCtx string, relProps map[string]string) int {
	linkedAmount := 0
	for _, singleFrom := range from {
		for _, singleTo := range to {
			// do we already have a relation between those too? if not we create it
			// If a relation already exists, we do not update its context or properties with this call.
			// Link is for creating new relations.
			if !s.RelationExistsUnsafe(singleFrom[0], singleFrom[1], singleTo[0], singleTo[1]) {
				// Deep copy properties to avoid shared map issues if relProps is reused
				propsCopy := make(map[string]string)
				if relProps != nil {
					for k, v := range relProps {
						propsCopy[k] = v
					}
				}
				s.CreateRelationUnsafe(singleFrom[0], singleFrom[1], singleTo[0], singleTo[1], types.StorageRelation{
					SourceType: singleFrom[0],
					SourceID:   singleFrom[1],
					TargetType: singleTo[0],
					TargetID:   singleTo[1],
					Context:    relCtx,
					Properties: propsCopy,
				})
				linkedAmount++
			}
		}
	}
	return linkedAmount
}

func (s *Storage) TraverseEnrich(entity *transport.TransportEntity, direction int, depth int) {
	if 1 > depth {
		// we reached max depth nuttin to do here
		return
	}
	depth--

	// retrive entity type map for further lookups
	entityTypeMap := s.GetEntityTypesUnsafe()
	currEntityTypeId, err := s.GetTypeIdByStringUnsafe(entity.Type)
	if nil != err {
		// source type entity type does not exist , can this even ever occure? Oo i dont know
		// trying to solve this error that should die alot earlier but just to be sure ###todo review
		// archivist.Error("This should be impossible to hit - run you fools")
		return
	}

	// collect adresses we know and we want to hit
	knownEntities := make(map[string]int)
	var related map[int]types.StorageRelation
	var iterator *[]transport.TransportRelation
	if DIRECTION_CHILD == direction {
		iterator = &(entity.ChildRelations)
		related, _ = s.GetChildRelationsBySourceTypeAndSourceIdUnsafe(currEntityTypeId, entity.ID, "")
	} else {
		iterator = &(entity.ParentRelations)
		related, _ = s.GetParentRelationsByTargetTypeAndTargetIdUnsafe(currEntityTypeId, entity.ID, "")
	}
	if 0 < len(*iterator) {
		for id, val := range *iterator {
			address := val.TargetType + ":" + strconv.Itoa(val.Target.ID)
			knownEntities[address] = id
		}
	}
	// now we go through all related datasets
	for _, rel := range related {
		// build the address to check if we already know it
		var relType int
		var relID int
		if DIRECTION_CHILD == direction {
			relType = rel.TargetType
			relID = rel.TargetID
		} else {
			relType = rel.SourceType
			relID = rel.SourceID
		}

		var iteratorIndex int
		if id, ok := knownEntities[entityTypeMap[relType]+":"+strconv.Itoa(relID)]; ok {
			// its an already known element
			iteratorIndex = id
		} else {
			// we ignore the error since it should be tecnicly impossible to occure.
			// we a re moving in a full locked storage and try to access an entity
			// whos address we just got from the storage. if this error would occure
			// there would be a major inconsistency in the storage itself ###todo review
			newEntity, _ := s.GetEntityByPathUnsafe(relType, relID, "")

			*iterator = append(*iterator, transport.TransportRelation{
				Target: transport.TransportEntity{
					Type:       entityTypeMap[newEntity.Type],
					ID:         newEntity.ID,
					Value:      newEntity.Value,
					Context:    newEntity.Context,
					Version:    newEntity.Version,
					Properties: newEntity.Properties,
				},
			})
			iteratorIndex = len(*iterator) - 1
		}
		s.TraverseEnrich(&((*iterator)[iteratorIndex].Target), direction, depth)
	}
}

// - - - - - - - - - - - - - - - - - - - - - - - - - -
// + + + + + + + + + +  PRIVATE  + + + + + + + + + + +
// - - - - - - - - - - - - - - - - - - - - - - - - - -
func (s *Storage) getRelationContextByAddressAndDirection(sourceType int, sourceID int, targetType int, targetID int, direction int) string {
	if 1 == direction {
		return s.RelationStorage[sourceType][sourceID][targetType][targetID].Context
	} else {
		return s.RelationStorage[targetType][targetID][sourceType][sourceID].Context
	}
}
func (s *Storage) getRelationPropertiesByAddressAndDirection(sourceType int, sourceID int, targetType int, targetID int, direction int) map[string]string {
	// relations need to be copied the current way can lead to thread collision ###todo
	if 1 == direction {
		return s.RelationStorage[sourceType][sourceID][targetType][targetID].Properties
	} else {
		return s.RelationStorage[targetType][targetID][sourceType][sourceID].Properties
	}
}

func (s *Storage) getRelationTargetIDsBySourceAddressAndTargetType(sourceType int, sourceID int, targetType int) []int {
	ret := make([]int, len(s.RelationStorage[sourceType][sourceID][targetType]))
	i := 0
	for key, _ := range s.RelationStorage[sourceType][sourceID][targetType] {
		ret[i] = key
		i++
	}
	return ret
}

func (s *Storage) getRRelationTargetIDsBySourceAddressAndTargetType(sourceType int, sourceID int, targetType int) []int {
	ret := make([]int, len(s.RelationRStorage[sourceType][sourceID][targetType]))
	i := 0
	for key, _ := range s.RelationRStorage[sourceType][sourceID][targetType] {
		ret[i] = key
		i++
	}
	return ret
}

func (s *Storage) matchGroup(filterGroup []int, conditions [][3]string, test string) bool {
	for _, filterGroupID := range filterGroup {
		if !s.Match(test, conditions[filterGroupID][1], conditions[filterGroupID][2]) {
			return false
		}
	}
	return true
}

// Match is an exported version of the internal match logic.
func (s *Storage) Match(alpha string, operator string, beta string) bool {
	switch operator {
	case "==":
		if alpha == beta {
			return true
		}
	case "!=":
		if alpha != beta {
			return true
		}
	case "prefix":
		// starts with
		if strings.HasPrefix(alpha, beta) {
			return true
		}
	case "suffix":
		// ends with
		if strings.HasSuffix(alpha, beta) {
			return true
		}
	case "contain":
		// string contains string
		if strings.Contains(alpha, beta) {
			return true
		}
	case ">":
		alphaInt, err := strconv.Atoi(alpha)
		if nil != err {
			return false
		}
		betaInt, err := strconv.Atoi(beta)
		if nil != err {
			return false
		}
		if alphaInt > betaInt {
			return true
		}
	case ">=":
		alphaInt, err := strconv.Atoi(alpha)
		if nil != err {
			return false
		}
		betaInt, err := strconv.Atoi(beta)
		if nil != err {
			return false
		}
		if alphaInt >= betaInt {
			return true
		}
	case "<":
		alphaInt, err := strconv.Atoi(alpha)
		if nil != err {
			return false
		}
		betaInt, err := strconv.Atoi(beta)
		if nil != err {
			return false
		}
		if alphaInt < betaInt {
			return true
		}
	case "<=":
		alphaInt, err := strconv.Atoi(alpha)
		if nil != err {
			return false
		}
		betaInt, err := strconv.Atoi(beta)
		if nil != err {
			return false
		}
		if alphaInt <= betaInt {
			return true
		}
	case "in":
		list := strings.Split(beta, ",")
		for _, value := range list {
			if alpha == value {
				return true
			}
		}
	}
	return false
}

func (s *Storage) deepCopyEntity(entity types.StorageEntity) types.StorageEntity {
	// first we copy the base values
	newEntity := types.StorageEntity{
		Type:    entity.Type,
		ID:      entity.ID,
		Value:   entity.Value,
		Context: entity.Context,
		Version: entity.Version,
	}

	// creat the base map ##todo check later if we can spare this out
	newEntity.Properties = make(map[string]string)

	// now we check for the properties map
	if nil != entity.Properties && 0 < len(entity.Properties) {
		for key, value := range entity.Properties {
			newEntity.Properties[key] = value
		}
	}

	return newEntity
}

func (s *Storage) deepCopyRelation(relation types.StorageRelation) types.StorageRelation {
	// first we copy the base values
	newRelation := types.StorageRelation{
		SourceType: relation.SourceType,
		SourceID:   relation.SourceID,
		TargetType: relation.TargetType,
		TargetID:   relation.TargetID,
		Context:    relation.Context,
		Version:    relation.Version,
	}

	// creat the base map ##todo check later if we can spare this out
	newRelation.Properties = make(map[string]string)

	// now we check for the properties map
	if nil != relation.Properties && 0 < len(relation.Properties) {
		for key, value := range relation.Properties {
			newRelation.Properties[key] = value
		}
	}

	return newRelation
}

// evaluateCondition recursively evaluates a complex condition tree against a given entity.
func (s *Storage) evaluateCondition(entity types.StorageEntity, condition cond.Condition) bool {
	if condition == nil {
		return true // Or handle as an error/false depending on desired behavior for nil conditions
	}

	var result bool

	switch c := condition.(type) {
	case *cond.MatchCondition:
		var fieldValue string
		switch c.Field {
		case "ID":
			fieldValue = strconv.Itoa(entity.ID)
		case "Value":
			fieldValue = entity.Value
		case "Context":
			fieldValue = entity.Context
		default:
			if strings.HasPrefix(c.Field, "Properties.") {
				propKey := strings.TrimPrefix(c.Field, "Properties.")
				if val, ok := entity.Properties[propKey]; ok {
					fieldValue = val
				} else {
					// Property does not exist on entity, so it cannot match
					result = false
					// Apply negation if MatchCondition itself is negated
					if c.IsNegated() {
						return !result
					}
					return result
				}
			} else {
				// Unknown field
				return false // Or handle error
			}
		}
		result = s.Match(fieldValue, c.Operator, c.Value)

	case *cond.ConditionGroup:
		if c.Type == cond.OpAnd {
			result = true // Assume true until an operand is false
			for _, operand := range c.Operands {
				if !s.evaluateCondition(entity, operand) {
					result = false
					break
				}
			}
		} else if c.Type == cond.OpOr {
			result = false // Assume false until an operand is true
			for _, operand := range c.Operands {
				if s.evaluateCondition(entity, operand) {
					result = true
					break
				}
			}
		}
	default:
		// Unknown condition type
		return false // Or handle error
	}

	// Apply negation if the condition itself (Match or Group) is negated
	if condition.IsNegated() {
		return !result
	}
	return result
}
