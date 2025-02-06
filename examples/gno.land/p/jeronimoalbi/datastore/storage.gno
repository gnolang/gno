package datastore

import (
	"strings"

	"gno.land/p/moul/collection"
)

func NewStorage(name string, options ...StorageOption) Storage {
	s := Storage{
		name:         name,
		currentIndex: collection.IDIndex,
		collection:   collection.New(),
		schema:       NewSchema(strings.Title(name)),
	}

	for _, apply := range options {
		apply(&s)
	}
	return s
}

type Storage struct {
	name         string
	currentIndex string
	collection   *collection.Collection
	schema       *Schema
}

func (s Storage) Name() string {
	return s.name
}

func (s Storage) CurrentIndex() string {
	return s.currentIndex
}

func (s Storage) Collection() *collection.Collection {
	return s.collection
}

func (s Storage) Schema() *Schema {
	return s.schema
}

func (s Storage) NewRecord() Record {
	r := &record{
		schema:     s.schema,
		collection: s.collection,
	}

	// Assign default record values if the schema defines them
	for i, name := range s.schema.Fields() {
		if v, found := s.schema.GetDefaultByIndex(i); found {
			r.Set(name, v)
		}
	}
	return r
}

func (s Storage) WithIndex(name string) Storage {
	s.currentIndex = name
	return s
}

func (s Storage) ForEach(fn func(Record) bool) bool {
	idx := s.collection.GetIndex(s.currentIndex)
	if idx == nil {
		return false
	}

	return idx.Iterate("", "", func(_ string, v interface{}) bool {
		return fn(v.(Record))
	})
}

func (s Storage) Get(key string) []Record {
	var (
		records []Record
		iter    = s.collection.Get(s.currentIndex, key)
	)
	for iter.Next() {
		records = append(records, iter.Value().Obj.(Record))
	}
	return records
}

func (s Storage) GetFirst(key string) Record {
	var (
		record Record
		iter   = s.collection.Get(s.currentIndex, key)
	)
	if iter.Next() {
		record = iter.Value().Obj.(Record)
	}
	return record
}

func (s Storage) Delete(id uint64) bool {
	return s.collection.Delete(id)
}
