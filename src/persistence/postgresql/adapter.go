package postgresql

import (
    `context`
    `database/sql`
    `encoding/json`
    `fmt`
    `net/url`
    `sync`

    _ `github.com/lib/pq`
    `github.com/voodooEntity/gits/src/persistence`
    `github.com/voodooEntity/gits/src/types`
)

const createTableSQL = `CREATE TABLE IF NOT EXISTS gits_persistence_state (id integer PRIMARY KEY, payload jsonb)`

type PostgresAdapter struct {
    dsn    string
    db     *sql.DB
    mu     sync.Mutex
    closed bool
    state  *persistence.Snapshot
}

func init() {
    if err := persistence.Register(persistence.AdapterKindPostgreSQL, New); err != nil {
        panic(err)
    }
}

func New(cfg *persistence.PersistenceConfig) (persistence.PersistenceAdapter, error) {
    if cfg == nil || cfg.PostgreSQL == nil {
        return nil, fmt.Errorf(`postgresql: config is required`)
    }
    dsn := dsnFromConfig(cfg.PostgreSQL)
    db, err := sql.Open(`postgres`, dsn)
    if err != nil {
        return nil, fmt.Errorf(`postgresql: open database: %w`, err)
    }
    return &PostgresAdapter{dsn: dsn, db: db, state: persistence.NewSnapshot()}, nil
}

func (a *PostgresAdapter) Name() string { return `postgresql:` + a.dsn }

func (a *PostgresAdapter) LoadAll(ctx context.Context) (*persistence.Snapshot, error) {
    a.mu.Lock()
    defer a.mu.Unlock()
    if err := a.loadLocked(ctx); err != nil {
        return nil, err
    }
    return a.snapshotLocked(), nil
}

func (a *PostgresAdapter) ApplyEvent(ctx context.Context, event *persistence.PersistenceEvent) error {
    a.mu.Lock()
    defer a.mu.Unlock()
    if a.closed {
        return fmt.Errorf(`postgresql: adapter is closed`)
    }
    switch event.Type {
    case persistence.EventTypeCreateEntityType:
        id := event.EntityTypeID
        name := event.EntityTypeName
        if name != `` {
            a.state.EntityTypes[id] = name
            a.state.EntityRTypes[name] = id
        }
        if id > a.state.EntityTypeIDMax {
            a.state.EntityTypeIDMax = id
        }
    case persistence.EventTypeCreateEntity, persistence.EventTypeUpdateEntity:
        e := event.Entity
        if _, ok := a.state.EntityTypes[e.Type]; !ok && event.EntityTypeName != `` {
            a.state.EntityTypes[e.Type] = event.EntityTypeName
            a.state.EntityRTypes[event.EntityTypeName] = e.Type
        }
        upsertEntity(a.state, e)
    case persistence.EventTypeDeleteEntity:
        removeEntity(a.state, event.Entity)
    case persistence.EventTypeCreateRelation, persistence.EventTypeUpdateRelation:
        upsertRelation(a.state, event.Relation)
    case persistence.EventTypeDeleteRelation:
        removeRelation(a.state, event.Relation)
    }
    return nil
}

func (a *PostgresAdapter) Flush(ctx context.Context) error {
    a.mu.Lock()
    defer a.mu.Unlock()
    if a.closed {
        return nil
    }
    return a.flushLocked(ctx)
}

func (a *PostgresAdapter) Close(ctx context.Context) error {
    a.mu.Lock()
    defer a.mu.Unlock()
    if a.closed {
        return nil
    }
    err := a.flushLocked(ctx)
    a.closed = true
    if closeErr := a.db.Close(); closeErr != nil && err == nil {
        err = closeErr
    }
    return err
}

func (a *PostgresAdapter) loadLocked(ctx context.Context) error {
    if _, err := a.db.ExecContext(ctx, createTableSQL); err != nil {
        return fmt.Errorf(`postgresql: ensure schema: %w`, err)
    }
    var payload []byte
    err := a.db.QueryRowContext(ctx, `SELECT payload FROM gits_persistence_state WHERE id = 1`).Scan(&payload)
    if err == sql.ErrNoRows {
        a.state = persistence.NewSnapshot()
        return nil
    }
    if err != nil {
        return fmt.Errorf(`postgresql: load state: %w`, err)
    }
    var loaded persistence.Snapshot
    if len(payload) > 0 {
        if err := json.Unmarshal(payload, &loaded); err != nil {
            return fmt.Errorf(`postgresql: decode state: %w`, err)
        }
    }
    if loaded.EntityTypes == nil {
        loaded.EntityTypes = make(map[int]string)
    }
    if loaded.EntityRTypes == nil {
        loaded.EntityRTypes = make(map[string]int)
    }
    if loaded.EntityIDMax == nil {
        loaded.EntityIDMax = make(map[int]int)
    }
    for id, name := range loaded.EntityTypes {
        if id > loaded.EntityTypeIDMax {
            loaded.EntityTypeIDMax = id
        }
        loaded.EntityRTypes[name] = id
    }
    for _, e := range loaded.Entities {
        if e.ID > loaded.EntityIDMax[e.Type] {
            loaded.EntityIDMax[e.Type] = e.ID
        }
    }
    a.state = &loaded
    return nil
}

func (a *PostgresAdapter) flushLocked(ctx context.Context) error {
    if a.state == nil {
        a.state = persistence.NewSnapshot()
    }
    payload, err := json.Marshal(a.state)
    if err != nil {
        return fmt.Errorf(`postgresql: encode state: %w`, err)
    }
    tx, err := a.db.BeginTx(ctx, nil)
    if err != nil {
        return fmt.Errorf(`postgresql: begin flush: %w`, err)
    }
    defer tx.Rollback()
    if _, err := tx.ExecContext(ctx, createTableSQL); err != nil {
        return fmt.Errorf(`postgresql: ensure schema: %w`, err)
    }
    const upsert = `INSERT INTO gits_persistence_state (id, payload) VALUES ($1, $2::jsonb) ON CONFLICT (id) DO UPDATE SET payload = EXCLUDED.payload`
    if _, err := tx.ExecContext(ctx, upsert, 1, string(payload)); err != nil {
        return fmt.Errorf(`postgresql: store state: %w`, err)
    }
    return tx.Commit()
}

func (a *PostgresAdapter) snapshotLocked() *persistence.Snapshot {
    s := persistence.NewSnapshot()
    for k, v := range a.state.EntityTypes {
        s.EntityTypes[k] = v
        if k > s.EntityTypeIDMax {
            s.EntityTypeIDMax = k
        }
    }
    for k, v := range a.state.EntityRTypes {
        s.EntityRTypes[k] = v
    }
    s.EntityTypeIDMax = a.state.EntityTypeIDMax
    for k, v := range a.state.EntityIDMax {
        s.EntityIDMax[k] = v
    }
    if len(a.state.Entities) > 0 {
        s.Entities = append(s.Entities, a.state.Entities...)
    }
    if len(a.state.Relations) > 0 {
        s.Relations = append(s.Relations, a.state.Relations...)
    }
    return s
}

func upsertEntity(s *persistence.Snapshot, e types.StorageEntity) {
    filtered := s.Entities[:0]
    for _, x := range s.Entities {
        if x.Type != e.Type || x.ID != e.ID {
            filtered = append(filtered, x)
        }
    }
    s.Entities = append(filtered, e)
    if e.ID > s.EntityIDMax[e.Type] {
        s.EntityIDMax[e.Type] = e.ID
    }
}

func removeEntity(s *persistence.Snapshot, e types.StorageEntity) {
    filtered := s.Entities[:0]
    for _, x := range s.Entities {
        if !(x.Type == e.Type && x.ID == e.ID) {
            filtered = append(filtered, x)
        }
    }
    s.Entities = filtered
}

func upsertRelation(s *persistence.Snapshot, r types.StorageRelation) {
    filtered := s.Relations[:0]
    for _, x := range s.Relations {
        if !sameRelation(x, r) {
            filtered = append(filtered, x)
        }
    }
    s.Relations = append(filtered, r)
}

func removeRelation(s *persistence.Snapshot, r types.StorageRelation) {
    filtered := s.Relations[:0]
    for _, x := range s.Relations {
        if !sameRelation(x, r) {
            filtered = append(filtered, x)
        }
    }
    s.Relations = filtered
}

func sameRelation(a, b types.StorageRelation) bool {
    return a.SourceType == b.SourceType && a.SourceID == b.SourceID && a.TargetType == b.TargetType && a.TargetID == b.TargetID
}

func dsnFromConfig(cfg *persistence.PostgresConfig) string {
    if cfg.DSN != `` {
        return cfg.DSN
    }
    u := url.URL{
        Scheme: `postgres`,
        User:   url.UserPassword(cfg.User, cfg.Password),
        Host:   fmt.Sprintf(`%s:%d`, cfg.Host, cfg.Port),
        Path:   `/` + cfg.Database,
    }
    q := u.Query()
    q.Set(`sslmode`, `disable`)
    u.RawQuery = q.Encode()
    return u.String()
}