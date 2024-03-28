package json

import (
	"testing"
)

func TestMarshal_Primitive(t *testing.T) {
	tests := []struct {
		name string
		node *Node
	}{
		{
			name: "null",
			node: NullNode(""),
		},
		{
			name: "true",
			node: BoolNode("", true),
		},
		{
			name: "false",
			node: BoolNode("", false),
		},
		{
			name: `"string"`,
			node: StringNode("", "string"),
		},
		{
			name: `"one \"encoded\" string"`,
			node: StringNode("", `one "encoded" string`),
		},
		{
			name: `{"foo":"bar"}`,
			node: ObjectNode("", map[string]*Node{
				"foo": StringNode("foo", "bar"),
			}),
		},
		{
			name: "42",
			node: NumberNode("", 42),
		},
		// TODO: fix output for not to use scientific notation
		{
			name: "1.005e+02",
			node: NumberNode("", 100.5),
		},
		{
			name: `[1,2,3]`,
			node: ArrayNode("", []*Node{
				NumberNode("0", 1),
				NumberNode("2", 2),
				NumberNode("3", 3),
			}),
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			value, err := Marshal(test.node)
			if err != nil {
				t.Errorf("unexpected error: %s", err)
			} else if string(value) != test.name {
				t.Errorf("wrong result: '%s', expected '%s'", value, test.name)
			}
		})
	}
}

func TestMarshal_Object(t *testing.T) {
	node := ObjectNode("", map[string]*Node{
		"foo": StringNode("foo", "bar"),
		"baz": NumberNode("baz", 100500),
		"qux": NullNode("qux"),
	})

	mustKey := []string{"foo", "baz", "qux"}

	value, err := Marshal(node)
	if err != nil {
		t.Errorf("unexpected error: %s", err)
	}

	// the order of keys in the map is not guaranteed
	// so we need to unmarshal the result and check the keys
	decoded, err := Unmarshal(value)
	if err != nil {
		t.Errorf("unexpected error: %s", err)
	}

	for _, key := range mustKey {
		if node, err := decoded.GetKey(key); err != nil {
			t.Errorf("unexpected error: %s", err)
		} else {
			if node == nil {
				t.Errorf("node is nil")
			} else if node.key == nil {
				t.Errorf("key is nil")
			} else if *node.key != key {
				t.Errorf("wrong key: '%s', expected '%s'", *node.key, key)
			}
		}
	}
}

func valueNode(prev *Node, key string, typ ValueType, val interface{}) *Node {
	curr := &Node{
		prev:     prev,
		data:     nil,
		key:      &key,
		borders:  [2]int{0, 0},
		value:    val,
		modified: true,
	}

	if val != nil {
		curr.nodeType = typ
	}

	return curr
}

func TestMarshal_Errors(t *testing.T) {
	tests := []struct {
		name string
		node func() (node *Node)
	}{
		{
			name: "nil",
			node: func() (node *Node) {
				return
			},
		},
		{
			name: "broken",
			node: func() (node *Node) {
				node = Must(Unmarshal([]byte(`{}`)))
				node.borders[1] = 0
				return
			},
		},
		{
			name: "Numeric",
			node: func() (node *Node) {
				return valueNode(nil, "", Number, false)
			},
		},
		{
			name: "String",
			node: func() (node *Node) {
				return valueNode(nil, "", String, false)
			},
		},
		{
			name: "Bool",
			node: func() (node *Node) {
				return valueNode(nil, "", Boolean, 1)
			},
		},
		{
			name: "Array_1",
			node: func() (node *Node) {
				node = ArrayNode("", nil)
				node.next["1"] = NullNode("1")
				return
			},
		},
		{
			name: "Array_2",
			node: func() (node *Node) {
				return ArrayNode("", []*Node{valueNode(nil, "", Boolean, 1)})
			},
		},
		{
			name: "Object",
			node: func() (node *Node) {
				return ObjectNode("", map[string]*Node{"key": valueNode(nil, "key", Boolean, 1)})
			},
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			value, err := Marshal(test.node())
			if err == nil {
				t.Errorf("expected error")
			} else if len(value) != 0 {
				t.Errorf("wrong result")
			}
		})
	}
}

func TestMarshal_Nil(t *testing.T) {
	_, err := Marshal(nil)
	if err == nil {
		t.Error("Expected error for nil node, but got nil")
	}
}

func TestMarshal_NotModified(t *testing.T) {
	node := &Node{}
	_, err := Marshal(node)
	if err == nil {
		t.Error("Expected error for not modified node, but got nil")
	}
}

func TestMarshalCycleReference(t *testing.T) {
	node1 := &Node{
		key:      stringPtr("node1"),
		nodeType: String,
		next: map[string]*Node{
			"next": nil,
		},
	}

	node2 := &Node{
		key:      stringPtr("node2"),
		nodeType: String,
		prev:     node1,
	}

	node1.next["next"] = node2

	_, err := Marshal(node1)
	if err == nil {
		t.Error("Expected error for cycle reference, but got nil")
	}
}

func TestMarshalNoCycleReference(t *testing.T) {
	node1 := &Node{
		key:      stringPtr("node1"),
		nodeType: String,
		value:    "value1",
		modified: true,
	}

	node2 := &Node{
		key:      stringPtr("node2"),
		nodeType: String,
		value:    "value2",
		modified: true,
	}

	_, err := Marshal(node1)
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}

	_, err = Marshal(node2)
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}
}

func stringPtr(s string) *string {
	return &s
}
