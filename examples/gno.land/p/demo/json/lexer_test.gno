package json

import (
	"errors"
	"testing"
)


func TestBlockEnd(t *testing.T) {
	tests := []struct {
		name     string
		data     []byte
		open     byte
		close    byte
		expected int
	}{
		{"Empty Object", []byte("{}"), '{', '}', 2},
		{"Nested Object", []byte(`{"key": {"nestedKey": "value"}}`), '{', '}', 31},
		{"Array", []byte(`["item1", "item2"]`), '[', ']', 18},
		{"Complex Object", []byte(`{"key": [1, 2, 3], "anotherKey": "value"}`), '{', '}', 41},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			result := blockEnd(test.data, test.open, test.close)
			if result != test.expected {
				t.Errorf("Failed %s: expected %d, got %d", test.name, test.expected, result)
			}
		})
	}
}

func TestFindKeyStart(t *testing.T) {
	tests := []struct {
		name     string
		data     []byte
		key      string
		expected int
	}{
		{
			name:     "Test 1",
			data:     []byte("{\n\t\"key\": \"value\",\n\t\"array\": [1, 2, 3]\n}"),
			key:      "key",
			expected: 3,
		},
		{
			name:     "Test 2",
			data:     []byte("{\n\t\"key\": \"value\",\n\t\"array\": [1, 2, 3]\n}"),
			key:      "array",
			expected: 20,
		},
		{
			name:     "Test 3",
			data:     []byte("{\n\t\"key\": \"value\",\n\t\"array\": [1, 2, 3]\n}"),
			key:      "value",
			expected: -1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, _ := findKeyStart(tt.data, tt.key)
			if got != tt.expected {
				t.Errorf("findKeyStart() = %v, want %v", got, tt.expected)
			}
		})
	}
}

func TestSearchKeys(t *testing.T) {
	tests := []struct {
		name     string
		data     []byte
		keys     []string
		expected int
	}{
		{
			name:     "Test 1",
			data:     []byte(`{"key1": "value1", "key2": "value2", "key3": "value3"}`),
			keys:     []string{"key1"},
			expected: 8,
		},
		{
			name:     "Test 2",
			data:     []byte(`{"key1": "value1", "key2": "value2", "key3": "value3"}`),
			keys:     []string{"key2"},
			expected: 26,
		},
		{
			name:     "Test 3",
			data:     []byte(`{"key1": "value1", "key2": "value2", "key3": "value3"}`),
			keys:     []string{"key4"},
			expected: -1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got, _ := findValueIndex(tt.data, tt.keys...); got != tt.expected {
				t.Errorf("findValueIndex() = %v, want %v", got, tt.expected)
			}
		})
	}
}

func TestGet(t *testing.T) {
	data := []byte(`{
		"person": {
			"name": {
				"first": "Alex",
				"last": "Johnson",
				"fullName": "Alex Johnson"
			},
			"github": {
				"handle": "alexj",
				"followers": 152
			}
		},
		"company": {
			"name": "company name",
		},
	}
	`)

	tests := []struct {
		path     []string
		expected string
	}{
		{[]string{"person", "name", "first"}, "Alex"},
		{[]string{"person", "name", "last"}, "Johnson"},
		{[]string{"person", "name", "fullName"}, "Alex Johnson"},
		{[]string{"person", "github", "handle"}, "alexj"},
		{[]string{"person", "github", "followers"}, "152"},
		{[]string{"company", "name"}, "company name"},
	}

	for _, test := range tests {
		value, _, _, err := Get(data, test.path...)
		if err != nil {
			t.Errorf("Got error: %v", err)
		}

		if string(value) != test.expected {
			t.Errorf("Expected '%s', got '%s'", test.expected, value)
		}
	}
}

func TestExtractKeys(t *testing.T) {
    tests := []struct {
        name     string
        jsonData []byte
        expected []string
    }{
        {
            name:     "Simple Object",
            jsonData: []byte(`{"name": "John", "age": 30}`),
            expected: []string{"name", "age"},
        },
        {
            name:     "Nested Object",
            jsonData: []byte(`{"person": {"name": "John", "age": 30}, "city": "New York"}`),
            expected: []string{"person", "name", "age", "city"},
        },
        {
            name:     "Array of Objects",
            jsonData: []byte(`[{"name": "John"}, {"age": 30}]`),
            expected: []string{"name", "age"},
        },
        {
            name:     "Empty Object",
            jsonData: []byte(`{}`),
            expected: []string{},
        },
        // Additional test cases can be added here.
    }

    for _, test := range tests {
        t.Run(test.name, func(t *testing.T) {
            result, err := ExtractKeys(test.jsonData)
            if err != nil {
                t.Errorf("name: %s error = %v", test.name, err)
                return
            }

			if len(result) != len(test.expected) {
				t.Errorf("name: %s, expected %d keys, got %d. contents: %s", test.name, len(test.expected), len(result), result)
				return
			}
        })
    }
}