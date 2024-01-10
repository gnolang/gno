package json

import (
	"errors"
	"fmt"
	"testing"
)

func TestParseStruct(t *testing.T) {
	data := `
        type Person struct {
            Name string ` + "`json:\"name\"`" + `
            Age int ` + "`json:\"age\"`" + `
        }
    `

	expectedFields := Fields{
		{Name: "Name", Type: "string", Tag: "name", Depth: 0},
		{Name: "Age", Type: "int", Tag: "age", Depth: 0},
	}

	fields := ParseStruct(data, 0)

	if len(fields) != len(expectedFields) {
		t.Errorf("Expected %d fields, but got %d", len(expectedFields), len(fields))
	}

	for i, field := range fields {
		expectedField := expectedFields[i]

		if field.Name != expectedField.Name {
			t.Errorf("Expected field name %s, but got %s", expectedField.Name, field.Name)
		}

		if field.Type != expectedField.Type {
			t.Errorf("Expected field type %s, but got %s", expectedField.Type, field.Type)
		}

		if field.Tag != expectedField.Tag {
			t.Errorf("Expected field tag %s, but got %s", expectedField.Tag, field.Tag)
		}

		if field.Depth != expectedField.Depth {
			t.Errorf("Expected field depth %d, but got %d", expectedField.Depth, field.Depth)
		}

		if field.IsNested != expectedField.IsNested {
			t.Errorf("Expected field IsNested %t, but got %t", expectedField.IsNested, field.IsNested)
		}
	}
}

func TestParseNestedStruct(t *testing.T) {
	data := `type Server struct {
        Name string ` + "`json:\"name\"`" + `
        Config struct {
            Enabled bool ` + "`json:\"enabled,omitempty\"`" + `
            Enabled2 bool ` + "`json:\"enabled2\"`" + `
        } ` + "`json:\"config\"`" + `
        Age int ` + "`json:\"age\"`" + `
    }`

	expected := []FieldMetaData{
		{Name: "Server", Type: "struct", Tag: "config", IsNested: true, Depth: 0},
		{Name: "Name", Type: "string", Tag: "name", IsNested: false, Depth: 1},
		{Name: "Enabled", Type: "bool", Tag: "enabled,omitempty", IsNested: false, Depth: 1},
		{Name: "Enabled2", Type: "bool", Tag: "enabled2", IsNested: false, Depth: 1},
		{Name: "Age", Type: "int", Tag: "age", IsNested: false, Depth: 0},
	}

	actual := ParseStruct(data, 0)

	if len(actual) != len(expected) {
		t.Errorf("Expected and actual slice lengths differ. Expected %d, got %d", len(expected), len(actual))
		return
	}

	for i, actualField := range actual {
		expectedField := expected[i]
		if actualField != expectedField {
			t.Errorf("Field %d - Expected %+v, got %+v", i, expectedField, actualField)
		}
	}
}

func TestExtractNames(t *testing.T) {
	tests := []struct {
		fields   Fields
		expected []string
	}{
		{
			fields: Fields{
				{Name: "Name", Type: "string", Tag: "name", Depth: 0},
				{Name: "Age", Type: "int", Tag: "age", Depth: 0},
			},
			expected: []string{"Name", "Age"},
		},
		{
			fields: Fields{
				{Name: "Server", Type: "struct", Tag: "config", IsNested: true, Depth: 0},
				{Name: "Name", Type: "string", Tag: "name", IsNested: false, Depth: 1},
				{Name: "Enabled", Type: "bool", Tag: "enabled,omitempty", IsNested: false, Depth: 1},
				{Name: "Enabled2", Type: "bool", Tag: "enabled2", IsNested: false, Depth: 1},
				{Name: "Age", Type: "int", Tag: "age", IsNested: false, Depth: 0},
			},
			expected: []string{"Server", "Name", "Enabled", "Enabled2", "Age"},
		},
	}

	for i, test := range tests {
		actual := test.fields.Names()
		if len(actual) != len(test.expected) {
			t.Errorf("Test %d - Expected and actual slice lengths differ. Expected %d, got %d", i, len(test.expected), len(actual))
			return
		}

		for j, actualName := range actual {
			expectedName := test.expected[j]
			if actualName != expectedName {
				t.Errorf("Test %d - Expected %s, got %s", i, expectedName, actualName)
			}
		}
	}
}

func TestZip(t *testing.T) {
	cases := []struct {
		name      string
		json      []byte
		keys      []string
		expected  map[string]interface{}
		expectErr bool
	}{
		{
			"Valid JSON with String, Number and Boolean",
			[]byte(`{"name":"test","age":30,"active":true, "memo": null}`),
			[]string{"name", "age", "active", "memo"},
			map[string]interface{}{"name": "test", "age": 30.0, "active": true, "memo": nil},
			false,
		},
		{
			"whitespace containing JSON",
			[]byte(`{"name":"value with whitespcae","age":30,"active":true, "memo": null}`),
			[]string{"name", "age", "active", "memo"},
			map[string]interface{}{"name": "value with whitespcae", "age": 30.0, "active": true, "memo": nil},
			false,
		},
		{
			"unknown type throw error",
			[]byte(`{"name":"test","age":30,"active":true, "memo": None}`),
			[]string{"name", "age", "active", "memo", "unknown"},
			map[string]interface{}{"name": "test", "age": 30.0, "active": true, "memo": nil},
			true,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			result, err := Zip(tc.json, tc.keys...)
			if (err != nil) != tc.expectErr {
				t.Errorf("Zip(%s) error = %v, expectErr %v", tc.name, err, tc.expectErr)
				return
			}

			if tc.expectErr {
				return
			}

			for key, value := range result {
				expectedValue, ok := tc.expected[key]
				if !ok {
					t.Errorf("Unexpected zipped field: %s", key)
				}

				if value != expectedValue {
					t.Errorf("Expected value %v for field %s, but got %v", expectedValue, key, value)
				}
			}
		})
	}
}

func TestMarshal(t *testing.T) {
	tests := []struct {
		name    string
		data    map[string]interface{}
		want    string
		wantErr bool
	}{
		{
			name: "Test with various data types",
			data: map[string]interface{}{
				"name":  "John Doe",
				"age":   30,
				"admin": true,
				"score": 99.5,
				"attributes": map[string]interface{}{
					"intelligence": 100,
					"strength":     80,
				},
				"friends": []interface{}{
					"Jane Doe",
					"Richard Roe",
				},
			},
			want:    `{"name": "John Doe", "age": 30, "admin": true, "score": 99.500000, "attributes": {"intelligence": 100, "strength": 80}, "friends": ["Jane Doe", "Richard Roe"]}`,
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := Marshal(tt.data)
			if (err != nil) != tt.wantErr {
				t.Errorf("Marshal() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("Marshal() = %v, want %v", got, tt.want)
			}
		})
	}
}
