package datastore

import (
	"strings"
	"testing"

	"gno.land/p/demo/uassert"
	"gno.land/p/demo/urequire"
	"gno.land/p/moul/collection"
)

func TestStorageDefaults(t *testing.T) {
	name := "foo"
	storage := NewStorage(name)

	uassert.Equal(t, name, storage.Name())
	uassert.Equal(t, collection.IDIndex, storage.CurrentIndex())
	uassert.NotEqual(t, nil, storage.Collection())
	uassert.Equal(t, 0, storage.Size())

	s := storage.Schema()
	uassert.NotEqual(t, nil, s)
	uassert.Equal(t, strings.Title(name), s.Name())
}

func TestStorageNewRecord(t *testing.T) {
	field := "status"
	defaultValue := "testing"
	s := NewSchema("Foo", WithDefaultField(field, defaultValue))
	storage := NewStorage("foo", WithSchema(s))

	r := storage.NewRecord()
	urequire.NotEqual(t, nil, r, "new record is not nil")
	uassert.Equal(t, uint64(0), r.ID())
	uassert.Equal(t, storage.Schema().Name(), r.Type())

	v, found := r.Get(field)
	urequire.True(t, found, "default value found")

	got, ok := v.(string)
	urequire.True(t, ok, "default value type")
	uassert.Equal(t, defaultValue, got)
}

// TODO: Finish and improve Storage unit tests
