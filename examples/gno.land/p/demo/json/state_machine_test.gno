package json

import (
	"bytes"
	"errors"
	"fmt"
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
		desc:    "read boolean 2",
		json:    `false`,
		isFound: true,
		data:    `false`,
	},
	{
		desc:    "read null",
		json:    `null`,
		isFound: true,
		data:    `null`,
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

func TestArrayEach(t *testing.T) {
	mock := []byte(`{"a": {        "b":[{"x": "1"} ,{"x":    2},{ "x":3}, {"x":4} ]}}`)
	count := 0

	ArrayEach(mock, func(value []byte, dataType ValueType, offset int, err error) {
		count++

		switch count {
		case 1:
			if string(value) != `{"x": "1"}` {
				t.Errorf("Wrong first item: %s", string(value))
			}
		case 2:
			if string(value) != `{"x":    2}` {
				t.Errorf("Wrong second item: %s", string(value))
			}
		case 3:
			if string(value) != `{ "x":3}` {
				t.Errorf("Wrong third item: %s", string(value))
			}
		case 4:
			if string(value) != `{"x":4}` {
				t.Errorf("Wrong forth item: %s", string(value))
			}
		default:
			t.Errorf("Should process only 4 items")
		}
	}, "a", "b")
}

func TestArrayEach2(t *testing.T) {
	test := []byte(`{"a": ["string", 42, true, false, null, [1, 2, 3]]}`)
	ArrayEach(test, func(value []byte, dataType ValueType, offset int, err error) {
		if string(value) != "string" && string(value) != "42" && string(value) != "true" && string(value) != "false" && string(value) != "null" && string(value) != `{"b": 1}` && string(value) != "[1, 2, 3]" {
			t.Errorf("Wrong item: %s", string(value))
		}
	}, "a")
}

func TestObjectEach(t *testing.T) {
	tests := []struct {
		name     string
		input    []byte
		expected map[string]string
		root     string
		isErr    bool
	}{
		{
			name:     "empty object",
			input:    []byte(`{}`),
			expected: map[string]string{},
			root:     "",
		},
		{
			name:     "single key",
			input:    []byte(`{"a": 1}`),
			expected: map[string]string{"a": "1"},
			root:     "",
		},
		{
			name:     "two keys",
			input:    []byte(`{"a": {"b": 1, "c": 2}}`),
			expected: map[string]string{"b": "1", "c": "2"},
			root:     "a",
		},
		{
			name:     "two keys 2",
			input:    []byte(`{"a": {"b": 1, "c": 2}, "d": 4}`),
			expected: map[string]string{"b": "1", "c": "2"},
			root:     "a",
		},
		{
			name:     "another set of keys",
			input:    []byte(`{"x": {"y": 3, "z": 4}}`),
			expected: map[string]string{"y": "3", "z": "4"},
			root:     "x",
		},
		{
			name: "multiple key-value object with many value types",
			input: []byte(`{
				"key1": null,
				"key2": true,
				"key3": 1.23,
				"key4": "string value",
				"key5": [1,2,3],
				"key6": {"a":"b"}
			}`),
			expected: map[string]string{
				"key1": "null",
				"key2": "true",
				"key3": "1.23",
				"key4": "string value",
				"key5": "[1,2,3]",
				"key6": `{"a":"b"}`,
			},
		},
		{
			name:     "so many white spaces",
			input:    []byte(`{"a"               :                 1}`),
			expected: map[string]string{"a": "1"},
		},
		/* Error Cases */
		{
			name:  "unmatched brace",
			input: []byte(`{`),
			isErr: true,
		},
		{
			name:  "unmatched brace 2",
			input: []byte(`{"a": 1`),
			root:  "a",
			isErr: true,
		},
		{
			name:  "unmatched brace 3",
			input: []byte(`}`),
			isErr: true,
		},
		{
			name:  "unmatched brace 3",
			input: []byte(`{{}}{`),
			isErr: true,
		},
		{
			name:  "no object present",
			input: []byte(`\t\n\r`),
			isErr: true,
		},
		{
			name:  "malformed key",
			input: []byte(`{"foo: 1}`),
			isErr: true,
		},
		{
			name:  "malformed key 2",
			input: []byte(`{"foo": 1, "bar: "11"}`),
			root:  "foo",
			isErr: true,
		},
		{
			name:  "malformed key in nested object",
			input: []byte(`{"foo": {"bar: 1}}`),
			root:  "foo",
			isErr: true,
		},
		{
			name:  "bad value",
			input: []byte(`{"foo": bar}`),
			isErr: true,
		},
		{
			name:  "no colon",
			input: []byte(`{"foo" "bar"}`),
			isErr: true,
		},
		{
			name:  "no colon 2",
			input: []byte(`{"foo""bar"}`),
			isErr: true,
		},
		{
			name:  "no colon 3",
			input: []byte(`{"foo"; "bar"}`),
			isErr: true,
		},
		{
			name:  "invalid colon",
			input: []byte(`{"foo":: "bar"}`),
			isErr: true,
		},
		{
			name:  "no trailing comma",
			input: []byte(`{"foo": "bar" "baz": "qux"}`),
			root:  "foo",
			isErr: true,
		},
		{
			name:  "key with unicode escape",
			input: []byte(`{"b": 10, "c": {"a\uD83D\uDE03b":1, "aa": false}}`),
			root:  "b",
			isErr: true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			err := ObjectEach(tc.input, func(key, value []byte, dataType ValueType, offset int) error {
				expectedValue, ok := tc.expected[string(key)]
				if !ok {
					t.Errorf("Unexpected key: %s", string(key))
				} else if string(value) != expectedValue {
					t.Errorf("Wrong value for key %s: got %s, want %s", string(key), string(value), expectedValue)
				}
				return nil
			}, tc.root)

			if (err != nil) != tc.isErr {
				t.Errorf("ObjectEach(%s) error = %v, wantErr %v", tc.name, err, tc.isErr)
			}
		})
	}
}
