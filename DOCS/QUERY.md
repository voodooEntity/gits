## Query Package Documentation

## Overview
The `query` package provides a flexible and powerful query language for querying data within a specific data model. This documentation aims to provide a comprehensive overview of the available methods and their usage.

## Query Methods
**1. Query Construction**
* **gits.NewQuery()**: Creates a new, empty query object.

**2. Setting Query Type**
* **Read(etype ...string)**:Sets the query type to read entities of the specified types.
* **Reduce(etype ...string)**: Sets the query type to reduce entities of the specified types (used in joins to reduce the results).
* **Update(etype ...string)**: Sets the query type to update entities of the specified types. Is only supported as root query.
* **Delete(etype ...string)**: Sets the query type to delete entities of the specified types. Is only supported as root query.
* **Link(etype ...string)**: Sets the query type to create links between entities of the specified types. Is only supported as root query.
* **Unlink(etype ...string)**: Sets the query type to remove links between entities of the specified types. Is only supported as root query.
* **Find(etype ...string)**: Sets the query type to find entities of the specified types. Is used in context of "Link()" and "Unlink()"

**3. Filtering and Matching**
* **Match(alpha string, operator string, beta string)**: Adds a condition to the query. The condition can be based on entity value, context, id or properties. Multiple match queries will be assumed as "AND".
* **OrMatch(alpha string, operator string, beta string)**: Adds an OR condition to the query match definitions. 

**4. Defining Relationships**
* **To(query *Query)**: Adds a child query to the current query.
* **From(query *Query)**: Adds a parent query to the current query.
* **CanTo(query *Query)**: Adds an optional child query to the current query.
* **CanFrom(query *Query)**: Adds an optional parent query to the current query.

**5. Modifying and Sorting**
* **Set(key string, value string)**: Sets a key-value pair for updating entity value,context or properties.
* **Order(field string, direction int, mode int)**: Specifies sorting criteria for the query results. Is only supported to modify the root query. Will sort results based on root level of results.

**6. Traversing Relationships**
* **TraverseOut(depth int)**: Traverses relationships outward (children) from the current entity up to a specified depth.
* **TraverseIn(depth int)**: Traverses relationships inward (parents) up to a specified depth.

**7. Executing the Query**
* **instance.ExecuteQuery(query *Query)**: Executes the query and returns the results.

In the next step, we will provide practical examples to illustrate how to use these methods to construct complex queries.

## Examples and Usage
In this examples we are assuming there is a instance created before starting to query.
```go
gitsInstance := gits.NewInstance("test")
```

### 1. Simple Read Query
```go
qry := gits.NewQuery().Read("Entity")
result := gitsInstance.ExecuteQuery(qry)
```
This query reads all entities of type "Entity".

### 2. Filtering by Value:
```go
qry := gits.NewQuery().Read("Entity").Match("Value", "==", "someValue")
result := gitsInstance.ExecuteQuery(qry)
```
This query reads all entities of type "Entity" where the "Value" property equals "someValue". [List of supported matching operators](#supported-match-operators)

### 3. Filtering by Context:
```go
qry := gits.NewQuery().Read("Entity").Match("Context", "==", "someContext")
result := gitsInstance.ExecuteQuery(qry)
```
This query reads all entities of type "Entity" where the "Context" property equals "someContext".  [List of supported matching operators](#supported-match-operators)

### 4. Filtering by Property:
```go
qry := gits.NewQuery().Read("Entity").Match("Properties.propertyKey", "==", "propertyValue")
result := gitsInstance.ExecuteQuery(qry)
```
This query reads all entities of type "Entity" where the "PropertyKey" property equals "propertyValue".  [List of supported matching operators](#supported-match-operators)

### 5. Combining Filters:
```go
qry := gits.NewQuery().Read("Entity").Match("Value", "==", "someValue").Match("Context", "==", "someContext")
result := gitsInstance.ExecuteQuery(qry)
```
This query reads all entities of type "Entity" where both the "Value" and "Context" properties match the specified values. Consecutive Match() statements are assumed as "AND".  [List of supported matching operators](#supported-match-operators)

### 6. Combining filters with OR Condition
```go
qry := gits.NewQuery().Read("Entity").Match("Context","==","Lorem").Match("Value", "==", "someValueA").OrMatch("Context", "==", "ipsum").Match("Value","==","finally")
result := gitsInstance.ExecuteQuery(qry)
```
This query reads all entities of type "Entity" where either ("Context" equals "Lorem" and "Value" equals "someValueA") OR ("Context" equals "ipsum" and "Value" equals "finally"). While multiple consecutive match conditions are assumed as AND, adding an "OrMatch()" will split the previous and following into different groups. You are allowed to use as many "OrMatch" as you need. Nesting of conditions is not supported right now.  [List of supported matching operators](#supported-match-operators)

### 7. Simple Child Join:
```go
qry := gits.NewQuery().Read("EntityA").To(gits.NewQuery().Read("EntityB"))
result := gitsInstance.ExecuteQuery(qry)
```
This query reads all entities of type "EntityA" and their child linked "EntityB" entities.

### 8. Filtered Join:
```go
qry := gits.NewQuery().Read("EntityA").To(gits.NewQuery().Read("EntityB").Match("Value", "==", "someValue"))
result := gitsInstance.ExecuteQuery(qry)
```
This query reads all entities of type "EntityA" and their linked "EntityB" entities where the "EntityB" entities have a "Value" property equal to "someValue".

### 9. Traversing out:
```go
qry := gits.NewQuery().Read("Entity").TraverseOut(3)
result := gitsInstance.ExecuteQuery(qry)
```
This query reads all entities of type "Entity", than it will traverse out (follow relations towards children) up to a depth of 3. Can be especially useful in abstract structures where properties are handled as child entities or abstracts without static structural definitions. Traverse is supported towards children "TraverseOut" and parents "TraverseIn". 

### 10. Update entities
```go
qry := gits.NewQuery().Update("Entity").Match("Value","==","old").Set("Value", "Lorem").Set("Context", "Ipsum").Set("Properties.dolor","appropinquare")
result := gitsInstance.Execute(qry)
```
This query will update all entities of type "Entity" which match ("Value" equals "old"). It will update "Context" to "Ipsum", "Value" to "Lorem" and the Property "dolor" to "appropinquare". This can affect a single or multiple entities, based on your filters. Update query must always be a root level query. Update can be used with "(Can)To" and "(Can)From" in order to reduce/filter the affected datasets. It is recommended to use "Reduce()" instead of "Read()" in such subqueries to minimize the amount of allocated memory.

### 11 Delete entities
```go
qry := gits.NewQuery().Delete("Entity").Match("Value", "==", "deleteme")
result := gitsInstance.Execute(qry)
```
This query will delete all entities of type "Entity" which match ("Value" equals "deleteme"). This can affect a single or multiple entities, based on your filters. Delete query must always be a root level query. Delete can be used with "(Can)To" and "(Can)From" in order to reduce/filter the affected datasets. It is recommended to use "Reduce()" instead of "Read()" in joins to minimize the amount of allocated memory.

### 12. Link entities
```go
qry := gits.NewQuery().Link("EntityA").Match("Value", "==", "alpha").To(
    gits.NewQuery().Find("EntityB").Match("Value", "==", "omega"),
)
gitsInstance.ExecuteQuery(qry)
```
This query will find all entities of type "EntityA" which match "Value" equals "alpha" and link (create a directed relation) the result list to result of the join which matches entities of type "EntityB" with "Value" equals "omega". As you can see the "To" definition is used in this context to define the direction of the "Link" action, in this case towards children. Also we use "Find" instead of "Read or Reduce" in order to provide the necessary dataset address list to our link function. You can use this to link any amount of entities. Link query must always be a root level query. 

### 13. Unlink entities
```go
qry := gits.NewQuery().Unlink("EntityA").Match("Value", "==", "alpha").To(
    gits.NewQuery().Find("EntityB").Match("Value", "==", "omega"),
)
gitsInstance.ExecuteQuery(qry)
```
This query will find all entities of type "EntityA" which match "Value" equals "alpha" and unlink (remove a directed relation) the result list to result of the join which matches entities of type "EntityB" with "Value" equals "omega". As you can see the "To" definition is used in this context to define the direction of the "Unlink" action, in this case towards children. Also we use "Find" instead of "Read or Reduce" in order to provide the necessary dataset address list to our unlink function. You can use this to unlink any amount of entities. Unlink query must always be a root level query.

### 14. Adjusting the result order
```go
qry := gits.NewQuery().Read("EntityA").To(
    gits.NewQuery().Read("EntityB"),
).Order("Value", ORDER_DIRECTION_ASC, ORDER_MODE_ALPHA)
result := gitsInstance.ExecuteQuery(qry)
```
This query will find all entities of type "EntityA" which are linked to entities of type "EntityB". Before returning the data, it will resort the order of the root level results by the field "Value" direction "ASC" (ascending) in mode "Alpha(numeric)". Order can only be applied on root level queries and will sort results only on root level results.

## Definitions
### Supported Match Operators
The following operators are supported in terms of matching actions.

| Operator | Description                      | Alpha Cast                      | Beta Cast |
|----------|----------------------------------|---------------------------------|-----------|
| ==       | alpha euqals beta                |                                 |           |
| !=       | alpha does not equal beta        |                                 |           |
| prefix   | beta is prefix of alpha          |                                 |           |
| suffix   | beta is suffix of alpha          |                                 |           |
| contain  | alpha contains beta              |                                 |           |
| >        | alpha is greater than beta       | int                             | int       |
| >=       | alpha is grater or equal to beta | int                             | int       |
| <        | alpha is lower than beta         | int                             | int       |
| <=       | alpha is lower or equal to beta  | int                             | int       |
| in       | if any alpha is equal to beta    | alpha is split by "," delimiter |           |
