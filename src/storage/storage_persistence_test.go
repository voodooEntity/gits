package storage

import (
	"context"
	"sort"
	"sync"
	"testing"

	"github.com/voodooEntity/gits/src/persistence"
	"github.com/voodooEntity/gits/src/types"
)

type persistenceTestAdapter struct {
	name     string
	snapshot *persistence.Snapshot

	mu      sync.Mutex
	applied []*persistence.PersistenceEvent
	flushed bool
	closed  bool
}

func (a *persistenceTestAdapter) Name() string { return a.name }

func (a *persistenceTestAdapter) LoadAll(ctx context.Context) (*persistence.Snapshot, error) {
	if a.snapshot != nil {
		return a.snapshot, nil
	}
	return persistence.NewSnapshot(), nil
}

func (a *persistenceTestAdapter) ApplyEvent(ctx context.Context, event *persistence.PersistenceEvent) error {
	a.mu.Lock()
	defer a.mu.Unlock()
	a.applied = append(a.applied, event)
	return nil
}

func (a *persistenceTestAdapter) Flush(ctx context.Context) error {
	a.mu.Lock()
	defer a.mu.Unlock()
	a.flushed = true
	return nil
}

func (a *persistenceTestAdapter) Close(ctx context.Context) error {
	a.mu.Lock()
	defer a.mu.Unlock()
	a.closed = true
	return nil
}

func (a *persistenceTestAdapter) appliedCount() int {
	a.mu.Lock()
	defer a.mu.Unlock()
	return len(a.applied)
}

func (a *persistenceTestAdapter) sequences() []uint64 {
	a.mu.Lock()
	defer a.mu.Unlock()
	out := make([]uint64, len(a.applied))
	for i, event := range a.applied {
		out[i] = event.Sequence
	}
	return out
}

func (a *persistenceTestAdapter) createEntityIDs() []int {
	a.mu.Lock()
	defer a.mu.Unlock()
	var ids []int
	for _, event := range a.applied {
		if event.Type == persistence.EventTypeCreateEntity {
			ids = append(ids, event.Entity.ID)
		}
	}
	return ids
}

func (a *persistenceTestAdapter) hasEventTypeAt(index int, want persistence.EventType) bool {
	a.mu.Lock()
	defer a.mu.Unlock()
	if index < 0 || index >= len(a.applied) {
		return false
	}
	return a.applied[index].Type == want
}

func (a *persistenceTestAdapter) isFlushed() bool {
	a.mu.Lock()
	defer a.mu.Unlock()
	return a.flushed
}

func (a *persistenceTestAdapter) isClosed() bool {
	a.mu.Lock()
	defer a.mu.Unlock()
	return a.closed
}

func TestStorageDefaultInMemoryMode(t *testing.T) {
	s := NewStorage()
	if s.PersistenceManager != nil {
		t.Fatalf("NewStorage should leave PersistenceManager nil, got %v", s.PersistenceManager.AdapterName())
	}

	typeID, err := s.CreateEntityType("thing")
	if err != nil {
		t.Fatalf("CreateEntityType: %v", err)
	}
	entityID, err := s.CreateEntity(types.StorageEntity{Type: typeID, Context: "ctx", Value: "v"})
	if err != nil {
		t.Fatalf("CreateEntity: %v", err)
	}

	s.EntityStorageMutex.RLock()
	entity, ok := s.EntityStorage[typeID][entityID]
	s.EntityStorageMutex.RUnlock()
	if !ok || entity.ID != entityID || entity.Type != typeID || entity.Value != "v" {
		t.Fatalf("in-memory storage did not retain created entity: present=%v entity=%+v", ok, entity)
	}
}

func TestStorageBootstrapAppliesSnapshotAndPreservesIDs(t *testing.T) {
	snapshot := persistence.NewSnapshot()
	snapshot.EntityTypes[1] = "thing"
	snapshot.EntityRTypes["thing"] = 1
	snapshot.EntityTypeIDMax = 1
	snapshot.EntityIDMax[1] = 42
	snapshot.Entities = append(snapshot.Entities, types.StorageEntity{
		ID:         42,
		Type:       1,
		Context:    "ctx",
		Value:      "v",
		Properties: map[string]string{"k": "v"},
		Version:    3,
	})
	snapshot.Relations = append(snapshot.Relations, types.StorageRelation{
		SourceType: 1,
		SourceID:   42,
		TargetType: 1,
		TargetID:   43,
		Context:    "rel",
		Properties: map[string]string{"rk": "rv"},
		Version:    7,
	})

	fake := &persistenceTestAdapter{name: "storage-bootstrap", snapshot: snapshot}
	s := NewStorage()
	s.PersistenceManager = persistence.NewManager(fake, 4)
	defer s.PersistenceManager.Close(context.Background())

	if err := s.PersistenceManager.Bootstrap(context.Background(), s.ApplySnapshot); err != nil {
		t.Fatalf("Bootstrap: %v", err)
	}

	s.EntityTypeMutex.RLock()
	typeName := s.EntityTypes[1]
	reverseTypeID := s.EntityRTypes["thing"]
	typeIDMax := s.EntityTypeIDMax
	s.EntityTypeMutex.RUnlock()
	if typeName != "thing" || reverseTypeID != 1 || typeIDMax < 1 {
		t.Fatalf("bootstrap did not restore entity types: name=%q reverse=%d max=%d", typeName, reverseTypeID, typeIDMax)
	}

	s.EntityStorageMutex.RLock()
	entity, ok := s.EntityStorage[1][42]
	maxEntityID := s.EntityIDMax[1]
	s.EntityStorageMutex.RUnlock()
	if !ok || entity.ID != 42 || entity.Type != 1 || entity.Value != "v" || entity.Version != 3 || entity.Properties["k"] != "v" {
		t.Fatalf("bootstrap did not preserve entity identity/state: present=%v entity=%+v", ok, entity)
	}
	if maxEntityID < 42 {
		t.Fatalf("bootstrap EntityIDMax[1] = %d, want at least 42", maxEntityID)
	}

	s.RelationStorageMutex.RLock()
	relation, ok := s.RelationStorage[1][42][1][43]
	reverseOK := s.RelationRStorage[1][43][1][42]
	s.RelationStorageMutex.RUnlock()
	if !ok || relation.SourceType != 1 || relation.SourceID != 42 || relation.TargetType != 1 || relation.TargetID != 43 ||
		relation.Context != "rel" || relation.Version != 7 || relation.Properties["rk"] != "rv" {
		t.Fatalf("bootstrap did not preserve relation identity/state: present=%v relation=%+v", ok, relation)
	}
	if !reverseOK {
		t.Fatal("bootstrap did not restore reverse relation index")
	}

	nextID, err := s.CreateEntity(types.StorageEntity{Type: 1, Context: "next", Value: "next"})
	if err != nil {
		t.Fatalf("CreateEntity after bootstrap: %v", err)
	}
	if nextID != 43 {
		t.Fatalf("new entity ID = %d, want 43 to prove persisted high-water mark was restored", nextID)
	}
}

func TestStoragePersistenceSequentialMutationOrder(t *testing.T) {
	fake := &persistenceTestAdapter{name: "storage-sequential"}
	s := NewStorage()
	s.PersistenceManager = persistence.NewManager(fake, 16)
	defer s.PersistenceManager.Close(context.Background())

	ctx := context.Background()
	typeID, err := s.CreateEntityType("thing")
	if err != nil {
		t.Fatalf("CreateEntityType: %v", err)
	}
	for i := 0; i < 5; i++ {
		if _, err := s.CreateEntity(types.StorageEntity{Type: typeID, Context: "ctx", Value: "v"}); err != nil {
			t.Fatalf("CreateEntity[%d]: %v", i, err)
		}
	}

	if err := s.PersistenceManager.Flush(ctx); err != nil {
		t.Fatalf("Flush: %v", err)
	}

	sequences := fake.sequences()
	wantTotal := 6
	if len(sequences) != wantTotal {
		t.Fatalf("applied event count = %d, want %d", len(sequences), wantTotal)
	}
	for i, sequence := range sequences {
		if sequence != uint64(i+1) {
			t.Fatalf("sequence[%d] = %d, want %d", i, sequence, i+1)
		}
	}
	if !fake.hasEventTypeAt(0, persistence.EventTypeCreateEntityType) {
		t.Fatal("first applied event should be entity type creation")
	}
	for i := 1; i <= 5; i++ {
		if !fake.hasEventTypeAt(i, persistence.EventTypeCreateEntity) {
			t.Fatalf("applied event[%d] should be entity creation", i)
		}
	}

	createIDs := fake.createEntityIDs()
	if len(createIDs) != 5 {
		t.Fatalf("create entity event count = %d, want 5", len(createIDs))
	}
	for i, id := range createIDs {
		if id != i+1 {
			t.Fatalf("persisted entity ID[%d] = %d, want %d", i, id, i+1)
		}
	}
}

func TestStoragePersistenceCloseDrainsQueuedMutations(t *testing.T) {
	fake := &persistenceTestAdapter{name: "storage-close"}
	s := NewStorage()
	manager := persistence.NewManager(fake, 4)
	s.PersistenceManager = manager

	ctx := context.Background()
	typeID, err := s.CreateEntityType("thing")
	if err != nil {
		t.Fatalf("CreateEntityType: %v", err)
	}
	for i := 0; i < 3; i++ {
		if _, err := s.CreateEntity(types.StorageEntity{Type: typeID, Context: "ctx", Value: "v"}); err != nil {
			t.Fatalf("CreateEntity[%d]: %v", i, err)
		}
	}

	if err := manager.Close(ctx); err != nil {
		t.Fatalf("Close: %v", err)
	}

	if count := fake.appliedCount(); count != 4 {
		t.Fatalf("applied event count after close = %d, want 4", count)
	}
	if !fake.isFlushed() {
		t.Fatal("adapter Flush was not called during Close")
	}
	if !fake.isClosed() {
		t.Fatal("adapter Close was not called during manager Close")
	}
	if err := manager.Close(ctx); err != nil {
		t.Fatalf("second Close: %v", err)
	}
}

func TestStoragePersistenceConcurrentMutationsAreRaceSafeAndOrdered(t *testing.T) {
	fake := &persistenceTestAdapter{name: "storage-concurrent"}
	s := NewStorage()
	s.PersistenceManager = persistence.NewManager(fake, 1024)
	defer s.PersistenceManager.Close(context.Background())

	ctx := context.Background()
	typeID, err := s.CreateEntityType("thing")
	if err != nil {
		t.Fatalf("CreateEntityType: %v", err)
	}

	const goroutines = 8
	const perGoroutine = 25
	expectedEntities := goroutines * perGoroutine

	var wg sync.WaitGroup
	var mu sync.Mutex
	var testErr error
	recordErr := func(err error) {
		if err == nil {
			return
		}
		mu.Lock()
		if testErr == nil {
			testErr = err
		}
		mu.Unlock()
	}

	for g := 0; g < goroutines; g++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for i := 0; i < perGoroutine; i++ {
				if _, err := s.CreateEntity(types.StorageEntity{Type: typeID, Context: "ctx", Value: "v"}); err != nil {
					recordErr(err)
					return
				}
			}
		}()
	}
	wg.Wait()
	if testErr != nil {
		t.Fatalf("concurrent CreateEntity error: %v", testErr)
	}

	if err := s.PersistenceManager.Flush(ctx); err != nil {
		t.Fatalf("Flush after concurrent mutations: %v", err)
	}

	wantTotal := expectedEntities + 1
	sequences := fake.sequences()
	if len(sequences) != wantTotal {
		t.Fatalf("applied event count = %d, want %d", len(sequences), wantTotal)
	}
	for i, sequence := range sequences {
		if sequence != uint64(i+1) {
			t.Fatalf("sequence[%d] = %d, want %d", i, sequence, i+1)
		}
	}

	createIDs := fake.createEntityIDs()
	if len(createIDs) != expectedEntities {
		t.Fatalf("create entity event count = %d, want %d", len(createIDs), expectedEntities)
	}
	sort.Ints(createIDs)
	for i, id := range createIDs {
		if id != i+1 {
			t.Fatalf("persisted entity ID[%d] = %d, want %d", i, id, i+1)
		}
	}

	s.EntityStorageMutex.RLock()
	count := len(s.EntityStorage[typeID])
	for id := 1; id <= expectedEntities; id++ {
		if _, ok := s.EntityStorage[typeID][id]; !ok {
			s.EntityStorageMutex.RUnlock()
			t.Fatalf("in-memory storage missing entity ID %d", id)
		}
	}
	s.EntityStorageMutex.RUnlock()
	if count != expectedEntities {
		t.Fatalf("in-memory entity count = %d, want %d", count, expectedEntities)
	}
}
