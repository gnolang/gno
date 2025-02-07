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

// WithDefaultIndex assigns an index to the storage and makes it the default one.
// Default indexes are used by default by a storage without the need to use the
// `Storage.WithIndex()` method to use a different index when searching or iterating.
func WithDefaultIndex(i Index) StorageOption {
	return func(st *Storage) {
		st.currentIndex = i.Name()
		st.collection.AddIndex(i.name, i.fn, i.options)
	}
}
