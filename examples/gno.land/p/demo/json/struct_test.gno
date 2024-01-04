package json

import (
	"fmt"
	"testing"
)

func TestParseStruct(t *testing.T) {
	data := []byte(`
		type Person struct {
			Name string ` + "`" + `json:"name"` + "`" + `
			Age int ` + "`" + `json:"age"` + "`" + `
		}
	`)

	expectedFields := []FieldMetaData{
		{
			Name:       "Name",
			Type:       "string",
			Tag:        "name",
			Depth:      0,
			IsNested:   false,
			IsOptional: false,
		},
		{
			Name:       "Age",
			Type:       "int",
			Tag:        "age",
			Depth:      0,
			IsNested:   false,
			IsOptional: false,
		},
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

		if field.IsOptional != expectedField.IsOptional {
			t.Errorf("Expected field IsOptional %t, but got %t", expectedField.IsOptional, field.IsOptional)
		}
	}
}

func TestParseNestedStruct(t *testing.T) {
	data := []byte(`type Server struct {
		Name string ` + "`json:\"name\"`" + `
		Config struct {
			Enabled bool ` + "`json:\"enabled,omitempty\"`" + `
			Enabled2 bool ` + "`json:\"enabled2\"`" + `
		} ` + "`json:\"config\"`" + `
		Age int ` + "`json:\"age\"`" + `
	}`)

	expected := []FieldMetaData{
		{Name: "Server", Type: "struct", Tag: "config", IsNested: true, IsOptional: false, Depth: 0},
		{Name: "Name", Type: "string", Tag: "name", IsNested: false, IsOptional: false, Depth: 1},
		{Name: "Enabled", Type: "bool", Tag: "enabled,omitempty", IsNested: false, IsOptional: true, Depth: 1},
		{Name: "Enabled2", Type: "bool", Tag: "enabled2", IsNested: false, IsOptional: false, Depth: 1},
		{Name: "Age", Type: "int", Tag: "age", IsNested: false, IsOptional: false, Depth: 0},
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
