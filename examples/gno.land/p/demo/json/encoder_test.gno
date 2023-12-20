package json

import (
	"errors"
	"strings"
	"testing"
)

func equal(a, b map[string]interface{}) bool {
	if len(a) != len(b) {
		return false
	}

	for k, v := range a {
		if b[k] != v {
			return false
		}
	}

	return true
}

func mapsAreEqual(a, b map[string]interface{}) bool {
	if len(a) != len(b) {
		return false
	}

	for ka, va := range a {
		vb, ok := b[ka]
		if !ok {
			return false
		}

		switch vaTyped := va.(type) {
		case map[string]interface{}:
			if vbTyped, ok := vb.(map[string]interface{}); ok {
				if !mapsAreEqual(vaTyped, vbTyped) {
					return false
				}
			} else {
				return false
			}
		default:
			if va != vb {
				return false
			}
		}
	}

	return true
}

func TestFlatten(t *testing.T) {
	testCases := []struct {
		name     string
		input    map[string]interface{}
		expected map[string]string
	}{
		{
			name: "Basic Test",
			input: map[string]interface{}{
				"user": map[string]interface{}{
					"name": "John Doe",
					"age":  30,
					"address": map[string]interface{}{
						"street": "123 Main St",
						"city":   "Anytown",
					},
					"favoriteNumbers": []interface{}{7, 13, 42},
				},
			},
			expected: map[string]string{
				"user.name":               "string:\"John Doe\"",
				"user.age":                "number:30",
				"user.address.street":     "string:\"123 Main St\"",
				"user.address.city":       "string:\"Anytown\"",
				"user.favoriteNumbers[0]": "number:7",
				"user.favoriteNumbers[1]": "number:13",
				"user.favoriteNumbers[2]": "number:42",
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result := Flatten(tc.input)
			if !mapsAreEqual(result, tc.expected) {
				t.Errorf("Expected %v, got %v", tc.expected, result)
			}
		})
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
			name: "Nested Object 2",
			jsonData: []byte(`{
				"a": {
					"b": {
						"c": 1,
						"d": 2
					}
				},
				"e": {
					"f": 3
				}
			}`),
			expected: []string{"a", "b", "c", "d", "e", "f"},
		},
		{
			name:     "Empty Object",
			jsonData: []byte(`{}`),
			expected: []string{},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			result, err := extractKeys(test.jsonData)
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

// func TestGetTypeFromKey(t *testing.T) {
// 	testCases := []struct {
// 		name     string
// 		data     []byte
// 		key      string
// 		expected ValueType
// 		err      error
// 	}{
// 		{
// 			name:     "Key Exists",
// 			data:     []byte(`{"name": "John", "age": 30}`),
// 			key:      "name",
// 			expected: String,
// 			err:      nil,
// 		},
// 		{
// 			name:     "Key Does Not Exist",
// 			data:     []byte(`{"name": "John", "age": 30}`),
// 			key:      "address",
// 			expected: NotExist,
// 			err:      errors.New("key not found"),
// 		},
// 		{
// 			name:     "Empty Data",
// 			data:     []byte(`{}`),
// 			key:      "name",
// 			expected: NotExist,
// 			err:      errors.New("key not found"),
// 		},
// 	}

// 	for _, tc := range testCases {
// 		t.Run(tc.name, func(t *testing.T) {
// 			result, err := getTypeFromKey(tc.data, tc.key)
// 			if result != tc.expected {
// 				t.Errorf("Expected %v, got %v", tc.expected, result)
// 			}
// 			if (err == nil && tc.err != nil) || (err != nil && tc.err == nil) || (err != nil && tc.err != nil && err.Error() != tc.err.Error()) {
// 				t.Errorf("Expected error: %v, got error: %v", tc.err, err)
// 			}
// 		})
// 	}
// }
