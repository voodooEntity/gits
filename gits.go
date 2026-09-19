package gits

import (
	"context"
	"fmt"
	"github.com/voodooEntity/gits/src/persistence"
	_ "github.com/voodooEntity/gits/src/persistence/filesystem"
	_ "github.com/voodooEntity/gits/src/persistence/postgresql"
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
}

type instanceOptions struct {
	persistenceCfg *persistence.PersistenceConfig
	adapterFactory *persistence.AdapterFactory
}

// InstanceOption customizes a named GITS instance during creation.
type InstanceOption func(*instanceOptions)

// WithPersistence selects an optional per-instance persistence backend and its adapter-specific settings.
func WithPersistence(cfg persistence.PersistenceConfig) InstanceOption {
	return func(o *instanceOptions) {
		o.persistenceCfg = &cfg
	}
}

// WithAdapterFactory overrides the default adapter factory used to construct the selected persistence backend.
func WithAdapterFactory(factory *persistence.AdapterFactory) InstanceOption {
	return func(o *instanceOptions) {
		if factory != nil {
			o.adapterFactory = factory
		}
	}
}

// NewInstance creates a new in-memory GITS instance with the given name.
func NewInstance(name string) *Gits {
	inst := &Gits{
		Name:    name,
		storage: storage.NewStorage(),
		logs:    log.Logger{},
	}
	instances.Add(name, inst)
	return instances.GetByName(name)
}

// NewInstanceWithOptions creates a new GITS instance with optional per-instance persistence configuration.
func NewInstanceWithOptions(name string, opts ...InstanceOption) (*Gits, error) {
	inst := &Gits{
		Name:    name,
		storage: storage.NewStorage(),
		logs:    log.Logger{},
	}

	if err := configureInstancePersistence(inst, opts...); err != nil {
		return nil, err
	}

	instances.Add(name, inst)
	ret := instances.GetByName(name)
	if ret == nil {
		return nil, fmt.Errorf("gits: instance %q was not registered", name)
	}

	// If the name already existed, Add keeps the existing instance. Close the newly created
	// persistence manager so its worker does not outlive the discarded storage.
	if ret != inst && inst.storage.PersistenceManager != nil {
		inst.storage.PersistenceManager.Close(context.Background())
	}

	return ret, nil
}

// NewInstanceWithPersistence creates a new GITS instance with an explicit per-instance persistence configuration.
func NewInstanceWithPersistence(name string, cfg persistence.PersistenceConfig) (*Gits, error) {
	return NewInstanceWithOptions(name, WithPersistence(cfg))
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

// deprecated -> use query adapter New() instead
func NewQuery() *query.Query {
	return query.New()
}

func (g *Gits) Storage() *storage.Storage {
	return g.storage
}

// PersistenceManager returns the optional persistence manager attached to this GITS instance.
func (g *Gits) PersistenceManager() *persistence.PersistenceManager {
	return g.storage.PersistenceManager
}

// Close gracefully shuts down the optional persistence backend for this GITS instance.
// It blocks until all queued mutations have been flushed and then closes the adapter.
func (g *Gits) Close(ctx context.Context) error {
	if pm := g.storage.PersistenceManager; pm != nil {
		return pm.Close(ctx)
	}
	return nil
}

// Shutdown is an alias for Close, provided as a more explicit graceful shutdown name.
func (g *Gits) Shutdown(ctx context.Context) error {
	return g.Close(ctx)
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
	if _, ok := instances[name]; !ok {
		instanceMutex.RUnlock()
		fmt.Println("Instance name not existing : '" + name + "'")
		return nil
	}
	ret := instances[name]
	instanceMutex.RUnlock()
	return ret
}

func (ii instanceIndex) Remove(name string) {
	instanceMutex.Lock()
	defer instanceMutex.Unlock()
	if _, ok := instances[name]; !ok {
		fmt.Println("Cant remove non existing instance")
		return
	}
	delete(instances, name)
	if defaultInstance != nil && defaultInstance.Name == name {
		fmt.Println("Removed active instance, fallback to first")
		defaultInstance = nil
		for _, gitsInst := range instances {
			defaultInstance = gitsInst
			fmt.Println("Instance set active", gitsInst.Name)
			return
		}
		if defaultInstance == nil {
			fmt.Println("No instances left setting active to nil pointer")
		}
	}
}

func configureInstancePersistence(inst *Gits, opts ...InstanceOption) error {
	var o instanceOptions
	for _, opt := range opts {
		if opt != nil {
			opt(&o)
		}
	}

	cfg := persistence.PersistenceConfig{}
	if o.persistenceCfg != nil {
		cfg = *o.persistenceCfg
	}

	if err := cfg.Validate(); err != nil {
		return err
	}
	if cfg.IsNone() {
		return nil
	}

	factory := persistence.DefaultAdapterFactory
	if o.adapterFactory != nil {
		factory = o.adapterFactory
	}

	adapter, err := factory.New(cfg)
	if err != nil {
		return err
	}

	inst.storage.PersistenceManager = persistence.NewManager(adapter, cfg.ChannelBufferSize)

	// Block until the configured backend has loaded all persisted state into memory.
	// This must happen before instances.Add publishes the GITS instance so callers never
	// observe a partially populated persistent storage.
	if err := inst.storage.PersistenceManager.Bootstrap(context.Background(), inst.storage.ApplySnapshot); err != nil {
		closeErr := inst.storage.PersistenceManager.Close(context.Background())
		if closeErr != nil {
			return fmt.Errorf("gits: bootstrap persistence for %q: %v; close: %w", inst.Name, err, closeErr)
		}
		return fmt.Errorf("gits: bootstrap persistence for %q: %w", inst.Name, err)
	}

	return nil
}
