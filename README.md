# [gits] graph in-memory threadsafe storage 

### Table of contents
- [About](#about)
- [Usage](#usage)
- [Contributing](#contributing)
- [Future plans](#future-plans)
- [FAQ](#faq)
- [Links](#links)

## About
Welcome to the home of gits - an in-memory graph storage. It's completely written in golang and provides access by using included custom query package or by directly accessing the storage. 

The main target of the storage is to have an "easy to use" high performance graph storage. In its current state the storage is shipped with basic functionality. I will extend the functionality over time based on needs and time.

Originally I started developing the storage as part of the [SlingshotDB](https://github.com/voodooEntity/slingshotdb) project. Back than it was my target to create an alternative to existing graph databases since none of them met my needs. Later on i had multiple use cases for the storage in other software tho I didn't want to access it via rest, so I extracted the base storage from SlingshotDB and created a standalone package - gits.

The code is written completely vanilla with only one dependency which is the [archivist](https://github.com/voodooEntity/archivist) package. This package is also developed and maintained by myself and provides a minimal and simple logger. This means you not going to have any compatibility issues due to tons of external library dependencies, but rather have only one maintainer to deal with.

While the database is focussed on high performance in memory operations, it also includes the option to enable persisting the storage on the disk in form of a log-like file format. The persistence is asynchronous, so it doesn't block the storage itself. For further information check the [usage](#usage). 

`Important!` Since the storage is in memory and focused on high performance it can use a lot of memory for indexes and data storage. Using the storage you should make sure to run it on a machine that provides enough memory to work fine. While there are several ways how i could reduce the memory usage i accepted this trade off to achieve a better performance.

Finally i wanne leave a special thanks to some friends that helped through the process of creating this software by listening to hours of rage/ideas and providing suggestions that lead the way to the software you are about to use.
* Maze 
* Luxer
* f0o

I hope you will enjoy the usage of gits and that this will just be the start of a great project.

Sincerely yours,
voodooEntity aka laughingman

## Usage
To use gits in your application first of all you need to require the package into your existing application. This can be done by executing the following command in your projects root directory.
```bash
go get github.com/voodooEntity/gits
```
This would fetch the latest release from pkg.go.dev - for a more stable use you should require a specific release like 
```bash
go get github.com/voodooEntity/gits@v0.0.19
```
After loading the package as dependency you can initialize the storage in your go code. This can be done as shown next
```go
gits.Init(types.PersistenceConfig{
    RotationEntriesMax:           1000000,
    Active:                       false,
    PersistenceChannelBufferSize: 10000000,
})
```
The "gits.Init" call will boot the storage and expects an instance of types.PersistenceConfig struct. To enable the persistence just change the Active flag from false to true. The storage will create a "storage" directory and necessary substructure. When persistence enabled, the storage will on init check for existing persistence logs and import them.

As mentioned in the [about](#about) section the persistence is implemented asynchronous which means that there always is a very small delay between in memory changes and on disk persistence. For more details about the way the persistence works check the [persistence](https://github.com/voodooEntity/gits/###) page.

After initializing the storage all the functions associated to the storage are global accessible to every package importing the gits packages. The storage is implemented fully threadsafe using multiple mutex. Therefor you can safely use this storage in multithreading applications without having to care about concurrency problems.

Now there are two main ways on how to use the storage. Follow the links for detailed information on 
- Using the custom [query package](https://github.com/voodooEntity/gits/docs/) (recommended for beginners)
- Using the [storage functions](https://github.com/voodooEntity/gits/docs/) (recommended for experienced & brave)

## Contributing

## Future plans

## FAQ

## Links