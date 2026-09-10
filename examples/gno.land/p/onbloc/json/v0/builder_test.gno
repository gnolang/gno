package json

import (
	"testing"
)

func TestNodeBuilder(t *testing.T) {
	tests := []struct {
		name     string
		build    func() *Node
		expected string
	}{
		{
			name: "plain object",
			build: func() *Node {
				return Builder().
					WriteString("name", "Alice").
					WriteNumber("age", 30).
					WriteBool("is_student", false).
					Node()
			},
			expected: `{"name":"Alice","age":30,"is_student":false}`,
		},
		{
			name: "nested object",
			build: func() *Node {
				return Builder().
					WriteString("name", "Alice").
					WriteObject("address", func(b *NodeBuilder) {
						b.WriteString("city", "New York").
							WriteNumber("zipcode", 10001)
					}).
					Node()
			},
			expected: `{"name":"Alice","address":{"city":"New York","zipcode":10001}}`,
		},
		{
			name: "null node",
			build: func() *Node {
				return Builder().WriteNull("foo").Node()
			},
			expected: `{"foo":null}`,
		},
		{
			name: "array node",
			build: func() *Node {
				return Builder().
					WriteArray("items", func(ab *ArrayBuilder) {
						ab.WriteString("item1").
							WriteString("item2").
							WriteString("item3")
					}).
					Node()
			},
			expected: `{"items":["item1","item2","item3"]}`,
		},
		{
			name: "array with objects",
			build: func() *Node {
				return Builder().
					WriteArray("users", func(ab *ArrayBuilder) {
						ab.WriteObject(func(b *NodeBuilder) {
							b.WriteString("name", "Bob").
								WriteNumber("age", 25)
						}).
							WriteObject(func(b *NodeBuilder) {
								b.WriteString("name", "Carol").
									WriteNumber("age", 27)
							})
					}).
					Node()
			},
			expected: `{"users":[{"name":"Bob","age":25},{"name":"Carol","age":27}]}`,
		},
		{
			name: "array with various types",
			build: func() *Node {
				return Builder().
					WriteArray("values", func(ab *ArrayBuilder) {
						ab.WriteString("item1").
							WriteNumber(123).
							WriteBool(true).
							WriteNull()
					}).
					Node()
			},
			expected: `{"values":["item1",123,true,null]}`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			node := tt.build()
			value, err := Marshal(node)
			if err != nil {
				t.Errorf("unexpected error: %s", err)
			}
			if string(value) != tt.expected {
				t.Errorf("expected %s, got %s", tt.expected, string(value))
			}
		})
	}
}
