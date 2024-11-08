# Storage Architecture

## Index
* [Overview](#overview)
* [Storage Types](#storage-types)
* [Data Storage Structure](#data-storage-structure)
  * [Entity Storage Structure](#entity-storage-structure)
  * [Relation Storage Structure](#relation-storage-structure)
  * [Reverse Relation Storage Structure](#reverse-relation-storage-structure)
  * [Other Storage Fundamentals](#other-storage-fundamentals)
    * [EntityTypeID](#entitytypeid)
* [Transport Definitions](#transport-definitions)
  * [Transport Entity](#transport-entity)
  * [Transport Relations](#transport-relations)
  * [Transport](#transport)
* [Key Points:](#key-points)

## Overview
This document should provide an architectural insight in the storage behind GITS. The concept for this storage was designed in order to have high performances access while still providing concurrency safety. While the storage could be used as a simple Typed Objekt Storage, it is designed to depict a directed graph storage. 

The design allows various forms of abstraction complexity in structural definitions. 

All values stored in the storage are strings. While this expects the applications using it to convert other types before storing and after retrieving, this allows for a much more streamlined way the storage works. 

## Storage Types
1. **StorageEntity:**
    - Represents a real-world object or concept.
    - Stored as a `StorageEntity` struct with properties like:
        - `ID`: A unique identifier within its *type*.
        - `Type`: Categorizes the entity (e.g., "user," "product").
        - `Context`: Additional context or metadata.
        - `Value`: The primary value or data associated with the entity.
        - `Properties`: Key-value pairs for extra information.
        - `Version`: Tracks changes to the entity.

2. **StorageRelation:**
    - Connects two entities, indicating a specific connection or association, is directed.
    - Stored as a `StorageRelation` struct with properties like:
        - `SourceType`: Type of the source entity.
        - `SourceID`: ID of the source entity.
        - `TargetType`: Type of the target entity.
        - `TargetID`: ID of the target entity.
        - `Context`: Additional context or metadata for the relationship.
        - `Properties`: Key-value pairs for extra information.
        - `Version`: Tracks changes to the relationship.

[to top](#storage-architecture)

## Data Storage Structure

The storage system uses a hierarchical map structure to efficiently store entities and relationships:

### Entity Storage Structure
```
EntityStorage: {
  EntityType1: {
    EntityType1ID1: StorageEntity1,
    EntityType1ID2: StorageEntity2,
    ...
  },
  EntityType2: {
    EntityType2ID1: StorageEntity3,
    ...
  },
  ...
}
```

- **Keys / Path:** 
  - (int) Entity Type ID
  - (int) Entity ID
- **Value:** `StorageEntity` struct

### Relation Storage Structure
```
RelationStorage: {
  SourceType1: {
    SourceID1: {
      TargetType1: {
        TargetID1: StorageRelation1,
        TargetID2: StorageRelation2,
        ...
      },
      TargetType2: {
        ...
      },
      ...
    },
    SourceID2: {
      ...
    },
    ...
  },
  SourceType2: {
    ...
  },
  ...
}
```
- **Keys / Path:** 
    - (int) Source Entity Type ID
    - (int) Source Entity ID
    - (int) Target Entity Type
    - (int) Target Entity ID
- **Value:** `StorageRelation` struct

### Reverse Relation Storage Structure
```
RelationRStorage: {
  TargetType1: {
    TargetID1: {
      SourceType1: {
        SourceID1: true,
        SourceID2: true,
        ...
      },
      SourceType2: {
        ...
      },
      ...
    },
    TargetID2: {
      ...
    },
    ...
  },
  TargetType2: {
    ...
  },
  ...
}
```
- **Keys / Path:**
    - (int) Target Entity Type ID
    - (int) Target Entity ID
    - (int) Source Entity Type
    - (int) Source Entity ID
- **Value:** `bool` true

[to top](#storage-architecture)

### Other Storage Fundamentals
#### EntityTypeID
* EntityIDMax          
  * Definition: 
    * `map[int]int`
  * Keys: 
    * (int) Entity Type ID
  * Description: 
    * Tracks the max ID for each Entity Type for auto increment purposes
* EntityTypes          
  * Definition: 
    * `map[int]string`
  * Keys:
    * (int) Entity Type ID
  * Description:
    * Tracks all entity types
* EntityRTypes         
  * Defintion: 
    * `map[string]int`
  * Keys:
    * Entity Type String
  * Description:
    * Reverse index for EntityTypes for easier and faster lookups
* EntityTypeIDMax      
  * Defintion: 
    * `int`
  * Description:
    * Tracks the Max ID of EntityTypes
* EntityStorageMutex   
  * Definition: 
    * `*sync.RWMutex`
  * Description:
    * RWMutex instances, used to Read/Write lock when working with the "EntityStorage". 
* EntityIDMaxMutex     
  * Definition: 
    * `*sync.RWMutex`
  * Description:
      * RWMutex instances, used to Read/Write lock when working with the "EntityIDMax".
* EntityTypeMutex      
  * Definition: 
    * `*sync.RWMutex`
        * RWMutex instances, used to Read/Write lock when working with the "EntityTypes","EntityRType","EntityTypeID" and "EntityTypeIDMax".
* RelationStorageMutex 
  * Definition: 
    * `*sync.RWMutex`
  * Description:
      * RWMutex instances, used to Read/Write lock when working with the "RelationStorage" and "RelationRStorage" .

[to top](#storage-architecture)
## Transport Definitions
To provide a better usability when handling datasets, the "transport" package has been implemented. It provides a more verbose representation of the data including handling the types as their string names and having the capability to be nested in order to create whole structures. 

**Important**: While this way of representing a directed graph offers an easy and fast way to work with the data, it will not prevent recursive queries (and that for possible duplicate data in the result). 

### Transport Entity
```go
type TransportEntity struct {
    Type            string
    ID              int
    Value           string
    Context         string
    Version         int
    Properties      map[string]string
    ChildRelations  []TransportRelation
    ParentRelations []TransportRelation
}
```

### Transport Relations
```go
type TransportRelation struct {
    Context    string
    Properties map[string]string
    Target     TransportEntity
    SourceType string
    SourceID   int
    TargetType string
    TargetID   int
    Version    int
}
```

### Transport
This format is used for Query Results. It was introduced in order to enable us to implement a flat representation of results rather than the nested one. For now only the nested representation exists.
```go
type Transport struct {
    Entities  []TransportEntity
    Relations []TransportRelation
    Amount    int
}
```


## Key Points:

- **Efficient Retrieval:** The nested map structure allows for quick retrieval of entities and relationships based on their types and IDs.
- **Scalability:** The storage can handle a large number of entities and relationships by efficiently indexing and organizing data.
- **Flexibility:** The system can accommodate various entity types and relationship types, making it adaptable to different use cases.
- **Versioning:** The versioning mechanism allows for tracking changes to entities and relationships.
- **Unique IDs per Type:** Entity IDs are unique within their respective types, ensuring no conflicts.

By understanding this core structure, you can effectively utilize the storage system to model complex data relationships and perform efficient queries and updates.

[to top](#storage-architecture) - 
[Docuemntation Overview](./README)