package gits

import (
	"fmt"
	"github.com/voodooEntity/gits/src/query"
	"github.com/voodooEntity/gits/src/storage"
	"github.com/voodooEntity/gits/src/transport"
	"log"
	"sync"
)

var instances = make(instanceIndex)
var instanceMutex = &sync.RWMutex{}
var defaultInstance *Gits

type Gits struct {
	Name    string
	storage *storage.Storage
	logs    log.Logger
	//config  *config.Config ### todo consider config
}

func NewInstance(name string) *Gits {
	inst := &Gits{
		Name:    name,
		storage: storage.NewStorage(),
		logs:    log.Logger{},
	}
	instances.Add(name, inst)
	return instances.GetByName(name)
}

func GetDefault() *Gits {
	return instances.GetDefault()
}

func GetByName(name string) *Gits {
	return instances.GetByName(name)
}

func SetDefault(name string) {
	instances.SetDefault(name)
}

func defunc_RemoveInstance(name string) {
	instances.defunc_Remove(name)
	return
}

// deprecated -> use query adapter New() instead
func NewQuery() *query.Query {
	return query.New()
}

func (g *Gits) Storage() *storage.Storage {
	return g.storage
}

func (g *Gits) ExecuteQuery(qry *query.Query) transport.Transport {
	return query.Execute(g.storage, qry)
}

func (g *Gits) MapData(data transport.TransportEntity) transport.TransportEntity {
	return g.storage.MapTransportData(data)
}

func (g *Gits) Query() *QueryAdapter {
	return &QueryAdapter{
		storage: g.storage,
	}
}

type QueryAdapter struct {
	storage *storage.Storage
}

func (qa *QueryAdapter) New() *query.Query {
	return query.New()
}

func (qa *QueryAdapter) Execute(qry *query.Query) transport.Transport {
	return query.Execute(qa.storage, qry)
}

type instanceIndex map[string]*Gits

func (ii instanceIndex) Add(name string, gitsInst *Gits) {
	instanceMutex.Lock()
	if _, ok := instances[name]; ok {
		instanceMutex.Unlock()
		fmt.Println("Name already in use : '" + name + "'")
		return
	}
	instances[name] = gitsInst
	if 1 == len(instances) {
		defaultInstance = gitsInst
	}
	instanceMutex.Unlock()
	return
}

func (ii instanceIndex) GetDefault() *Gits {
	instanceMutex.RLock()
	ret := defaultInstance
	instanceMutex.RUnlock()
	return ret
}

func (ii instanceIndex) SetDefault(name string) {
	instanceMutex.Lock()
	if _, ok := instances[name]; !ok {
		instanceMutex.Unlock()
		fmt.Println("Instance name not existing : '" + name + "'")
		return
	}
	defaultInstance = instances[name]
	instanceMutex.Unlock()
}

func (ii instanceIndex) GetByName(name string) *Gits {
	instanceMutex.RLock()
	ret := instances[name]
	instanceMutex.RUnlock()
	return ret
}

func (ii instanceIndex) Remove(name string) {
	instanceMutex.Lock()
	if _, ok := instances[name]; !ok {
		fmt.Println("Cant remove non existing instance")
		return
	}
	delete(instances, name)
	if defaultInstance.Name == name {
		fmt.Println("Removed active instance, fallback to first")
		if 0 < len(instances) {
			defaultInstance = nil
			for _, gitsInst := range instances {
				defaultInstance = gitsInst
				fmt.Println("Instance set active", gitsInst.Name)
				instanceMutex.Unlock()
				return
			}
		} else {
			defaultInstance = nil
			fmt.Println("No instances left setting active to nil pointer")
		}
	}
}
