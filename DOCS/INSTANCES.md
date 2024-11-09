# Instance Handling
GITS allows the user to create as many "instances" as he pleases to. In most use cases one instance should be sufficient, but in this document i will explain the possibilities and options GITS provides in terms of instance handling.


## Definition
Each gits.Gits instance will hold an *storage.Storage to a newly create Storage instance. It also exposes several methods to interact with said storage, those are explained in detail in the other topics.

Means, with every instance you create, you will create a complete storage instance underlying that is separated from the others. The Gits instance itself holds this Storage and provides an easy combined interface for Mapping / Query'ing Data while also providing the option to directly access the Storage directly.

## Index
* [Definition](#definition)
* [API](#api)
* [Usage](#usage)
* [FAQ](#faq)


## API
The GITS package provides the following public functions in terms of instance handling
```go
func NewInstance(name string) *Gits 
func GetDefault() *Gits 
func GetByName(name string) *Gits 
func SetDefault(name string) 
func GetQueryBuilder() *query.Query 
```

## Usage
So the first step when working with GITS is to create an instance. 

Note: the first instance that is created when running the application will automatically be marked as "default" - this can be changed.
```go
myGitsInstance := gits.NewInstance("main")
```
as shown in the API listing, this will return a *gits.Gits

This pointer can be stored for later usage tho it is not the only way to access an existing storage.

When for example using
```go
newVariable := gits.GetDefault()
```
we would retrieve a *gits.Gits pointer towards the current default instance. Since we only created one this would be the "main".

Another way to access an instance globally is by its defined name
```go
anotherVariable := gits.GetByName("main")
```
as you can see we passed the name "main" which we also used when creating the instance in first place. This way you can retrieve a pointer to all existing instances you just are required to provide the name.

For the next example we first create a second storage
```go
fourthVar := gits.NewInstance("secondary")
```
Note: In contrast to the first time we create a new instance, this one will not be automatically set default. We only set the first instance to default, afterwards this has to be altered manually.

Working with multiple GITS instances, you may need to adjust the instance which is set as default. You can achieve this by calling
```go
gits.SetDefault("secondary")
```
this would now change the default instance to the freshly created one - impacting that calling 
```go
secondaryInstance := gits.GetDefault()
```
would return a pointer to the "secondary" named instance since we just changed the default.

So you basically can always access your storages from wherever you want by addressing them with their name - you can handle a default instance and you can always store and handle the pointers in your application.

The system is designed to keep you as free in your choice of usage as possible.


## FAQ
Q: Are instance names unique?
A: Yes, you can not have two instance with the same name.

Q: Why is the light blue
A: .....



[top](#instance-handling) - 
[Documentation Overview](README.md)