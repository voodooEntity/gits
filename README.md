<h1 align="center">
  <br>
  <a href="https://github.com/voodooEntity/go-tachicrypt/"><img src="DOCS/IMAGES/gits_logo.jpeg" alt="GITS" width="300"></a>
  <br>
  GITS
  <br>
</h1>

<h4 align="center">A <span style="color:#35b9e9">G</span>raph <span style="color:#35b9e9">I</span>n-memory <span style="color:#35b9e9">T</span>hread-safe <span style="color:#35b9e9">S</span>torage for easy and fast concurrency safe data handling.</h4>


<p align="center">
  <a href="#key-features">Key Features</a> •
  <a href="#about">About</a> •
  <a href="#use-cases">Use Cases</a> •
  <a href="#how-to-use">How To Use</a> •
  <a href="#roadmap">Roadmap</a> •
  <a href="DOCS/INDEX.md">Documentation</a> •
  <a href="#license">License</a>
</p>

## Key Features
- Full in-memory handling
- Concurrency safe and optimized for multithreading applications
- Native directed graph structure support
- Full access via direct storage api
- Simple builder based query language (json compatible)
- Supports multiple parallel storages (factory)
- Global accessibility of storage index
- Small dependency footprint (using only my own libraries)
- Option to map nested structures at once (json compatible)


## About
<span style="color:#35b9e9">GITS</span> has been developed in order to enable developers to use complex data structures in their golang applications. Due to the nature of <span style="color:#35b9e9">GITS</span> handling all storage and operations in memory, it allows for very fast processing of large amounts of datasets and structures. The library also is designed for multithreading purposes and therefor full concurrency safe. While providing a simple  query interface, which probably suits most of the use cases, <span style="color:#35b9e9">GITS</span> also exposes the storage API so the developer can optimize his application without any restrictions.

## Use Cases
The following use cases are example either of applications in which i used GITS or ideas that came to my mind in which using GITS could be beneficial. Apart from that, GITS can be used in any environment that can benefit from the features it provides.

- Corona Dashboard
  - Back in the main phase of corona i created my own dashboard hosting and showing the current infection/etc statistics for countries/states/ and other subnational units. This data was fully kept in memory till i hit my servers limits (20gb ~). 
- Webcrawler
  - Due to GITS offering the mapping of data in graph structure, you can create complex networks - in this case websites/pages and the interlinking in between those. Due to the index being in memory - lookups for already existing entries and such are very fast. 
- File importer
  - You got a very large XML export with different nested sections and ID relations in between those. You can easily use GITS to map the data and than simply retrieve the complete structures for exporting/processing instead of having to use multiple lookups every time.
- NFT trader
  - GITS was used to keep track of NFT trading transactions and evaluate possible arbitrary trades based on the data kept in memory. 
- go-cyberbrain
  - A longterm project of myself (which i initially created SlingshotDB for). It's a processing/computing framework which allows for self supervising/automated processing of data.  


## How to use
### Setup
The setup is really simple and only consists of requiring the library into your existing go project
```bash
go get github.com/voodooEntity/gits@0.8.0
```


### Create an instance
To use GITS first we need to create a new instance
```go
myGitsInstance := gits.NewInstance("main")
```
myGistInstance is a *gits.Gits than can be used to map new data, query the storage or use the direct storage api. 

GITS also keeps track of all storages. The first created storage will always be automatically set as "default" and globally retrievable using gits.GetDefault().

You are free to create as many instances of GITS as you like, in the following examples we are going to focus on having one storage to work with.

### Datasets
A dataset in GITS is considered an entity, or in graph speech a "node/vertice". When interacting with the data mapping function or query results, the type of structure used is the transport.TransportEntity
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
Each dataset in GITS will be primarily defined by its Type (string), ID (auto increment per type).

While providing a Value:string field to store data, it also holds a Context:string field and the possibility to store further data using the map[string]string Properties field. The Version field is used to keep track of versions when using GITS in a multithreaded environment and should usually not be tempered with. ChildRelations and ParentRelations are pretty self explaining and can hold any number (amount and depth) of further transport.TransportEntities.

This format is mainly used for input/output purposes and the actual stored dataset (types.StorageEntity) is a lot smaller. You will only encounter instances of said type when directly interacting with the storage API.

Also, since GITS is a graph structured storage, it by default has includes the ability to link any dataset with what we call a "relation" or in graph speech an "edge".


### Create new data
To make creating new data as easy as possible, GITS provides the "MapTransport" method which only needs to be provided with an instance of transport.TransportEntity and will take care of the rest. 
```go
rootIntID := myGitsInstance.MapTransport(transport.TransportEntity{
		ID:    storage.MAP_FORCE_CREATE,
		Type:  "Alpha",
		Value: "Something",
		Context: "Example",
		ChildRelations: []transport.TransportRelation{
			{
				Target: transport.TransportEntity{
					ID:    storage.MAP_FORCE_CREATE,
					Type:  "Beta"
					Value: "Else"
					Context: "Example",
				},
			},
		},
	})
```
As you can see in the example, we are passing a nested structure to the MapTransport function. You can map a single dataset or any amount of nested datasets. Since we are using the MAP_FORCE_CREATE constant both datasets will definitely be created. The return of MapTransport is always the ID of the root dataset passed to the method.

For more detailed information on the capabilities of MapTransport please check "MapTransport Examples & Details" addlink

### Use the Query language
The most simple way to access data in a GITS storage is to use the inbuilt query language builder. While the GITS custom query builder options may be limited, they include the most important options to cover a wide range of use cases. 

An example of how to use it to read the previously mapped data
```go
// Retrieve a query builder, since the query builder is unrelated to a storage
// the method is available globally
myQueryBuilder := gits.GetQueryBuilder()

// now we create a query to read the desired data
finalQuery := myQueryBuilder.Read("Alpha").Match("Value","==","Something").To(
	    gits.GetQueryBuilder().Read("Beta").Match("Value","==","Else")
	)

// finally we execute the query to retrieve the results. ExecuteQuery 
// is called on the previously created instance of *gits.Gits 
result := myGitsInstance.ExecuteQuery(finalQuery)
```

For more information please refer to "Query Language Reference and Examples" addlink


### Use store API
While it is recommended to primary use the query language when accessing the storage, you are always able to directly interact with the storage.

To access the storage you simply call
```go
storageInstance := myGitsInstance.Storage()
```
and afterwards you can interact with the storage directly
```go
allEntityTypes := storageInstance.GetEntityTypes()
```

For more information please refer to "Storage API Reference and Examples" addlink


## Roadmap
Right now im at a point of major restructuring. GITS, which initially started as a standalone database (SlingshotDB) and transformed into a library - is undergoing a probably final big restructuring. 

While in its initial form GITS was planned as one big storage, the library has been adjusted to provide as many storages as you need as part of the Restructuring Part 1. In this step the original form of persistence (custom) has been kicked out. While the persistence was working fine, keeping it up to date turned out to be to much of a "sideproject" that i decided to remove it. 

Instead, work towards an implementation that allows storages like mysql, pgsql etc to be used by a simple plugin system - this will be part of Restructuring Part 2. These two major steps in reforming GITS should allow it to be even more modular and finally simplify the "interface" on how to be used. 

These two Restructuring projects are priority above other optimizations. 


- [x] Restructure Part 1
  - Enable multiple storages due to factory pattern
  - Add a global tracking of storages and easy accessibility
  - Move storage storage.go package and make gits.go the main interface
  - Add simple interface in gits.go to enable query'ing and storage access
  - Remove old persistence implementation
  - Adjust README.md and add more necessary documents for full documentation
- [ ] Restructure Part 2
  - Add CHANGELOG.md
  - Implement plugin based persistence system
  - Add config system for persistence etc.
  - Adjust logger used to standard format
- [ ] Enhance test coverage


## Changelog

coming soon

## License
[Apache License Version 2.0](./LICENSE)

---

> [laughingman.dev](https://blog.laughingman.dev) &nbsp;&middot;&nbsp;
> GitHub [@voodooEntity](https://github.com/voodooEntity)

