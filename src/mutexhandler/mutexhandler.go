package mutexhandler

import (
	"github.com/voodooEntity/gits"
)

const (
	EntityTypeLock       = 1
	EntityTypeRLock      = 2
	EntityStorageLock    = 3
	EntityStorageRLock   = 4
	RelationStorageLock  = 5
	RelationStorageRLock = 6
)

type MutexHandler struct {
	Applied []int
}

func New() *MutexHandler {
	tmp := MutexHandler{}
	return &tmp
}

func (self *MutexHandler) Apply(muident int) *MutexHandler {
	// prepare applied flag
	applied := false
	// apply mmutex
	switch muident {
	case EntityTypeLock:
		gits.EntityTypeMutex.Lock()
		applied = true
	case EntityTypeRLock:
		gits.EntityTypeMutex.RLock()
		applied = true
	case EntityStorageLock:
		gits.EntityStorageMutex.Lock()
		applied = true
	case EntityStorageRLock:
		gits.EntityStorageMutex.RLock()
		applied = true
	case RelationStorageLock:
		gits.RelationStorageMutex.Lock()
		applied = true
	case RelationStorageRLock:
		gits.RelationStorageMutex.RLock()
		applied = true
	}
	// if a Mutex was applied, add the muname to our Applied list
	if applied {
		self.Applied = append(self.Applied, muident)
	}
	return self
}

func (self *MutexHandler) Release() {
	for _, muident := range self.Applied {
		// apply mmutex
		switch muident {
		case EntityTypeLock:
			gits.EntityTypeMutex.Unlock()
		case EntityTypeRLock:
			gits.EntityTypeMutex.RUnlock()
		case EntityStorageLock:
			gits.EntityStorageMutex.Unlock()
		case EntityStorageRLock:
			gits.EntityStorageMutex.RUnlock()
		case RelationStorageLock:
			gits.RelationStorageMutex.Unlock()
		case RelationStorageRLock:
			gits.RelationStorageMutex.RUnlock()
		}
	}
}
