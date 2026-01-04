package datastore

import (
	"strings"
	"testing"

	"gno.land/p/nt/uassert"
	"gno.land/p/nt/urequire"
)

func TestStorageDefaults(t *testing.T) {
	name := "foo"
	storage := NewStorage(name)

	uassert.Equal(t, name, storage.Name())
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

func TestStorageQuery(t *testing.T) {
	index := NewIndex("tag", func(r Record) string {
		if v, found := r.Get("tag"); found {
			return v.(string)
		}
		return ""
	})

	cases := []struct {
		name    string
		options []QueryOption
		results []uint64
		setup   func() *Storage
		errMsg  string
	}{
		{
			name:    "default query",
			results: []uint64{1, 2},
			setup: func() *Storage {
				s := NewStorage("foo")
				s.NewRecord().Save()
				s.NewRecord().Save()
				return &s
			},
		},
		{
			name:    "with size",
			results: []uint64{1},
			options: []QueryOption{WithSize(1)},
			setup: func() *Storage {
				s := NewStorage("foo")
				s.NewRecord().Save()
				s.NewRecord().Save()
				return &s
			},
		},
		{
			name:    "with offset",
			results: []uint64{2},
			options: []QueryOption{WithOffset(1)},
			setup: func() *Storage {
				s := NewStorage("foo")
				s.NewRecord().Save()
				s.NewRecord().Save()
				return &s
			},
		},
		{
			name:    "with offset overflow",
			options: []QueryOption{WithOffset(4)},
			setup: func() *Storage {
				s := NewStorage("foo")
				s.NewRecord().Save()
				return &s
			},
		},
		{
			name:    "with size and offset",
			results: []uint64{2, 3},
			options: []QueryOption{WithSize(2), WithOffset(1)},
			setup: func() *Storage {
				s := NewStorage("foo")
				s.NewRecord().Save()
				s.NewRecord().Save()
				s.NewRecord().Save()
				s.NewRecord().Save()
				return &s
			},
		},
		{
			name:    "custom index",
			options: []QueryOption{UseIndex("tag", "A")},
			results: []uint64{1, 3},
			setup: func() *Storage {
				s := NewStorage("foo", WithIndex(index))

				r := s.NewRecord()
				r.Set("tag", "A")
				r.Save()

				r = s.NewRecord()
				r.Set("tag", "B")
				r.Save()

				r = s.NewRecord()
				r.Set("tag", "A")
				r.Save()

				return &s
			},
		},
		{
			name:    "custom index with offset",
			options: []QueryOption{UseIndex("tag", "B"), WithOffset(1)},
			results: []uint64{3, 4},
			setup: func() *Storage {
				s := NewStorage("foo", WithIndex(index))

				r := s.NewRecord()
				r.Set("tag", "B")
				r.Save()

				r = s.NewRecord()
				r.Set("tag", "A")
				r.Save()

				r = s.NewRecord()
				r.Set("tag", "B")
				r.Save()

				r = s.NewRecord()
				r.Set("tag", "B")
				r.Save()

				return &s
			},
		},
		{
			name:    "custom index with offset and size",
			options: []QueryOption{UseIndex("tag", "B"), WithOffset(1), WithSize(1)},
			results: []uint64{3},
			setup: func() *Storage {
				s := NewStorage("foo", WithIndex(index))

				r := s.NewRecord()
				r.Set("tag", "B")
				r.Save()

				r = s.NewRecord()
				r.Set("tag", "A")
				r.Save()

				r = s.NewRecord()
				r.Set("tag", "B")
				r.Save()

				r = s.NewRecord()
				r.Set("tag", "B")
				r.Save()

				return &s
			},
		},
		{
			name:    "custom index not found",
			options: []QueryOption{UseIndex("foo", "B")},
			setup: func() *Storage {
				s := NewStorage("foo")
				return &s
			},
			errMsg: "storage index for query not found: foo",
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			// Arrange
			storage := tc.setup()

			// Act
			rs, err := storage.Query(tc.options...)

			// Assert
			if tc.errMsg != "" {
				uassert.ErrorContains(t, err, tc.errMsg, "expect error")
				return
			}

			urequire.NoError(t, err, "expect no error")
			urequire.NotEqual(t, nil, rs, "new record is not nil")
			urequire.Equal(t, len(tc.results), rs.Size(), "expect query results count to match")

			var i int
			rs.Iterate(func(r Record) bool {
				urequire.Equal(t, tc.results[i], r.ID(), "expect result IDs to match")
				i++
				return false
			})
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
			storage := NewStorage("foo", WithIndex(index))
			if tc.setup != nil {
				tc.setup(&storage)
			}

			r, found := storage.Get(index.Name(), tc.key)

			if tc.recordID == 0 {
				uassert.Equal(t, nil, r, "expect no record")
				uassert.False(t, found, "expect record not found")
				return
			}

			uassert.True(t, found, "expect record found")
			urequire.NotEqual(t, nil, r, "expect record to be found")
			uassert.Equal(t, tc.recordID, r.ID(), "expect ID to match")
		})
	}
}

func TestStorageGetByID(t *testing.T) {
	cases := []struct {
		name     string
		recordID uint64
		found    bool
		setup    func(*Storage)
	}{
		{
			name:     "single record",
			recordID: 1,
			found:    true,
			setup: func(s *Storage) {
				s.NewRecord().Save()
			},
		},
		{
			name:     "multiple records",
			recordID: 2,
			found:    true,
			setup: func(s *Storage) {
				s.NewRecord().Save()
				s.NewRecord().Save()
				s.NewRecord().Save()
			},
		},
		{
			name:     "record not found",
			recordID: 3,
			setup: func(s *Storage) {
				s.NewRecord().Save()
				s.NewRecord().Save()
			},
		},
		{
			name:     "empty storage",
			recordID: 1,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			storage := NewStorage("foo")
			if tc.setup != nil {
				tc.setup(&storage)
			}

			r, found := storage.GetByID(tc.recordID)

			if !tc.found {
				uassert.Equal(t, nil, r, "expect no record")
				uassert.False(t, found, "expect record not found")
				return
			}

			uassert.True(t, found, "expect record found")
			urequire.NotEqual(t, nil, r, "expect record to be found")
			uassert.Equal(t, tc.recordID, r.ID(), "expect ID to match")
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
