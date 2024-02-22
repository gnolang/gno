package json

import (
	"errors"
	"strconv"

	"gno.land/p/demo/ufmt"
)

type NodeValue struct {
	value interface{}
	typ   ValueType
}

func newNodeValue(value interface{}) *NodeValue {
	return &NodeValue{
		value: value,
		typ:   typeOf(value),
	}
}

func (nv *NodeValue) Store(value interface{}) {
	nv.value = value
	nv.typ = typeOf(value)
}

func (nv *NodeValue) Load() interface{} {
	return nv.value
}

type Node struct {
	prev     *Node
	next     map[string]*Node
	key      string
	data     []byte
	value    *NodeValue
	index    int
	borders  [2]int  // start, end
	modified bool
}


func (n *Node) Key() string {
	if n == nil || n.key == "" {
		return ""
	}

	return n.key
}

func (n *Node) Type() string {
	return n.value.typ.String()
}

func (n *Node) Value() interface{} {
	return n.value.Load()
}

func (n *Node) Size() uint {
	if n == nil {
		return 0
	}

	return uint(len(n.next))
}

func (n *Node) Keys() []string {
	if n == nil {
		return nil
	}

	result := make([]string, 0, len(n.next))
	for key := range n.next {
		result = append(result, key)
	}

	return result
}

func NullNode(key string) *Node {
	return &Node{
		key:   key,
		value: &NodeValue{value: nil, typ: Null},
		modified: true,
	}
}

func NumberNode(key string, value float64) *Node {
	return &Node{
		key:   key,
		value: &NodeValue{
			value: value,
			typ:   Number,
		},
		modified: true,
	}
}

func StringNode(key string, value string) *Node {
	val := newNodeValue(value)

	return &Node{
		key:   key,
		value: val,
		modified: true,
	}
}

func BoolNode(key string, value bool) *Node {
	val := newNodeValue(value)

	return &Node{
		key:   key,
		value: val,
		modified: true,
	}
}

func ArrayNode(key string, value []*Node) *Node {
	val := newNodeValue(value)
	curr := &Node{
		key:   key,
		value: val,
		modified: true,
	}

	curr.next = make(map[string]*Node, len(value))
	if value != nil {
		curr.value.Store(value)

		for i, v := range value {
			idx := i
			curr.next[strconv.Itoa(i)] = v

			v.prev = curr
			v.index = idx
		}
	}

	return curr
}

func (n *Node) ready() bool {
	return n.borders[1] != 0
}

func (n *Node) Source() []byte {
	if n == nil {
		return nil
	}

	if n.ready() && !n.modified && n.data != nil {
		return (n.data)[n.borders[0]:n.borders[1]]
	}

	return nil
}
