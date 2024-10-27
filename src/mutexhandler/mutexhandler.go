package mutexhandler

import (
	"github.com/voodooEntity/archivist"
	"github.com/voodooEntity/gits/src/storage"
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
	Storage *storage.Storage
	Applied []int
}

func New(store *storage.Storage) *MutexHandler {
	tmp := MutexHandler{
		Storage: store,
	}
	return &tmp
}

func (self *MutexHandler) Apply(muident int) *MutexHandler {
	// first we check if this is locked already, this should not be neccesary but we running in an issue atm that might be caused due to this ###
	if 0 < len(self.Applied) {
		for _, val := range self.Applied {
			if val == muident {
				archivist.Debug("Trying to multi-apply same lock in MutexHandler")
				return self
			}
		}
	}

	// prepare applied flag
	applied := false
	// apply mmutex
	switch muident {
	case EntityTypeLock:
		self.Storage.EntityTypeMutex.Lock()
		applied = true
	case EntityTypeRLock:
		self.Storage.EntityTypeMutex.RLock()
		applied = true
	case EntityStorageLock:
		self.Storage.EntityStorageMutex.Lock()
		applied = true
	case EntityStorageRLock:
		self.Storage.EntityStorageMutex.RLock()
		applied = true
	case RelationStorageLock:
		self.Storage.RelationStorageMutex.Lock()
		applied = true
	case RelationStorageRLock:
		self.Storage.RelationStorageMutex.RLock()
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
			self.Storage.EntityTypeMutex.Unlock()
		case EntityTypeRLock:
			self.Storage.EntityTypeMutex.RUnlock()
		case EntityStorageLock:
			self.Storage.EntityStorageMutex.Unlock()
		case EntityStorageRLock:
			self.Storage.EntityStorageMutex.RUnlock()
		case RelationStorageLock:
			self.Storage.RelationStorageMutex.Unlock()
		case RelationStorageRLock:
			self.Storage.RelationStorageMutex.RUnlock()
		}
	}
}
