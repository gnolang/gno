package datastore

import (
	"errors"

	"gno.land/p/demo/avl"
	"gno.land/p/demo/seqid"
	"gno.land/p/moul/collection"
)

var ErrUndefinedField = errors.New("undefined field")

type (
	ReadOnlyRecord interface {
		ID() uint64
		Key() string
		Type() string
		Fields() []string
		IsEmpty() bool
		HasField(name string) bool
		Get(field string) (value interface{}, found bool)
		MustGet(field string) interface{}
	}

	Record interface {
		ReadOnlyRecord

		Set(field string, value interface{}) error
		Save() bool
	}
)

type record struct {
	id         uint64
	schema     *Schema
	collection *collection.Collection
	values     avl.Tree // string(field index) -> interface{}
}

func (r record) ID() uint64 {
	return r.id
}

func (r record) Key() string {
	return seqid.ID(r.id).String()
}

func (r record) Type() string {
	return r.schema.Name()
}

func (r record) Fields() []string {
	return r.schema.Fields()
}

func (r record) IsEmpty() bool {
	return r.values.Size() == 0
}

func (r record) HasField(name string) bool {
	return r.schema.HasField(name)
}

func (r *record) Set(field string, value interface{}) error {
	i := r.schema.GetFieldIndex(field)
	if i == -1 {
		if r.schema.IsStrict() {
			return ErrUndefinedField
		}

		i, _ = r.schema.AddField(field, nil)
	}

	key := castIntToKey(i)
	r.values.Set(key, value)
	return nil
}

func (r record) Get(field string) (value interface{}, found bool) {
	i := r.schema.GetFieldIndex(field)
	if i == -1 {
		return nil, false
	}

	key := castIntToKey(i)
	return r.values.Get(key)
}

func (r record) MustGet(field string) interface{} {
	v, found := r.Get(field)
	if !found {
		panic("field not found: " + field)
	}
	return v
}

func (r *record) Save() bool {
	if r.id == 0 {
		r.id = r.collection.Set(r)
		return r.id != 0
	}
	return r.collection.Update(r.id, r)
}
