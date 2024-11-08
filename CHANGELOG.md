# Changelog

## Major restructuring : 0.9.0  `8.11.2024`
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

## Small Bugfix Latest `6.8.2024`
* Introducing a temporary variable when running recursive queries to collect tmpResultAddresses and adding them afterwards to the total list in order to prevent exponential allocation of addresses - in certain cases this could result in the application using extreme amounts of memory.
* adding some manual testcode

##  Minor adjustment `21.4.2024`
* Adjusting some test functions
* Changing a log statement from Info to Debug

## Bugfix  `24.9.2023`
* Bugfix for DeleteParentRelationsUnsafe which did non unsafe calls resulting in mutex deadlock

##  Bugfix `24.9.2023`
* Bugfix for deleteEntityUnsafe which did non unsafe calls resulting in mutex deadlock

##  Bugfixes `1.5.2023`
* Bugfixing an issue that prevented from using gits.MapTransportData to map towards parents 
* Bugfixing an issue that prevented from correctly getting the return data on subquerying towards parents

##  Adding Traverse functionality to query `8.1.2023`
* Adding TraverseIn(depth) and TraverseOut(depth) functionality to query.
* Also added some tests for query.go package.

##  Fixing Fatal bug in update `17.10.2022`
* Fixing a fatal error that could occur on duplicate release of same mutex

##  Fixing deadlock issue on mapTransportData method `6.10.2022`
* Fixing a possible mutex deadlock that could occure while using the mapTransportData when providing a map by value which value is not existing yet.

##  Add map by value case to mapTransportData `1.10.2022`
* Adding a new behaviour to the mapTransportData function. Additional to creating a dataset by setting the Entities ID to -1 , 0 is now a valid value for an entity's ID in a structure given to mapTransportData. If the ID is 0 gits will search for an Entity of given Type with given value, and if found map to it. Should there not be an entity available with given Type and Value the mapTransportData will behave the same way as on ID == -1 and create a new entity.

##  Minor refactoring `1.10.2022`
* Adding Version field to transport.TransportRelation struct

##  Minor refactoring `1.10.2022`
* Refactor GetEntityTypes from array to map return to preserve correct ID keys (same with unsafe variant)

## Minor refactoring `1.10.2022`
* Minor refactoring - changing return of GetTypeStringById and GetTypeStringByIdUnsafe from *string to string

##  Adding Query and mapdata `7.9.2022`
* Adding basic query functionality and mapdata provided by gits.

## Fix possible concurrency flaw `6.7.2022`
*  getEntityByPath did Unlock the storage mutex to early which possibly could create a concurrency flaw. So deepcopied the entity before unlocking and returning the copy.

##  Hotfix release `3.7.2022`
* Hotfix. In case that application using gits whas shutdown before persistency could complete a write operation it might happen that a relation existsd from or to an entity that doesnt exist. that results in a hard shutdown/exit on import. Hotfixing this by skipping the relation in case this happens

##  Mitigate rising import times `3.7.2022`
* Mitigate the rising import times by optimizing persistent data to parse in further imports while doing import.

##  Adding some simple storage statistics getters `16.8.2022`
* Adding following function to the gits storage
  * GetAmountPersistencePayloadsPending | returns INT count of Payloads still in the persistence channel waiting to be stored on disk
  * GetEntityAmount | returns INT count of total amount of entities existing
  * GetEntityAmountByType | returns INT count of total amount of entities existing of a specific entityType

##  Extending CreateEntityUniqueValue return `10.8.2022`
* Adding a bool return flag to CreateEntityUniqueValue telling if a new dataset has been created, since when an entity with the exact same value already exists the current behaviour is the function will return the existing datasets ID
* Maybe rename the function in future to something like  'CreateEntityUniqueValueIfNotExists'

##   Adding new helper method `8.6.2022`
* adding new helper method

##  fixed flaw in recovering persistent relations `8.6.2022`
* fixed relation persistance recovery flaw

##  Updated dependency on logger `8.6.2022`
* changing version dependency on archivist

## Initial alpha `8.6.2022`
* some additions 
* remove debug printing 
* fixing restore bug