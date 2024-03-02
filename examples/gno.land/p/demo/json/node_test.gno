package json

import (
	"strconv"
	"testing"
)

func TestNodeValue(t *testing.T) {
	tests := []struct {
		name       string
		value      interface{}
		expectType ValueType
	}{
		{"string", "test value", String},
		{"integer", 123, Number},
		{"float", 1.23, Float},
		{"boolean", true, Boolean},
		{"null", nil, Null},
		{"array", []interface{}{1, 2, 3}, Array},
		{"object", map[string]interface{}{"key": "value"}, Object},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			nv := &NodeValue{}
			nv.Store(tt.value)

			if nv.value != tt.value {
				t.Errorf("Store: %s want %v got %v", tt.name, tt.value, nv.value)
			}

			typ := typeOf(tt.value)
			if typ != tt.expectType {
				t.Errorf("Store: %s want type %v got %v", tt.name, tt.expectType.String(), typ.String())
			}

			loadedValue := nv.Load()
			if loadedValue != tt.value {
				t.Errorf("Load: %s want %v got %v", tt.name, tt.value, loadedValue)
			}
		})
	}
}

func TestNodeCreation(t *testing.T) {
	tests := []struct {
		name     string
		testFunc func() *Node
		expected Node
	}{
		{
			name: "NullNode",
			testFunc: func() *Node {
				return NullNode("testNull")
			},
			expected: Node{
				key:      "testNull",
				value:    &NodeValue{value: nil, typ: Null},
				modified: true,
			},
		},
		{
			name: "NumberNode",
			testFunc: func() *Node {
				return NumberNode("testNumber", 42.0)
			},
			expected: Node{
				key:      "testNumber",
				value:    &NodeValue{value: 42.0, typ: Number},
				modified: true,
			},
		},
		{
			name: "StringNode",
			testFunc: func() *Node {
				return StringNode("testString", "Hello")
			},
			expected: Node{
				key:      "testString",
				value:    &NodeValue{value: "Hello", typ: String},
				modified: true,
			},
		},
		{
			name: "BoolNode",
			testFunc: func() *Node {
				return BoolNode("testBool", true)
			},
			expected: Node{
				key:      "testBool",
				value:    &NodeValue{value: true, typ: Boolean},
				modified: true,
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			actual := tc.testFunc()
			if actual.key != tc.expected.key {
				t.Errorf("expected key %s, got %s", tc.expected.key, actual.key)
			}
			if actual.value.typ != tc.expected.value.typ {
				t.Errorf("expected type %v, got %v", tc.expected.value.typ.String(), actual.value.typ.String())
			}
			if actual.modified != tc.expected.modified {
				t.Errorf("expected modified %t, got %t", tc.expected.modified, actual.modified)
			}
			// Special handling for comparing slices and maps
			if tc.name == "ArrayNode" {
				if len(actual.next) != len(tc.expected.next) {
					t.Errorf("expected next length %d, got %d", len(tc.expected.next), len(actual.next))
				}
				for k, v := range tc.expected.next {
					if _, ok := actual.next[k]; !ok {
						t.Errorf("expected next to have key %s", k)
					}
					if actual.next[k].key != v.key {
						t.Errorf("expected next node with key %s, got %s", v.key, actual.next[k].key)
					}
				}
			}
		})
	}
}

func TestNodeSource(t *testing.T) {
	tests := []struct {
		name     string
		node     *Node
		expected []byte
	}{
		{
			name:     "nil node",
			node:     nil,
			expected: nil,
		},
		{
			name: "ready unmodified node with data",
			node: &Node{
				data:     []byte("test data"),
				borders:  [2]int{0, 9},
				modified: false,
			},
			expected: []byte("test data"),
		},
		{
			name: "ready modified node with data",
			node: &Node{
				data:     []byte("test data"),
				borders:  [2]int{0, 9},
				modified: true,
			},
			expected: nil,
		},
		{
			name: "not ready node",
			node: &Node{
				data:     nil,
				borders:  [2]int{0, 0},
				modified: false,
			},
			expected: nil,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			actual := tc.node.Source()
			if string(actual) != string(tc.expected) {
				t.Errorf("expected %v, got %v", tc.expected, actual)
			}
		})
	}
}
