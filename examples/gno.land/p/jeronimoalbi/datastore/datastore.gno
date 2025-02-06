package datastore

import (
	"errors"

	"gno.land/p/demo/avl"
)

// TODO: Write unit test for Datastore

var ErrStorageExists = errors.New("a storage with the same name exists")

type Datastore struct {
	storages avl.Tree // string(name) -> *Storage
}

func (ds *Datastore) CreateStorage(name string, options ...StorageOption) *Storage {
	if ds.storages.Has(name) {
		return nil
	}

	s := NewStorage(name, options...)
	ds.storages.Set(name, &s)
	return &s
}

func (ds Datastore) HasStorage(name string) bool {
	return ds.storages.Has(name)
}

func (ds Datastore) GetStorage(name string) *Storage {
	if v, found := ds.storages.Get(name); found {
		return v.(*Storage)
	}
	return nil
}
