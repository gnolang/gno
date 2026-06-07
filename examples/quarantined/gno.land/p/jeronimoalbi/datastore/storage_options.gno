package datastore

// StorageOption configures storages.
type StorageOption func(*Storage)

// WithSchema assigns a schema to the storage.
func WithSchema(s *Schema) StorageOption {
	return func(st *Storage) {
		if s != nil {
			st.schema = s
		}
	}
}

// WithIndex assigns an index to the storage.
func WithIndex(i Index) StorageOption {
	return func(st *Storage) {
		st.collection.AddIndex(i.name, i.fn, i.options)
	}
}
