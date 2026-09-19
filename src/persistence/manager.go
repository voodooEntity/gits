package persistence

import (
	"context"
	"errors"
	"fmt"
	"sync"
)

// ErrManagerClosed is returned when an event is submitted after Close has started.
var ErrManagerClosed = errors.New("persistence manager is closed")

const defaultQueueSize = 256

// PersistenceManager owns the asynchronous persistence worker for one storage instance.
type PersistenceManager struct {
	adapter    PersistenceAdapter
	queue      chan queueItem
	workerDone chan struct{}

	mu        sync.Mutex
	seq       uint64
	closed    bool
	closeWait chan struct{}
	closeErr  error
	applyErr  error
}

// NewManager creates a manager for adapter. A nil adapter produces a safe no-op manager.
func NewManager(adapter PersistenceAdapter, bufferSize int) *PersistenceManager {
	if adapter == nil {
		return &PersistenceManager{}
	}
	if bufferSize <= 0 {
		bufferSize = defaultQueueSize
	}
	m := &PersistenceManager{
		adapter:    adapter,
		queue:      make(chan queueItem, bufferSize),
		workerDone: make(chan struct{}),
	}
	go m.runWorker()
	return m
}

// Enabled reports whether the manager has an active persistence adapter.
func (m *PersistenceManager) Enabled() bool {
	return m != nil && m.adapter != nil
}

// AdapterName returns the configured adapter name, or empty for no-op mode.
func (m *PersistenceManager) AdapterName() string {
	if !m.Enabled() {
		return ""
	}
	return m.adapter.Name()
}

// Enqueue assigns a sequence number and schedules event for asynchronous application.
// Callers should invoke this after mutating in-memory state but still under the storage lock(s).
func (m *PersistenceManager) Enqueue(event *PersistenceEvent) error {
	if !m.Enabled() || event == nil {
		return nil
	}

	m.mu.Lock()
	defer m.mu.Unlock()

	if m.closed {
		return ErrManagerClosed
	}

	m.seq++
	cloned := cloneEvent(event)
	cloned.Sequence = m.seq

	select {
	case m.queue <- queueItem{event: cloned}:
		return nil
	case <-m.workerDone:
		return ErrManagerClosed
	}
}

// Bootstrap blocks until adapter.LoadAll completes. If apply is non-nil, it is invoked with the loaded snapshot.
func (m *PersistenceManager) Bootstrap(ctx context.Context, apply func(*Snapshot) error) error {
	if ctx == nil {
		ctx = context.Background()
	}
	if !m.Enabled() {
		return nil
	}

	snapshot, err := m.adapter.LoadAll(ctx)
	if err != nil {
		return fmt.Errorf("persistence bootstrap load: %w", err)
	}
	if snapshot == nil {
		snapshot = NewSnapshot()
	}
	if apply != nil {
		if err := apply(snapshot); err != nil {
			return fmt.Errorf("persistence bootstrap apply: %w", err)
		}
	}
	return nil
}

// Flush blocks until all events enqueued before the call have been applied by the worker.
func (m *PersistenceManager) Flush(ctx context.Context) error {
	if ctx == nil {
		ctx = context.Background()
	}
	if !m.Enabled() {
		return nil
	}

	done := make(chan struct{})
	m.mu.Lock()
	if m.closed {
		m.mu.Unlock()
		return nil
	}
	select {
	case m.queue <- queueItem{done: done}:
	case <-m.workerDone:
		m.mu.Unlock()
		return nil
	}
	m.mu.Unlock()

	select {
	case <-done:
		return nil
	case <-ctx.Done():
		return ctx.Err()
	}
}

// Close drains the queue, flushes the adapter, and closes it. It is safe to call multiple times.
func (m *PersistenceManager) Close(ctx context.Context) error {
	if ctx == nil {
		ctx = context.Background()
	}
	if !m.Enabled() {
		return nil
	}

	m.mu.Lock()
	if m.closed {
		wait := m.closeWait
		m.mu.Unlock()
		if wait != nil {
			<-wait
		}
		m.mu.Lock()
		err := m.closeErr
		m.mu.Unlock()
		return err
	}

	m.closed = true
	m.closeWait = make(chan struct{})
	close(m.queue)
	adapter := m.adapter
	m.mu.Unlock()

	select {
	case <-m.workerDone:
	case <-ctx.Done():
		m.mu.Lock()
		m.closeErr = ctx.Err()
		m.mu.Unlock()
		close(m.closeWait)
		return ctx.Err()
	}

	m.mu.Lock()
	err := m.applyErr
	m.mu.Unlock()

	if flushErr := adapter.Flush(ctx); flushErr != nil {
		err = combinePersistenceErrors(err, fmt.Errorf("persistence flush: %w", flushErr))
	}
	if closeErr := adapter.Close(ctx); closeErr != nil {
		err = combinePersistenceErrors(err, fmt.Errorf("persistence close: %w", closeErr))
	}

	m.mu.Lock()
	m.closeErr = err
	m.mu.Unlock()
	close(m.closeWait)
	return err
}

func cloneEvent(event *PersistenceEvent) *PersistenceEvent {
	cloned := *event
	cloned.Entity.Properties = copyStringMap(cloned.Entity.Properties)
	cloned.Relation.Properties = copyStringMap(cloned.Relation.Properties)
	return &cloned
}

func copyStringMap(values map[string]string) map[string]string {
	if len(values) == 0 {
		return nil
	}
	out := make(map[string]string, len(values))
	for key, value := range values {
		out[key] = value
	}
	return out
}

func combinePersistenceErrors(first error, second error) error {
	if first == nil {
		return second
	}
	if second == nil {
		return first
	}
	return fmt.Errorf("%v; %w", first, second)
}
