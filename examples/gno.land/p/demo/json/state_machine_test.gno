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

// func TestGet(t *testing.T) {
// 	data := []byte(`{
// 		"person": {
// 			"name": {
// 				"first": "Alex",
// 				"last": "Johnson",
// 				"fullName": "Alex Johnson"
// 			},
// 			"github": {
// 				"handle": "alexj",
// 				"followers": 152
// 			}
// 		},
// 		"company": {
// 			"name": "company name",
// 		},
// 	}
// 	`)

// 	tests := []struct {
// 		path     []string
// 		expected string
// 	}{
// 		{[]string{"person", "name", "first"}, "Alex"},
// 		{[]string{"person", "name", "last"}, "Johnson"},
// 		{[]string{"person", "name", "fullName"}, "Alex Johnson"},
// 		{[]string{"person", "github", "handle"}, "alexj"},
// 		{[]string{"person", "github", "followers"}, "152"},
// 		{[]string{"company", "name"}, "company name"},
// 	}

// 	for _, test := range tests {
// 		value, _, _, err := Get(data, test.path...)
// 		if err != nil {
// 			t.Errorf("Got error: %v", err)
// 		}

// 		if string(value) != test.expected {
// 			t.Errorf("Expected '%s', got '%s'", test.expected, value)
// 		}
// 	}
// }

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

type GetTest struct {
	desc string
	json string
	path []string

	isErr   bool
	isFound bool

	data interface{}
}

var getTests = []GetTest{
	// Trivial tests
	{
		desc:    "read string",
		json:    `""`,
		isFound: true,
		data:    ``,
	},
	{
		desc:    "read number",
		json:    `0`,
		isFound: true,
		data:    `0`,
	},
	{
		desc:    "read object",
		json:    `{}`,
		isFound: true,
		data:    `{}`,
	},
	{
		desc:    "read array",
		json:    `[]`,
		isFound: true,
		data:    `[]`,
	},
	{
		desc:    "read boolean",
		json:    `true`,
		isFound: true,
		data:    `true`,
	},
	{
		desc: "read boolean 2",
		json: `false`,
		isFound: true,
		data: `false`,
	},
	{
		desc:   "read null",
		json:   `null`,
		isFound: true,
		data:   `null`,
	},

	// Found key tests
	{
		desc:    "handling multiple nested keys with same name",
		json:    `{"a":[{"b":1},{"b":2},3],"c":{"d":[1,2]}} }`,
		path:    []string{"c", "d"},
		isFound: true,
		data:    `[1,2]`,
	},
	{
		desc:    "read basic key",
		json:    `{"a":"b"}`,
		path:    []string{"a"},
		isFound: true,
		data:    `b`,
	},
	{
		desc:    "read basic key with space",
		json:    `{"a": "b"}`,
		path:    []string{"a"},
		isFound: true,
		data:    `b`,
	},
	{
		desc:    "read composite key",
		json:    `{"a": { "b":{"c":"d" }}}`,
		path:    []string{"a", "b", "c"},
		isFound: true,
		data:    `d`,
	},
	{
		desc:    `read numberic value as string`,
		json:    `{"a": "b", "c": 1}`,
		path:    []string{"c"},
		isFound: true,
		data:    `1`,
	},
	{
		desc:    `handle multiple nested keys with same name`,
		json:    `{"a":[{"b":1},{"b":2},3],"c":{"c":[1,2]}} }`,
		path:    []string{"c", "c"},
		isFound: true,
		data:    `[1,2]`,
	},
	// {
	// 	desc:    `read string values with quotes`,
	// 	json:    `{"a": "string\"with\"quotes"}`,
	// 	path:    []string{"a"},
	// 	isFound: true,
	// 	data:    `string\"with\"quotes`,
	// },
	{
		desc:    `read object`,
		json:    `{"a": { "b":{"c":"d" }}}`,
		path:    []string{"a", "b"},
		isFound: true,
		data:    `{"c":"d" }`,
	},
	{
		desc:    `empty path`,
		json:    `{"c":"d" }`,
		path:    []string{},
		isFound: true,
		data:    `{"c":"d" }`,
	},
	{
		desc:    `formatted JSON value`,
		json:    "{\n  \"a\": \"b\"\n}",
		path:    []string{"a"},
		isFound: true,
		data:    `b`,
	},
	{
		desc:    `formatted JSON value 2`,
		json:    "{\n  \"a\":\n    {\n\"b\":\n   {\"c\":\"d\",\n\"e\": \"f\"}\n}\n}",
		path:    []string{"a", "b"},
		isFound: true,
		data:    "{\"c\":\"d\",\n\"e\": \"f\"}",
	},
	{
		desc:    `whitespace`,
		json:    " \n\r\t{ \n\r\t\"whitespace\" \n\r\t: \n\r\t333 \n\r\t} \n\r\t",
		path:    []string{"whitespace"},
		isFound: true,
		data:    "333",
	},
	// {
	// 	desc:    `escaped backslash quote`,
	// 	json:    `{"a": "\\\""}`,
	// 	path:    []string{"a"},
	// 	isFound: true,
	// 	data:    `\\\"`,
	// },
	{
		desc:    `unescaped backslash quote`,
		json:    `{"a": "\\"}`,
		path:    []string{"a"},
		isFound: true,
		data:    `\\`,
	},
	{
		desc:    `unicode in JSON`,
		json:    `{"a": "15°C"}`,
		path:    []string{"a"},
		isFound: true,
		data:    `15°C`,
	},
	{
		desc:    `no padding + nested`,
		json:    `{"a":{"a":"1"},"b":2}`,
		path:    []string{"b"},
		isFound: true,
		data:    `2`,
	},
	{
		desc:    `no padding + nested + array`,
		json:    `{"a":{"b":[1,2]},"c":3}`,
		path:    []string{"c"},
		isFound: true,
		data:    `3`,
	},
	{
		desc:    `empty key`,
		json:    `{"":{"":{"":true}}}`,
		path:    []string{"", "", ""},
		isFound: true,
		data:    `true`,
	},

	// Escaped key tests
	{
		desc:    `key with simple escape`,
		json:    `{"a\\b":1}`,
		path:    []string{"a\\b"},
		isFound: true,
		data:    `1`,
	},
	{
		desc:    `key and value with whitespace escapes`,
		json:    `{"key\b\f\n\r\tkey":"value\b\f\n\r\tvalue"}`,
		path:    []string{"key\b\f\n\r\tkey"},
		isFound: true,
		data:    `value\b\f\n\r\tvalue`, // value is not unescaped since this is Get(), but the key should work correctly
	},
	{
		desc:    `key with Unicode escape`,
		json:    `{"a\u00B0b":1}`,
		path:    []string{"a\u00B0b"},
		isFound: true,
		data:    `1`,
	},
	{
		desc:    `key with complex escape`,
		json:    `{"a\uD83D\uDE03b":1}`,
		path:    []string{"a\U0001F603b"},
		isFound: true,
		data:    `1`,
	},

	{ // This test returns a match instead of a parse error, as checking for the malformed JSON would reduce performance
		desc:    `malformed with trailing whitespace`,
		json:    `{"a":1 `,
		path:    []string{"a"},
		isFound: true,
		data:    `1`,
	},
	{ // This test returns a match instead of a parse error, as checking for the malformed JSON would reduce performance
		desc:    `malformed with wrong closing bracket`,
		json:    `{"a":1]`,
		path:    []string{"a"},
		isFound: true,
		data:    `1`,
	},

	// Not found key tests     ** current: throw reflect error. should be fixed
	// {
	// 	desc:    `empty input`,
	// 	json:    ``,
	// 	path:    []string{"a"},
	// 	isFound: false,
	// },
	// {
	// 	desc:    "non-existent key 1",
	// 	json:    `{"a":"b"}`,
	// 	path:    []string{"c"},
	// 	isFound: false,
	// },
	// {
	// 	desc:    "non-existent key 2",
	// 	json:    `{"a":"b"}`,
	// 	path:    []string{"b"},
	// 	isFound: false,
	// },
	// {
	// 	desc:    "non-existent key 3",
	// 	json:    `{"aa":"b"}`,
	// 	path:    []string{"a"},
	// 	isFound: false,
	// },
	// {
	// 	desc:    "apply scope of parent when search for nested key",
	// 	json:    `{"a": { "b": 1}, "c": 2 }`,
	// 	path:    []string{"a", "b", "c"},
	// 	isFound: false,
	// },
	// {
	// 	desc:    `apply scope to key level`,
	// 	json:    `{"a": { "b": 1}, "c": 2 }`,
	// 	path:    []string{"b"},
	// 	isFound: false,
	// },
	// {
	// 	desc:    `handle escaped quote in key name in JSON`,
	// 	json:    `{"key\"key": 1}`,
	// 	path:    []string{"key"},
	// 	isFound: false,
	// },
	// {
	// 	desc:    "handling multiple keys with different name",
	// 	json:    `{"a":{"a":1},"b":{"a":3,"c":[1,2]}}`,
	// 	path:    []string{"a", "c"},
	// 	isFound: false,
	// },
	// {
	// 	desc:    "handling nested json",
	// 	json:    `{"a":{"b":{"c":1},"d":4}}`,
	// 	path:    []string{"a", "d"},
	// 	isFound: true,
	// 	data:    `4`,
	// },
	// { // Issue #148
	// 	desc:    `missing key in different key same level`,
	// 	json:    `{"s":"s","ic":2,"r":{"o":"invalid"}}`,
	// 	path:    []string{"ic", "o"},
	// 	isFound: false,
	// },

	// Error/invalid tests
	// {
	// 	desc:    `handle escaped quote in key name in JSON`,
	// 	json:    `{"key\"key": 1}`,
	// 	path:    []string{"key"},
	// 	isFound: false,
	// },
	{
		desc:    `missing closing brace, but can still find key`,
		json:    `{"a":"b"`,
		path:    []string{"a"},
		isFound: true,
		data:    `b`,
	},
	{
		desc:  `missing value closing quote`,
		json:  `{"a":"b`,
		path:  []string{"a"},
		isErr: true,
	},
	{
		desc:  `missing value closing curly brace`,
		json:  `{"a": { "b": "c"`,
		path:  []string{"a"},
		isErr: true,
	},
	{
		desc:  `missing value closing square bracket`,
		json:  `{"a": [1, 2, 3 }`,
		path:  []string{"a"},
		isErr: true,
	},
	{
		desc:  `missing value 1`,
		json:  `{"a":`,
		path:  []string{"a"},
		isErr: true,
	},
	{
		desc:  `missing value 2`,
		json:  `{"a": `,
		path:  []string{"a"},
		isErr: true,
	},
	{
		desc:  `missing value 3`,
		json:  `{"a":}`,
		path:  []string{"a"},
		isErr: true,
	},
	{
		desc:    `malformed array (no closing brace)`,
		json:    `{"a":[, "b":123}`,
		path:    []string{"b"},
		isFound: false,
		isErr:   true,
	},
	// { // Issue #81
	// 	desc:    `missing key in object in array`,
	// 	json:    `{"p":{"a":[{"u":"abc","t":"th"}]}}`,
	// 	path:    []string{"p", "a", "[0]", "x"},
	// },
	// { // Issue #81 counter test
	// 	desc:    `existing key in object in array`,
	// 	json:    `{"p":{"a":[{"u":"abc","t":"th"}]}}`,
	// 	path:    []string{"p", "a", "[0]", "u"},
	// 	isFound: true,
	// 	data:    "abc",
	// },
	// { // This test returns not found instead of a parse error, as checking for the malformed JSON would reduce performance
	// 	desc:    "malformed key (followed by comma followed by colon)",
	// 	json:    `{"a",:1}`,
	// 	path:    []string{"a"},
	// 	isFound: false,
	// },
	{ // This test returns a match instead of a parse error, as checking for the malformed JSON would reduce performance (this is not ideal)
		desc:    "malformed 'colon chain', lookup first string",
		json:    `{"a":"b":"c"}`,
		path:    []string{"a"},
		isFound: true,
		data:    "b",
	},
	{ // This test returns a match instead of a parse error, as checking for the malformed JSON would reduce performance (this is not ideal)
		desc:    "malformed 'colon chain', lookup second string",
		json:    `{"a":"b":"c"}`,
		path:    []string{"b"},
		isFound: true,
		data:    "c",
	},
	// Array index paths
	// {
	// 	desc:    "last key in path is index",
	// 	json:    `{"a":[{"b":1},{"b":"2"}, 3],"c":{"c":[1,2]}}`,
	// 	path:    []string{"a", "[1]"},
	// 	isFound: true,
	// 	data:    `{"b":"2"}`,
	// },
	// {
	// 	desc:    "get string from array",
	// 	json:    `{"a":[{"b":1},"foo", 3],"c":{"c":[1,2]}}`,
	// 	path:    []string{"a", "[1]"},
	// 	isFound: true,
	// 	data:    "foo",
	// },
	// {
	// 	desc:    "key in path is index",
	// 	json:    `{"a":[{"b":"1"},{"b":"2"},3],"c":{"c":[1,2]}}`,
	// 	path:    []string{"a", "[0]", "b"},
	// 	isFound: true,
	// 	data:    `1`,
	// },
	// {
	// 	desc: "last key in path is an index to value in array (formatted json)",
	// 	json: `{
	// 	    "a": [
	// 		{
	// 		    "b": 1
	// 		},
	// 		{"b":"2"},
	// 		3
	// 	    ],
	// 	    "c": {
	// 		"c": [
	// 		    1,
	// 		    2
	// 		]
	// 	    }
	// 	}`,
	// 	path:    []string{"a", "[1]"},
	// 	isFound: true,
	// 	data:    `{"b":"2"}`,
	// },
	// {
	// 	desc: "key in path is index (formatted json)",
	// 	json: `{
	// 	"a": [
	// 		{"b": 1},
	// 		{"b": "2"},
	// 		3
	// 	],
	// 	"c": {
	// 			"c": [1, 2]
	// 		}
	// 	}`,
	// 	path:    []string{"a", "[0]", "b"},
	// 	isFound: true,
	// 	data:    `1`,
	// },
	// {
	// 	// Issue #178: Crash in searchKeys
	// 	desc:    `invalid json`,
	// 	json:    `{{{"":`,
	// 	path:    []string{"a", "b"},
	// 	isFound: false,
	// },
	// {
	// 	desc:    `opening brace instead of closing and without key`,
	// 	json:    `{"a":1{`,
	// 	path:    []string{"b"},
	// 	isFound: false,
	// },
}

func TestGet(t *testing.T) {
	for _, test := range getTests {
		t.Run(test.desc, func(t *testing.T) {
			data := []byte(test.json)
			value, _, _, err := Get(data, test.path...)
			if test.isErr {
				if err == nil {
					t.Errorf("Expected error, got nil")
				}
			} else {
				if err != nil {
					t.Errorf("Got error: %v", err)
				}
			}

			if test.isFound {
				if string(value) != test.data {
					t.Errorf("Expected '%s', got '%s'", test.data, value)
				}
			} else {
				if value != nil {
					t.Errorf("Expected nil, got '%s'", value)
				}
			}
		})
	}
}