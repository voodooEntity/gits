# Creating/Mapping Data
The easiest and recommended way in GITS to add new data to your storage is to use the *MapData* method available on your [gits.Gits Instance](INSTANCES.md).

While this method can be used to store a single dataset, it can also be used to store multiple depth of nested datasets at once.

At this point it is recommended to have a look at the [transport.Transport* struct definitions](STORAGE_ARCHITECTURE.md).

## Examples
### Single new dataset mapping
Lets define a simple minimum dataset and put it in our storage.

```go
alphaGits := gits.NewInstance("alpha")
datasetToStore := transport.TransportEntity{
	ID: storage.MAP_FORCE_CREATE,
	Type: "Example",
	Value: "Lorem ipsum dolor",
	
}
newId := alphaGits.MapData(datasetToStore)
```

In our example, we create a new gits instance, defined a minimal dataset in form of a transport.TransportEntity and Mapped it using the MapData function provided by our alpha instance.

As you might notice i passed the constant "storage.MAP_FORCE_CREATE" as value of the ID. This is because right now there are three supported modes in terms of defining the ID. In our case we wanted to have a new dataset created, therefor we use the MAP_FORCE_CREATE.

It is also important to notice that we received newId (int) as return from MapData. This is the ID of the newly created dataset. Since in GITS (for details [check the Storage Architecture](STORAGE_ARCHITECTURE.md)) the unique identifier for a dataset is the combination of Type and ID, we can now directly access this dataset using the given information.


### Nested new dataset mapping
As the title might suggest, GITS is also able to map nested datasets. Since GITS by nature is structured as a directed graph storage, relations between two entities always have a parent and a child. Therefor an entity can have parents and children.

To create such a nested structure we will need the transport.TransportEntity but also transport.TransportRelation.

An example of mapping such a dataset could look like
```go
alphaGits := gits.NewInstance("alpha")
datasetToStore := transport.TransportEntity{
	ID: storage.MAP_FORCE_CREATE,
	Type: "Example",
	Value: "Lorem ipsum dolor",
	ChildRelations: []transport.TransportRelation{
        {
            Target: transport.TransportEntity{
                ID:    storage.MAP_FORCE_CREATE,
                Type:  "Another",
                Value: "Great content",
            },
        }, {
            Target: transport.TransportEntity{
                ID:    storage.MAP_FORCE_CREATE,
                Type:  "or just",
                Value: "another example",
            },
        },
    },
	
}
newId := alphaGits.MapData(datasetToStore)
```

In this example we created an entity of type Example, then we create two more entities of type "Another" and "or just". Finally we mapped 2 relations from the "Example" entity towards the other two newly created once.

This means with one call you created 3 datasets with 2 directed relations.

There is no direct restriction in how width/far you can nest your data - lets say the callstack is the limit.

Also, as you may see we still get one id returned. This will be the ID of the most parent dataset passed to the MapData function.

This type of nesting works not just towards Children but also towards Parents. 

### Map by Value (and Context)
The next important point is the ability to "upsert" datasets based on Value (and optional Context).

If, instead of using the storage.MAP_FORCE_CREATE, we use the second constant - storage.MAP_USERT_VALUE_CONTEXT.

This constant will trigger a specific logic when trying to map this entity into our storage.

Instead of just creating the dataset, the MapData will first check if a similar Entity (same Type and Value) already exists. If yes, the entity will not be recreated. If it does not exist, the entity will be created. If you define a Context in your entity that you Map with storage.MAP_USERT_VALUE_CONTEXT than the entities Context will also be used to check if this entity already exists in the storage.


