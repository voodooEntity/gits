package gits

import (
    "context"
    "fmt"
    "sync/atomic"
    "testing"

    "github.com/voodooEntity/gits/src/persistence"
    "github.com/voodooEntity/gits/src/types"
)

type selectionFakeAdapter struct {
    name string
}

func (a *selectionFakeAdapter) Name() string { return a.name }

func (a *selectionFakeAdapter) LoadAll(ctx context.Context) (*persistence.Snapshot, error) {
    return persistence.NewSnapshot(), nil
}

func (a *selectionFakeAdapter) ApplyEvent(ctx context.Context, event *persistence.PersistenceEvent) error {
    return nil
}

func (a *selectionFakeAdapter) Flush(ctx context.Context) error { return nil }

func (a *selectionFakeAdapter) Close(ctx context.Context) error { return nil }

type bootstrapFakeAdapter struct {
    name        string
    snapshot    *persistence.Snapshot
    loadErr     error
    loadStarted chan struct{}
    releaseLoad chan struct{}
    loaded      int32
    closed      int32
}

func (a *bootstrapFakeAdapter) Name() string { return a.name }

func (a *bootstrapFakeAdapter) LoadAll(ctx context.Context) (*persistence.Snapshot, error) {
    if a.loadStarted != nil {
        close(a.loadStarted)
    }
    select {
    case <-a.releaseLoad:
    case <-ctx.Done():
        return nil, ctx.Err()
    }
    atomic.StoreInt32(&a.loaded, 1)
    return a.snapshot, a.loadErr
}

func (a *bootstrapFakeAdapter) ApplyEvent(ctx context.Context, event *persistence.PersistenceEvent) error {
    return nil
}

func (a *bootstrapFakeAdapter) Flush(ctx context.Context) error { return nil }

func (a *bootstrapFakeAdapter) Close(ctx context.Context) error {
    atomic.StoreInt32(&a.closed, 1)
    return nil
}

func TestGitsPerStoragePersistenceSelection(t *testing.T) {
    factory := persistence.NewAdapterFactory()
    if err := factory.Register(persistence.AdapterKindFilesystem, func(cfg *persistence.PersistenceConfig) (persistence.PersistenceAdapter, error) {
        return &selectionFakeAdapter{name: "filesystem:" + cfg.Filesystem.Directory}, nil
    }); err != nil {
        t.Fatalf("register filesystem constructor: %v", err)
    }
    if err := factory.Register(persistence.AdapterKindPostgreSQL, func(cfg *persistence.PersistenceConfig) (persistence.PersistenceAdapter, error) {
        return &selectionFakeAdapter{name: fmt.Sprintf("postgresql:%s:%d/%s", cfg.PostgreSQL.Host, cfg.PostgreSQL.Port, cfg.PostgreSQL.Database)}, nil
    }); err != nil {
        t.Fatalf("register postgresql constructor: %v", err)
    }

    inMemoryName := "gits-persist-selection-inmemory"
    defer instances.Remove(inMemoryName)
    inMem := NewInstance(inMemoryName)
    if inMem == nil {
        t.Fatalf("NewInstance returned nil")
    }
    if inMem.PersistenceManager() != nil {
        t.Fatalf("default instance should have no persistence manager, got %v", inMem.PersistenceManager().AdapterName())
    }

    dir := t.TempDir()
    fsExpected := "filesystem:" + dir
    fsName := "gits-persist-selection-filesystem"
    defer instances.Remove(fsName)
    fs, err := NewInstanceWithOptions(
        fsName,
        WithPersistence(persistence.PersistenceConfig{
            Enabled:           true,
            Kind:              persistence.AdapterKindFilesystem,
            ChannelBufferSize: 1,
            Filesystem:        &persistence.FilesystemConfig{Directory: dir},
        }),
        WithAdapterFactory(factory),
    )
    if err != nil {
        t.Fatalf("NewInstanceWithOptions filesystem: %v", err)
    }
    if fs.PersistenceManager() == nil {
        t.Fatalf("filesystem instance should have persistence manager")
    }
    defer fs.PersistenceManager().Close(context.Background())
    if got := fs.PersistenceManager().AdapterName(); got != fsExpected {
        t.Fatalf("filesystem adapter name = %q, want %q", got, fsExpected)
    }

    pgExpected := "postgresql:localhost:5432/gitsdb"
    pgName := "gits-persist-selection-postgresql"
    defer instances.Remove(pgName)
    pg, err := NewInstanceWithOptions(
        pgName,
        WithPersistence(persistence.PersistenceConfig{
            Enabled:           true,
            Kind:              persistence.AdapterKindPostgreSQL,
            ChannelBufferSize: 1,
            PostgreSQL: &persistence.PostgresConfig{
                Host:     "localhost",
                Port:     5432,
                User:     "gits_user",
                Password: "secret",
                Database: "gitsdb",
            },
        }),
        WithAdapterFactory(factory),
    )
    if err != nil {
        t.Fatalf("NewInstanceWithOptions postgresql: %v", err)
    }
    if pg.PersistenceManager() == nil {
        t.Fatalf("postgresql instance should have persistence manager")
    }
    defer pg.PersistenceManager().Close(context.Background())
    if got := pg.PersistenceManager().AdapterName(); got != pgExpected {
        t.Fatalf("postgresql adapter name = %q, want %q", got, pgExpected)
    }

    noneName := "gits-persist-selection-none-convenience"
    defer instances.Remove(noneName)
    none, err := NewInstanceWithPersistence(noneName, persistence.PersistenceConfig{})
    if err != nil {
        t.Fatalf("NewInstanceWithPersistence none: %v", err)
    }
    if none.PersistenceManager() != nil {
        t.Fatalf("none convenience instance should have no persistence manager")
    }

    invalidName := "gits-persist-selection-invalid"
    if _, err := NewInstanceWithOptions(
        invalidName,
        WithPersistence(persistence.PersistenceConfig{Enabled: true, Kind: persistence.AdapterKindFilesystem}),
        WithAdapterFactory(factory),
    ); err == nil {
        t.Fatalf("expected invalid filesystem config to fail")
    }
    if instances.GetByName(invalidName) != nil {
        t.Fatalf("invalid instance should not be registered")
    }
}

func TestGitsBootstrapBlocksUntilLoadCompletesAndPreservesIDs(t *testing.T) {
    dir := t.TempDir()
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

    adapter := &bootstrapFakeAdapter{
        name:        "filesystem:" + dir,
        snapshot:    snapshot,
        loadStarted: make(chan struct{}),
        releaseLoad: make(chan struct{}),
    }
    factory := persistence.NewAdapterFactory()
    if err := factory.Register(persistence.AdapterKindFilesystem, func(cfg *persistence.PersistenceConfig) (persistence.PersistenceAdapter, error) {
        return adapter, nil
    }); err != nil {
        t.Fatalf("register filesystem constructor: %v", err)
    }

    name := "gits-persist-bootstrap-block"
    defer instances.Remove(name)
    created := make(chan struct{})
    var inst *Gits
    var createErr error
    go func() {
        inst, createErr = NewInstanceWithOptions(
            name,
            WithPersistence(persistence.PersistenceConfig{
                Enabled:           true,
                Kind:              persistence.AdapterKindFilesystem,
                ChannelBufferSize: 1,
                Filesystem:        &persistence.FilesystemConfig{Directory: dir},
            }),
            WithAdapterFactory(factory),
        )
        close(created)
    }()

    <-adapter.loadStarted
    if atomic.LoadInt32(&adapter.loaded) != 0 {
        t.Fatalf("bootstrap LoadAll completed before release signal")
    }
    close(adapter.releaseLoad)
    <-created
    if createErr != nil {
        t.Fatalf("NewInstanceWithOptions bootstrap: %v", createErr)
    }
    if inst == nil {
        t.Fatalf("NewInstanceWithOptions returned nil instance")
    }
    defer inst.PersistenceManager().Close(context.Background())

    storage := inst.Storage()
    typeMap := storage.GetEntityTypes()
    if got, ok := typeMap[1]; !ok || got != "thing" {
        t.Fatalf("entity type 1 = %q (present=%v), want thing", got, ok)
    }
    reverseTypeMap := storage.GetEntityRTypes()
    if got := reverseTypeMap["thing"]; got != 1 {
        t.Fatalf("reverse entity type thing = %d, want 1", got)
    }

    entity, err := storage.GetEntityByPath(1, 42, "ctx")
    if err != nil {
        t.Fatalf("GetEntityByPath after bootstrap: %v", err)
    }
    if entity.ID != 42 || entity.Type != 1 || entity.Value != "v" || entity.Version != 3 || entity.Properties["k"] != "v" {
        t.Fatalf("bootstrap did not preserve entity identity/state: %+v", entity)
    }

    relation, err := storage.GetRelation(1, 42, 1, 43)
    if err != nil {
        t.Fatalf("GetRelation after bootstrap: %v", err)
    }
    if relation.SourceType != 1 || relation.SourceID != 42 || relation.TargetType != 1 || relation.TargetID != 43 ||
        relation.Context != "rel" || relation.Version != 7 || relation.Properties["rk"] != "rv" {
        t.Fatalf("bootstrap did not preserve relation identity/state: %+v", relation)
    }

    nextID, err := storage.CreateEntity(types.StorageEntity{Type: 1, Context: "next", Value: "next"})
    if err != nil {
        t.Fatalf("CreateEntity after bootstrap: %v", err)
    }
    if nextID != 43 {
        t.Fatalf("new entity ID = %d, want 43 to prove persisted high-water mark was restored", nextID)
    }
}

func TestGitsBootstrapErrorDoesNotPublishInstance(t *testing.T) {
    dir := t.TempDir()
    adapter := &bootstrapFakeAdapter{
        name:        "filesystem:" + dir,
        snapshot:    persistence.NewSnapshot(),
        loadErr:     fmt.Errorf("load failed"),
        loadStarted: make(chan struct{}),
        releaseLoad: make(chan struct{}),
    }
    factory := persistence.NewAdapterFactory()
    if err := factory.Register(persistence.AdapterKindFilesystem, func(cfg *persistence.PersistenceConfig) (persistence.PersistenceAdapter, error) {
        return adapter, nil
    }); err != nil {
        t.Fatalf("register filesystem constructor: %v", err)
    }

    name := "gits-persist-bootstrap-error"
    defer instances.Remove(name)
    created := make(chan struct{})
    var inst *Gits
    var createErr error
    go func() {
        inst, createErr = NewInstanceWithOptions(
            name,
            WithPersistence(persistence.PersistenceConfig{
                Enabled:           true,
                Kind:              persistence.AdapterKindFilesystem,
                ChannelBufferSize: 1,
                Filesystem:        &persistence.FilesystemConfig{Directory: dir},
            }),
            WithAdapterFactory(factory),
        )
        close(created)
    }()

    <-adapter.loadStarted
    close(adapter.releaseLoad)
    <-created
    if createErr == nil {
        t.Fatalf("expected bootstrap error, got nil")
    }
    if inst != nil {
        t.Fatalf("bootstrap failure should return nil instance, got %v", inst.Name)
    }
    if instances.GetByName(name) != nil {
        t.Fatalf("failed bootstrap instance should not be registered")
    }
    if atomic.LoadInt32(&adapter.closed) != 1 {
        t.Fatalf("expected persistence manager to close adapter after failed bootstrap")
    }
}

type shutdownFakeAdapter struct {
	name         string
	snapshot     *persistence.Snapshot
	appliedCount int32
	lastEntityID int64
	flushes      int32
	closed       int32
}

func (a *shutdownFakeAdapter) Name() string { return a.name }

func (a *shutdownFakeAdapter) LoadAll(ctx context.Context) (*persistence.Snapshot, error) {
	return a.snapshot, nil
}

func (a *shutdownFakeAdapter) ApplyEvent(ctx context.Context, event *persistence.PersistenceEvent) error {
	atomic.AddInt32(&a.appliedCount, 1)
	if event.Entity.ID != 0 {
		atomic.StoreInt64(&a.lastEntityID, int64(event.Entity.ID))
	}
	return nil
}

func (a *shutdownFakeAdapter) Flush(ctx context.Context) error {
	atomic.AddInt32(&a.flushes, 1)
	return nil
}

func (a *shutdownFakeAdapter) Close(ctx context.Context) error {
	atomic.StoreInt32(&a.closed, 1)
	return nil
}

func TestGitsShutdownFlushesQueuedMutations(t *testing.T) {
	dir := t.TempDir()
	snapshot := persistence.NewSnapshot()
	snapshot.EntityTypes[1] = "thing"
	snapshot.EntityRTypes["thing"] = 1
	snapshot.EntityTypeIDMax = 1
	snapshot.EntityIDMax[1] = 42

	adapter := &shutdownFakeAdapter{name: "filesystem:" + dir, snapshot: snapshot}
	factory := persistence.NewAdapterFactory()
	if err := factory.Register(persistence.AdapterKindFilesystem, func(cfg *persistence.PersistenceConfig) (persistence.PersistenceAdapter, error) {
		return adapter, nil
	}); err != nil {
		t.Fatalf("register filesystem constructor: %v", err)
	}

	name := "gits-persist-shutdown-flush"
	defer instances.Remove(name)
	inst, err := NewInstanceWithOptions(
		name,
		WithPersistence(persistence.PersistenceConfig{
			Enabled:           true,
			Kind:              persistence.AdapterKindFilesystem,
			ChannelBufferSize: 1,
			Filesystem:        &persistence.FilesystemConfig{Directory: dir},
		}),
		WithAdapterFactory(factory),
	)
	if err != nil {
		t.Fatalf("NewInstanceWithOptions shutdown: %v", err)
	}

	entityID, err := inst.Storage().CreateEntity(types.StorageEntity{Type: 1, Context: "ctx", Value: "v"})
	if err != nil {
		t.Fatalf("CreateEntity before shutdown: %v", err)
	}
	if entityID != 43 {
		t.Fatalf("new entity ID = %d, want 43 from restored high-water mark", entityID)
	}

	ctx := context.Background()
	if err := inst.Shutdown(ctx); err != nil {
		t.Fatalf("Shutdown: %v", err)
	}
	if atomic.LoadInt32(&adapter.appliedCount) == 0 {
		t.Fatalf("expected queued mutation to be applied before Shutdown returned")
	}
	if got := atomic.LoadInt64(&adapter.lastEntityID); int(got) != entityID {
		t.Fatalf("applied entity ID = %d, want %d", got, entityID)
	}
	if atomic.LoadInt32(&adapter.flushes) == 0 {
		t.Fatalf("expected adapter Flush to be called during Shutdown")
	}
	if atomic.LoadInt32(&adapter.closed) != 1 {
		t.Fatalf("expected adapter Close to be called during Shutdown")
	}
	if err := inst.Close(ctx); err != nil {
		t.Fatalf("Close after Shutdown: %v", err)
	}

	memName := "gits-persist-shutdown-inmemory"
	defer instances.Remove(memName)
	mem := NewInstance(memName)
	if mem == nil {
		t.Fatalf("NewInstance returned nil")
	}
	if err := mem.Shutdown(ctx); err != nil {
		t.Fatalf("in-memory Shutdown: %v", err)
	}
	if err := mem.Close(ctx); err != nil {
		t.Fatalf("in-memory Close: %v", err)
	}
}
