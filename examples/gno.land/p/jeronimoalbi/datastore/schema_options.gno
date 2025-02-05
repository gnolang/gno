package datastore

import "strings"

type SchemaOption func(*Schema)

func WithField(name string) SchemaOption {
	return func(s *Schema) {
		name = strings.TrimSpace(name)
		if name != "" {
			s.fields.Append(name)
		}
	}
}

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

func Strict() SchemaOption {
	return func(s *Schema) {
		s.strict = true
	}
}
