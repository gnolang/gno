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
				borders:  [2]uint{0, 9},
				modified: false,
			},
			expected: []byte("test data"),
		},
		{
			name: "ready modified node with data",
			node: &Node{
				data:     []byte("test data"),
				borders:  [2]uint{0, 9},
				modified: true,
			},
			expected: nil,
		},
		{
			name: "not ready node",
			node: &Node{
				data:     nil,
				borders:  [2]uint{0, 0},
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
