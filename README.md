<h1 align="center">
  <br>
  <a href="https://github.com/voodooEntity/go-tachicrypt/"><img src="DOCS/IMAGES/gits_logo.jpeg" alt="GITS" width="300"></a>
  <br>
  GITS
  <br>
</h1>

<h4 align="center">A <span style="color:#35b9e9">G</span>raph <span style="color:#35b9e9">I</span>n-memory <span style="color:#35b9e9">T</span>hread-safe <span style="color:#35b9e9">S</span>torage library for easy and fast concurrency safe data handling.</h4>


<p align="center">
  <a href="#key-features">Key Features</a> •
  <a href="#about">About</a> •
  <a href="#use-cases-aka-why-gits">Use Cases</a> •
  <a href="#how-to-use">How To Use</a> •
  <a href="#roadmap">Roadmap</a> •
  <a href="DOCS/README.md">Documentation</a> •
  <a href="#license">License</a>
</p>

## Key Features
- Full in-memory handling
- Concurrency safe and optimized for multithreading applications
- Easy integration as a standalone library
- No third party dependencies
- Simple builder based query language (json compatible)
- Option to create/map nested structures at once (json compatible)
- Native directed graph structure support
- Full access via direct storage api
- Supports multiple parallel storages (factory)
- Global accessibility of storage index


## About
GITS is designed to simplify the management of complex data structures in Go applications, eliminating the need for manual concurrency handling. By storing and operating on data entirely in memory, GITS enables rapid processing of large datasets and intricate structures. The library is inherently thread-safe, making it suitable for multithreaded environments.

While offering a straightforward query interface for most use cases, GITS also exposes an underlying storage API for advanced optimization. This flexibility allows GITS to accommodate a broad spectrum of applications, from high-performance in-memory object storage to intricate information network mapping.

## Use Cases Examples aka Why GITS?

### 1. Real-time Processing & Live Analytics (Volatile Data)

GITS is perfect for scenarios where data is constantly flowing, quickly generated, or only relevant for a short period, allowing for immediate insights without the overhead of disk I/O.

* **Real-time Anomaly Detection in Streaming Data:**
  * **How:** Ingest event streams (e.g., financial transactions, IoT sensor data, system logs), building a temporary graph of interconnected entities (users, devices, IPs) and their interactions within a sliding time window.
  * **Why GITS:** Rapidly traverse these dynamic relationships to identify unusual patterns, suspicious activity, or system deviations as they happen. The graph is ephemeral, focusing on the current state for immediate action, not historical storage.
* **Dynamic Workflow Execution & Dependency Resolution:**
  * **How:** For complex, short-lived computational workflows or batch processing jobs, model individual tasks and their execution dependencies (e.g., "Task B requires output from Task A," "Task C can run in parallel with Task D").
  * **Why GITS:** Concurrently update task states and quickly query the graph to determine the next executable tasks, manage conditional branching, and identify bottlenecks in real-time. The workflow graph exists only for the duration of its execution.
* **Live Network Flow Analysis:**
  * **How:** Ingest real-time network packet data, building a graph of connections between IP addresses, ports, and protocols for a rolling time window (e.g., the last 5 minutes).
  * **Why GITS:** Instantly detect anomalies, identify denial-of-service (DoS) attacks, or visualize active connections. The in-memory nature allows for extremely high data ingestion and query throughput, and older flow data can be gracefully discarded as new data arrives.

### 2. Temporary Data Staging & Complex Transformations

Use GITS as a powerful, intermediate workspace for processing highly connected data before it's moved, stored elsewhere, or discarded.

* **Advanced Data Ingestion & ETL Pipelines:**
  * **How:** Ingest complex, semi-structured data (e.g., large XML exports, sensor logs, historical transaction batches) into GITS. Use its graph capabilities to clean, enrich, deduplicate (by finding linked duplicate records), or restructure data by traversing relationships, before exporting it to a traditional database or data warehouse.
  * **Why GITS:** Perform intricate, graph-based transformations and lookups that are cumbersome or slow with traditional relational databases, acting as a high-speed, transient staging area for the "Transform" phase of Extract, Transform, Load (ETL) processes.
* **Web Crawler Session Graph:**
  * **How:** Map crawled web pages, their internal and external links, and encountered assets (images, scripts) during a *single crawling session*.
  * **Why GITS:** Rapidly check for already visited pages, identify link structures, or find new paths to explore without hitting disk I/O. The graph is built for the current crawl, and its findings (e.g., unique URLs, link statistics) are exported or summarized at the end.
* **Complex Document Parsing & Flattening:**
  * **How:** Load a deeply nested or interconnected document (e.g., a massive JSON/XML report with many internal references) into GITS, creating entities for sections and relations for references.
  * **Why GITS:** Easily navigate and "flatten" complex document structures by traversing relations, allowing for quick extraction of specific sub-graphs or reconstruction of complete objects, without needing to maintain the full memory of nested pointers.

### 3. Graph-based Caching for Rapid Lookups

GITS can serve as an ultra-fast, in-memory cache for pre-computed or frequently accessed graph data loaded from a slower, persistent source.

* **Pre-calculated Recommendation Cache:**
  * **How:** An offline recommendation engine (running on a persistent database) pre-calculates a large graph of "users who might like X" or "items frequently bought together." This graph is loaded into GITS on application startup.
  * **Why GITS:** Serve personalized recommendations with lightning-fast latency directly from memory, handling high request volumes. If the application restarts, the cache can be quickly reloaded from its persistent source.
* **Dynamic Authorization & Policy Cache:**
  * **How:** Load frequently accessed security policies, user roles, resource permissions, and complex access rule relationships from a persistent database into GITS on application startup.
  * **Why GITS:** Perform rapid authorization checks across multiple concurrent requests. The in-memory graph provides sub-millisecond lookup times for complex "who can access what" questions, avoiding repeated, slow database queries.



## How to use
### Setup
Since GITS is a library, the setup is really simple and only consists of requiring GITS into your existing go project
```bash
go get github.com/voodooEntity/gits@0.9.7
```

### Usage
The following examples provide a sneak peak into the usage of GITS. For a more detailed overview please [visit the documentation](DOCS/README.md).

#### Creating an instance
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

#### Creating new data
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

For more detailed information on the capabilities of MapTransport please check [MapTransport Examples & Details](DOCS/DATA_MAPPING.md#examples)

#### Using the Query Builder
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


#### Using the Storage API
While it is recommended to primary use the query builder when accessing the storage, you are always able to directly interact with the storage.

To access the storage you simply call
```go
storageInstance := myGitsInstance.Storage()
```
and afterwards you can interact with the storage directly
```go
allEntityTypes := storageInstance.GetEntityTypes()
```

For more information please refer to ["Storage API Reference and Examples"](DOCS/STORAGE_API.md)

#### Understanding the Storage Architecture
While GITS is designed for optimal performance through in-memory handling and optimized indexing, understanding the underlying storage architecture can further enhance its usage. For a detailed explanation of this architecture, please refer to the following documentation: [Storage Architecture](DOCS/STORAGE_ARCHITECTURE.md)

## Roadmap
The following list contains topics that will be the focus of future major updates. This list is not ordered. These are planned after 1.0.0 stable release. 

- [ ] Persistence option
  - Reimplement (optional) asynchronous persistence via interface, while also providing a package delivering a simple file based persistence  
- [ ] Enhance test coverage
  - As for now, only the query builder/parser is mostly fully tested (91%). This should be extended to cover as much code as (usefully) possible.
- [ ] Enhance query capabilities
  - Right now the query builder/language has certain limitations especially in context of methods like "Link","Unlink" and the possibilities to create complex Match(Filter) conditions. This should be enhanced by reworking the query parser and enhancing the query.Query functionality.
  - The current return format for linked data is nested struct instances. While this is very comfortable to work with, it also has certain limitations. Therefor it should also be possible to get the return in a flat format.


## Changelog
### Latest Release Changes

* Adding CascadeIn(depth int) and CascadeOut(depth) mthods to Query struct, which can be used to have deletes cascade over multiple levels.
* Adding tests for CascadeIn and CascadeOut
* Adding base info for Cascade methods to docs



[Full Changelog](CHANGELOG.md) - [Latest Release](https://github.com/voodooEntity/gits/releases/tag/v0.9.7)

## License
[GNU General Public License v3.0](./LICENSE)

---

> [laughingman.dev](https://blog.laughingman.dev) &nbsp;&middot;&nbsp;
> GitHub [@voodooEntity](https://github.com/voodooEntity)

