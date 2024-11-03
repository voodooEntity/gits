# Storage API
GITS exposes its internal storage api to the developer. While it is recommended to primary use [queries](./QUERY.md) and [data mapper](DATA_MAPPING.md) there might be certain situations in which direct usage of the storage might be better.

It is important to notice that the storage works with a different kind of datasets than [query](./QUERY.md) and [data mapper](./DATA_MAPPING.md) - "StorageEntity" and "StorageRelation". You might want to have a peek at the [struct definitions](./STORAGE_ARCHITECTURE.md).

The two major differences between the datasets are:
1. Storage datasets dont support nesting
2. In most cases - entity types are handled by ID not by String

**Important:** Many but not all functions of the storage have an so known 'unsafe counterpart'. For example there is "CreateEntityType" and "CreateEntityTypeUnsafe". While GITS is designed to be fully concurrency safe, you are able to call the core functions without the concurrency protection. Its discouraged to do so, and there really is just a handful of cases in which this is useful. 

## API Definitions
### Core Functions

* **NewStorage()**
  * Creates a new Storage instance.
  * **Returns:** *Storage*
* **CreateEntityType(name string)**
  * Creates a new entity type with the given name.
  * **Returns:** *int, error*
  * *Note: Has an unsafe counterpart.*
* **CreateEntity(entity types.StorageEntity)**
  * Creates a new entity.
  * **Returns:** *int, error*
  * *Note: Has an unsafe counterpart.*
* **CreateEntityUniqueValue(entity types.StorageEntity)**
  * Creates a new entity, ensuring uniqueness based on a specific value.
  * **Returns:** *int, bool, error*
  * *Note: Has an unsafe counterpart.*
* **GetEntityByPath(Type int, id int, context string)**
  * Retrieves an entity by its type and ID.
  * **Returns:** *types.StorageEntity, error*
  * *Note: Has an unsafe counterpart.*
* **GetEntitiesByType(Type string, context string)**
  * Retrieves entities of a specific type.
  * **Returns:** *map[int]types.StorageEntity, error*
  * *Note: Has an unsafe counterpart.*
* **GetEntitiesByValue(value string, mode string, context string)**
  * Retrieves entities based on a specific value and mode.
  * **Returns:** *map[int]types.StorageEntity, error*
  * *Note: Has an unsafe counterpart.*
* **GetEntitiesByTypeAndValue(Type string, value string, mode string, context string)**
  * Retrieves entities based on type, value, and mode.
  * **Returns:** *map[int]types.StorageEntity, error*
  * *Note: Has an unsafe counterpart.*
* **UpdateEntity(entity types.StorageEntity)**
  * Updates an existing entity.
  * **Returns:** *error*
  * *Note: Has an unsafe counterpart.*
* **DeleteEntity(Type int, id int)**
  * Deletes an entity.
  * **Returns:** *nil*
  * *Note: Has an unsafe counterpart.*

### Relation Functions

* **GetRelation(srcType int, srcID int, targetType int, targetID int)**
  * Retrieves a relation between two entities.
  * **Returns:** *types.StorageRelation, error*
  * *Note: Has an unsafe counterpart.*
* **RelationExists(srcType int, srcID int, targetType int, targetID int)**
  * Checks if a relation exists between two entities.
  * **Returns:** *bool*
  * *Note: Has an unsafe counterpart.*
* **DeleteRelationList(relationList map[int]types.StorageRelation)**
  * Deletes a list of relations.
  * **Returns:** *nil*
  * *Note: Has an unsafe counterpart.*
* **DeleteRelation(sourceType int, sourceID int, targetType int, targetID int)**
  * Deletes a specific relation.
  * **Returns:** *nil*
  * *Note: Has an unsafe counterpart.*
* **DeleteChildRelations(Type int, id int)**
  * Deletes all child relations of an entity.
  * **Returns:** *error*
  * *Note: Has an unsafe counterpart.*
* **DeleteParentRelations(Type int, id int)**
  * Deletes all parent relations of an entity.
  * **Returns:** *error*
  * *Note: Has an unsafe counterpart.*
* **CreateRelation(srcType int, srcID int, targetType int, targetID int, relation types.StorageRelation)**
  * Creates a new relation between two entities.
  * **Returns:** *bool, error*
  * *Note: Has an unsafe counterpart.*
* **UpdateRelation(srcType int, srcID int, targetType int, targetID int, relation types.StorageRelation)**
  * Updates an existing relation.
  * **Returns:** *types.StorageRelation, error*
  * *Note: Has an unsafe counterpart.*
* **GetChildRelationsBySourceTypeAndSourceId(Type int, id int, context string)**
  * Retrieves child relations of an entity.
  * **Returns:** *map[int]types.StorageRelation, error*
  * *Note: Has an unsafe counterpart.*
* **GetParentEntitiesByTargetTypeAndTargetIdAndSourceType(targetType int, targetID int, sourceType int, context string)**
  * Retrieves parent entities of an entity.
  * **Returns:** *map[int]types.StorageEntity*
  * *Note: Has an unsafe counterpart.*
* **GetParentRelationsByTargetTypeAndTargetId(targetType int, targetID int, context string)**
  * Retrieves parent relations of an entity.
  * **Returns:** *map[int]types.StorageRelation, error*
  * *Note: Has an unsafe counterpart.*

### Type and Entity Management

* **GetEntityTypes()**
  * Retrieves a map of entity types and their IDs.
  * **Returns:** *map[int]string*
  * *Note: Has an unsafe counterpart.*
* **GetEntityRTypes()**
  * Retrieves a map of entity type names and their IDs.
  * **Returns:** *map[string]int*
  * *Note: Has an unsafe counterpart.*
* **TypeExists(strType string)**
  * Checks if an entity type exists.
  * **Returns:** *bool*
  * *Note: Has an unsafe counterpart.*
* **EntityExists(Type int, id int)**
  * Checks if an entity exists.
  * **Returns:** *bool*
  * *Note: Has an unsafe counterpart.*
* **TypeIdExists(id int)**
  * Checks if an entity type ID exists.
  * **Returns:** *bool*
  * *Note: Has an unsafe counterpart.*
* **GetTypeIdByString(strType string)**
  * Retrieves the ID of an entity type given its name.
  * **Returns:** *int, error*
  * *Note: Has an unsafe counterpart.*
* **GetTypeStringById(intType int)**
  * Retrieves the name of an entity type given its ID.
  * **Returns:** *string, error*
  * *Note: Has an unsafe counterpart.*
* **GetEntityAmount()**
  * Retrieves the total number of entities.
  * **Returns:** *int*
* **GetEntityAmountByType(intType int)**
  * Retrieves the number of entities of a specific type.
  * **Returns:** *int, error*

### Additional Functions / Mainly build for query interpreter

* **MapTransportData(data transport.TransportEntity)**
  * Maps data ins transport.* format to storage. This method is exposed via an interface function directly by ur GITS instance [and documented here](./DATA_MAPPING.md).
  * **Returns:** *transport.TransportEntity*
* **GetEntitiesByQueryFilter(...)**
  * Retrieves entities based on a query filter.
  * **Returns:** *...*
* **GetEntitiesByQueryFilterAndSourceAddress(...)**
  * Retrieves entities based on a query filter and source address.
  * **Returns:** *...*
* **BatchUpdateAddressList(addressList [][2]int, values map[string]string)**
  * Batch updates addresses.
  * **Returns:** *nil*
* **BatchDeleteAddressList(addressList [][2]int)**
  * Batch deletes addresses.
  * **Returns:** *nil*
* **LinkAddressLists(from [][2]int, to [][2]int)**
  * Links address lists.
  * **Returns:** *int*
* **TraverseEnrich(entity *transport.TransportEntity, direction int, depth int)**
  * Traverses and enriches an entity.
  * **Returns:** *nil*
