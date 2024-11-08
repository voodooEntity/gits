# Creating/Mapping Data
The easiest and recommended way in GITS to add new data to your storage is to use the *MapData* method available on your [gits.Gits Instance](INSTANCES.md).

While this method can be used to store a single dataset, it can also be used to store multiple depth of nested datasets at once.

At this point it is recommended to have a look at the [transport.Transport* struct definitions](STORAGE_ARCHITECTURE.md).

## Index
* [Examples](#examples)
  * [Single new dataset mapping](#single-new-dataset-mapping)
  * [Nested new dataset mapping](#nested-new-dataset-mapping)
  * [Map by Value (and Context)](#map-by-value-and-context)
  * [Map to existing](#map-to-existing)
  * [Complex datasets](#complex-datasets)

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
The next important point is the ability to "create if not exists" datasets based on Type and Value (and optional Context).

If, instead of using the storage.MAP_FORCE_CREATE, we use the second constant - storage.MAP_IF_NOT_EXISTS.

This constant will trigger a specific logic when trying to map this entity into our storage.

Instead of just creating the dataset, the MapData will first check if a similar Entity (same Type and Value) already exists. If yes, the entity will not be recreated. If it does not exist, the entity will be created. If you also define a Context in said new entity it will also be taking into account when checking if it's already existing .

```go
alphaGits := gits.NewInstance("alpha")
datasetToStore := transport.TransportEntity{
	ID: storage.MAP_IF_NOT_EXISTS,
	Type: "Example",
	Value: "Lorem ipsum dolor",
}
newOrExistingId := alphaGits.MapData(datasetToStore)
```
As you can see the call is very similar just differing in terms of the constant passed for the ID. Also, as usual the MapData function will return the most root dataset ID, in this case it will either return the ID of the newly created dataset - or in case a similar dataset already exists, will return the ID of said already existing one.

This mode can also be used in nested structures such as shown above.

### Map to existing
This method is used to map new data to existing datasets/structures. If you want to insert data mapped to existing datasets, you may simply address them in your structure via Type and ID.

First we create a dataset 
```go
alphaGits := gits.NewInstance("alpha")
datasetToStore := transport.TransportEntity{
	ID: storage.MAP_FORCE_CREATE,
	Type: "Example",
	Value: "Lorem ipsum dolor",
}
newId := alphaGits.MapData(datasetToStore)
```

Now we will map a nested structure in which the just created instance will be mapped as child to a new created dataset, using the newId variable.
```go
alphaGits := gits.NewInstance("alpha")
datasetToStore := transport.TransportEntity{
	ID: storage.MAP_FORCE_CREATE,
	Type: "Something",
	Value: "appropinquare",
	ChildRelations: []transport.TransportRelation{
        {
            Target: transport.TransportEntity{
                ID:    newId,
                Type:  "Example",
            },
        },
    },
}
secondNewId := alphaGits.MapData(datasetToStore)
```

This code would create the new entity of type "Something" and map it to the existing entity of type Example with ID newId.

This way of addressing already existing entities can be used in any nested structure to enhance your existing data.

### Complex datasets
Till this point we always created a minimum entity. Tho an entity is capable of holding more data than just a Value. While the idea is that the Value may define an entity to a certain degree, sometimes you will need some additional information. While every additional information could be added using more entities, a more pragmatic way is to add some properties. That for an entity also has a map[string]string to also hold such properties. Also as mentioned earlier a dataset can hold a Context string. A more complex entity definition might look like the following

```go
transport.TransportEntity{
    ID:         storage.MAP_FORCE_CREATE,
    Type:       "ExampleType",
    Value:      "SomeGreatValue",
    Context:    "TheGreaterPicture",
    Properties: map[string]string{"some": "properties","you":"might","need":"later"},
}
```

This enables the developer to vary the level of data abstraction based on actual practical needs. Either using different entities for your properties, or enhance the entities directly. It's up to you.

### Final notes
The MapData function should provide an easy way to map structures combining existing datasets and new datasets. 

If you dont specify an ID in your transport.TransportEntity, the default will be MAP_IF_NOT_EXISTS.

## FAQ

Q: Is it possible to update an existing entity using MapData?

A: No, MapData is only able to create new and/or map inbetween new and existing datasets. To update you either need to use a [query](QUERY.md) or directly the [storage api](STORAGE_API.md).


[top](#creatingmapping-data) - 
[Documentation Overview](README)