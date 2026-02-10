package datastore

import (
	"testing"

	"gno.land/p/nt/uassert"
	"gno.land/p/nt/urequire"
)

func TestSchemaNew(t *testing.T) {
	cases := []struct {
		name    string
		options []SchemaOption
		fields  []string
		strict  bool
	}{
		{
			name: "default",
		},
		{
			name:    "strict",
			options: []SchemaOption{Strict()},
			strict:  true,
		},
		{
			name: "with fields",
			options: []SchemaOption{
				WithField("foo"),
				WithField("bar"),
				WithDefaultField("baz", 42),
			},
			fields: []string{"foo", "bar", "baz"},
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			s := NewSchema("Foo", tc.options...)

			uassert.Equal(t, "Foo", s.Name())
			uassert.Equal(t, tc.strict, s.IsStrict())
			urequire.Equal(t, len(tc.fields), s.Size(), "field count")

			for i, name := range s.Fields() {
				uassert.Equal(t, tc.fields[i], name)
				uassert.True(t, s.HasField(name))
			}
		})
	}
}

func TestSchemaAddField(t *testing.T) {
	cases := []struct {
		name       string
		options    []SchemaOption
		fieldName  string
		fieldIndex int
		fields     []string
		success    bool
	}{
		{
			name:       "new only field",
			fieldName:  "foo",
			fieldIndex: 0,
			fields:     []string{"foo"},
			success:    true,
		},
		{
			name: "new existing fields",
			options: []SchemaOption{
				WithField("foo"),
				WithField("bar"),
			},
			fieldName:  "baz",
			fieldIndex: 2,
			fields:     []string{"foo", "bar", "baz"},
			success:    true,
		},
		{
			name:       "duplicated field",
			options:    []SchemaOption{WithField("foo")},
			fieldName:  "foo",
			fieldIndex: -1,
			fields:     []string{"foo"},
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			s := NewSchema("Foo", tc.options...)

			index, added := s.AddField(tc.fieldName, nil)

			if tc.success {
				uassert.Equal(t, tc.fieldIndex, index)
				uassert.True(t, added)
			} else {
				uassert.Equal(t, -1, index)
				uassert.False(t, added)
			}

			urequire.Equal(t, len(tc.fields), s.Size(), "field count")

			for i, name := range s.Fields() {
				uassert.Equal(t, tc.fields[i], name)
				uassert.True(t, s.HasField(name))
			}
		})
	}
}

func TestSchemaGetFieldIndex(t *testing.T) {
	s := NewSchema("Foo")
	s.AddField("foo", nil)
	s.AddField("bar", nil)
	s.AddField("baz", nil)

	uassert.Equal(t, 0, s.GetFieldIndex("foo"))
	uassert.Equal(t, 1, s.GetFieldIndex("bar"))
	uassert.Equal(t, 2, s.GetFieldIndex("baz"))

	uassert.Equal(t, -1, s.GetFieldIndex(""))
	uassert.Equal(t, -1, s.GetFieldIndex("unknown"))
}

func TestSchemaGetFieldName(t *testing.T) {
	s := NewSchema("Foo")
	s.AddField("foo", nil)
	s.AddField("bar", nil)
	s.AddField("baz", nil)

	name, found := s.GetFieldName(0)
	uassert.Equal(t, "foo", name)
	uassert.True(t, found)

	name, found = s.GetFieldName(1)
	uassert.Equal(t, "bar", name)
	uassert.True(t, found)

	name, found = s.GetFieldName(2)
	uassert.Equal(t, "baz", name)
	uassert.True(t, found)

	name, found = s.GetFieldName(404)
	uassert.Equal(t, "", name)
	uassert.False(t, found)
}

func TestSchemaGetDefault(t *testing.T) {
	s := NewSchema("Foo")
	s.AddField("foo", nil)
	s.AddField("bar", 42)

	_, found := s.GetDefault("foo")
	uassert.False(t, found)

	v, found := s.GetDefault("bar")
	uassert.True(t, found)

	got, ok := v.(int)
	urequire.True(t, ok, "default field value")
	uassert.Equal(t, 42, got)
}

func TestSchemaGetDefaultByIndex(t *testing.T) {
	s := NewSchema("Foo")
	s.AddField("foo", nil)
	s.AddField("bar", 42)

	_, found := s.GetDefaultByIndex(0)
	uassert.False(t, found)

	_, found = s.GetDefaultByIndex(404)
	uassert.False(t, found)

	v, found := s.GetDefaultByIndex(1)
	uassert.True(t, found)

	got, ok := v.(int)
	urequire.True(t, ok, "default field value")
	uassert.Equal(t, 42, got)
}

func TestSchemaRenameField(t *testing.T) {
	s := NewSchema("Foo")
	s.AddField("foo", nil)
	s.AddField("bar", nil)

	renamed := s.RenameField("foo", "bar")
	uassert.False(t, renamed)

	renamed = s.RenameField("", "baz")
	uassert.False(t, renamed)

	renamed = s.RenameField("foo", "")
	uassert.True(t, renamed)

	renamed = s.RenameField("", "foo")
	uassert.True(t, renamed)

	renamed = s.RenameField("foo", "foobar")
	uassert.True(t, renamed)

	urequire.Equal(t, 2, s.Size(), "field count")
	fields := []string{"foobar", "bar"}
	for i, name := range s.Fields() {
		uassert.Equal(t, fields[i], name)
		uassert.True(t, s.HasField(name))
	}
}
