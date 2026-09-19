package persistence

import "context"

// PersistenceAdapter is the extension point for storage backends.
type PersistenceAdapter interface {
	Name() string
	LoadAll(ctx context.Context) (*Snapshot, error)
	ApplyEvent(ctx context.Context, event *PersistenceEvent) error
	Flush(ctx context.Context) error
	Close(ctx context.Context) error
}
