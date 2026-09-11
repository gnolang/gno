package json

import (
	"bytes"
	"testing"
)

func TestIndentJSON(t *testing.T) {
	tests := []struct {
		name     string
		input    []byte
		indent   string
		expected []byte
	}{
		{
			name:     "empty object",
			input:    []byte(`{}`),
			indent:   "  ",
			expected: []byte(`{}`),
		},
		{
			name:     "empty array",
			input:    []byte(`[]`),
			indent:   "  ",
			expected: []byte(`[]`),
		},
		{
			name:     "nested object",
			input:    []byte(`{{}}`),
			indent:   "\t",
			expected: []byte("{\n\t\t{}\n}"),
		},
		{
			name:     "nested array",
			input:    []byte(`[[[]]]`),
			indent:   "\t",
			expected: []byte("[[\n\t\t[\n\t\t\t\t\n\t\t]\n]]"),
		},
		{
			name:     "top-level array",
			input:    []byte(`["apple","banana","cherry"]`),
			indent:   "\t",
			expected: []byte(`["apple","banana","cherry"]`),
		},
		{
			name:     "array of arrays",
			input:    []byte(`["apple",["banana","cherry"],"date"]`),
			indent:   "  ",
			expected: []byte("[\"apple\",[\n    \"banana\",\n    \"cherry\"\n],\"date\"]"),
		},

		{
			name:     "nested array in object",
			input:    []byte(`{"fruits":["apple",["banana","cherry"],"date"]}`),
			indent:   "  ",
			expected: []byte("{\n    \"fruits\": [\"apple\",[\n        \"banana\",\n        \"cherry\"\n    ],\"date\"]\n}"),
		},
		{
			name:     "complex nested structure",
			input:    []byte(`{"data":{"array":[1,2,3],"bool":true,"nestedArray":[["a","b"],"c"]}}`),
			indent:   "  ",
			expected: []byte("{\n    \"data\": {\n        \"array\": [1,2,3],\"bool\": true,\"nestedArray\": [[\n            \"a\",\n            \"b\"\n        ],\"c\"]\n    }\n}"),
		},
		{
			name:     "custom ident character",
			input:    []byte(`{"fruits":["apple",["banana","cherry"],"date"]}`),
			indent:   "*",
			expected: []byte("{\n**\"fruits\": [\"apple\",[\n****\"banana\",\n****\"cherry\"\n**],\"date\"]\n}"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			actual, err := Indent(tt.input, tt.indent)
			if err != nil {
				t.Errorf("IndentJSON() error = %v", err)
				return
			}
			if !bytes.Equal(actual, tt.expected) {
				t.Errorf("IndentJSON() = %q, want %q", actual, tt.expected)
			}
		})
	}
}
