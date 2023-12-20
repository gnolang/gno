package json

import (
	"testing"
)

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
func TestMapKeyTypes(t *testing.T) {
	tests := []struct {
		name         string
		data         []byte
		expectedMap  map[string]ValueType
		expectedErr  error
	}{
		{
			name: "Test 1",
			data: []byte(`{"key1": "value1", "key2": 123, "key3": true}`),
			expectedMap: map[string]ValueType{
				"key1": String,
				"key2": Number,
				"key3": Boolean,
			},
			expectedErr: nil,
		},
		// {
		// 	name: "Test 2",
		// 	data: []byte(`{"key1": "value1", "key2": [1, 2, 3], "key3": {"nested": "value"}}`),
		// 	expectedMap: map[string]ValueType{
		// 		"key1": String,
		// 		"key2": Array,
		// 		"key3": Object,
		// 	},
		// 	expectedErr: nil,
		// },
		{
			name: "Test 3",
			data: []byte(`{"key1": "value1", "key2": null, "key3": 3.14}`),
			expectedMap: map[string]ValueType{
				"key1": String,
				"key2": Null,
				"key3": Number,
			},
			expectedErr: nil,
		},
		// {
		// 	name: "Test 4",
		// 	data: []byte(`{"key1": "value1", "key2": {"nested": [1, 2, 3]}, "key3": true}`),
		// 	expectedMap: map[string]ValueType{
		// 		"key1": String,
		// 		"key2": Object,
		// 		"key3": Boolean,
		// 	},
		// 	expectedErr: nil,
		// },
		{
			name: "Test 5",
			data: []byte(`{"key1": "value1", "key2": "value2", "key3": "value3"}`),
			expectedMap: map[string]ValueType{
				"key1": String,
				"key2": String,
				"key3": String,
			},
			expectedErr: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotMap, gotErr := MapKeyTypes(tt.data)

			if gotMap == nil {
				t.Errorf("MapKeyTypes() gotMap = %v, expectedMap = %v", gotMap, tt.expectedMap)
			}

			if gotErr != tt.expectedErr {
				t.Errorf("MapKeyTypes() gotErr = %v, expectedErr = %v", gotErr, tt.expectedErr)
			}

			for key, value := range gotMap {
				if tt.expectedMap[key] != value {
					t.Errorf("MapKeyTypes() gotMap = %v, expectedMap = %v", gotMap, tt.expectedMap)
				}
			}
		})
	}
}