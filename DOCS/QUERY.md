## Query Package Documentation

### Overview
The `query` package provides a flexible and powerful query language for querying data within a specific data model. This documentation aims to provide a comprehensive overview of the available methods and their usage.

### Query Methods and Their Usage

**1. Query Construction**
* **New()**: Creates a new, empty query object.

**2. Setting Query Type**
* **Read(etype ...string)**: Sets the query type to read entities of the specified types.
* **Reduce(etype ...string)**: Sets the query type to reduce entities of the specified types (e.g., aggregation).
* **Update(etype ...string)**: Sets the query type to update entities of the specified types.
* **Delete(etype ...string)**: Sets the query type to delete entities of the specified types.
* **Link(etype ...string)**: Sets the query type to create links between entities of the specified types.
* **Unlink(etype ...string)**: Sets the query type to remove links between entities of the specified types.
* **Find(etype ...string)**: Sets the query type to find entities of the specified types.

**3. Filtering and Matching**
* **Match(alpha string, operator string, beta string)**: Adds a condition to the query. The condition can be based on entity properties or relationships.
* **OrMatch(alpha string, operator string, beta string)**: Adds an OR condition to the query.

**4. Defining Relationships**
* **To(query *Query)**: Adds a child query to the current query.
* **From(query *Query)**: Adds a parent query to the current query.
* **CanTo(query *Query)**: Adds an optional child query to the current query.
* **CanFrom(query *Query)**: Adds an optional parent query to the current query.

**5. Modifying and Sorting**
* **Modify(properties ...string)**: Specifies properties to modify in an update operation.
* **SetDirection(direction int)**: Sets the direction for traversing relationships (parent or child).
* **Set(key string, value string)**: Sets a key-value pair for updating entity properties.
* **Order(field string, direction int, mode int)**: Specifies sorting criteria for the query results.

**6. Traversing Relationships**
* **TraverseOut(depth int)**: Traverses relationships outward from the current entity to a specified depth.
* **TraverseIn(depth int)**: Traverses relationships inward to a specified depth.

**7. Executing the Query**
* **Execute(query *Query)**: Executes the query and returns the results.

**Please note:** This is a preliminary list. We will refine and expand upon this documentation as we delve deeper into the package's capabilities.

In the next step, we will provide practical examples to illustrate how to use these methods to construct complex queries.

## Examples and Usage
### 1. Simple Read Query
```go
qry := New().Read("Entity")
result := Execute(qry)
```
This query reads all entities of type "Entity".


### 2. Filtering by Value:
```go
qry := New().Read("Entity").Match("Value", "==", "someValue")
result := Execute(qry)
```
This query reads all entities of type "Entity" where the "Value" property equals "someValue".

### 3. Filtering by Context:
```go
qry := New().Read("Entity").Match("Context", "==", "someContext")
result := Execute(qry)
```
This query reads all entities of type "Entity" where the "Context" property equals "someContext".

### 4. Filtering by Property:
```go
qry := New().Read("Entity").Match("Properties.propertyKey", "==", "propertyValue")
result := Execute(qry)
```
This query reads all entities of type "Entity" where the "propertyKey" property equals "propertyValue".

### 5. Combining Filters:
```go
qry := New().Read("Entity").Match("Value", "==", "someValue").Match("Context", "==", "someContext")
result := Execute(qry)
```
This query reads all entities of type "Entity" where both the "Value" and "Context" properties match the specified values.

### 6. Simple Child Join:
```go
qry := New().Read("EntityA").To(New().Read("EntityB"))
result := Execute(qry)
```
This query reads all entities of type "EntityA" and their child linked "EntityB" entities.

### 7. Filtered Join:
```go
qry := New().Read("EntityA").To(New().Read("EntityB").Match("Value", "==", "someValue"))
result := Execute(qry)
```
This query reads all entities of type "EntityA" and their linked "EntityB" entities where the "EntityB" entities have a "Value" property equal to "someValue".
