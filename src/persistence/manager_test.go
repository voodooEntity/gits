package persistence

import (
	"context"
	"sync"
	"testing"
	"time"

	"github.com/voodooEntity/gits/src/types"
)

type fakeAdapter struct {
	name             string
	loadStarted      chan struct{}
	loadRelease      chan struct{}
	applyStarted     chan struct{}
	applyRelease     chan struct{}
	applySignalOnce  sync.Once
	applyReleaseOnce sync.Once

	mu      sync.Mutex
	applied []*PersistenceEvent
	flushed bool
	closed  bool
}

func (f *fakeAdapter) Name() string { return f.name }

func (f *fakeAdapter) LoadAll(ctx context.Context) (*Snapshot, error) {
	if f.loadStarted != nil {
		close(f.loadStarted)
		if f.loadRelease != nil {
			<-f.loadRelease
		}
	}
	return NewSnapshot(), nil
}

func (f *fakeAdapter) ApplyEvent(ctx context.Context, event *PersistenceEvent) error {
	if f.applyStarted != nil {
		f.applySignalOnce.Do(func() { close(f.applyStarted) })
		if f.applyRelease != nil {
			<-f.applyRelease
		}
	}

	f.mu.Lock()
	f.applied = append(f.applied, event)
	f.mu.Unlock()
	return nil
}

func (f *fakeAdapter) Flush(ctx context.Context) error {
	f.mu.Lock()
	f.flushed = true
	f.mu.Unlock()
	return nil
}

func (f *fakeAdapter) Close(ctx context.Context) error {
	f.mu.Lock()
	f.closed = true
	f.mu.Unlock()
	return nil
}

func (f *fakeAdapter) appliedCount() int {
	f.mu.Lock()
	defer f.mu.Unlock()
	return len(f.applied)
}

func (f *fakeAdapter) sequences() []uint64 {
	f.mu.Lock()
	defer f.mu.Unlock()
	out := make([]uint64, len(f.applied))
	for i, event := range f.applied {
		out[i] = event.Sequence
	}
	return out
}

func (f *fakeAdapter) entityIDs() []int {
	f.mu.Lock()
	defer f.mu.Unlock()
	out := make([]int, len(f.applied))
	for i, event := range f.applied {
		out[i] = event.Entity.ID
	}
	return out
}

func (f *fakeAdapter) isFlushed() bool {
	f.mu.Lock()
	defer f.mu.Unlock()
	return f.flushed
}

func (f *fakeAdapter) isClosed() bool {
	f.mu.Lock()
	defer f.mu.Unlock()
	return f.closed
}

func (f *fakeAdapter) releaseApply() {
	f.applyReleaseOnce.Do(func() {
		if f.applyRelease != nil {
			close(f.applyRelease)
		}
	})
}

func TestPersistenceManagerOrderedEnqueue(t *testing.T) {
	fake := &fakeAdapter{name: "ordered"}
	m := NewManager(fake, 8)

	for i := 0; i < 5; i++ {
		event := &PersistenceEvent{
			Type:   EventTypeCreateEntity,
			Entity: types.StorageEntity{ID: i + 1},
		}
		if err := m.Enqueue(event); err != nil {
			t.Fatalf("Enqueue() error = %v", err)
		}
	}

	if err := m.Flush(context.Background()); err != nil {
		t.Fatalf("Flush() error = %v", err)
	}

	sequences := fake.sequences()
	if len(sequences) != 5 {
		t.Fatalf("applied event count = %d, want 5", len(sequences))
	}
	for i, sequence := range sequences {
		if sequence != uint64(i+1) {
			t.Fatalf("sequence[%d] = %d, want %d", i, sequence, i+1)
		}
	}

	entityIDs := fake.entityIDs()
	for i, id := range entityIDs {
		if id != i+1 {
			t.Fatalf("entity ID[%d] = %d, want %d", i, id, i+1)
		}
	}

	if err := m.Close(context.Background()); err != nil {
		t.Fatalf("Close() error = %v", err)
	}
}

func TestPersistenceManagerBootstrapBlocksUntilLoadAll(t *testing.T) {
	fake := &fakeAdapter{
		name:        "bootstrap",
		loadStarted: make(chan struct{}),
		loadRelease: make(chan struct{}),
	}
	m := NewManager(fake, 4)
	defer m.Close(context.Background())

	applyDone := make(chan struct{})
	bootstrapDone := make(chan error, 1)
	go func() {
		err := m.Bootstrap(context.Background(), func(snapshot *Snapshot) error {
			close(applyDone)
			return nil
		})
		bootstrapDone <- err
	}()

	select {
	case <-fake.loadStarted:
	case <-time.After(time.Second):
		t.Fatal("LoadAll did not start")
	}

	select {
	case <-applyDone:
		t.Fatal("snapshot applied before LoadAll completed")
	default:
	}

	close(fake.loadRelease)

	select {
	case err := <-bootstrapDone:
		if err != nil {
			t.Fatalf("Bootstrap() error = %v", err)
		}
	case <-time.After(time.Second):
		t.Fatal("Bootstrap did not complete")
	}

	select {
	case <-applyDone:
	case <-time.After(time.Second):
		t.Fatal("snapshot was not applied after LoadAll completed")
	}
}

func TestPersistenceManagerFlushWaitsForPriorEvents(t *testing.T) {
	fake := &fakeAdapter{
		name:         "flush",
		applyStarted: make(chan struct{}),
		applyRelease: make(chan struct{}),
	}
	m := NewManager(fake, 4)
	defer m.Close(context.Background())

	if err := m.Enqueue(&PersistenceEvent{Type: EventTypeCreateEntity}); err != nil {
		t.Fatalf("Enqueue() error = %v", err)
	}

	flushDone := make(chan error, 1)
	go func() {
		flushDone <- m.Flush(context.Background())
	}()

	select {
	case <-fake.applyStarted:
	case <-time.After(time.Second):
		fake.releaseApply()
		t.Fatal("ApplyEvent did not start")
	}

	if count := fake.appliedCount(); count != 0 {
		fake.releaseApply()
		t.Fatalf("applied event count before release = %d, want 0", count)
	}

	fake.releaseApply()

	select {
	case err := <-flushDone:
		if err != nil {
			t.Fatalf("Flush() error = %v", err)
		}
	case <-time.After(time.Second):
		t.Fatal("Flush did not complete")
	}

	if count := fake.appliedCount(); count != 1 {
		t.Fatalf("applied event count after flush = %d, want 1", count)
	}
}

func TestPersistenceManagerCloseDrainsQueueAndFlushesAdapter(t *testing.T) {
	fake := &fakeAdapter{name: "close"}
	m := NewManager(fake, 4)

	for i := 0; i < 3; i++ {
		if err := m.Enqueue(&PersistenceEvent{Type: EventTypeCreateEntity}); err != nil {
			t.Fatalf("Enqueue() error = %v", err)
		}
	}

	if err := m.Close(context.Background()); err != nil {
		t.Fatalf("Close() error = %v", err)
	}

	if count := fake.appliedCount(); count != 3 {
		t.Fatalf("applied event count after close = %d, want 3", count)
	}
	if !fake.isFlushed() {
		t.Fatal("adapter Flush was not called")
	}
	if !fake.isClosed() {
		t.Fatal("adapter Close was not called")
	}

	if err := m.Enqueue(&PersistenceEvent{Type: EventTypeCreateEntity}); err != ErrManagerClosed {
		t.Fatalf("Enqueue after close error = %v, want ErrManagerClosed", err)
	}
	if err := m.Close(context.Background()); err != nil {
		t.Fatalf("second Close() error = %v", err)
	}
}

func TestPersistenceManagerNilAdapterNoOp(t *testing.T) {
	m := NewManager(nil, 0)
	if m.Enabled() {
		t.Fatal("nil adapter manager should be disabled")
	}

	if err := m.Enqueue(&PersistenceEvent{Type: EventTypeCreateEntity}); err != nil {
		t.Fatalf("Enqueue() error = %v", err)
	}

	bootstrapCalled := false
	err := m.Bootstrap(context.Background(), func(snapshot *Snapshot) error {
		bootstrapCalled = true
		return nil
	})
	if err != nil {
		t.Fatalf("Bootstrap() error = %v", err)
	}
	if bootstrapCalled {
		t.Fatal("nil adapter Bootstrap should not invoke apply")
	}

	if err := m.Flush(context.Background()); err != nil {
		t.Fatalf("Flush() error = %v", err)
	}
	if err := m.Close(context.Background()); err != nil {
		t.Fatalf("Close() error = %v", err)
	}
}
