package persistence

import (
	"fmt"
	"sort"
	"sync"
)

// AdapterConstructor creates a PersistenceAdapter from validated configuration.
type AdapterConstructor func(cfg *PersistenceConfig) (PersistenceAdapter, error)

// AdapterFactory holds registered constructors for named adapter kinds.
type AdapterFactory struct {
	mu           sync.RWMutex
	constructors map[AdapterKind]AdapterConstructor
}

// NewAdapterFactory returns an empty adapter factory.
func NewAdapterFactory() *AdapterFactory {
	return &AdapterFactory{
		constructors: make(map[AdapterKind]AdapterConstructor),
	}
}

// DefaultAdapterFactory is the package-level registry used by built-in adapters and convenience helpers.
var DefaultAdapterFactory = NewAdapterFactory()

// Register adds or replaces a constructor for an adapter kind.
func (f *AdapterFactory) Register(kind AdapterKind, constructor AdapterConstructor) error {
	if kind == "" {
		return fmt.Errorf("persistence: adapter kind is required")
	}
	if constructor == nil {
		return fmt.Errorf("persistence: adapter constructor for %q is required", kind)
	}

	f.mu.Lock()
	defer f.mu.Unlock()
	f.constructors[kind] = constructor
	return nil
}

// Has reports whether a constructor is registered for the given adapter kind.
func (f *AdapterFactory) Has(kind AdapterKind) bool {
	f.mu.RLock()
	defer f.mu.RUnlock()
	_, ok := f.constructors[kind]
	return ok
}

// Kinds returns the sorted set of registered adapter kinds.
func (f *AdapterFactory) Kinds() []AdapterKind {
	f.mu.RLock()
	defer f.mu.RUnlock()

	kinds := make([]AdapterKind, 0, len(f.constructors))
	for kind := range f.constructors {
		kinds = append(kinds, kind)
	}
	sort.Slice(kinds, func(i, j int) bool { return kinds[i] < kinds[j] })
	return kinds
}

// New validates the configuration and creates the selected adapter.
// It returns nil, nil when persistence is disabled or explicitly set to none.
func (f *AdapterFactory) New(cfg PersistenceConfig) (PersistenceAdapter, error) {
	if err := cfg.Validate(); err != nil {
		return nil, err
	}
	if cfg.IsNone() {
		return nil, nil
	}

	f.mu.RLock()
	constructor := f.constructors[cfg.Kind]
	f.mu.RUnlock()

	if constructor == nil {
		return nil, fmt.Errorf("persistence: no adapter registered for kind %q", cfg.Kind)
	}
	return constructor(&cfg)
}

// Register adds or replaces a constructor in DefaultAdapterFactory.
func Register(kind AdapterKind, constructor AdapterConstructor) error {
	return DefaultAdapterFactory.Register(kind, constructor)
}

// NewAdapter creates an adapter using DefaultAdapterFactory.
func NewAdapter(cfg PersistenceConfig) (PersistenceAdapter, error) {
	return DefaultAdapterFactory.New(cfg)
}
