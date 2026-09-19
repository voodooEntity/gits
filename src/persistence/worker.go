package persistence

import (
	"context"
	"fmt"
)

// queueItem is a unit of work for the persistence worker.
type queueItem struct {
	event *PersistenceEvent
	done  chan struct{}
}

// runWorker consumes queued events in order and applies them to the adapter.
func (m *PersistenceManager) runWorker() {
	defer close(m.workerDone)
	ctx := context.Background()

	for item := range m.queue {
		if item.event != nil {
			err := m.adapter.ApplyEvent(ctx, item.event)
			if err != nil {
				m.mu.Lock()
				m.applyErr = combinePersistenceErrors(m.applyErr, fmt.Errorf("persistence apply %s: %w", item.event.Type, err))
				m.mu.Unlock()
			}
		}

		if item.done != nil {
			close(item.done)
		}
	}
}
