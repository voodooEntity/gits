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
  <a href="DOCS/README.md">Documentation</a> •
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
- No third party dependencies
- Option to map nested structures at once (json compatible)


## About
<span style="color:#35b9e9">GITS</span> has been developed in order to enable developers to easily handle complex data structures in their golang applications without having to worry about concurrency. Due to the nature of <span style="color:#35b9e9">GITS</span> handling all storage and operations in memory, it allows for very fast processing of large amounts of datasets and structures. The library also is designed for multithreading purposes and therefor full concurrency safe. While providing a simple  query interface, which probably suits most of the use cases, <span style="color:#35b9e9">GITS</span> also exposes the storage API so the developer can optimize his application without any restrictions.

## Use Cases
The following use cases are example either applications in which i used GITS or ideas that came to my mind in which using GITS could be beneficial. Apart from that, GITS can be used in any environment that can benefit from the features it provides.

- Corona Dashboard
  - Back in the main phase of corona i created my own dashboard hosting and showing the current infection/etc statistics for countries/states/ and other subnational units. A crawler collected information from official sources at an hourly rate and mapped them into the storage. This data was fully kept in memory till i hit my servers limits (20gb ~). 
- Webcrawler
  - Due to GITS offering the mapping of data in graph structure, you can create complex networks - in this case websites/pages and the interlinking in between those. Due to the index being in memory - lookups for already existing entries and such are very fast. 
- File importer
  - You got a very large XML export with different nested sections and ID relations in between those. You can easily use GITS to map the data and than simply retrieve the complete structures for exporting/processing instead of having to use multiple lookups every time.
- NFT trader
  - GITS was used to keep track of NFT trading transactions and evaluate possible arbitrary trades based on the data kept in memory. Since such arbitrary trading is very time sensitive, an in memory storage was optimal. 
- go-cyberbrain
  - A longterm project of myself (which i initially created SlingshotDB for). It's a processing/computing framework which allows for self supervising/automated processing of data.  


## How to use
### Setup
Since GITS is a library, the setup is really simple and only consists of requiring GITS into your existing go project
```bash
go get github.com/voodooEntity/gits@0.8.0
```

### Usage
The following examples provide a sneak peak into the usage of GITS. For a more detailed overview please [visit the documentation](DOCS/README.md).

#### Create an instance
To use GITS first we need to create a new instance
```go
myGitsInstance := gits.NewInstance("main")
```
myGistInstance is a *gits.Gits than can be used to map new data, query the storage or use the direct storage api. 

GITS also keeps track of all storages. The first created storage will always be automatically set as "default" and globally retrievable using gits.GetDefault().

You are free to create as many instances of GITS as you like, in the following examples we are going to focus on having one storage to work with.

#### Datasets
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

Also, since GITS is a graph structured storage, it by default has includes the ability to link any dataset with what we call a "relation" or in graph speech an "edge". Such a relation is structured as following
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

#### Create new data
To make creating new data as easy as possible, GITS provides the "MapData" method which only needs to be provided with an instance of transport.TransportEntity and will take care of the rest. 
```go
rootIntID := myGitsInstance.MapData(transport.TransportEntity{
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

#### Use the Query language
The most simple way to access data in a GITS storage is to use the inbuilt query language builder. While the GITS custom query builder options may be limited, they include the most important options to cover a wide range of use cases. 

An example of how to use it to read the previously mapped data
```go
// Retrieve a query adapter instance. while its not required to store the adapter 
// (you may also just always use myGitsInstance.Query() before calling query methods) 
// it might be handy to reduce the amount of typing
qryAdapter := myGitsInstance.Query()

// now we create a query to read the desired data
finalQuery := qryAdapter.New().Read("Alpha").Match("Value","==","Something").To(
    qryAdapter.New().Read("Beta").Match("Value","==","Else"),
)

// finally we execute the query to retrieve the results. ExecuteQuery 
// is called on the previously created instance of *gits.Gits 
result := qryAdapter.Execute(finalQuery)
```

For more information please refer to ["Query Language Reference and Examples"](DOCS/QUERY.md)


#### Use store API
While it is recommended to primary use the query language when accessing the storage, you are always able to directly interact with the storage.

To access the storage you simply call
```go
storageInstance := myGitsInstance.Storage()
```
and afterwards you can interact with the storage directly
```go
allEntityTypes := storageInstance.GetEntityTypes()
```

For more information please refer to ["Storage API Reference and Examples"](DOCS/STORAGE_API.md)


## Roadmap
The following list contains topics that will be the focus of future major updates. This list is not ordered. 

- [ ] Enhance query capabilities
  - Right now the query builder/language has certain limitations especially in context of methods like "Link","Unlink" and the possibilities to create complex Match(Filter) conditions. This should be enhanced by reworking the query parser and enhancing the query.Query functionality.
  - The current return format for linked data is nested struct instances. While this is very comfortable to work with, it also has certain limitations. Therefor it should also be possible to get the return in a flat format.
- [ ] Enhance test coverage
  - As for now, only the query builder/parser is fully tested. This should be extended to cover as much code as (usefully) possible.
- [ ] Reimplement data persistence capabilities
  - In the past GITS had an in-build custom persistence storage which has been removed. Instead there should be the option to add an adapter for existing persistent storages like pg, mysql or any other. 


## Changelog
### Latest Release Changes
* Massive restructuring of gits in order to make it more accessible and easier to get into for newcomers. This changes are leading towards the first stabled release and include:
  * Enable multiple storages due to factory pattern
  * Add a global tracking of storages and easy accessibility
  * Move the storage to storage.go package and make gits.go the main interface
  * Add simple interface in gits.go to enable query'ing and storage access
  * Remove old persistence implementation
  * Rework README.md and add [full Documentation](./DOCS/README.md)
* Bugfixes
  * Fixing an issue on which a array out of index error could occure on nested Reduce() queries
  * Adjusting Match() and OrMatch() when used with "Properties." to require the results to have the searched Property name.
  * Bugfix addresslist memory exception 

[Full Changelog](CHANGELOG.md) - [Latest Release](https://github.com/voodooEntity/gits/releases/tag/0.0.9)

## License
[Apache License Version 2.0](./LICENSE)

---

> [laughingman.dev](https://blog.laughingman.dev) &nbsp;&middot;&nbsp;
> GitHub [@voodooEntity](https://github.com/voodooEntity)

