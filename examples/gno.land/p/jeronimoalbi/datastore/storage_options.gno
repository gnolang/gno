package datastore

type StorageOption func(*Storage)

func WithSchema(s *Schema) StorageOption {
	return func(st *Storage) {
		if s != nil {
			st.schema = s
		}
	}
}

func WithIndex(i Index) StorageOption {
	return func(st *Storage) {
		st.collection.AddIndex(i.name, i.fn, i.options)
	}
}

func WithDefaultIndex(i Index) StorageOption {
	return func(st *Storage) {
		st.currentIndex = i.Name()
		st.collection.AddIndex(i.name, i.fn, i.options)
	}
}
