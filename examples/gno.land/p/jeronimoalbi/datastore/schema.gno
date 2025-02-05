package datastore

import (
	"encoding/binary"
	"strings"

	"gno.land/p/demo/avl"
	"gno.land/p/demo/avl/list"
)

// TODO: Support versioning

type Schema struct {
	name     string
	strict   bool
	fields   list.List // int(field index) -> string(field name)
	defaults avl.Tree  // string(field index) -> interface{}
}

func NewSchema(name string, options ...SchemaOption) *Schema {
	s := &Schema{name: name}
	for _, apply := range options {
		apply(s)
	}
	return s
}

func (s Schema) Name() string {
	return s.name
}

func (s Schema) Fields() []string {
	fields := make([]string, s.fields.Len())
	s.fields.ForEach(func(i int, v interface{}) bool {
		fields[i] = v.(string)
		return false
	})
	return fields
}

func (s Schema) Size() int {
	return s.fields.Len()
}

func (s Schema) IsStrict() bool {
	return s.strict
}

func (s Schema) HasField(name string) bool {
	return s.GetFieldIndex(name) >= 0
}

func (s *Schema) AddField(name string, defaultValue interface{}) (index int, added bool) {
	if s.HasField(name) {
		return -1, false
	}

	s.fields.Append(name)
	index = s.fields.Len() - 1
	if defaultValue != nil {
		key := castIntToKey(index)
		s.defaults.Set(key, defaultValue)
	}
	return index, true
}

func (s Schema) GetFieldIndex(name string) int {
	index := -1
	s.fields.ForEach(func(i int, v interface{}) bool {
		if name != v.(string) {
			return false
		}

		index = i
		return true
	})
	return index
}

func (s Schema) GetFieldName(index int) (name string, found bool) {
	v := s.fields.Get(index)
	if v == nil {
		return "", false
	}
	return v.(string), true
}

func (s Schema) GetDefault(name string) (value interface{}, found bool) {
	i := s.GetFieldIndex(name)
	if i == -1 {
		return nil, false
	}
	return s.GetDefaultByIndex(i)
}

func (s Schema) GetDefaultByIndex(index int) (value interface{}, found bool) {
	key := castIntToKey(index)
	v, found := s.defaults.Get(key)
	if !found {
		return nil, false
	}

	if fn, ok := v.(func() interface{}); ok {
		return fn(), true
	}
	return v, true
}

func (s *Schema) RenameField(name, newName string) (renamed bool) {
	if s.HasField(newName) {
		return false
	}

	i := s.GetFieldIndex(name)
	if i == -1 {
		return false
	}

	s.fields.Set(i, newName)
	return true
}

func castIntToKey(i int) string {
	buf := make([]byte, 8)
	binary.BigEndian.PutUint64(buf, uint64(i))
	return string(buf)
}
