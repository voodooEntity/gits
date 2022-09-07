package gits

// handle all the imports
import (
	"errors"
	"github.com/voodooEntity/archivist"
	"github.com/voodooEntity/gits/src/persistence"
	"github.com/voodooEntity/gits/src/transport"
	"github.com/voodooEntity/gits/src/types"
	"regexp"
	"strconv"
	"strings"
	"sync"
)

// - - - - - - - - - - - - - - - - - - - - - - - - - -
// entity storage map            [Type] [ID]
var EntityStorage = make(map[int]map[int]types.StorageEntity)

// entity storage master mutex
var EntityStorageMutex = &sync.RWMutex{}

// - - - - - - - - - - - - - - - - - - - - - - - - - -
// entity storage id max         [Type]
var EntityIDMax = make(map[int]int)

// master mutexd for EntityIdMax
var EntityIDMaxMutex = &sync.RWMutex{}

// - - - - - - - - - - - - - - - - - - - - - - - - - -
// maps to translate Types to their INT and reverse
var EntityTypes = make(map[int]string)
var EntityRTypes = make(map[string]int)

// and a fitting max ID
var EntityTypeIDMax int = 0

// entity Type mutex (for adding and deleting Type types)
var EntityTypeMutex = &sync.RWMutex{}

// - - - - - - - - - - - - - - - - - - - - - - - - - -
// s prefix = source
// t prefix = target
// relation storage map             [sType][sId]   [tType][tId]
var RelationStorage = make(map[int]map[int]map[int]map[int]types.StorageRelation)

// and relation reverse storage map
// (for faster queries)              [tType][Tid]   [sType][sId]
var RelationRStorage = make(map[int]map[int]map[int]map[int]bool)

// relation storage master mutex
var RelationStorageMutex = &sync.RWMutex{}

// direction constants
const (
	DIRECTION_NONE   = -1
	DIRECTION_PARENT = 0
	DIRECTION_CHILD  = 1
)

// - - - - - - - - - - - - - - - - - - - - - - - - - -
// + + + + + + + + + +  PUBLIC  + + + + + + + + + + +
// - - - - - - - - - - - - - - - - - - - - - - - - - -

// - - - - - - - - - - - - - - - - - - - - - - - - - -
// init/construct function for storage package
func Init(persistenceCfg types.PersistenceConfig) {
	// check for persistence.go
	if true == persistenceCfg.Active {
		// init the persistence on need
		importChan := persistence.Init(persistenceCfg)
		if nil != importChan {
			handleImport(importChan)
		}
	}
}

// - - - - - - - - - - - - - - - - - - - - - - - - - -
// Create an entity Type
func CreateEntityType(name string) (int, error) {
	// first of allw e lock
	EntityTypeMutex.Lock()

	// lets check if the Type allready exists
	// if it does we just return the ID
	if id, ok := EntityRTypes[name]; ok {
		// dont forget to unlock
		EntityTypeMutex.Unlock()
		return id, nil
	}

	// ok entity doesnt exist yet, lets
	// upcount our ID Max and copy it
	// into another variable so we can be sure
	// between unlock of the ressource and return
	// it doesnt get upcounted
	EntityTypeIDMax++
	var newID = EntityTypeIDMax

	// finally create the new Type in our
	// EntityTypes index and reverse index
	EntityTypes[newID] = name
	EntityRTypes[name] = newID

	// and create mutex for EntityStorage Type+type
	EntityStorageMutex.Lock()

	// now we prepare the submaps in the entity
	// storage itseöf....
	EntityStorage[newID] = make(map[int]types.StorageEntity)

	// set the maxID for the new
	// Type type
	EntityIDMax[newID] = 0
	EntityStorageMutex.Unlock()

	// create the base maps in relation storage
	RelationStorageMutex.Lock()
	RelationStorage[newID] = make(map[int]map[int]map[int]types.StorageRelation)
	RelationRStorage[newID] = make(map[int]map[int]map[int]bool)
	RelationStorageMutex.Unlock()

	// and create the basic submaps for
	// the relation storage
	// now we unlock the mutex
	// and return the new id
	// - - - - - - - - - - - - - - - - -
	// persistence.go handling
	if true == persistence.PersistenceFlag {
		persistence.PersistenceChan <- types.PersistencePayload{
			Type:        "EntityType",
			EntityTypes: EntityTypes,
		}
	}
	// - - - - - - - - - - - - - - - - -
	EntityTypeMutex.Unlock()
	return newID, nil
}

func CreateEntityTypeUnsafe(name string) (int, error) {
	// lets check if the Type allready exists
	// if it does we just return the ID
	if id, ok := EntityRTypes[name]; ok {
		// dont forget to unlock
		return id, nil
	}

	// ok entity doesnt exist yet, lets
	// upcount our ID Max and copy it
	// into another variable so we can be sure
	// between unlock of the ressource and return
	// it doesnt get upcounted
	EntityTypeIDMax++
	var newID = EntityTypeIDMax

	// finally create the new Type in our
	// EntityTypes index and reverse index
	EntityTypes[newID] = name
	EntityRTypes[name] = newID

	// now we prepare the submaps in the entity
	// storage itseöf....
	EntityStorage[newID] = make(map[int]types.StorageEntity)

	// set the maxID for the new
	// Type type
	EntityIDMax[newID] = 0

	// create the base maps in relation storage
	RelationStorage[newID] = make(map[int]map[int]map[int]types.StorageRelation)
	RelationRStorage[newID] = make(map[int]map[int]map[int]bool)

	// and create the basic submaps for
	// the relation storage
	// now we unlock the mutex
	// and return the new id
	// - - - - - - - - - - - - - - - - -
	// persistence.go handling
	if true == persistence.PersistenceFlag {
		persistence.PersistenceChan <- types.PersistencePayload{
			Type:        "EntityType",
			EntityTypes: EntityTypes,
		}
	}
	// - - - - - - - - - - - - - - - - -
	return newID, nil
}

func CreateEntity(entity types.StorageEntity) (int, error) {
	//types.PrintMemUsage()
	// first we lock the entity Type mutex
	// to make sure while we check for the
	// existence it doesnt get deletet, this
	// may sound like a very rare upcoming case,
	//but better be safe than sorry
	EntityTypeMutex.RLock()

	// now
	if _, ok := EntityTypes[entity.Type]; !ok {
		// the Type doest exist, lets unlock
		// the Type mutex and return -1 for fail0r
		EntityTypeMutex.RUnlock()
		return -1, errors.New("CreateEntity.Entity Type not existing")
	}
	// the Type seems to exist, now lets lock the
	// storage mutex before Unlocking the Entity
	// Type mutex to prevent the Type beeing
	// deleted before we start locking (small
	// timing still possible )
	EntityTypeMutex.RUnlock()

	// upcount our ID Max and copy it
	// into another variable so we can be sure
	// between unlock of the ressource and return
	// it doesnt get upcounted
	// and set the IDMaxMutex on write Lock
	// lets upcount the entity id max fitting to
	//         [Type]
	EntityStorageMutex.Lock()
	EntityIDMax[entity.Type]++
	var newID = EntityIDMax[entity.Type]

	//EntityIDMaxMasterMutex.Lock()
	// and tell the entity its own id
	entity.ID = newID

	// and set the version to 1
	entity.Version = 1

	// - - - - - - - - - - - - - - - - -
	// persistence.go handling
	if true == persistence.PersistenceFlag {
		persistence.PersistenceChan <- types.PersistencePayload{
			Type:   "Entity",
			Method: "Create",
			Entity: entity,
		}
	}
	// - - - - - - - - - - - - - - - - -

	// now we store the entity element
	// in the EntityStorage
	EntityStorage[entity.Type][newID] = entity

	//printMutexActions("CreateEntity.EntityStorageMutex.Unlock");
	EntityStorageMutex.Unlock()

	// create the mutex for our ressource on
	// relation. we have to create the sub maps too
	// golang things....
	RelationStorageMutex.Lock()
	RelationStorage[entity.Type][newID] = make(map[int]map[int]types.StorageRelation)
	RelationRStorage[entity.Type][newID] = make(map[int]map[int]bool)
	RelationStorageMutex.Unlock()

	// since we now stored the entity and created all
	// needed ressources we can unlock
	// the storage ressource and return the ID (or err)
	return newID, nil
}

func CreateEntityUnsafe(entity types.StorageEntity) (int, error) {
	//types.PrintMemUsage()
	// first we lock the entity Type mutex
	// to make sure while we check for the
	// existence it doesnt get deletet, this
	// may sound like a very rare upcoming case,
	//but better be safe than sorry

	// now
	if _, ok := EntityTypes[entity.Type]; !ok {
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
	EntityIDMax[entity.Type]++
	var newID = EntityIDMax[entity.Type]

	//EntityIDMaxMasterMutex.Lock()
	// and tell the entity its own id
	entity.ID = newID

	// and set the version to 1
	entity.Version = 1

	// - - - - - - - - - - - - - - - - -
	// persistence.go handling
	if true == persistence.PersistenceFlag {
		persistence.PersistenceChan <- types.PersistencePayload{
			Type:   "Entity",
			Method: "Create",
			Entity: entity,
		}
	}
	// - - - - - - - - - - - - - - - - -

	// now we store the entity element
	// in the EntityStorage
	EntityStorage[entity.Type][newID] = entity

	// create the mutex for our ressource on
	// relation. we have to create the sub maps too
	// golang things....
	RelationStorage[entity.Type][newID] = make(map[int]map[int]types.StorageRelation)
	RelationRStorage[entity.Type][newID] = make(map[int]map[int]bool)

	// since we now stored the entity and created all
	// needed ressources we can unlock
	// the storage ressource and return the ID (or err)
	return newID, nil
}

// bool return = has a new dataset been created
func CreateEntityUniqueValue(entity types.StorageEntity) (int, bool, error) {
	//types.PrintMemUsage()
	// first we lock the entity Type mutex
	// to make sure while we check for the
	// existence it doesnt get deletet, this
	// may sound like a very rare upcoming case,
	//but better be safe than sorry
	EntityTypeMutex.RLock()

	// now we will cache the stype string due to the
	// special hack implementation of createEntityUniqueValue
	// for what we created the unsafe retrieval version  getEntitiesByTypeAndValueUnsafe()
	// that expects a string instread of the usualy on create neccesary id.
	var stype string
	if val, ok := EntityTypes[entity.Type]; !ok {
		// the Type doest exist, lets unlock
		// the Type mutex and return -1 for fail0r
		EntityTypeMutex.RUnlock()
		return -1, false, errors.New("CreateEntityUniqueValue.Entity Type not existing")
	} else {
		stype = val
	}
	// the Type seems to exist, now lets lock the
	// storage mutex before Unlocking the Entity
	// Type mutex to prevent the Type beeing
	// deleted before we start locking (small
	// timing still possible )
	EntityTypeMutex.RUnlock()

	// since this is the UniqueValue variant
	// we have to lock and make sure the type:value combination
	// doesnt exist. thatfor we call getEntitiesByTypeAndValueUnsafe()
	// which doesnt have any locking implemented and thatfor will be able
	// to see if we can retrieve any entity fitting
	EntityStorageMutex.Lock()
	entities, err := GetEntitiesByTypeAndValueUnsafe(stype, entity.Value, "match", entity.Context)
	if nil != err {
		EntityStorageMutex.Unlock()
		return -1, false, err
	}
	// ### think about update logic since collection properties might change
	if 0 < len(entities) {
		EntityStorageMutex.Unlock()
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
	EntityIDMax[entity.Type]++
	var newID = EntityIDMax[entity.Type]

	//EntityIDMaxMasterMutex.Lock()
	// and tell the entity its own id
	entity.ID = newID

	// and set the version to 1
	entity.Version = 1

	// - - - - - - - - - - - - - - - - -
	// persistance handling
	if true == persistence.PersistenceFlag {
		persistence.PersistenceChan <- types.PersistencePayload{
			Type:   "Entity",
			Method: "Create",
			Entity: entity,
		}
	}
	// - - - - - - - - - - - - - - - - -

	// now we store the entity element
	// in the EntityStorage
	EntityStorage[entity.Type][newID] = entity

	//printMutexActions("CreateEntity.EntityStorageMutex.Unlock");
	EntityStorageMutex.Unlock()

	// create the mutex for our ressource on
	// relation. we have to create the sub maps too
	// golang things....
	RelationStorageMutex.Lock()
	RelationStorage[entity.Type][newID] = make(map[int]map[int]types.StorageRelation)
	RelationRStorage[entity.Type][newID] = make(map[int]map[int]bool)
	RelationStorageMutex.Unlock()

	// since we now stored the entity and created all
	// needed ressources we can unlock
	// the storage ressource and return the ID (or err)
	return newID, true, nil
}

// bool return = has a new dataset been created
func CreateEntityUniqueValueUnsafe(entity types.StorageEntity) (int, bool, error) {
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
	if val, ok := EntityTypes[entity.Type]; !ok {
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
	entities, err := GetEntitiesByTypeAndValueUnsafe(stype, entity.Value, "match", entity.Context)
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
	EntityIDMax[entity.Type]++
	var newID = EntityIDMax[entity.Type]

	//EntityIDMaxMasterMutex.Lock()
	// and tell the entity its own id
	entity.ID = newID

	// and set the version to 1
	entity.Version = 1

	// - - - - - - - - - - - - - - - - -
	// persistance handling
	if true == persistence.PersistenceFlag {
		persistence.PersistenceChan <- types.PersistencePayload{
			Type:   "Entity",
			Method: "Create",
			Entity: entity,
		}
	}
	// - - - - - - - - - - - - - - - - -

	// now we store the entity element
	// in the EntityStorage
	EntityStorage[entity.Type][newID] = entity

	//printMutexActions("CreateEntity.EntityStorageMutex.Unlock");

	// create the mutex for our ressource on
	// relation. we have to create the sub maps too
	// golang things....
	RelationStorage[entity.Type][newID] = make(map[int]map[int]types.StorageRelation)
	RelationRStorage[entity.Type][newID] = make(map[int]map[int]bool)

	// since we now stored the entity and created all
	// needed ressources we can unlock
	// the storage ressource and return the ID (or err)
	return newID, true, nil
}

func GetEntityByPath(Type int, id int, context string) (types.StorageEntity, error) {
	// lets check if entity witrh the given path exists
	EntityStorageMutex.Lock()
	if entity, ok := EntityStorage[Type][id]; ok {
		// if yes we return the entity
		// and nil for error
		if "" == context || entity.Context == context {
			ret := deepCopyEntity(entity)
			EntityStorageMutex.Unlock()
			return ret, nil
		}
	}

	EntityStorageMutex.Unlock()

	// the path seems to transport empty , so
	// we throw an error
	return types.StorageEntity{}, errors.New("Entity on given path does not exist.")
}

func GetEntityByPathUnsafe(Type int, id int, context string) (types.StorageEntity, error) {
	// lets check if entity with the given path exists
	if entity, ok := EntityStorage[Type][id]; ok {
		// if yes we return the entity
		// and nil for error
		if "" == context || entity.Context == context {
			return deepCopyEntity(entity), nil
		}
	}

	// the path seems to transport empty , so
	// we throw an error
	return types.StorageEntity{}, errors.New("Entity on given path does not exist.")
}

func GetEntitiesByType(Type string, context string) (map[int]types.StorageEntity, error) {
	// retrieve the fitting id
	entityTypeID, _ := GetTypeIdByString(Type)

	// lock retrieve und unlock the storage
	mapRet := make(map[int]types.StorageEntity)
	i := 0
	EntityStorageMutex.RLock()
	for _, entity := range EntityStorage[entityTypeID] {
		// preset add with true
		add := true

		// check if context is set , if yes and it doesnt
		// fit we dont add
		if context != "" && entity.Context != context {
			add = false
		}

		// finally if everything is fine we add the dataset
		if add {
			mapRet[i] = deepCopyEntity(entity)
			i++
		}
	}

	// unlock the storage again
	EntityStorageMutex.RUnlock()

	// return the entity map
	return mapRet, nil
}

func GetEntitiesByTypeUnsafe(Type string, context string) (map[int]types.StorageEntity, error) {
	// retrieve the fitting id
	entityTypeID, _ := GetTypeIdByString(Type)

	// lock retrieve und unlock the storage
	mapRet := make(map[int]types.StorageEntity)
	i := 0
	for _, entity := range EntityStorage[entityTypeID] {
		// preset add with true
		add := true

		// check if context is set , if yes and it doesnt
		// fit we dont add
		if context != "" && entity.Context != context {
			add = false
		}

		// finally if everything is fine we add the dataset
		if add {
			mapRet[i] = deepCopyEntity(entity)
			i++
		}
	}

	// return the entity map
	return mapRet, nil
}

func GetEntitiesByValue(value string, mode string, context string) (map[int]types.StorageEntity, error) {
	// lets prepare the return map, counter and regex r
	entities := make(map[int]types.StorageEntity)
	i := 0
	var r *regexp.Regexp
	var err error = nil

	// first we lock the storage
	EntityStorageMutex.RLock()

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
	if 0 < len(EntityStorage) {
		for typeID := range EntityStorage {
			if 0 < len(EntityStorage[typeID]) {
				for _, entity := range EntityStorage[typeID] {
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
								entities[i] = deepCopyEntity(entity)
								i++
							}
						case "prefix":
							// starts with
							if strings.HasPrefix(entity.Value, value) {
								entities[i] = deepCopyEntity(entity)
								i++
							}
						case "suffix":
							// ends with
							if strings.HasSuffix(entity.Value, value) {
								entities[i] = deepCopyEntity(entity)
								i++
							}
						case "contain":
							// string contains string
							if strings.Contains(entity.Value, value) {
								entities[i] = deepCopyEntity(entity)
								i++
							}
						case "regex":
							// string matches regex
							if r.MatchString(entity.Value) {
								entities[i] = deepCopyEntity(entity)
								i++
							}
						}
					}
				}
			}
		}
	}

	// unlock storage again and return
	EntityStorageMutex.RUnlock()
	return entities, nil
}

func GetEntitiesByValueUnsafe(value string, mode string, context string) (map[int]types.StorageEntity, error) {
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
	if 0 < len(EntityStorage) {
		for typeID := range EntityStorage {
			if 0 < len(EntityStorage[typeID]) {
				for _, entity := range EntityStorage[typeID] {
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
								entities[i] = deepCopyEntity(entity)
								i++
							}
						case "prefix":
							// starts with
							if strings.HasPrefix(entity.Value, value) {
								entities[i] = deepCopyEntity(entity)
								i++
							}
						case "suffix":
							// ends with
							if strings.HasSuffix(entity.Value, value) {
								entities[i] = deepCopyEntity(entity)
								i++
							}
						case "contain":
							// string contains string
							if strings.Contains(entity.Value, value) {
								entities[i] = deepCopyEntity(entity)
								i++
							}
						case "regex":
							// string matches regex
							if r.MatchString(entity.Value) {
								entities[i] = deepCopyEntity(entity)
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

func GetEntitiesByTypeAndValue(Type string, value string, mode string, context string) (map[int]types.StorageEntity, error) {
	// lets prepare the return map, counter and regex r
	entities := make(map[int]types.StorageEntity)
	i := 0
	var r *regexp.Regexp
	var err error = nil

	// first we lock the storage
	EntityStorageMutex.RLock()

	// retrieve the fitting id
	entityTypeID, _ := GetTypeIdByString(Type)

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
	if 0 < len(EntityStorage) {
		if 0 < len(EntityStorage[entityTypeID]) {
			for _, entity := range EntityStorage[entityTypeID] {
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
							entities[i] = deepCopyEntity(entity)
							i++
						}
					case "prefix":
						// starts with
						if strings.HasPrefix(entity.Value, value) {
							entities[i] = deepCopyEntity(entity)
							i++
						}
					case "suffix":
						// ends with
						if strings.HasSuffix(entity.Value, value) {
							entities[i] = deepCopyEntity(entity)
							i++
						}
					case "contain":
						// contains
						if strings.Contains(entity.Value, value) {
							entities[i] = deepCopyEntity(entity)
							i++
						}
					case "regex":
						// matches regex
						if r.MatchString(entity.Value) {
							entities[i] = deepCopyEntity(entity)
							i++
						}
					}
				}
			}
		}
	}

	// unlock storage again and return
	EntityStorageMutex.RUnlock()
	return entities, nil
}

func GetEntitiesByTypeAndValueUnsafe(Type string, value string, mode string, context string) (map[int]types.StorageEntity, error) {
	// lets prepare the return map, counter and regex r
	entities := make(map[int]types.StorageEntity)
	i := 0
	var r *regexp.Regexp
	var err error = nil

	// retrieve the fitting id
	entityTypeID, _ := GetTypeIdByString(Type)

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
	if 0 < len(EntityStorage) {
		if 0 < len(EntityStorage[entityTypeID]) {
			for _, entity := range EntityStorage[entityTypeID] {
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
							entities[i] = deepCopyEntity(entity)
							i++
						}
					case "prefix":
						// starts with
						if strings.HasPrefix(entity.Value, value) {
							entities[i] = deepCopyEntity(entity)
							i++
						}
					case "suffix":
						// ends with
						if strings.HasSuffix(entity.Value, value) {
							entities[i] = deepCopyEntity(entity)
							i++
						}
					case "contain":
						// contains
						if strings.Contains(entity.Value, value) {
							entities[i] = deepCopyEntity(entity)
							i++
						}
					case "regex":
						// matches regex
						if r.MatchString(entity.Value) {
							entities[i] = deepCopyEntity(entity)
							i++
						}
					}
				}
			}
		}
	}

	return entities, nil
}

func UpdateEntity(entity types.StorageEntity) error {
	// - - - - - - - - - - - - - - - - -
	// lock the storage for concurrency
	EntityStorageMutex.Lock()
	if check, ok := EntityStorage[entity.Type][entity.ID]; ok {
		// - - - - - - - - - - - - - - - - -
		// lets check if the version is up to date
		if entity.Version != check.Version {
			EntityStorageMutex.Unlock()
			return errors.New("Mismatch of version.")
		}
		entity.Version++

		// - - - - - - - - - - - - - - - - -
		// persistence.go handling
		if true == persistence.PersistenceFlag {
			persistence.PersistenceChan <- types.PersistencePayload{
				Type:   "Entity",
				Method: "Update",
				Entity: entity,
			}
		}
		// - - - - - - - - - - - - - - - - -
		EntityStorage[entity.Type][entity.ID] = entity
		EntityStorageMutex.Unlock()
		return nil
	}

	// unlock the storage and return an error in case we get here
	EntityStorageMutex.Unlock()
	return errors.New("Cant update non existing entity")
}

func UpdateEntityUnsafe(entity types.StorageEntity) error {
	// - - - - - - - - - - - - - - - - -
	// lock the storage for concurrency
	if check, ok := EntityStorage[entity.Type][entity.ID]; ok {
		// - - - - - - - - - - - - - - - - -
		// lets check if the version is up to date
		if entity.Version != check.Version {
			return errors.New("Mismatch of version.")
		}
		entity.Version++

		// - - - - - - - - - - - - - - - - -
		// persistence.go handling
		if true == persistence.PersistenceFlag {
			persistence.PersistenceChan <- types.PersistencePayload{
				Type:   "Entity",
				Method: "Update",
				Entity: entity,
			}
		}
		// - - - - - - - - - - - - - - - - -
		EntityStorage[entity.Type][entity.ID] = entity
		return nil
	}

	// unlock the storage and return an error in case we get here
	return errors.New("Cant update non existing entity")
}

func DeleteEntity(Type int, id int) {
	// we gonne lock the mutex and
	// delete the element
	EntityStorageMutex.Lock()
	// - - - - - - - - - - - - - - - - -
	// persistence.go handling
	if true == persistence.PersistenceFlag {
		persistence.PersistenceChan <- types.PersistencePayload{
			Type:   "Entity",
			Method: "Delete",
			Entity: types.StorageEntity{
				ID:   id,
				Type: Type,
			},
		}
	}
	// - - - - - - - - - - - - - - - - -
	delete(EntityStorage[Type], id)
	EntityStorageMutex.Unlock()
	// now we delete the relations from and to this entity
	// first child
	DeleteChildRelations(Type, id)
	// than parent
	DeleteParentRelations(Type, id)
}

func DeleteEntityUnsafe(Type int, id int) {
	// we gonne lock the mutex and
	// delete the element
	// - - - - - - - - - - - - - - - - -
	// persistence.go handling
	if true == persistence.PersistenceFlag {
		persistence.PersistenceChan <- types.PersistencePayload{
			Type:   "Entity",
			Method: "Delete",
			Entity: types.StorageEntity{
				ID:   id,
				Type: Type,
			},
		}
	}
	// - - - - - - - - - - - - - - - - -
	delete(EntityStorage[Type], id)
	// now we delete the relations from and to this entity
	// first child
	DeleteChildRelations(Type, id)
	// than parent
	DeleteParentRelations(Type, id)
}

func GetRelation(srcType int, srcID int, targetType int, targetID int) (types.StorageRelation, error) {
	// first we lock the relation storage
	RelationStorageMutex.RLock()
	if _, firstOk := RelationStorage[srcType]; firstOk {
		if _, secondOk := RelationStorage[srcType][srcID]; secondOk {
			if _, thirdOk := RelationStorage[srcType][srcID][targetType]; thirdOk {
				if relation, fourthOk := RelationStorage[srcType][srcID][targetType][targetID]; fourthOk {
					RelationStorageMutex.RUnlock()
					return deepCopyRelation(relation), nil
				}
			}
		}
	}
	RelationStorageMutex.RUnlock()
	return types.StorageRelation{}, errors.New("Non existing relation requested")
}

func GetRelationUnsafe(srcType int, srcID int, targetType int, targetID int) (types.StorageRelation, error) {
	// first we lock the relation storage
	if _, firstOk := RelationStorage[srcType]; firstOk {
		if _, secondOk := RelationStorage[srcType][srcID]; secondOk {
			if _, thirdOk := RelationStorage[srcType][srcID][targetType]; thirdOk {
				if relation, fourthOk := RelationStorage[srcType][srcID][targetType][targetID]; fourthOk {
					return deepCopyRelation(relation), nil
				}
			}
		}
	}
	return types.StorageRelation{}, errors.New("Non existing relation requested")
}

// maybe deprecated, check later
func RelationExists(srcType int, srcID int, targetType int, targetID int) bool {
	// first we lock the relation storage
	RelationStorageMutex.RLock()
	if srcTypeMap, firstOk := RelationStorage[srcType]; firstOk {
		if srcIDMap, secondOk := srcTypeMap[srcID]; secondOk {
			if targetTypeMap, thirdOk := srcIDMap[targetType]; thirdOk {
				if _, fourthOk := targetTypeMap[targetID]; fourthOk {
					RelationStorageMutex.RUnlock()
					return true
				}
			}
		}
	}
	RelationStorageMutex.RUnlock()
	return false
}

func RelationExistsUnsafe(srcType int, srcID int, targetType int, targetID int) bool {
	// first we lock the relation storage
	if srcTypeMap, firstOk := RelationStorage[srcType]; firstOk {
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

func DeleteRelationList(relationList map[int]types.StorageRelation) {
	// lets walk through the iterations and delete all
	// corrosponding Relation & RRelation index entries
	if 0 < len(relationList) {
		for _, relation := range relationList {
			DeleteRelation(relation.SourceType, relation.SourceID, relation.TargetType, relation.TargetID)
		}
	}
	return
}

func DeleteRelationListUnsafe(relationList map[int]types.StorageRelation) {
	// lets walk through the iterations and delete all
	// corrosponding Relation & RRelation index entries
	if 0 < len(relationList) {
		for _, relation := range relationList {
			DeleteRelationUnsafe(relation.SourceType, relation.SourceID, relation.TargetType, relation.TargetID)
		}
	}
	return
}

func DeleteRelation(sourceType int, sourceID int, targetType int, targetID int) {
	RelationStorageMutex.Lock()
	// - - - - - - - - - - - - - - - - -
	// persistence.go handling
	if true == persistence.PersistenceFlag {
		persistence.PersistenceChan <- types.PersistencePayload{
			Type:   "Relation",
			Method: "Delete",
			Relation: types.StorageRelation{
				SourceID:   sourceID,
				SourceType: sourceType,
				TargetID:   targetID,
				TargetType: targetType,
			},
		}
	}
	// - - - - - - - - - - - - - - - - -
	delete(RelationStorage[sourceType][sourceID][targetType], targetID)
	delete(RelationRStorage[targetType][targetID][sourceType], sourceID)
	RelationStorageMutex.Unlock()
}

func DeleteRelationUnsafe(sourceType int, sourceID int, targetType int, targetID int) {
	// - - - - - - - - - - - - - - - - -
	// persistence.go handling
	if true == persistence.PersistenceFlag {
		persistence.PersistenceChan <- types.PersistencePayload{
			Type:   "Relation",
			Method: "Delete",
			Relation: types.StorageRelation{
				SourceID:   sourceID,
				SourceType: sourceType,
				TargetID:   targetID,
				TargetType: targetType,
			},
		}
	}
	// - - - - - - - - - - - - - - - - -
	delete(RelationStorage[sourceType][sourceID][targetType], targetID)
	delete(RelationRStorage[targetType][targetID][sourceType], sourceID)
}

func DeleteChildRelations(Type int, id int) error {
	childRelations, err := GetChildRelationsBySourceTypeAndSourceId(Type, id, "")
	if nil != err {
		return err
	}
	DeleteRelationList(childRelations)
	return nil
}

func DeleteChildRelationsUnsafe(Type int, id int) error {
	childRelations, err := GetChildRelationsBySourceTypeAndSourceIdUnsafe(Type, id, "")
	if nil != err {
		return err
	}
	DeleteRelationListUnsafe(childRelations)
	return nil
}

func DeleteParentRelations(Type int, id int) error {
	parentRelations, err := GetParentRelationsByTargetTypeAndTargetId(Type, id, "")
	if nil != err {
		return err
	}
	DeleteRelationList(parentRelations)
	return nil
}

func DeleteParentRelationsUnsafe(Type int, id int) error {
	parentRelations, err := GetParentRelationsByTargetTypeAndTargetId(Type, id, "")
	if nil != err {
		return err
	}
	DeleteRelationListUnsafe(parentRelations)
	return nil
}

func CreateRelation(srcType int, srcID int, targetType int, targetID int, relation types.StorageRelation) (bool, error) {
	// first we Readlock the EntityTypeMutex
	//printMutexActions("CreateRelation.EntityTypeMutex.RLock");
	EntityTypeMutex.RLock()
	// lets make sure the source Type exist
	if _, ok := EntityTypes[srcType]; !ok {
		//printMutexActions("CreateRelation.EntityTypeMutex.RUnlock");
		EntityTypeMutex.RUnlock()
		return false, errors.New("Source Type not existing")
	}
	// and the target Type exists too
	if _, ok := EntityTypes[targetType]; !ok {
		//printMutexActions("CreateRelation.EntityTypeMutex.RUnlock");
		EntityTypeMutex.RUnlock()
		return false, errors.New("Target Type not existing")
	}
	// finally unlock the TypeMutex again if both checks were successfull
	EntityTypeMutex.RUnlock()
	//// - - - - - - - - - - - - - - - - -
	// now we lock the relation mutex
	//printMutexActions("CreateRelation.RelationStorageMutex.Lock");
	RelationStorageMutex.Lock()
	// lets check if their exists a map for our
	// source entity to the target Type if not
	// create it.... golang things...
	if _, ok := RelationStorage[srcType][srcID][targetType]; !ok {
		RelationStorage[srcType][srcID][targetType] = make(map[int]types.StorageRelation)
		// if the map doesnt exist in this direction
		// it wont exist in the other as in reverse
		// map either so we should create it too
		// but we will store a pointer to the other
		// maps Relation instead of the complete
		// relation twice - defunct, refactor later (may create more problems then help)
		//RelationStorage[targetType][targetID][srcType] = make(map[int]Relation)
	}
	// now we prepare the reverse storage if necessary
	if _, ok := RelationRStorage[targetType][targetID][srcType]; !ok {
		RelationRStorage[targetType][targetID][srcType] = make(map[int]bool)
	}
	// set version to 1
	relation.Version = 1
	// now we store the relation
	RelationStorage[srcType][srcID][targetType][targetID] = relation
	// - - - - - - - - - - - - - - - - -
	// persistence.go handling
	if true == persistence.PersistenceFlag {
		persistence.PersistenceChan <- types.PersistencePayload{
			Type:     "Relation",
			Method:   "Create",
			Relation: relation,
		}
	}
	// - - - - - - - - - - - - - - - - -
	// and an entry into the reverse index, its existence
	// allows us to use the coords in the normal index to revtrieve
	// the Relation. We dont create a pointer because golang doesnt
	// allow pointer on submaps in nested maps
	RelationRStorage[targetType][targetID][srcType][srcID] = true
	// we are done now we can unlock the entity Types
	//// - - - - - - - - - - - - - - - -
	//and finally unlock the relation Type and return
	//printMutexActions("CreateRelation.RelationStorageMutex.Unlock");
	RelationStorageMutex.Unlock()
	return true, nil
}

func CreateRelationUnsafe(srcType int, srcID int, targetType int, targetID int, relation types.StorageRelation) (bool, error) {
	// first we Readlock the EntityTypeMutex
	//printMutexActions("CreateRelation.EntityTypeMutex.RLock");
	// lets make sure the source Type exist
	if _, ok := EntityTypes[srcType]; !ok {
		//printMutexActions("CreateRelation.EntityTypeMutex.RUnlock");
		return false, errors.New("Source Type not existing")
	}
	// and the target Type exists too
	if _, ok := EntityTypes[targetType]; !ok {
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
	if _, ok := RelationStorage[srcType][srcID][targetType]; !ok {
		RelationStorage[srcType][srcID][targetType] = make(map[int]types.StorageRelation)
		// if the map doesnt exist in this direction
		// it wont exist in the other as in reverse
		// map either so we should create it too
		// but we will store a pointer to the other
		// maps Relation instead of the complete
		// relation twice - defunct, refactor later (may create more problems then help)
		//RelationStorage[targetType][targetID][srcType] = make(map[int]Relation)
	}
	// now we prepare the reverse storage if necessary
	if _, ok := RelationRStorage[targetType][targetID][srcType]; !ok {
		RelationRStorage[targetType][targetID][srcType] = make(map[int]bool)
	}
	// set version to 1
	relation.Version = 1
	// now we store the relation
	RelationStorage[srcType][srcID][targetType][targetID] = relation
	// - - - - - - - - - - - - - - - - -
	// persistence.go handling
	if true == persistence.PersistenceFlag {
		persistence.PersistenceChan <- types.PersistencePayload{
			Type:     "Relation",
			Method:   "Create",
			Relation: relation,
		}
	}
	// - - - - - - - - - - - - - - - - -
	// and an entry into the reverse index, its existence
	// allows us to use the coords in the normal index to revtrieve
	// the Relation. We dont create a pointer because golang doesnt
	// allow pointer on submaps in nested maps
	RelationRStorage[targetType][targetID][srcType][srcID] = true
	// we are done now we can unlock the entity Types
	//// - - - - - - - - - - - - - - - -
	//and finally unlock the relation Type and return
	//printMutexActions("CreateRelation.RelationStorageMutex.Unlock");
	return true, nil
}

func UpdateRelation(srcType int, srcID int, targetType int, targetID int, relation types.StorageRelation) (types.StorageRelation, error) {
	// first we lock the relation storage
	RelationStorageMutex.Lock()
	if _, firstOk := RelationStorage[srcType]; firstOk {
		if _, secondOk := RelationStorage[srcType][srcID]; secondOk {
			if _, thirdOk := RelationStorage[srcType][srcID][targetType]; thirdOk {
				if rel, fourthOk := RelationStorage[srcType][srcID][targetType][targetID]; fourthOk {
					// check if the version is fine
					if rel.Version != relation.Version {
						RelationStorageMutex.Unlock()
						return types.StorageRelation{}, errors.New("Mismatch of version.")
					}
					rel.Version++

					// - - - - - - - - - - - - - - - - -
					// persistence.go handling
					if true == persistence.PersistenceFlag {
						persistence.PersistenceChan <- types.PersistencePayload{
							Type:     "Relation",
							Method:   "Create",
							Relation: rel,
						}
					}

					// - - - - - - - - - - - - - - - - -
					// update the data itself
					rel.Context = relation.Context
					rel.Properties = relation.Properties
					RelationStorage[srcType][srcID][targetType][targetID] = rel
					RelationStorageMutex.Unlock()
					return relation, nil
				}
			}
		}
	}
	RelationStorageMutex.Unlock()
	return types.StorageRelation{}, errors.New("Cant update non existing relation")
}

func UpdateRelationUnsafe(srcType int, srcID int, targetType int, targetID int, relation types.StorageRelation) (types.StorageRelation, error) {
	// first we lock the relation storage
	if _, firstOk := RelationStorage[srcType]; firstOk {
		if _, secondOk := RelationStorage[srcType][srcID]; secondOk {
			if _, thirdOk := RelationStorage[srcType][srcID][targetType]; thirdOk {
				if rel, fourthOk := RelationStorage[srcType][srcID][targetType][targetID]; fourthOk {
					// check if the version is fine
					if rel.Version != relation.Version {
						return types.StorageRelation{}, errors.New("Mismatch of version.")
					}
					rel.Version++

					// - - - - - - - - - - - - - - - - -
					// persistence.go handling
					if true == persistence.PersistenceFlag {
						persistence.PersistenceChan <- types.PersistencePayload{
							Type:     "Relation",
							Method:   "Create",
							Relation: rel,
						}
					}

					// - - - - - - - - - - - - - - - - -
					// update the data itself
					rel.Context = relation.Context
					rel.Properties = relation.Properties
					RelationStorage[srcType][srcID][targetType][targetID] = rel
					return relation, nil
				}
			}
		}
	}
	return types.StorageRelation{}, errors.New("Cant update non existing relation")
}

func GetChildRelationsBySourceTypeAndSourceId(Type int, id int, context string) (map[int]types.StorageRelation, error) {
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
	RelationStorageMutex.Lock()
	var pool = RelationStorage[Type][id]
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
				mapRet[cnt] = deepCopyRelation(relation)
				cnt++
			}
		}
	}
	RelationStorageMutex.Unlock()
	return mapRet, nil
}

func GetChildRelationsBySourceTypeAndSourceIdUnsafe(Type int, id int, context string) (map[int]types.StorageRelation, error) {
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
	var pool = RelationStorage[Type][id]
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
				mapRet[cnt] = deepCopyRelation(relation)
				cnt++
			}
		}
	}
	return mapRet, nil
}

func GetParentEntitiesByTargetTypeAndTargetIdAndSourceType(targetType int, targetID int, sourceType int, context string) map[int]types.StorageEntity {
	// initialice the return map
	var mapRet = make(map[int]types.StorageEntity)

	// set counter for the loop
	var cnt = 0

	// we lock the RelationStorage and EntityStorage
	// mutex with.  this allows us to proceed
	// faster since we just block to copy instead
	// of blocking for the whole process
	EntityStorageMutex.RLock()
	RelationStorageMutex.RLock()
	for sourceID, _ := range RelationRStorage[targetType][targetID][sourceType] {
		entity := EntityStorage[sourceType][sourceID]
		add := true
		if "" != context && context != entity.Context {
			add = false
		}
		// copy the relation into the return map
		// and upcount the int
		if true == add {
			mapRet[cnt] = deepCopyEntity(EntityStorage[sourceType][sourceID])
			cnt++
		}

	}
	RelationStorageMutex.RUnlock()
	EntityStorageMutex.RUnlock()

	return mapRet
}

func GetParentEntitiesByTargetTypeAndTargetIdAndSourceTypeUnsafe(targetType int, targetID int, sourceType int, context string) map[int]types.StorageEntity {
	// initialice the return map
	var mapRet = make(map[int]types.StorageEntity)

	// set counter for the loop
	var cnt = 0

	// we lock the RelationStorage and EntityStorage
	// mutex with.  this allows us to proceed
	// faster since we just block to copy instead
	// of blocking for the whole process
	for sourceID, _ := range RelationRStorage[targetType][targetID][sourceType] {
		entity := EntityStorage[sourceType][sourceID]
		add := true
		if "" != context && context != entity.Context {
			add = false
		}
		// copy the relation into the return map
		// and upcount the int
		if true == add {
			mapRet[cnt] = deepCopyEntity(EntityStorage[sourceType][sourceID])
			cnt++
		}

	}

	return mapRet
}

func GetParentRelationsByTargetTypeAndTargetId(targetType int, targetID int, context string) (map[int]types.StorageRelation, error) {
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
	RelationStorageMutex.RLock()
	var pool = RelationRStorage[targetType][targetID]
	// for each possible targtType
	for sourceTypeID, targetTypeMap := range pool {
		// for each possible targetId per targetType
		for sourceRelationID, _ := range targetTypeMap {
			// context handling, default is adding
			add := true
			if "" != context && context != RelationStorage[sourceTypeID][sourceRelationID][targetType][targetID].Context {
				add = false
			}
			// copy the relation into the return map
			// and upcount the int
			if true == add {
				mapRet[cnt] = deepCopyRelation(RelationStorage[sourceTypeID][sourceRelationID][targetType][targetID])
				cnt++
			}
		}
	}
	RelationStorageMutex.RUnlock()

	return mapRet, nil
}

func GetParentRelationsByTargetTypeAndTargetIdUnsafe(targetType int, targetID int, context string) (map[int]types.StorageRelation, error) {
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
	var pool = RelationRStorage[targetType][targetID]
	// for each possible targtType
	for sourceTypeID, targetTypeMap := range pool {
		// for each possible targetId per targetType
		for sourceRelationID, _ := range targetTypeMap {
			// context handling, default is adding
			add := true
			if "" != context && context != RelationStorage[sourceTypeID][sourceRelationID][targetType][targetID].Context {
				add = false
			}
			// copy the relation into the return map
			// and upcount the int
			if true == add {
				mapRet[cnt] = deepCopyRelation(RelationStorage[sourceTypeID][sourceRelationID][targetType][targetID])
				cnt++
			}
		}
	}
	return mapRet, nil
}

func GetEntityTypes() []string {
	// prepare the return array
	types := []string{}

	// now we lock the storage
	EntityTypeMutex.RLock()
	for _, Type := range EntityTypes {
		types = append(types, Type)
	}

	// unlock the mutex and return
	EntityTypeMutex.RUnlock()
	return types
}

func GetEntityTypesUnsafe() []string {
	// prepare the return array
	types := []string{}

	// now we lock the storage
	for _, Type := range EntityTypes {
		types = append(types, Type)
	}

	// unlock the mutex and return
	return types
}

func GetEntityRTypes() map[string]int {
	// prepare the return array
	types := make(map[string]int)

	// now we lock the storage
	EntityTypeMutex.RLock()
	for Type, id := range EntityRTypes {
		types[Type] = id
	}

	// unlock the mutex and return
	EntityTypeMutex.RUnlock()
	return types
}

func GetEntityRTypesUnsafe() map[string]int {
	// prepare the return array
	types := make(map[string]int)

	for Type, id := range EntityRTypes {
		types[Type] = id
	}

	return types
}

func TypeExists(strType string) bool {
	EntityTypeMutex.RLock()
	// lets check if this Type exists
	if _, ok := EntityRTypes[strType]; ok {
		// it does lets return it
		EntityTypeMutex.RUnlock()
		return true
	}

	EntityTypeMutex.RUnlock()
	return false
}

func TypeExistsUnsafe(strType string) bool {
	// lets check if this Type exists
	if _, ok := EntityRTypes[strType]; ok {
		// it does lets return it
		return true
	}

	return false
}

func EntityExists(Type int, id int) bool {
	EntityStorageMutex.RLock()
	// lets check if this Type exists
	if _, ok := EntityStorage[Type][id]; ok {
		// it does lets return it
		EntityStorageMutex.RUnlock()
		return true
	}

	EntityStorageMutex.RUnlock()
	return false
}

func EntityExistsUnsafe(Type int, id int) bool {
	// lets check if this Type exists
	if _, ok := EntityStorage[Type][id]; ok {
		// it does lets return it
		return true
	}

	return false
}

func TypeIdExists(id int) bool {
	EntityTypeMutex.RLock()
	// lets check if this Type exists
	if _, ok := EntityTypes[id]; ok {
		// it does lets return it
		EntityTypeMutex.RUnlock()
		return true
	}

	EntityTypeMutex.RUnlock()
	return false
}

func TypeIdExistsUnsafe(id int) bool {
	// lets check if this Type exists
	if _, ok := EntityTypes[id]; ok {
		// it does lets return it
		return true
	}

	return false
}

func GetTypeIdByString(strType string) (int, error) {
	EntityTypeMutex.RLock()
	// lets check if this Type exists
	if id, ok := EntityRTypes[strType]; ok {
		// it does lets return it
		EntityTypeMutex.RUnlock()
		return id, nil
	}

	EntityTypeMutex.RUnlock()
	return -1, errors.New("Entity Type string does not exist")
}

func GetTypeIdByStringUnsafe(strType string) (int, error) {
	// lets check if this Type exists
	if id, ok := EntityRTypes[strType]; ok {
		// it does lets return it
		return id, nil
	}

	return -1, errors.New("Entity Type string does not exist")
}

func GetTypeStringById(intType int) (*string, error) {
	EntityTypeMutex.RLock()
	// lets check if this Type exists
	if strType, ok := EntityTypes[intType]; ok {
		// it does lets return it
		EntityTypeMutex.RUnlock()
		return &strType, nil
	}

	EntityTypeMutex.RUnlock()
	return nil, errors.New("Entity Type string does not exist")
}

func GetTypeStringByIdUnsafe(intType int) (*string, error) {
	// lets check if this Type exists
	if strType, ok := EntityTypes[intType]; ok {
		// it does lets return it
		return &strType, nil
	}

	return nil, errors.New("Entity Type string does not exist")
}

func GetAmountPersistencePayloadsPending() int {
	return len(persistence.PersistenceChan)
}

func GetEntityAmount() int {
	amount := 0
	EntityStorageMutex.RLock()
	for key, _ := range EntityStorage {
		amount += len(EntityStorage[key])
	}
	EntityStorageMutex.RUnlock()
	return amount
}

func GetEntityAmountByType(intType int) (int, error) {
	EntityStorageMutex.RLock()
	// lets check if this Type exists
	if _, ok := EntityStorage[intType]; ok {
		// it does lets return
		amount := len(EntityStorage[intType])
		EntityStorageMutex.RUnlock()
		return amount, nil
	}

	EntityStorageMutex.RUnlock()
	return -1, errors.New("Entity Type does not exist")
}

func MapTransportData(data transport.TransportEntity) transport.TransportEntity {
	// first we lock all the storages
	EntityTypeMutex.Lock()
	EntityStorageMutex.Lock()
	RelationStorageMutex.Lock()

	// lets start recursive mapping of the data
	newID := mapRecursive(data, -1, -1, DIRECTION_NONE)

	// now we unlock all the mutexes again
	EntityTypeMutex.Unlock()
	EntityStorageMutex.Unlock()
	RelationStorageMutex.Unlock()

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

func mapRecursive(entity transport.TransportEntity, relatedType int, relatedID int, direction int) int {
	// first we get the right TypeID
	var TypeID int
	TypeID, err := GetTypeIdByStringUnsafe(entity.Type)
	if nil != err {
		TypeID, _ = CreateEntityTypeUnsafe(entity.Type)
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
		mapID, _ = CreateEntityUnsafe(tmpEntity)
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
			mapRecursive(childRelation.Target, TypeID, mapID, DIRECTION_CHILD)
		}
	}
	// than map the parent elements
	if len(entity.ParentRelations) != 0 {
		// there are children lets iteater over
		// the map
		for _, childRelation := range entity.ChildRelations {
			// pas the child entity and the parent coords to
			// create the relation after inserting the entity
			mapRecursive(childRelation.Target, TypeID, mapID, DIRECTION_PARENT)
		}
	}
	// now lets check if ourparent Type and id
	// are not -1 , if so we need to create
	// a relation
	if relatedType != -1 && relatedID != -1 {
		// lets create the relation to our parent
		if DIRECTION_CHILD == direction {
			// first we make sure the relation doesnt already exist (because we allow mapped existing data inside a to map json)
			if !RelationExistsUnsafe(relatedType, relatedID, TypeID, mapID) {
				tmpRelation := types.StorageRelation{
					SourceType: relatedType,
					SourceID:   relatedID,
					TargetType: TypeID,
					TargetID:   mapID,
					Version:    1,
				}
				CreateRelationUnsafe(relatedType, relatedID, TypeID, mapID, tmpRelation)
			}
		} else if DIRECTION_PARENT == direction {
			// first we make sure the relation doesnt already exist (because we allow mapped existing data inside a to map json)
			if !RelationExistsUnsafe(TypeID, mapID, relatedType, relatedID) {
				// or relation towards the child
				tmpRelation := types.StorageRelation{
					SourceType: TypeID,
					SourceID:   mapID,
					TargetType: relatedType,
					TargetID:   relatedID,
					Version:    1,
				}
				CreateRelationUnsafe(TypeID, mapID, relatedType, relatedID, tmpRelation)
			}
		}
	}
	// only the first return is interesting since it
	// returns the most parent id
	return mapID
}

func GetEntitiesByQueryFilter(
	typePool []string,
	conditions [][][3]string,
	idFilter [][]int,
	valueFilter [][]int,
	contextFilter [][]int,
	propertyList []map[string][]int,
	returnDataFlag bool,
) (
	[]transport.TransportEntity,
	[][2]int,
	int,
) {

	// check the pools given
	typeList := []int{}
	for _, eType := range typePool {
		if val, ok := EntityRTypes[eType]; ok {
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
		for entityID, entity := range EntityStorage[typeID] {
			add := false
			// if there are matchgroups
			if 0 < len(conditions) {
				for conditionGroupKey, conditionGroup := range conditions {
					// first we check if there is an ID filter
					// ### could have a special case for == on
					// id since this can be resolved very fast
					if 0 < len(idFilter[conditionGroupKey]) && !matchGroup(idFilter[conditionGroupKey], conditionGroup, strconv.Itoa(entityID)) {
						continue
					}
					// now we value
					if 0 < len(valueFilter[conditionGroupKey]) && !matchGroup(valueFilter[conditionGroupKey], conditionGroup, entity.Value) {
						continue
					}
					// than context
					if 0 < len(contextFilter[conditionGroupKey]) && !matchGroup(contextFilter[conditionGroupKey], conditionGroup, entity.Context) {
						continue
					}
					// and now the properties
					contGroupLoop := false
					for propertyKey, propertyConditions := range propertyList[conditionGroupKey] {
						if _, ok := entity.Properties[propertyKey]; ok {
							if !matchGroup(propertyConditions, conditionGroup, entity.Properties[propertyKey]) {
								contGroupLoop = true // ### refactor this i dont like it a bit but dont see a better way right now
								break
							}
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
						Type:            EntityTypes[entity.Type],
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

func GetEntitiesByQueryFilterAndSourceAddress(
	typePool []string,
	conditions [][][3]string,
	idFilter [][]int,
	valueFilter [][]int,
	contextFilter [][]int,
	propertyList []map[string][]int,
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
		if val, ok := EntityRTypes[eType]; ok {
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
			relPool[typeID] = getRelationTargetIDsBySourceAddressAndTargetType(sourceAddress[0], sourceAddress[1], typeID)
		} else {
			// else is -1 -> towards parents
			relPool[typeID] = getRRelationTargetIDsBySourceAddressAndTargetType(sourceAddress[0], sourceAddress[1], typeID)
		}
	}

	// now we know which IDs we have to check, so lets iterate through them
	for targetType, targetIDlist := range relPool {
		for _, targetID := range targetIDlist {
			add := false
			entity := EntityStorage[targetType][targetID]
			if 0 < len(conditions) {
				for conditionGroupKey, conditionGroup := range conditions {
					// first we check if there is an ID filter
					// ### could have a special case for == on
					// id since this can be resolved very fast
					if !matchGroup(idFilter[conditionGroupKey], conditionGroup, strconv.Itoa(targetID)) {
						continue
					}
					// now we value
					if !matchGroup(valueFilter[conditionGroupKey], conditionGroup, entity.Value) {
						continue
					}
					// than context
					if !matchGroup(contextFilter[conditionGroupKey], conditionGroup, entity.Context) {
						continue
					}
					// and now the properties
					contGroupLoop := false
					for propertyKey, propertyConditions := range propertyList[conditionGroupKey] {
						if _, ok := entity.Properties[propertyKey]; ok {
							if !matchGroup(propertyConditions, conditionGroup, entity.Properties[propertyKey]) {
								contGroupLoop = true // ### refactor this i dont like it a bit but dont see a better way right now
								break
							}
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
						Context:    getRelationContextByAddressAndDirection(sourceAddress[0], sourceAddress[1], targetType, targetID, direction),
						Properties: getRelationPropertiesByAddressAndDirection(sourceAddress[0], sourceAddress[1], targetType, targetID, direction),
						Target: transport.TransportEntity{
							Type:            EntityTypes[entity.Type],
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

func BatchUpdateAddressList(addressList [][2]int, values map[string]string) {
	for _, address := range addressList {
		entity, _ := GetEntityByPathUnsafe(address[0], address[1], "")
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
		UpdateEntityUnsafe(entity)
	}
}

func BatchDeleteAddressList(addressList [][2]int) {
	for _, address := range addressList {
		DeleteEntityUnsafe(address[0], address[1])
	}
}

func LinkAddressLists(from [][2]int, to [][2]int) int {
	linkedAmount := 0
	for _, singleFrom := range from {
		for _, singleTo := range to {
			// do we already have a relation between those too?`if not we create it
			if !RelationExistsUnsafe(singleFrom[0], singleFrom[1], singleTo[0], singleTo[1]) {
				CreateRelationUnsafe(singleFrom[0], singleFrom[1], singleTo[0], singleTo[1], types.StorageRelation{
					SourceType: singleFrom[0],
					SourceID:   singleFrom[1],
					TargetType: singleTo[0],
					TargetID:   singleTo[1],
				})
				archivist.Info("Creating link from to ", singleFrom[0], singleFrom[1], singleTo[0], singleTo[1])
				linkedAmount++
			}
		}
	}
	return linkedAmount
}

// - - - - - - - - - - - - - - - - - - - - - - - - - -
// + + + + + + + + + +  PRIVATE  + + + + + + + + + + +
// - - - - - - - - - - - - - - - - - - - - - - - - - -
func getRelationContextByAddressAndDirection(sourceType int, sourceID int, targetType int, targetID int, direction int) string {
	if 1 == direction {
		return RelationStorage[sourceType][sourceID][targetType][targetID].Context
	} else {
		return RelationStorage[targetType][targetID][sourceType][sourceID].Context
	}
}
func getRelationPropertiesByAddressAndDirection(sourceType int, sourceID int, targetType int, targetID int, direction int) map[string]string {
	// relations need to be copied the current way can lead to thread collision ###todo
	if 1 == direction {
		return RelationStorage[sourceType][sourceID][targetType][targetID].Properties
	} else {
		return RelationStorage[targetType][targetID][sourceType][sourceID].Properties
	}
}

func getRelationTargetIDsBySourceAddressAndTargetType(sourceType int, sourceID int, targetType int) []int {
	ret := make([]int, len(RelationStorage[sourceType][sourceID][targetType]))
	i := 0
	for key, _ := range RelationStorage[sourceType][sourceID][targetType] {
		ret[i] = key
		i++
	}
	return ret
}

func getRRelationTargetIDsBySourceAddressAndTargetType(sourceType int, sourceID int, targetType int) []int {
	ret := make([]int, len(RelationRStorage[sourceType][sourceID][targetType]))
	i := 0
	for key, _ := range RelationRStorage[sourceType][sourceID][targetType] {
		ret[i] = key
		i++
	}
	return ret
}

func matchGroup(filterGroup []int, conditions [][3]string, test string) bool {
	for _, filterGroupID := range filterGroup {
		if !match(test, conditions[filterGroupID][1], conditions[filterGroupID][2]) {
			return false
		}
	}
	return true
}

func match(alpha string, operator string, beta string) bool {
	switch operator {
	case "==":
		if alpha == beta {
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

func handleImport(importChan chan types.PersistencePayload) {
	var importEntities = 0
	var importRelations = 0
	for elem := range importChan {
		switch elem.Type {
		case "Entity":
			importEntity(elem)
			importEntities++
		case "Relation":
			importRelation(elem)
			importRelations++
		case "EntityTypes":
			importEntityTypes(elem)
		case "Done":
			// if we reach this we imported all data so we can close the channel
			// and return
			close(importChan)
			return
		}
	}
}

func importEntityTypes(payload types.PersistencePayload) {
	// than we lock the entity type mutex and relationstorage mutex
	EntityTypeMutex.Lock()
	EntityStorageMutex.Lock()
	RelationStorageMutex.Lock()
	//presets
	maxID := 0

	// first we build the corrosponding
	// reverse index and determine the maxID
	reverseMap := make(map[string]int)
	for key, value := range payload.EntityTypes {
		// store the reverse index
		reverseMap[value] = key

		// if bigger replace the max ID
		if maxID < key {
			maxID = key
		}

		// and prepare the relation storage
		RelationStorage[key] = make(map[int]map[int]map[int]types.StorageRelation)
		RelationRStorage[key] = make(map[int]map[int]map[int]bool)

		// same as the entity storage
		EntityStorage[key] = make(map[int]types.StorageEntity)
	}

	// store typemap, rmap and max id
	EntityTypeIDMax = maxID
	EntityTypes = payload.EntityTypes
	EntityRTypes = reverseMap

	//  unlock the mutex's again
	RelationStorageMutex.Unlock()
	EntityStorageMutex.Unlock()
	EntityTypeMutex.Unlock()
}

func importEntity(payload types.PersistencePayload) {
	// first we handle the ID max
	EntityIDMaxMutex.Lock()
	if EntityIDMax[payload.Entity.Type] < payload.Entity.ID {
		EntityIDMax[payload.Entity.Type] = payload.Entity.ID
	}
	EntityIDMaxMutex.Unlock()

	// now we create the entity themself
	// first we lock the storage
	EntityStorageMutex.Lock()

	// and put the entity in the EntityStorage
	EntityStorage[payload.Entity.Type][payload.Entity.ID] = payload.Entity

	// than unlock the entity storage again
	EntityStorageMutex.Unlock()

	// now we handle the relations , prepare the maps
	RelationStorageMutex.Lock()

	// create all the maps
	RelationStorage[payload.Entity.Type][payload.Entity.ID] = make(map[int]map[int]types.StorageRelation)
	RelationRStorage[payload.Entity.Type][payload.Entity.ID] = make(map[int]map[int]bool)

	// and unlock the relation storage again
	RelationStorageMutex.Unlock()
}

func importRelation(payload types.PersistencePayload) {
	//// - - - - - - - - - - - - - - - - -
	// now we lock the relation mutex
	RelationStorageMutex.Lock()

	// we check the case that the "from" history doesnt exist. in this case the relation makes no sense
	if _, ok := RelationStorage[payload.Relation.SourceType][payload.Relation.SourceID]; !ok {
		// unlock the mutex and stop processing
		RelationStorageMutex.Unlock()
		return
	}
	// same hotfix check see babove
	if _, ok := RelationRStorage[payload.Relation.TargetType][payload.Relation.TargetID]; !ok {
		// unlock the mutex and stop processing
		RelationStorageMutex.Unlock()
		return
	}

	// lets check if their exists a map for our
	// source entity to the target Type if not
	// create it.... golang things...
	if _, ok := RelationStorage[payload.Relation.SourceType][payload.Relation.SourceID][payload.Relation.TargetType]; !ok {
		RelationStorage[payload.Relation.SourceType][payload.Relation.SourceID][payload.Relation.TargetType] = make(map[int]types.StorageRelation)
	}

	// now we prepare the reverse storage if necessary
	if _, ok := RelationRStorage[payload.Relation.TargetType][payload.Relation.TargetID][payload.Relation.SourceType]; !ok {
		RelationRStorage[payload.Relation.TargetType][payload.Relation.TargetID][payload.Relation.SourceType] = make(map[int]bool)
	}

	// now we store the relation
	RelationStorage[payload.Relation.SourceType][payload.Relation.SourceID][payload.Relation.TargetType][payload.Relation.TargetID] = payload.Relation

	// - - - - - - - - - - - - - - - - -
	// and an entry into the reverse index, its existence
	// allows us to use the coords in the normal index to revtrieve
	// the Relation. We dont create a pointer because golang doesnt
	// allow pointer on submaps in nested maps
	RelationRStorage[payload.Relation.TargetType][payload.Relation.TargetID][payload.Relation.SourceType][payload.Relation.SourceID] = true

	// we are done now we can unlock the entity Types
	//// - - - - - - - - - - - - - - - -
	//and finally unlock the relation Type and return
	RelationStorageMutex.Unlock()
}

func deepCopyEntity(entity types.StorageEntity) types.StorageEntity {
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

func deepCopyRelation(relation types.StorageRelation) types.StorageRelation {
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
