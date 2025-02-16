package datastore

import (
	"testing"

	"gno.land/p/demo/uassert"
	"gno.land/p/demo/urequire"
)

var (
	_ Record    = (*record)(nil)
	_ Recordset = (*recordset)(nil)
)

func TestRecordDefaults(t *testing.T) {
	storage := NewStorage("foo")
	r := storage.NewRecord()

	uassert.Equal(t, uint64(0), r.ID())
	uassert.Equal(t, "0000000", r.Key())
	uassert.Equal(t, "Foo", r.Type())
	uassert.Equal(t, nil, r.Fields())
	uassert.True(t, r.IsEmpty())
}

func TestRecordHasField(t *testing.T) {
	storage := NewStorage("foo")
	r := storage.NewRecord()

	s := storage.Schema()
	s.AddField("foo", nil)

	uassert.True(t, r.HasField("foo"))
	uassert.False(t, r.HasField("undefined"))
}

func TestRecordSet(t *testing.T) {
	cases := []struct {
		name        string
		options     []SchemaOption
		field       string
		fieldsCount int
		value       int
		err         error
	}{
		{
			name:        "first new field",
			field:       "test",
			fieldsCount: 1,
			value:       42,
		},
		{
			name: "new extra field",
			options: []SchemaOption{
				WithField("foo"),
				WithField("bar"),
			},
			field:       "test",
			fieldsCount: 3,
			value:       42,
		},
		{
			name: "existing field",
			options: []SchemaOption{
				WithField("test"),
			},
			field:       "test",
			fieldsCount: 1,
			value:       42,
		},
		{
			name:        "undefined field",
			options:     []SchemaOption{Strict()},
			field:       "test",
			fieldsCount: 1,
			value:       42,
			err:         ErrUndefinedField,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			s := NewSchema("Foo", tc.options...)
			storage := NewStorage("foo", WithSchema(s))
			r := storage.NewRecord()

			err := r.Set(tc.field, tc.value)

			if tc.err != nil {
				urequire.ErrorIs(t, err, tc.err)
				return
			}

			urequire.NoError(t, err)
			uassert.True(t, r.HasField("test"))
			uassert.False(t, r.IsEmpty())
			uassert.Equal(t, tc.fieldsCount, len(r.Fields()))
		})
	}
}

func TestRecordGet(t *testing.T) {
	storage := NewStorage("foo")
	r := storage.NewRecord()
	r.Set("foo", "bar")
	r.Set("test", 42)

	v, found := r.Get("test")
	urequire.True(t, found, "get setted value")

	got, ok := v.(int)
	urequire.True(t, ok, "setted value type")
	uassert.Equal(t, 42, got)

	_, found = r.Get("unknown")
	uassert.False(t, found)
}

func TestRecordSave(t *testing.T) {
	index := NewIndex("name", func(r Record) string {
		return r.MustGet("name").(string)
	}).Unique().CaseInsensitive()

	storage := NewStorage("foo", WithIndex(index))
	cases := []struct {
		name            string
		id              uint64
		fieldValue, key string
		storageSize     int
		setup           func(Storage) Record
	}{
		{
			name:        "create first record",
			id:          1,
			key:         "0000001",
			fieldValue:  "foo",
			storageSize: 1,
			setup:       func(s Storage) Record { return s.NewRecord() },
		},
		{
			name:        "create second record",
			id:          2,
			key:         "0000002",
			fieldValue:  "bar",
			storageSize: 2,
			setup:       func(s Storage) Record { return s.NewRecord() },
		},
		{
			name:        "update second record",
			id:          2,
			key:         "0000002",
			fieldValue:  "baz",
			storageSize: 2,
			setup: func(s Storage) Record {
				r, _ := storage.GetRecord(index.Name(), "bar")
				return r
			},
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			r := tc.setup(storage)
			r.Set("name", tc.fieldValue)

			_, found := storage.GetRecord(index.Name(), tc.fieldValue)
			urequire.False(t, found, "record not found")
			urequire.True(t, r.Save(), "save success")
			uassert.Equal(t, tc.storageSize, storage.Size())

			r, found = storage.GetRecord(index.Name(), tc.fieldValue)
			urequire.True(t, found, "record found")
			uassert.Equal(t, tc.id, r.ID())
			uassert.Equal(t, tc.key, r.Key())
			uassert.Equal(t, tc.fieldValue, r.MustGet("name"))
		})
	}
}
