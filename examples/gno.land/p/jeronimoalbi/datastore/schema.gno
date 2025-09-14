package datastore

import (
	"encoding/binary"

	"gno.land/p/nt/avl"
	"gno.land/p/nt/avl/list"
)

// TODO: Support versioning

// Schema contains information about fields and default field values.
// It also offers the possibility to configure it as static to indicate
// that only configured fields should be allowed.
type Schema struct {
	name     string
	strict   bool
	fields   list.List // int(field index) -> string(field name)
	defaults avl.Tree  // string(field index) -> interface{}
}

// NewSchema creates a new schema.
func NewSchema(name string, options ...SchemaOption) *Schema {
	s := &Schema{name: name}
	for _, apply := range options {
		apply(s)
	}
	return s
}

// Name returns schema's name.
func (s Schema) Name() string {
	return s.name
}

// Fields returns the list field names that are defined in the schema.
func (s Schema) Fields() []string {
	fields := make([]string, s.fields.Len())
	s.fields.ForEach(func(i int, v interface{}) bool {
		fields[i] = v.(string)
		return false
	})
	return fields
}

// Size returns the number of fields the schema has.
func (s Schema) Size() int {
	return s.fields.Len()
}

// IsStrict check if the schema is configured as a strict one.
func (s Schema) IsStrict() bool {
	return s.strict
}

// HasField check is a field has been defined in the schema.
func (s Schema) HasField(name string) bool {
	return s.GetFieldIndex(name) >= 0
}

// AddField adds a new field to the schema.
// A default field value can be specified, otherwise `defaultValue` must be nil.
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

// GetFieldIndex returns the index number of a schema field.
//
// Field index indicates the order the field has within the schema.
// When defined fields are added they get an index starting from
// field index 0.
//
// Fields are internally referenced by index number instead of the name
// to be able to rename fields easily.
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

// GetFieldName returns the name of a field for a specific field index.
func (s Schema) GetFieldName(index int) (name string, found bool) {
	v := s.fields.Get(index)
	if v == nil {
		return "", false
	}
	return v.(string), true
}

// GetDefault returns the default value for a field.
func (s Schema) GetDefault(name string) (value interface{}, found bool) {
	i := s.GetFieldIndex(name)
	if i == -1 {
		return nil, false
	}
	return s.GetDefaultByIndex(i)
}

// GetDefaultByIndex returns the default value for a field by it's index.
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

// RenameField renames a field.
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
