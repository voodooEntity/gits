# [gits] graph in-memory threadsafe storage 

### Table of contents
- [About](#about)
- [Usage]
- [Contributing]
- [Future plans]
- [FAQ]
- [Links]

## About
Welcome to the home of gits - an in-memory graph storage. It's completly written in golang and provides access by using included custom query package or by directly accessing the storage. 

The main target of the storage is to have a easy to use high performance graph storage. In it's current state the storage is shipped with basic functionality. I will extend the functionality over time based on needs and time.

`Important!` Since the storage is in memory and focused on high performance it can use a lot of memory for indexes and data storage. Using the storage you should make sure to run it on a machine that provides enough memory to work fine. While there are several ways how i could reduce the memory usage i accepted this trade off to achieve a better performance.

Finally i wanne leave a special thanks to some friends that helped through the process of creating this software by listening to hours of rage/ideas and providing suggestions that lead the way to the software you are about to use.
* Maze 
* Luxer
* f0o

I hope you will enjoy the usage of gits and that this will just be the start of a great project.

Sincerely yours,
voodooEntity aka laughingman

