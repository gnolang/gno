package datastore

import "strings"

// StorageOption configures schemas.
type SchemaOption func(*Schema)

// WithField assign a new field to the schema definition.
func WithField(name string) SchemaOption {
	return func(s *Schema) {
		name = strings.TrimSpace(name)
		if name != "" {
			s.fields.Append(name)
		}
	}
}

// WithDefaultField assign a new field with a default value to the schema definition.
// Default value is assigned to newly created records asociated to to schema.
func WithDefaultField(name string, value interface{}) SchemaOption {
	return func(s *Schema) {
		name = strings.TrimSpace(name)
		if name != "" {
			s.fields.Append(name)

			key := castIntToKey(s.fields.Len() - 1)
			s.defaults.Set(key, value)
		}
	}
}

// Strict configures the schema as a strict one.
// By default schemas should allow the creation of any user defined field,
// making them strict limits the allowed record fields to the ones pre-defined
// in the schema. Fields are pre-defined using `WithField`, `WithDefaultField`
// or by calling `Schema.AddField()`.
func Strict() SchemaOption {
	return func(s *Schema) {
		s.strict = true
	}
}
