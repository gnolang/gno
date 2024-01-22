package json

import (
	"errors"
	"testing"
)

var dummyStruct = &Struct{
	fields: []Field{
		{name: "field1", value: "value1"},
		{name: "field2", value: "value2"},
		{name: "field3", value: "value3"},
	},
}

func TestStructSearch(t *testing.T) {
	s := dummyStruct
	tests := []struct {
		name         string
		searchName   string
		expectedName string
		expectedErr  error
	}{
		{
			name:         "Test 1",
			searchName:   "field1",
			expectedName: "field1",
			expectedErr:  nil,
		},
		{
			name:         "Test 2",
			searchName:   "field2",
			expectedName: "field2",
			expectedErr:  nil,
		},
		{
			name:         "Test 3",
			searchName:   "field4",
			expectedName: "",
			expectedErr:  errors.New("Field name not found"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			field, err := s.Search(tt.searchName)

			if err != nil && tt.expectedErr == nil {
				t.Errorf("Search() error = %v, expectedErr %v", err, tt.expectedErr)
				return
			}

			if field != nil && field.name != tt.expectedName {
				t.Errorf("Search() got = %v, want %v", field.name, tt.expectedName)
			}
		})
	}
}

func TestRemoveField(t *testing.T) {
	tests := []struct {
		name           string
		fields         []Field
		remove         string
		expectedFields []string
	}{
		{
			name: "remove existing field",
			fields: []Field{
				{name: "a", typ: String, value: "apple"},
				{name: "b", typ: String, value: "banana"},
			},
			remove:         "a",
			expectedFields: []string{"b"},
		},
		{
			name: "remove non-existing field",
			fields: []Field{
				{name: "a", typ: String, value: "apple"},
			},
			remove:         "b",
			expectedFields: []string{"a"},
		},
		{
			name:           "remove from empty struct",
			fields:         []Field{},
			remove:         "a",
			expectedFields: []string{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := NewStruct()
			for _, f := range tt.fields {
				s.addField(f.name, f.typ, f.value)
			}

			s.Remove(tt.remove)

			if !fieldsMatch(s.fields, tt.expectedFields) {
				t.Errorf("Remove() got = %v, want %v", fieldNames(s.fields), tt.expectedFields)
			}
		})
	}
}

func TestUpdateField(t *testing.T) {
	tests := []struct {
		name           string
		init           []Field
		fName          string
		newValue       interface{}
		expectedFields []Field
		expectError    bool
	}{
		{
			name: "update existing field",
			init: []Field{
				{name: "a", typ: String, value: "apple"},
				{name: "b", typ: Number, value: 10},
			},
			fName:    "a",
			newValue: "avocado",
			expectedFields: []Field{
				{name: "a", typ: String, value: "avocado"},
				{name: "b", typ: Number, value: 10},
			},
			expectError: false,
		},
		{
			name: "update non-existing field",
			init: []Field{
				{name: "a", typ: String, value: "apple"},
			},
			fName:    "b",
			newValue: "banana",
			expectedFields: []Field{
				{name: "a", typ: String, value: "apple"},
			},
			expectError: true,
		},
		{
			name: "type mismatch update",
			init: []Field{
				{name: "a", typ: String, value: "apple"},
			},
			fName:    "a",
			newValue: 100, // int instead of string
			expectedFields: []Field{
				{name: "a", typ: String, value: "apple"},
			},
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := NewStruct()
			for _, f := range tt.init {
				s.addField(f.name, f.typ, f.value)
			}

			err := s.Update(tt.fName, tt.newValue)
			if (err != nil) != tt.expectError {
				t.Errorf("Update() error = %v, expectErr %v", err, tt.expectError)
				return
			}

			// check if fields are equal
			for _, ef := range tt.expectedFields {
				f, err := s.Search(ef.name)
				if err != nil {
					t.Errorf("Update() error = %v, expectErr %v", err, tt.expectError)
					return
				}

				if f.value != ef.value {
					t.Errorf("Update() got = %v, want %v", f.value, ef.value)
				}

				if f.typ != ef.typ {
					t.Errorf("Update() got = %v, want %v", f.typ, ef.typ)
				}
			}
		})
	}
}
