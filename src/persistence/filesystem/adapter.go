package filesystem

import ("context"; "encoding/json"; "fmt"; "os"; "path/filepath"; "sync"; "github.com/voodooEntity/gits/src/persistence"; "github.com/voodooEntity/gits/src/types")

const stateFileName = "state.json"

type FilesystemAdapter struct {
  dir    string
  mu     sync.Mutex
  closed bool
  state  *persistence.Snapshot
}

func init() { if err := persistence.Register(persistence.AdapterKindFilesystem, New); err != nil { panic(err) } }

func New(cfg *persistence.PersistenceConfig) (persistence.PersistenceAdapter, error) {
  if cfg == nil || cfg.Filesystem == nil || cfg.Filesystem.Directory == "" { return nil, fmt.Errorf("filesystem: directory is required") }
  dir := cfg.Filesystem.Directory
  if err := os.MkdirAll(dir, 0755); err != nil { return nil, fmt.Errorf("filesystem: create directory %q: %w", dir, err) }
  return &FilesystemAdapter{dir: dir, state: persistence.NewSnapshot()}, nil
}

func (a *FilesystemAdapter) Name() string { return "filesystem:" + a.dir }

func (a *FilesystemAdapter) LoadAll(ctx context.Context) (*persistence.Snapshot, error) {
  a.mu.Lock(); defer a.mu.Unlock()
  if err := a.loadLocked(); err != nil { return nil, err }
  return a.snapshotLocked(), nil
}

func (a *FilesystemAdapter) ApplyEvent(ctx context.Context, event *persistence.PersistenceEvent) error {
  a.mu.Lock(); defer a.mu.Unlock()
  if a.closed { return fmt.Errorf("filesystem: adapter is closed") }
  switch event.Type {
  case persistence.EventTypeCreateEntityType:
    id := event.EntityTypeID; name := event.EntityTypeName
    if name != "" { a.state.EntityTypes[id] = name; a.state.EntityRTypes[name] = id }
    if id > a.state.EntityTypeIDMax { a.state.EntityTypeIDMax = id }
  case persistence.EventTypeCreateEntity, persistence.EventTypeUpdateEntity:
    e := event.Entity
    if _, ok := a.state.EntityTypes[e.Type]; !ok && event.EntityTypeName != "" {
      a.state.EntityTypes[e.Type] = event.EntityTypeName; a.state.EntityRTypes[event.EntityTypeName] = e.Type
    }
    upsertEntity(a.state, e)
  case persistence.EventTypeDeleteEntity: removeEntity(a.state, event.Entity)
  case persistence.EventTypeCreateRelation, persistence.EventTypeUpdateRelation: upsertRelation(a.state, event.Relation)
  case persistence.EventTypeDeleteRelation: removeRelation(a.state, event.Relation)
  }
  return nil
}

func (a *FilesystemAdapter) Flush(ctx context.Context) error { a.mu.Lock(); defer a.mu.Unlock(); if a.closed { return nil }; return a.flushLocked() }

func (a *FilesystemAdapter) Close(ctx context.Context) error { a.mu.Lock(); defer a.mu.Unlock(); if a.closed { return nil }; err := a.flushLocked(); a.closed = true; return err }

func (a *FilesystemAdapter) loadLocked() error {
  s := persistence.NewSnapshot()
  path := filepath.Join(a.dir, stateFileName)
  data, err := os.ReadFile(path)
  if os.IsNotExist(err) { a.state = s; return nil }
  if err != nil { return fmt.Errorf("filesystem: read %q: %w", path, err) }
  var loaded persistence.Snapshot
  if err := json.Unmarshal(data, &loaded); err != nil { return fmt.Errorf("filesystem: decode %q: %w", path, err) }
  if loaded.EntityTypes == nil { loaded.EntityTypes = make(map[int]string) }
  if loaded.EntityRTypes == nil { loaded.EntityRTypes = make(map[string]int) }
  if loaded.EntityIDMax == nil { loaded.EntityIDMax = make(map[int]int) }
  for id, name := range loaded.EntityTypes { if id > loaded.EntityTypeIDMax { loaded.EntityTypeIDMax = id }; loaded.EntityRTypes[name] = id }
  for _, e := range loaded.Entities { if e.ID > loaded.EntityIDMax[e.Type] { loaded.EntityIDMax[e.Type] = e.ID } }
  a.state = &loaded
  return nil
}

func (a *FilesystemAdapter) flushLocked() error {
  if a.state == nil { a.state = persistence.NewSnapshot() }
  data, err := json.Marshal(a.state)
  if err != nil { return fmt.Errorf("filesystem: encode state: %w", err) }
  tmp, err := os.CreateTemp(a.dir, ".state-*.json")
  if err != nil { return fmt.Errorf("filesystem: create temp file: %w", err) }
  name := tmp.Name(); defer os.Remove(name)
  if _, err := tmp.Write(data); err != nil { tmp.Close(); return fmt.Errorf("filesystem: write state: %w", err) }
  if err := tmp.Sync(); err != nil { tmp.Close(); return fmt.Errorf("filesystem: sync state: %w", err) }
  if err := tmp.Close(); err != nil { return fmt.Errorf("filesystem: close state: %w", err) }
  return os.Rename(name, filepath.Join(a.dir, stateFileName))
}

func (a *FilesystemAdapter) snapshotLocked() *persistence.Snapshot {
  s := persistence.NewSnapshot()
  for k, v := range a.state.EntityTypes { s.EntityTypes[k] = v; if k > s.EntityTypeIDMax { s.EntityTypeIDMax = k } }
  for k, v := range a.state.EntityRTypes { s.EntityRTypes[k] = v }
  s.EntityTypeIDMax = a.state.EntityTypeIDMax
  for k, v := range a.state.EntityIDMax { s.EntityIDMax[k] = v }
  if len(a.state.Entities) > 0 { s.Entities = append(s.Entities, a.state.Entities...) }
  if len(a.state.Relations) > 0 { s.Relations = append(s.Relations, a.state.Relations...) }
  return s
}

func upsertEntity(s *persistence.Snapshot, e types.StorageEntity) {
  filtered := s.Entities[:0]
  for _, x := range s.Entities { if x.Type != e.Type || x.ID != e.ID { filtered = append(filtered, x) } }
  s.Entities = append(filtered, e)
  if e.ID > s.EntityIDMax[e.Type] { s.EntityIDMax[e.Type] = e.ID }
}

func removeEntity(s *persistence.Snapshot, e types.StorageEntity) {
  filtered := s.Entities[:0]
  for _, x := range s.Entities { if !(x.Type == e.Type && x.ID == e.ID) { filtered = append(filtered, x) } }
  s.Entities = filtered
}

func upsertRelation(s *persistence.Snapshot, r types.StorageRelation) {
  filtered := s.Relations[:0]
  for _, x := range s.Relations { if !sameRelation(x, r) { filtered = append(filtered, x) } }
  s.Relations = append(filtered, r)
}

func removeRelation(s *persistence.Snapshot, r types.StorageRelation) {
  filtered := s.Relations[:0]
  for _, x := range s.Relations { if !sameRelation(x, r) { filtered = append(filtered, x) } }
  s.Relations = filtered
}

func sameRelation(a, b types.StorageRelation) bool { return a.SourceType == b.SourceType && a.SourceID == b.SourceID && a.TargetType == b.TargetType && a.TargetID == b.TargetID }
