package filesystem_test

import (
	"context"
	"testing"

	"github.com/voodooEntity/gits/src/persistence"
	"github.com/voodooEntity/gits/src/persistence/filesystem"
	"github.com/voodooEntity/gits/src/storage"
	"github.com/voodooEntity/gits/src/types"
)

func TestFilesystemStorageRoundTripPreservesIDs(t *testing.T) {
	ctx := context.Background()
	dir := t.TempDir()
	cfg := &persistence.PersistenceConfig{
		Enabled:           true,
		Kind:              persistence.AdapterKindFilesystem,
		ChannelBufferSize: 1,
		Filesystem:        &persistence.FilesystemConfig{Directory: dir},
	}

	adapter, err := filesystem.New(cfg)
	if err != nil {
		t.Fatalf("create first adapter: %v", err)
	}
	stg := storage.NewStorage()
	manager := persistence.NewManager(adapter, cfg.ChannelBufferSize)
	stg.PersistenceManager = manager
	defer func() { _ = manager.Close(ctx) }()

	if err := manager.Bootstrap(ctx, stg.ApplySnapshot); err != nil {
		t.Fatalf("bootstrap first storage: %v", err)
	}

	typeID, err := stg.CreateEntityType("person")
	if err != nil {
		t.Fatalf("CreateEntityType: %v", err)
	}
	if typeID != 1 {
		t.Fatalf("first entity type ID = %d, want 1", typeID)
	}

	entityA, err := stg.CreateEntity(types.StorageEntity{Type: typeID, Context: "ctx-a", Value: "value-a", Properties: map[string]string{"k": "v"}})
	if err != nil {
		t.Fatalf("CreateEntity A: %v", err)
	}
	if entityA != 1 {
		t.Fatalf("entity A ID = %d, want 1", entityA)
	}

	entityB, err := stg.CreateEntity(types.StorageEntity{Type: typeID, Context: "ctx-b", Value: "value-b", Properties: map[string]string{"k": "v"}})
	if err != nil {
		t.Fatalf("CreateEntity B: %v", err)
	}
	if entityB != 2 {
		t.Fatalf("entity B ID = %d, want 2", entityB)
	}

	entityC, err := stg.CreateEntity(types.StorageEntity{Type: typeID, Context: "ctx-c", Value: "value-c", Properties: map[string]string{"k": "v"}})
	if err != nil {
		t.Fatalf("CreateEntity C: %v", err)
	}
	if entityC != 3 {
		t.Fatalf("entity C ID = %d, want 3", entityC)
	}

	created, err := stg.CreateRelation(typeID, entityA, typeID, entityB, types.StorageRelation{Context: "rel", Properties: map[string]string{"rk": "rv"}})
	if !created || err != nil {
		t.Fatalf("CreateRelation: created=%v err=%v", created, err)
	}

	updatedA := types.StorageEntity{Type: typeID, ID: entityA, Context: "ctx-a", Value: "updated-a", Properties: map[string]string{"k": "v"}, Version: 1}
	if err := stg.UpdateEntity(updatedA); err != nil {
		t.Fatalf("UpdateEntity A: %v", err)
	}

	stg.DeleteEntity(typeID, entityC)

	if err := manager.Close(ctx); err != nil {
		t.Fatalf("shutdown first storage: %v", err)
	}

	adapter2, err := filesystem.New(cfg)
	if err != nil {
		t.Fatalf("create second adapter: %v", err)
	}
	stg2 := storage.NewStorage()
	manager2 := persistence.NewManager(adapter2, cfg.ChannelBufferSize)
	stg2.PersistenceManager = manager2
	defer func() { _ = manager2.Close(ctx) }()

	if err := manager2.Bootstrap(ctx, stg2.ApplySnapshot); err != nil {
		t.Fatalf("bootstrap second storage: %v", err)
	}

	typeName, err := stg2.GetTypeStringById(typeID)
	if err != nil {
		t.Fatalf("GetTypeStringById after reload: %v", err)
	}
	if typeName != "person" {
		t.Fatalf("reloaded type name = %q, want person", typeName)
	}

	gotTypeID, err := stg2.GetTypeIdByString("person")
	if err != nil {
		t.Fatalf("GetTypeIdByString after reload: %v", err)
	}
	if gotTypeID != typeID {
		t.Fatalf("reloaded type ID for person = %d, want %d", gotTypeID, typeID)
	}

	a, err := stg2.GetEntityByPath(typeID, entityA, "ctx-a")
	if err != nil {
		t.Fatalf("GetEntityByPath A after reload: %v", err)
	}
	if a.ID != entityA || a.Type != typeID || a.Value != "updated-a" || a.Version != 2 || a.Properties["k"] != "v" {
		t.Fatalf("reloaded entity A mismatch: %+v", a)
	}

	b, err := stg2.GetEntityByPath(typeID, entityB, "ctx-b")
	if err != nil {
		t.Fatalf("GetEntityByPath B after reload: %v", err)
	}
	if b.ID != entityB || b.Type != typeID || b.Value != "value-b" || b.Version != 1 || b.Properties["k"] != "v" {
		t.Fatalf("reloaded entity B mismatch: %+v", b)
	}

	if stg2.EntityExists(typeID, entityC) {
		t.Fatalf("deleted entity C present after reload")
	}

	amount, err := stg2.GetEntityAmountByType(typeID)
	if err != nil {
		t.Fatalf("GetEntityAmountByType after reload: %v", err)
	}
	if amount != 2 {
		t.Fatalf("entity amount for type %d = %d, want 2", typeID, amount)
	}

	rel, err := stg2.GetRelation(typeID, entityA, typeID, entityB)
	if err != nil {
		t.Fatalf("GetRelation after reload: %v", err)
	}
	if rel.SourceType != typeID || rel.SourceID != entityA || rel.TargetType != typeID || rel.TargetID != entityB ||
		rel.Context != "rel" || rel.Version != 1 || rel.Properties["rk"] != "rv" {
		t.Fatalf("reloaded relation mismatch: %+v", rel)
	}

	nextEntityID, err := stg2.CreateEntity(types.StorageEntity{Type: typeID, Context: "ctx-d", Value: "value-d"})
	if err != nil {
		t.Fatalf("CreateEntity after reload: %v", err)
	}
	if nextEntityID != 4 {
		t.Fatalf("next entity ID = %d, want 4 to prove high-water mark preserved", nextEntityID)
	}

	nextTypeID, err := stg2.CreateEntityType("other")
	if err != nil {
		t.Fatalf("CreateEntityType after reload: %v", err)
	}
	if nextTypeID != 2 {
		t.Fatalf("next entity type ID = %d, want 2 to prove type high-water mark preserved", nextTypeID)
	}
}
