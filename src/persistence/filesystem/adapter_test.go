package filesystem_test

import (
	"context"
	"testing"

	"github.com/voodooEntity/gits/src/persistence"
	"github.com/voodooEntity/gits/src/persistence/filesystem"
	"github.com/voodooEntity/gits/src/types"
)

func TestFilesystemAdapterLoadEmptyDirectory(t *testing.T) {
	ctx := context.Background()
	dir := t.TempDir()

	cfg := &persistence.PersistenceConfig{
		Enabled:    true,
		Kind:       persistence.AdapterKindFilesystem,
		Filesystem: &persistence.FilesystemConfig{Directory: dir},
	}

	adapter, err := filesystem.New(cfg)
	if err != nil {
		t.Fatalf("create adapter: %v", err)
	}
	defer func() {
		if err := adapter.Close(ctx); err != nil {
			t.Errorf("close adapter: %v", err)
		}
	}()

	snap, err := adapter.LoadAll(ctx)
	if err != nil {
		t.Fatalf("load empty directory: %v", err)
	}
	if snap.EntityTypes == nil || snap.EntityRTypes == nil || snap.EntityIDMax == nil {
		t.Fatalf("expected non-nil snapshot maps")
	}
	if len(snap.Entities) != 0 || len(snap.Relations) != 0 {
		t.Fatalf("expected empty snapshot, got %d entities and %d relations", len(snap.Entities), len(snap.Relations))
	}
}

func TestFilesystemAdapterRoundTripPreservesIDs(t *testing.T) {
	ctx := context.Background()
	dir := t.TempDir()

	cfg := &persistence.PersistenceConfig{
		Enabled:    true,
		Kind:       persistence.AdapterKindFilesystem,
		Filesystem: &persistence.FilesystemConfig{Directory: dir},
	}

	adapter, err := filesystem.New(cfg)
	if err != nil {
		t.Fatalf("create adapter: %v", err)
	}

	typeID := 7
	typeName := "person"
	if err := adapter.ApplyEvent(ctx, &persistence.PersistenceEvent{
		Sequence:       1,
		Type:           persistence.EventTypeCreateEntityType,
		EntityTypeID:   typeID,
		EntityTypeName: typeName,
	}); err != nil {
		t.Fatalf("apply create entity type: %v", err)
	}

	entity42 := types.StorageEntity{
		ID:         42,
		Type:       typeID,
		Context:    "ctx-42",
		Value:      "value-42",
		Properties: map[string]string{"k": "v"},
		Version:    1,
	}
	if err := adapter.ApplyEvent(ctx, &persistence.PersistenceEvent{
		Sequence:       2,
		Type:           persistence.EventTypeCreateEntity,
		EntityTypeName: typeName,
		Entity:         entity42,
	}); err != nil {
		t.Fatalf("apply create entity 42: %v", err)
	}

	entity43 := types.StorageEntity{
		ID:         43,
		Type:       typeID,
		Context:    "ctx-43",
		Value:      "value-43",
		Properties: map[string]string{"k": "v"},
		Version:    1,
	}
	if err := adapter.ApplyEvent(ctx, &persistence.PersistenceEvent{
		Sequence:       3,
		Type:           persistence.EventTypeCreateEntity,
		EntityTypeName: typeName,
		Entity:         entity43,
	}); err != nil {
		t.Fatalf("apply create entity 43: %v", err)
	}

	relation := types.StorageRelation{
		SourceType: typeID,
		SourceID:   42,
		TargetType: typeID,
		TargetID:   43,
		Context:    "rel",
		Properties: map[string]string{"rk": "rv"},
		Version:    1,
	}
	if err := adapter.ApplyEvent(ctx, &persistence.PersistenceEvent{
		Sequence: 4,
		Type:     persistence.EventTypeCreateRelation,
		Relation: relation,
	}); err != nil {
		t.Fatalf("apply create relation: %v", err)
	}

	updated42 := entity42
	updated42.Value = "updated-42"
	updated42.Version = 2
	if err := adapter.ApplyEvent(ctx, &persistence.PersistenceEvent{
		Sequence:       5,
		Type:           persistence.EventTypeUpdateEntity,
		EntityTypeName: typeName,
		Entity:         updated42,
	}); err != nil {
		t.Fatalf("apply update entity 42: %v", err)
	}

	entity44 := types.StorageEntity{
		ID:         44,
		Type:       typeID,
		Context:    "ctx-44",
		Value:      "value-44",
		Properties: map[string]string{},
		Version:    1,
	}
	if err := adapter.ApplyEvent(ctx, &persistence.PersistenceEvent{
		Sequence:       6,
		Type:           persistence.EventTypeCreateEntity,
		EntityTypeName: typeName,
		Entity:         entity44,
	}); err != nil {
		t.Fatalf("apply create entity 44: %v", err)
	}
	if err := adapter.ApplyEvent(ctx, &persistence.PersistenceEvent{
		Sequence: 7,
		Type:     persistence.EventTypeDeleteEntity,
		Entity:   types.StorageEntity{ID: 44, Type: typeID},
	}); err != nil {
		t.Fatalf("apply delete entity 44: %v", err)
	}

	if err := adapter.Flush(ctx); err != nil {
		t.Fatalf("flush adapter: %v", err)
	}
	if err := adapter.Close(ctx); err != nil {
		t.Fatalf("close adapter: %v", err)
	}

	loadedAdapter, err := filesystem.New(cfg)
	if err != nil {
		t.Fatalf("create loaded adapter: %v", err)
	}
	snap, err := loadedAdapter.LoadAll(ctx)
	if err != nil {
		t.Fatalf("load persisted state: %v", err)
	}
	if err := loadedAdapter.Close(ctx); err != nil {
		t.Errorf("close loaded adapter: %v", err)
	}

	if snap.EntityTypes[typeID] != typeName {
		t.Fatalf("EntityTypes[%d] = %q, want %q", typeID, snap.EntityTypes[typeID], typeName)
	}
	if snap.EntityRTypes[typeName] != typeID {
		t.Fatalf("EntityRTypes[%q] = %d, want %d", typeName, snap.EntityRTypes[typeName], typeID)
	}
	if snap.EntityTypeIDMax < typeID {
		t.Fatalf("EntityTypeIDMax = %d, want at least %d", snap.EntityTypeIDMax, typeID)
	}
	if got := snap.EntityIDMax[typeID]; got != 44 {
		t.Fatalf("EntityIDMax[%d] = %d, want 44", typeID, got)
	}

	loadedUpdated42, ok := findStorageEntity(snap.Entities, typeID, 42)
	if !ok {
		t.Fatalf("entity type %d id 42 not found after reload", typeID)
	}
	if loadedUpdated42.Value != "updated-42" || loadedUpdated42.Version != 2 || loadedUpdated42.Properties["k"] != "v" {
		t.Fatalf("reloaded entity 42 mismatch: %+v", loadedUpdated42)
	}

	loaded43, ok := findStorageEntity(snap.Entities, typeID, 43)
	if !ok {
		t.Fatalf("entity type %d id 43 not found after reload", typeID)
	}
	if loaded43.Value != "value-43" || loaded43.Version != 1 {
		t.Fatalf("reloaded entity 43 mismatch: %+v", loaded43)
	}

	for _, e := range snap.Entities {
		if e.Type == typeID && e.ID == 44 {
			t.Fatalf("deleted entity type %d id 44 present after reload", typeID)
		}
	}

	if len(snap.Relations) != 1 {
		t.Fatalf("expected 1 relation, got %d", len(snap.Relations))
	}
	r := snap.Relations[0]
	if r.SourceType != typeID || r.SourceID != 42 || r.TargetType != typeID || r.TargetID != 43 {
		t.Fatalf("reloaded relation endpoints mismatch: %+v", r)
	}
	if r.Context != "rel" || r.Properties["rk"] != "rv" || r.Version != 1 {
		t.Fatalf("reloaded relation metadata mismatch: %+v", r)
	}
}

func findStorageEntity(entities []types.StorageEntity, typ int, id int) (types.StorageEntity, bool) {
	for _, e := range entities {
		if e.Type == typ && e.ID == id {
			return e, true
		}
	}
	return types.StorageEntity{}, false
}
