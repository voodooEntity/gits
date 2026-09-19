package postgresql

import (
    "context"
    "os"
    "testing"

    "github.com/voodooEntity/gits/src/persistence"
    "github.com/voodooEntity/gits/src/types"
)

func TestPostgresDSNFromConfig(t *testing.T) {
    cfg := &persistence.PostgresConfig{
        Host:     "db.example",
        Port:     5432,
        User:     "alice",
        Password: "secret",
        Database: "gits",
    }
    got := dsnFromConfig(cfg)
    want := "postgres://alice:secret@db.example:5432/gits?sslmode=disable"
    if got != want {
        t.Fatalf("dsnFromConfig() = %q, want %q", got, want)
    }

    cfg.DSN = "postgres://custom@example.com:6432/custom?sslmode=require"
    got = dsnFromConfig(cfg)
    want = "postgres://custom@example.com:6432/custom?sslmode=require"
    if got != want {
        t.Fatalf("dsnFromConfig() with DSN override = %q, want %q", got, want)
    }
}

func TestPostgresSchemaSQLUsesSingleJSONBRow(t *testing.T) {
    want := `CREATE TABLE IF NOT EXISTS gits_persistence_state (id integer PRIMARY KEY, payload jsonb)`
    if createTableSQL != want {
        t.Fatalf("createTableSQL = %q, want %q", createTableSQL, want)
    }
}

func TestPostgresNewRequiresConfig(t *testing.T) {
    if _, err := New(nil); err == nil {
        t.Fatal("New(nil) expected error")
    }

    cfg := &persistence.PersistenceConfig{Enabled: true, Kind: persistence.AdapterKindPostgreSQL}
    if _, err := New(cfg); err == nil {
        t.Fatal("New() without PostgreSQL config expected error")
    }
}

func TestPostgresAdapterNameIncludesDSN(t *testing.T) {
    dsn := "postgres://user:pass@localhost:5432/gits"
    cfg := &persistence.PersistenceConfig{
        Enabled: true,
        Kind:    persistence.AdapterKindPostgreSQL,
        PostgreSQL: &persistence.PostgresConfig{DSN: dsn},
    }

    adapter, err := New(cfg)
    if err != nil {
        t.Fatalf("New() error = %v", err)
    }
    a := adapter.(*PostgresAdapter)
    defer a.db.Close()

    want := "postgresql:" + dsn
    if got := a.Name(); got != want {
        t.Fatalf("Name() = %q, want %q", got, want)
    }
}

func TestPostgresAdapterApplyEventsPreservesIDsInMemory(t *testing.T) {
    cfg := &persistence.PersistenceConfig{
        Enabled: true,
        Kind:    persistence.AdapterKindPostgreSQL,
        PostgreSQL: &persistence.PostgresConfig{
            DSN: "postgres://user:pass@localhost:5432/gits",
        },
    }

    adapter, err := New(cfg)
    if err != nil {
        t.Fatalf("New() error = %v", err)
    }
    a := adapter.(*PostgresAdapter)
    defer a.db.Close()

    ctx := context.Background()
    events := []*persistence.PersistenceEvent{
        {Type: persistence.EventTypeCreateEntityType, EntityTypeID: 1, EntityTypeName: "thing"},
        {Type: persistence.EventTypeCreateEntity, Entity: types.StorageEntity{Type: 1, ID: 42, Value: "v1", Version: 1}},
        {Type: persistence.EventTypeUpdateEntity, Entity: types.StorageEntity{Type: 1, ID: 42, Value: "v2", Version: 2}},
        {Type: persistence.EventTypeCreateEntity, Entity: types.StorageEntity{Type: 1, ID: 43, Value: "temp", Version: 1}},
        {Type: persistence.EventTypeDeleteEntity, Entity: types.StorageEntity{Type: 1, ID: 43}},
        {Type: persistence.EventTypeCreateRelation, Relation: types.StorageRelation{SourceType: 1, SourceID: 42, TargetType: 1, TargetID: 44, Context: "link", Properties: map[string]string{"k": "v"}, Version: 7}},
    }

    for _, event := range events {
        if err := a.ApplyEvent(ctx, event); err != nil {
            t.Fatalf("ApplyEvent(%s) error = %v", event.Type, err)
        }
    }

    snap := a.snapshotLocked()
    if snap.EntityTypes[1] != "thing" {
        t.Fatalf("EntityTypes[1] = %q, want thing", snap.EntityTypes[1])
    }
    if snap.EntityRTypes["thing"] != 1 {
        t.Fatalf("EntityRTypes[thing] = %d, want 1", snap.EntityRTypes["thing"])
    }
    if snap.EntityTypeIDMax < 1 {
        t.Fatalf("EntityTypeIDMax = %d, want >= 1", snap.EntityTypeIDMax)
    }

    entity, ok := findEntity(snap.Entities, 1, 42)
    if !ok {
        t.Fatal("entity type=1 id=42 not found")
    }
    if entity.Value != "v2" || entity.Version != 2 {
        t.Fatalf("entity = %+v, want updated value v2 version 2", entity)
    }

    if _, ok := findEntity(snap.Entities, 1, 43); ok {
        t.Fatal("deleted entity type=1 id=43 still present")
    }

    relation, ok := findRelation(snap.Relations, 1, 42, 1, 44)
    if !ok {
        t.Fatal("relation 1:42 -> 1:44 not found")
    }
    if relation.Context != "link" || relation.Properties["k"] != "v" || relation.Version != 7 {
        t.Fatalf("relation = %+v, want preserved context/properties/version", relation)
    }
}

func TestPostgresAdapterSnapshotPreservesHighWaterMarks(t *testing.T) {
    cfg := &persistence.PersistenceConfig{
        Enabled: true,
        Kind:    persistence.AdapterKindPostgreSQL,
        PostgreSQL: &persistence.PostgresConfig{
            DSN: "postgres://user:pass@localhost:5432/gits",
        },
    }

    adapter, err := New(cfg)
    if err != nil {
        t.Fatalf("New() error = %v", err)
    }
    a := adapter.(*PostgresAdapter)
    defer a.db.Close()

    ctx := context.Background()
    events := []*persistence.PersistenceEvent{
        {Type: persistence.EventTypeCreateEntityType, EntityTypeID: 1, EntityTypeName: "thing"},
        {Type: persistence.EventTypeCreateEntityType, EntityTypeID: 2, EntityTypeName: "other"},
        {Type: persistence.EventTypeCreateEntity, Entity: types.StorageEntity{Type: 1, ID: 10, Value: "a", Version: 1}},
        {Type: persistence.EventTypeCreateEntity, Entity: types.StorageEntity{Type: 1, ID: 20, Value: "b", Version: 1}},
        {Type: persistence.EventTypeUpdateEntity, Entity: types.StorageEntity{Type: 1, ID: 20, Value: "updated-b", Version: 2}},
        {Type: persistence.EventTypeDeleteEntity, Entity: types.StorageEntity{Type: 1, ID: 10}},
    }

    for _, event := range events {
        if err := a.ApplyEvent(ctx, event); err != nil {
            t.Fatalf("ApplyEvent(%s) error = %v", event.Type, err)
        }
    }

    snap := a.snapshotLocked()
    if snap.EntityTypeIDMax != 2 {
        t.Fatalf("EntityTypeIDMax = %d, want 2", snap.EntityTypeIDMax)
    }
    if got := snap.EntityIDMax[1]; got != 20 {
        t.Fatalf("EntityIDMax[1] = %d, want 20", got)
    }
    if len(snap.Entities) != 1 {
        t.Fatalf("len(Entities) = %d, want 1", len(snap.Entities))
    }

    entity, ok := findEntity(snap.Entities, 1, 20)
    if !ok {
        t.Fatal("entity type=1 id=20 not found")
    }
    if entity.Value != "updated-b" || entity.Version != 2 {
        t.Fatalf("entity = %+v, want updated value and version", entity)
    }

    if _, ok := findEntity(snap.Entities, 1, 10); ok {
        t.Fatal("deleted entity type=1 id=10 still present")
    }
}

func TestPostgresAdapterClosedStateBehaviors(t *testing.T) {
    dsn := "postgres://user:pass@localhost:5432/gits"
    cfg := &persistence.PersistenceConfig{
        Enabled: true,
        Kind:    persistence.AdapterKindPostgreSQL,
        PostgreSQL: &persistence.PostgresConfig{DSN: dsn},
    }

    adapter, err := New(cfg)
    if err != nil {
        t.Fatalf("New() error = %v", err)
    }
    a := adapter.(*PostgresAdapter)
    defer a.db.Close()

    ctx := context.Background()
    a.mu.Lock()
    a.closed = true
    a.mu.Unlock()

    if err := a.ApplyEvent(ctx, &persistence.PersistenceEvent{Type: persistence.EventTypeCreateEntity, Entity: types.StorageEntity{Type: 1, ID: 1}}); err == nil {
        t.Fatal("ApplyEvent after close expected error")
    }
    if err := a.Flush(ctx); err != nil {
        t.Fatalf("Flush() after close = %v, want nil", err)
    }
    if err := adapter.Close(ctx); err != nil {
        t.Fatalf("Close() after close = %v, want nil", err)
    }
}

func TestPostgresAdapterRoundTripPreservesIDs(t *testing.T) {
    dsn := os.Getenv("GITS_POSTGRES_TEST_DSN")
    if dsn == "" {
        t.Skip("set GITS_POSTGRES_TEST_DSN to run PostgreSQL integration tests")
    }

    ctx := context.Background()
    cfg := &persistence.PersistenceConfig{
        Enabled: true,
        Kind:    persistence.AdapterKindPostgreSQL,
        PostgreSQL: &persistence.PostgresConfig{DSN: dsn},
    }

    adapter, err := New(cfg)
    if err != nil {
        t.Fatalf("New() error = %v", err)
    }
    a := adapter.(*PostgresAdapter)

    if _, err := a.db.ExecContext(ctx, createTableSQL); err != nil {
        t.Fatalf("ensure schema: %v", err)
    }
    if _, err := a.db.ExecContext(ctx, `DELETE FROM gits_persistence_state WHERE id = 1`); err != nil {
        t.Fatalf("clean state row: %v", err)
    }

    empty, err := adapter.LoadAll(ctx)
    if err != nil {
        t.Fatalf("LoadAll() empty error = %v", err)
    }
    if len(empty.Entities) != 0 || len(empty.Relations) != 0 {
        t.Fatalf("expected empty snapshot, got entities=%d relations=%d", len(empty.Entities), len(empty.Relations))
    }

    events := []*persistence.PersistenceEvent{
        {Type: persistence.EventTypeCreateEntityType, EntityTypeID: 1, EntityTypeName: "thing"},
        {Type: persistence.EventTypeCreateEntity, Entity: types.StorageEntity{Type: 1, ID: 42, Value: "v1", Version: 1}},
        {Type: persistence.EventTypeUpdateEntity, Entity: types.StorageEntity{Type: 1, ID: 42, Value: "v2", Version: 2}},
        {Type: persistence.EventTypeCreateRelation, Relation: types.StorageRelation{SourceType: 1, SourceID: 42, TargetType: 1, TargetID: 43, Context: "link", Properties: map[string]string{"k": "v"}, Version: 7}},
    }
    for _, event := range events {
        if err := a.ApplyEvent(ctx, event); err != nil {
            t.Fatalf("ApplyEvent(%s) error = %v", event.Type, err)
        }
    }

    if err := adapter.Flush(ctx); err != nil {
        t.Fatalf("Flush() error = %v", err)
    }
    if err := adapter.Close(ctx); err != nil {
        t.Fatalf("Close() error = %v", err)
    }

    reloaded, err := New(cfg)
    if err != nil {
        t.Fatalf("New() reloaded error = %v", err)
    }
    a2 := reloaded.(*PostgresAdapter)
    defer func() {
        _, _ = a2.db.ExecContext(ctx, `DELETE FROM gits_persistence_state WHERE id = 1`)
        _ = a2.db.Close()
    }()

    snap, err := reloaded.LoadAll(ctx)
    if err != nil {
        t.Fatalf("LoadAll() reloaded error = %v", err)
    }

    if snap.EntityTypes[1] != "thing" || snap.EntityRTypes["thing"] != 1 {
        t.Fatalf("type mappings not preserved: %+v", snap)
    }
    entity, ok := findEntity(snap.Entities, 1, 42)
    if !ok {
        t.Fatal("reloaded entity type=1 id=42 not found")
    }
    if entity.Value != "v2" || entity.Version != 2 {
        t.Fatalf("reloaded entity = %+v, want value v2 version 2", entity)
    }
    relation, ok := findRelation(snap.Relations, 1, 42, 1, 43)
    if !ok {
        t.Fatal("reloaded relation 1:42 -> 1:43 not found")
    }
    if relation.Context != "link" || relation.Properties["k"] != "v" || relation.Version != 7 {
        t.Fatalf("reloaded relation = %+v, want preserved context/properties/version", relation)
    }
}

func findEntity(entities []types.StorageEntity, typ int, id int) (types.StorageEntity, bool) {
    for _, entity := range entities {
        if entity.Type == typ && entity.ID == id {
            return entity, true
        }
    }
    return types.StorageEntity{}, false
}

func findRelation(relations []types.StorageRelation, sourceType int, sourceID int, targetType int, targetID int) (types.StorageRelation, bool) {
    for _, relation := range relations {
        if relation.SourceType == sourceType && relation.SourceID == sourceID && relation.TargetType == targetType && relation.TargetID == targetID {
            return relation, true
        }
    }
    return types.StorageRelation{}, false
}
