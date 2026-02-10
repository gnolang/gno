package json

import (
	"errors"
	"strconv"
	"strings"

	"gno.land/p/nt/ufmt"
)

// Node represents a JSON node.
type Node struct {
	prev     *Node            // prev is the parent node of the current node.
	next     map[string]*Node // next is the child nodes of the current node.
	key      *string          // key holds the key of the current node in the parent node.
	data     []byte           // byte slice of JSON data
	value    any              // value holds the value of the current node.
	nodeType ValueType        // NodeType holds the type of the current node. (Object, Array, String, Number, Boolean, Null)
	index    *int             // index holds the index of the current node in the parent array node.
	borders  [2]int           // borders stores the start and end index of the current node in the data.
	modified bool             // modified indicates the current node is changed or not.
}

// NewNode creates a new node instance with the given parent node, buffer, type, and key.
func NewNode(prev *Node, b *buffer, typ ValueType, key **string) (*Node, error) {
	curr := &Node{
		prev:     prev,
		data:     b.data,
		borders:  [2]int{b.index, 0},
		key:      *key,
		nodeType: typ,
		modified: false,
	}

	if typ == Object || typ == Array {
		curr.next = make(map[string]*Node)
	}

	if prev != nil {
		if prev.IsArray() {
			size := len(prev.next)
			curr.index = &size

			prev.next[strconv.Itoa(size)] = curr
		} else if prev.IsObject() {
			if key == nil {
				return nil, errKeyRequired
			}

			prev.next[**key] = curr
		} else {
			return nil, errors.New("invalid parent type")
		}
	}

	return curr, nil
}

// load retrieves the value of the current node.
func (n *Node) load() any {
	return n.value
}

// Changed checks the current node is changed or not.
func (n *Node) Changed() bool {
	return n.modified
}

// Key returns the key of the current node.
func (n *Node) Key() string {
	if n == nil || n.key == nil {
		return ""
	}

	return *n.key
}

// HasKey checks the current node has the given key or not.
func (n *Node) HasKey(key string) bool {
	if n == nil {
		return false
	}

	_, ok := n.next[key]
	return ok
}

// GetKey returns the value of the given key from the current object node.
func (n *Node) GetKey(key string) (*Node, error) {
	if n == nil {
		return nil, errNilNode
	}

	if n.Type() != Object {
		return nil, ufmt.Errorf("target node is not object type. got: %s", n.Type().String())
	}

	value, ok := n.next[key]
	if !ok {
		return nil, ufmt.Errorf("key not found: %s", key)
	}

	return value, nil
}

// MustKey returns the value of the given key from the current object node.
func (n *Node) MustKey(key string) *Node {
	val, err := n.GetKey(key)
	if err != nil {
		panic(err)
	}

	return val
}

// UniqueKeyLists traverses the current JSON nodes and collects all the unique keys.
func (n *Node) UniqueKeyLists() []string {
	var collectKeys func(*Node) []string
	collectKeys = func(node *Node) []string {
		if node == nil || !node.IsObject() {
			return nil
		}

		result := make(map[string]bool)
		for key, childNode := range node.next {
			result[key] = true
			childKeys := collectKeys(childNode)
			for _, childKey := range childKeys {
				result[childKey] = true
			}
		}

		keys := make([]string, 0, len(result))
		for key := range result {
			keys = append(keys, key)
		}
		return keys
	}

	return collectKeys(n)
}

// Empty returns true if the current node is empty.
func (n *Node) Empty() bool {
	if n == nil {
		return false
	}

	return len(n.next) == 0
}

// Type returns the type (ValueType) of the current node.
func (n *Node) Type() ValueType {
	return n.nodeType
}

// Value returns the value of the current node.
//
// Usage:
//
//	root := Unmarshal([]byte(`{"key": "value"}`))
//	val, err := root.MustKey("key").Value()
//	if err != nil {
//		t.Errorf("Value returns error: %v", err)
//	}
//
//	result: "value"
func (n *Node) Value() (value any, err error) {
	value = n.load()

	if value == nil {
		switch n.nodeType {
		case Null:
			return nil, nil

		case Number:
			value, err = strconv.ParseFloat(string(n.source()), 64)
			if err != nil {
				return nil, err
			}

			n.value = value

		case String:
			var ok bool
			value, ok = Unquote(n.source(), doubleQuote)
			if !ok {
				return "", errInvalidStringValue
			}

			n.value = value

		case Boolean:
			if len(n.source()) == 0 {
				return nil, errEmptyBooleanNode
			}

			b := n.source()[0]
			value = b == 't' || b == 'T'
			n.value = value

		case Array:
			elems := make([]*Node, len(n.next))

			for _, e := range n.next {
				elems[*e.index] = e
			}

			value = elems
			n.value = value

		case Object:
			obj := make(map[string]*Node, len(n.next))

			for k, v := range n.next {
				obj[k] = v
			}

			value = obj
			n.value = value
		}
	}

	return value, nil
}

// Delete removes the current node from the parent node.
//
// Usage:
//
//	root := Unmarshal([]byte(`{"key": "value"}`))
//	if err := root.MustKey("key").Delete(); err != nil {
//		t.Errorf("Delete returns error: %v", err)
//	}
//
//	result: {} (empty object)
func (n *Node) Delete() error {
	if n == nil {
		return errors.New("can't delete nil node")
	}

	if n.prev == nil {
		return nil
	}

	return n.prev.remove(n)
}

// Size returns the size (length) of the current array node.
//
// Usage:
//
//	root := ArrayNode("", []*Node{StringNode("", "foo"), NumberNode("", 1)})
//	if root == nil {
//		t.Errorf("ArrayNode returns nil")
//	}
//
//	if root.Size() != 2 {
//		t.Errorf("ArrayNode returns wrong size: %d", root.Size())
//	}
func (n *Node) Size() int {
	if n == nil {
		return 0
	}

	return len(n.next)
}

// Index returns the index of the current node in the parent array node.
//
// Usage:
//
//	root := ArrayNode("", []*Node{StringNode("", "foo"), NumberNode("", 1)})
//	if root == nil {
//		t.Errorf("ArrayNode returns nil")
//	}
//
//	if root.MustIndex(1).Index() != 1 {
//		t.Errorf("Index returns wrong index: %d", root.MustIndex(1).Index())
//	}
//
// We can also use the index to the byte slice of the JSON data directly.
//
// Example:
//
//	root := Unmarshal([]byte(`["foo", 1]`))
//	if root == nil {
//		t.Errorf("Unmarshal returns nil")
//	}
//
//	if string(root.MustIndex(1).source()) != "1" {
//		t.Errorf("source returns wrong result: %s", root.MustIndex(1).source())
//	}
func (n *Node) Index() int {
	if n == nil || n.index == nil {
		return -1
	}

	return *n.index
}

// MustIndex returns the array element at the given index.
//
// If the index is negative, it returns the index is from the end of the array.
// Also, it panics if the index is not found.
//
// check the Index method for detailed usage.
func (n *Node) MustIndex(expectIdx int) *Node {
	val, err := n.GetIndex(expectIdx)
	if err != nil {
		panic(err)
	}

	return val
}

// GetIndex returns the array element at the given index.
//
// if the index is negative, it returns the index is from the end of the array.
func (n *Node) GetIndex(idx int) (*Node, error) {
	if n == nil {
		return nil, errNilNode
	}

	if !n.IsArray() {
		return nil, errNotArrayNode
	}

	if idx > n.Size() {
		return nil, errors.New("input index exceeds the array size")
	}

	if idx < 0 {
		idx += len(n.next)
	}

	child, ok := n.next[strconv.Itoa(idx)]
	if !ok {
		return nil, errIndexNotFound
	}

	return child, nil
}

// DeleteIndex removes the array element at the given index.
func (n *Node) DeleteIndex(idx int) error {
	node, err := n.GetIndex(idx)
	if err != nil {
		return err
	}

	return n.remove(node)
}

// NullNode creates a new null type node.
//
// Usage:
//
//	_ := NullNode("")
func NullNode(key string) *Node {
	return &Node{
		key:      &key,
		value:    nil,
		nodeType: Null,
		modified: true,
	}
}

// NumberNode creates a new number type node.
//
// Usage:
//
//	root := NumberNode("", 1)
//	if root == nil {
//		t.Errorf("NumberNode returns nil")
//	}
func NumberNode(key string, value float64) *Node {
	return &Node{
		key:      &key,
		value:    value,
		nodeType: Number,
		modified: true,
	}
}

// StringNode creates a new string type node.
//
// Usage:
//
//	root := StringNode("", "foo")
//	if root == nil {
//		t.Errorf("StringNode returns nil")
//	}
func StringNode(key string, value string) *Node {
	return &Node{
		key:      &key,
		value:    value,
		nodeType: String,
		modified: true,
	}
}

// BoolNode creates a new given boolean value node.
//
// Usage:
//
//	root := BoolNode("", true)
//	if root == nil {
//		t.Errorf("BoolNode returns nil")
//	}
func BoolNode(key string, value bool) *Node {
	return &Node{
		key:      &key,
		value:    value,
		nodeType: Boolean,
		modified: true,
	}
}

// ArrayNode creates a new array type node.
//
// If the given value is nil, it creates an empty array node.
//
// Usage:
//
//	root := ArrayNode("", []*Node{StringNode("", "foo"), NumberNode("", 1)})
//	if root == nil {
//		t.Errorf("ArrayNode returns nil")
//	}
func ArrayNode(key string, value []*Node) *Node {
	curr := &Node{
		key:      &key,
		nodeType: Array,
		modified: true,
	}

	curr.next = make(map[string]*Node, len(value))
	if value != nil {
		curr.value = value

		for i, v := range value {
			idx := i
			curr.next[strconv.Itoa(i)] = v

			v.prev = curr
			v.index = &idx
		}
	}

	return curr
}

// ObjectNode creates a new object type node.
//
// If the given value is nil, it creates an empty object node.
//
// next is a map of key and value pairs of the object.
func ObjectNode(key string, value map[string]*Node) *Node {
	curr := &Node{
		nodeType: Object,
		key:      &key,
		next:     value,
		modified: true,
	}

	if value != nil {
		curr.value = value

		for key, val := range value {
			vkey := key
			val.prev = curr
			val.key = &vkey
		}
	} else {
		curr.next = make(map[string]*Node)
	}

	return curr
}

// IsArray returns true if the current node is array type.
func (n *Node) IsArray() bool {
	return n.nodeType == Array
}

// IsObject returns true if the current node is object type.
func (n *Node) IsObject() bool {
	return n.nodeType == Object
}

// IsNull returns true if the current node is null type.
func (n *Node) IsNull() bool {
	return n.nodeType == Null
}

// IsBool returns true if the current node is boolean type.
func (n *Node) IsBool() bool {
	return n.nodeType == Boolean
}

// IsString returns true if the current node is string type.
func (n *Node) IsString() bool {
	return n.nodeType == String
}

// IsNumber returns true if the current node is number type.
func (n *Node) IsNumber() bool {
	return n.nodeType == Number
}

// ready checks the current node is ready or not.
//
// the meaning of ready is the current node is parsed and has a valid value.
func (n *Node) ready() bool {
	return n.borders[1] != 0
}

// source returns the source of the current node.
func (n *Node) source() []byte {
	if n == nil {
		return nil
	}

	if n.ready() && !n.modified && n.data != nil {
		return (n.data)[n.borders[0]:n.borders[1]]
	}

	return nil
}

// root returns the root node of the current node.
func (n *Node) root() *Node {
	if n == nil {
		return nil
	}

	curr := n
	for curr.prev != nil {
		curr = curr.prev
	}

	return curr
}

// GetNull returns the null value if current node is null type.
//
// Usage:
//
//	root := Unmarshal([]byte("null"))
//	val, err := root.GetNull()
//	if err != nil {
//		t.Errorf("GetNull returns error: %v", err)
//	}
//	if val != nil {
//		t.Errorf("GetNull returns wrong result: %v", val)
//	}
func (n *Node) GetNull() (any, error) {
	if n == nil {
		return nil, errNilNode
	}

	if !n.IsNull() {
		return nil, errNotNullNode
	}

	return nil, nil
}

// MustNull returns the null value if current node is null type.
//
// It panics if the current node is not null type.
func (n *Node) MustNull() any {
	v, err := n.GetNull()
	if err != nil {
		panic(err)
	}

	return v
}

// GetNumeric returns the numeric (int/float) value if current node is number type.
//
// Usage:
//
//	root := Unmarshal([]byte("10.5"))
//	val, err := root.GetNumeric()
//	if err != nil {
//		t.Errorf("GetNumeric returns error: %v", err)
//	}
//	println(val) // 10.5
func (n *Node) GetNumeric() (float64, error) {
	if n == nil {
		return 0, errNilNode
	}

	if n.nodeType != Number {
		return 0, errNotNumberNode
	}

	val, err := n.Value()
	if err != nil {
		return 0, err
	}

	v, ok := val.(float64)
	if !ok {
		return 0, errNotNumberNode
	}

	return v, nil
}

// MustNumeric returns the numeric (int/float) value if current node is number type.
//
// It panics if the current node is not number type.
func (n *Node) MustNumeric() float64 {
	v, err := n.GetNumeric()
	if err != nil {
		panic(err)
	}

	return v
}

// GetString returns the string value if current node is string type.
//
// Usage:
//
//	root, err := Unmarshal([]byte("foo"))
//	if err != nil {
//		t.Errorf("Error on Unmarshal(): %s", err)
//	}
//
//	str, err := root.GetString()
//	if err != nil {
//		t.Errorf("should retrieve string value: %s", err)
//	}
//
//	println(str) // "foo"
func (n *Node) GetString() (string, error) {
	if n == nil {
		return "", errEmptyStringNode
	}

	if !n.IsString() {
		return "", errNotStringNode
	}

	val, err := n.Value()
	if err != nil {
		return "", err
	}

	v, ok := val.(string)
	if !ok {
		return "", errNotStringNode
	}

	return v, nil
}

// MustString returns the string value if current node is string type.
//
// It panics if the current node is not string type.
func (n *Node) MustString() string {
	v, err := n.GetString()
	if err != nil {
		panic(err)
	}

	return v
}

// GetBool returns the boolean value if current node is boolean type.
//
// Usage:
//
//	root := Unmarshal([]byte("true"))
//	val, err := root.GetBool()
//	if err != nil {
//		t.Errorf("GetBool returns error: %v", err)
//	}
//	println(val) // true
func (n *Node) GetBool() (bool, error) {
	if n == nil {
		return false, errNilNode
	}

	if n.nodeType != Boolean {
		return false, errNotBoolNode
	}

	val, err := n.Value()
	if err != nil {
		return false, err
	}

	v, ok := val.(bool)
	if !ok {
		return false, errNotBoolNode
	}

	return v, nil
}

// MustBool returns the boolean value if current node is boolean type.
//
// It panics if the current node is not boolean type.
func (n *Node) MustBool() bool {
	v, err := n.GetBool()
	if err != nil {
		panic(err)
	}

	return v
}

// GetArray returns the array value if current node is array type.
//
// Usage:
//
//		root := Must(Unmarshal([]byte(`["foo", 1]`)))
//		arr, err := root.GetArray()
//		if err != nil {
//			t.Errorf("GetArray returns error: %v", err)
//		}
//
//		for _, val := range arr {
//			println(val)
//		}
//
//	 result: "foo", 1
func (n *Node) GetArray() ([]*Node, error) {
	if n == nil {
		return nil, errNilNode
	}

	if n.nodeType != Array {
		return nil, errNotArrayNode
	}

	val, err := n.Value()
	if err != nil {
		return nil, err
	}

	v, ok := val.([]*Node)
	if !ok {
		return nil, errNotArrayNode
	}

	return v, nil
}

// MustArray returns the array value if current node is array type.
//
// It panics if the current node is not array type.
func (n *Node) MustArray() []*Node {
	v, err := n.GetArray()
	if err != nil {
		panic(err)
	}

	return v
}

// AppendArray appends the given values to the current array node.
//
// If the current node is not array type, it returns an error.
//
// Example 1:
//
//	root := Must(Unmarshal([]byte(`[{"foo":"bar"}]`)))
//	if err := root.AppendArray(NullNode("")); err != nil {
//		t.Errorf("should not return error: %s", err)
//	}
//
//	result: [{"foo":"bar"}, null]
//
// Example 2:
//
//	root := Must(Unmarshal([]byte(`["bar", "baz"]`)))
//	err := root.AppendArray(NumberNode("", 1), StringNode("", "foo"))
//	if err != nil {
//		t.Errorf("AppendArray returns error: %v", err)
//	 }
//
//	result: ["bar", "baz", 1, "foo"]
func (n *Node) AppendArray(value ...*Node) error {
	if !n.IsArray() {
		return errInvalidAppend
	}

	for _, val := range value {
		if err := n.append(nil, val); err != nil {
			return err
		}
	}

	n.mark()
	return nil
}

// ArrayEach executes the callback for each element in the JSON array.
//
// Usage:
//
//	jsonArrayNode.ArrayEach(func(i int, valueNode *Node) {
//	    ufmt.Println(i, valueNode)
//	})
func (n *Node) ArrayEach(callback func(i int, target *Node)) {
	if n == nil || !n.IsArray() {
		return
	}

	for idx := 0; idx < len(n.next); idx++ {
		element, err := n.GetIndex(idx)
		if err != nil {
			continue
		}

		callback(idx, element)
	}
}

// GetObject returns the object value if current node is object type.
//
// Usage:
//
//	root := Must(Unmarshal([]byte(`{"key": "value"}`)))
//	obj, err := root.GetObject()
//	if err != nil {
//		t.Errorf("GetObject returns error: %v", err)
//	}
//
//	result: map[string]*Node{"key": StringNode("key", "value")}
func (n *Node) GetObject() (map[string]*Node, error) {
	if n == nil {
		return nil, errNilNode
	}

	if !n.IsObject() {
		return nil, errNotObjectNode
	}

	val, err := n.Value()
	if err != nil {
		return nil, err
	}

	v, ok := val.(map[string]*Node)
	if !ok {
		return nil, errNotObjectNode
	}

	return v, nil
}

// MustObject returns the object value if current node is object type.
//
// It panics if the current node is not object type.
func (n *Node) MustObject() map[string]*Node {
	v, err := n.GetObject()
	if err != nil {
		panic(err)
	}

	return v
}

// AppendObject appends the given key and value to the current object node.
//
// If the current node is not object type, it returns an error.
func (n *Node) AppendObject(key string, value *Node) error {
	if !n.IsObject() {
		return errInvalidAppend
	}

	if err := n.append(&key, value); err != nil {
		return err
	}

	n.mark()
	return nil
}

// ObjectEach executes the callback for each key-value pair in the JSON object.
//
// Usage:
//
//	jsonObjectNode.ObjectEach(func(key string, valueNode *Node) {
//	    ufmt.Println(key, valueNode)
//	})
func (n *Node) ObjectEach(callback func(key string, value *Node)) {
	if n == nil || !n.IsObject() {
		return
	}

	for key, child := range n.next {
		callback(key, child)
	}
}

// String converts the node to a string representation.
func (n *Node) String() string {
	if n == nil {
		return ""
	}

	if n.ready() && !n.modified {
		return string(n.source())
	}

	val, err := Marshal(n)
	if err != nil {
		return "error: " + err.Error()
	}

	return string(val)
}

// Path builds the path of the current node.
//
// For example:
//
//	{ "key": { "sub": [ "val1", "val2" ] }}
//
// The path of "val2" is: $.key.sub[1]
func (n *Node) Path() string {
	if n == nil {
		return ""
	}

	var sb strings.Builder

	if n.prev == nil {
		sb.WriteString("$")
	} else {
		sb.WriteString(n.prev.Path())

		if n.key != nil {
			sb.WriteString("['" + n.Key() + "']")
		} else {
			sb.WriteString("[" + strconv.Itoa(n.Index()) + "]")
		}
	}

	return sb.String()
}

// mark marks the current node as modified.
func (n *Node) mark() {
	node := n
	for node != nil && !node.modified {
		node.modified = true
		node = node.prev
	}
}

// isContainer checks the current node type is array or object.
func (n *Node) isContainer() bool {
	return n.IsArray() || n.IsObject()
}

// remove removes the value from the current container type node.
func (n *Node) remove(v *Node) error {
	if !n.isContainer() {
		return ufmt.Errorf(
			"can't remove value from non-array or non-object node. got=%s",
			n.Type().String(),
		)
	}

	if v.prev != n {
		return errors.New("invalid parent node")
	}

	n.mark()
	if n.IsArray() {
		delete(n.next, strconv.Itoa(*v.index))
		n.dropIndex(*v.index)
	} else {
		delete(n.next, *v.key)
	}

	v.prev = nil
	return nil
}

// dropIndex rebase the index of current array node values.
func (n *Node) dropIndex(idx int) {
	for i := idx + 1; i <= len(n.next); i++ {
		prv := i - 1
		if curr, ok := n.next[strconv.Itoa(i)]; ok {
			curr.index = &prv
			n.next[strconv.Itoa(prv)] = curr
		}

		delete(n.next, strconv.Itoa(i))
	}
}

// append is a helper function to append the given value to the current container type node.
func (n *Node) append(key *string, val *Node) error {
	if n.isSameOrParentNode(val) {
		return errInvalidAppendCycle
	}

	if val.prev != nil {
		if err := val.prev.remove(val); err != nil {
			return err
		}
	}

	val.prev = n
	val.key = key

	if key == nil {
		size := len(n.next)
		val.index = &size
		n.next[strconv.Itoa(size)] = val
	} else {
		if old, ok := n.next[*key]; ok {
			if err := n.remove(old); err != nil {
				return err
			}
		}
		n.next[*key] = val
	}

	return nil
}

func (n *Node) isSameOrParentNode(nd *Node) bool {
	return n == nd || n.isParentNode(nd)
}

func (n *Node) isParentNode(nd *Node) bool {
	if n == nil {
		return false
	}

	for curr := nd.prev; curr != nil; curr = curr.prev {
		if curr == n {
			return true
		}
	}

	return false
}

// cptrs returns the pointer of the given string value.
func cptrs(cpy *string) *string {
	if cpy == nil {
		return nil
	}

	val := *cpy

	return &val
}

// cptri returns the pointer of the given integer value.
func cptri(i *int) *int {
	if i == nil {
		return nil
	}

	val := *i
	return &val
}

// Must panics if the given node is not fulfilled the expectation.
// Usage:
//
//	node := Must(Unmarshal([]byte(`{"key": "value"}`))
func Must(root *Node, expect error) *Node {
	if expect != nil {
		panic(expect)
	}

	return root
}
