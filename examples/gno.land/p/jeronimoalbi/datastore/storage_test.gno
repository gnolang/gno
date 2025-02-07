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

func TestStorageWithIndex(t *testing.T) {
	index := "bar"
	storage1 := NewStorage("foo")
	storage2 := storage1.WithIndex(index)

	uassert.Equal(t, collection.IDIndex, storage1.CurrentIndex())
	uassert.Equal(t, storage1.Name(), storage2.Name())
	uassert.Equal(t, index, storage2.CurrentIndex())
}

func TestStorageIterate(t *testing.T) {
	index := NewIndex("name", func(r Record) string {
		if v, found := r.Get("name"); found {
			return v.(string)
		}
		return ""
	})

	cases := []struct {
		name      string
		recordIDs []uint64
		setup     func(*Storage)
	}{
		{
			name:      "single record",
			recordIDs: []uint64{1},
			setup: func(s *Storage) {
				r := s.NewRecord()
				r.Set("name", "foobar")
				r.Save()
			},
		},
		{
			name:      "two records",
			recordIDs: []uint64{1, 2},
			setup: func(s *Storage) {
				r := s.NewRecord()
				r.Set("name", "foobar")
				r.Save()

				r = s.NewRecord()
				r.Set("name", "foobar")
				r.Save()
			},
		},
		{
			name: "empty storage",
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			storage := NewStorage("foo", WithDefaultIndex(index))
			if tc.setup != nil {
				tc.setup(&storage)
			}

			var records []Record
			storage.Iterate("", "", func(r Record) bool {
				records = append(records, r)
				return false
			})

			urequire.Equal(t, len(tc.recordIDs), len(records), "results count")
			for i, r := range records {
				uassert.Equal(t, tc.recordIDs[i], r.ID())
			}
		})
	}
}

func TestStorageReverseIterate(t *testing.T) {
	index := NewIndex("name", func(r Record) string {
		if v, found := r.Get("name"); found {
			return v.(string)
		}
		return ""
	})

	cases := []struct {
		name      string
		recordIDs []uint64
		setup     func(*Storage)
	}{
		{
			name:      "single record",
			recordIDs: []uint64{1},
			setup: func(s *Storage) {
				r := s.NewRecord()
				r.Set("name", "foobar")
				r.Save()
			},
		},
		{
			name:      "two records",
			recordIDs: []uint64{2, 1},
			setup: func(s *Storage) {
				r := s.NewRecord()
				r.Set("name", "foobar")
				r.Save()

				r = s.NewRecord()
				r.Set("name", "foobar")
				r.Save()
			},
		},
		{
			name: "empty storage",
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			storage := NewStorage("foo", WithDefaultIndex(index))
			if tc.setup != nil {
				tc.setup(&storage)
			}

			var records []Record
			storage.ReverseIterate("", "", func(r Record) bool {
				records = append(records, r)
				return false
			})

			urequire.Equal(t, len(tc.recordIDs), len(records), "results count")
			for i, r := range records {
				uassert.Equal(t, tc.recordIDs[i], r.ID())
			}
		})
	}
}

func TestStorageGet(t *testing.T) {
	index := NewIndex("name", func(r Record) string {
		if v, found := r.Get("name"); found {
			return v.(string)
		}
		return ""
	})

	cases := []struct {
		name      string
		key       string
		recordIDs []uint64
		setup     func(*Storage)
	}{
		{
			name:      "single record",
			key:       "foobar",
			recordIDs: []uint64{1},
			setup: func(s *Storage) {
				r := s.NewRecord()
				r.Set("name", "foobar")
				r.Save()
			},
		},
		{
			name:      "two records",
			key:       "foobar",
			recordIDs: []uint64{1, 2},
			setup: func(s *Storage) {
				r := s.NewRecord()
				r.Set("name", "foobar")
				r.Save()

				r = s.NewRecord()
				r.Set("name", "foobar")
				r.Save()
			},
		},
		{
			name: "no records found",
			key:  "unknown",
			setup: func(s *Storage) {
				r := s.NewRecord()
				r.Set("name", "foobar")
				r.Save()
			},
		},
		{
			name: "empty storage",
			key:  "unknown",
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			storage := NewStorage("foo", WithDefaultIndex(index))
			if tc.setup != nil {
				tc.setup(&storage)
			}

			records := storage.Get(tc.key)

			urequire.Equal(t, len(tc.recordIDs), len(records), "results count")
			for i, r := range records {
				uassert.Equal(t, tc.recordIDs[i], r.ID())
			}
		})
	}
}

func TestStorageGetFirst(t *testing.T) {
	index := NewIndex("name", func(r Record) string {
		if v, found := r.Get("name"); found {
			return v.(string)
		}
		return ""
	})

	cases := []struct {
		name     string
		key      string
		recordID uint64
		setup    func(*Storage)
	}{
		{
			name:     "single record",
			key:      "foobar",
			recordID: 1,
			setup: func(s *Storage) {
				r := s.NewRecord()
				r.Set("name", "foobar")
				r.Save()
			},
		},
		{
			name:     "two records",
			key:      "foobar",
			recordID: 1,
			setup: func(s *Storage) {
				r := s.NewRecord()
				r.Set("name", "foobar")
				r.Save()

				r = s.NewRecord()
				r.Set("name", "foobar")
				r.Save()

				r = s.NewRecord()
				r.Set("name", "extra")
				r.Save()
			},
		},
		{
			name: "record not found",
			key:  "unknown",
			setup: func(s *Storage) {
				r := s.NewRecord()
				r.Set("name", "foobar")
				r.Save()
			},
		},
		{
			name: "empty storage",
			key:  "foobar",
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			storage := NewStorage("foo", WithDefaultIndex(index))
			if tc.setup != nil {
				tc.setup(&storage)
			}

			r := storage.GetFirst(tc.key)

			if tc.recordID == 0 {
				urequire.Equal(t, nil, r, "record not found")
				return
			}

			urequire.NotEqual(t, nil, r, "record found")
			uassert.Equal(t, tc.recordID, r.ID())
		})
	}
}

func TestStorageDelete(t *testing.T) {
	storage := NewStorage("foo")
	r := storage.NewRecord()
	r.Save()

	deleted := storage.Delete(r.ID())
	uassert.True(t, deleted)

	deleted = storage.Delete(r.ID())
	uassert.False(t, deleted)
}
