package datastore

import (
	"errors"

	"gno.land/p/nt/avl/v0"
)

// ErrStorageExists indicates that a storage exists with the same name.
var ErrStorageExists = errors.New("a storage with the same name exists")

// Datastore is a store that can contain multiple named storages.
// A storage is a collection of records.
//
// Example usage:
//
//	// Create an empty storage to store user records
//	var db Datastore
//	storage := db.CreateStorage("users")
//
//	// Get a storage that has been created before
//	storage = db.GetStorage("profiles")
type Datastore struct {
	storages avl.Tree // string(name) -> *Storage
}

// CreateStorage creates a new named storage within the data store.
func (ds *Datastore) CreateStorage(name string, options ...StorageOption) *Storage {
	if ds.storages.Has(name) {
		return nil
	}

	s := NewStorage(name, options...)
	ds.storages.Set(name, &s)
	return &s
}

// HasStorage checks if data store contains a storage with a specific name.
func (ds Datastore) HasStorage(name string) bool {
	return ds.storages.Has(name)
}

// GetStorage returns a storage that has been created with a specific name.
// It returns nil when a storage with the specified name is not found.
func (ds Datastore) GetStorage(name string) *Storage {
	if v, found := ds.storages.Get(name); found {
		return v.(*Storage)
	}
	return nil
}
