package jsonutil

import (
	"std"
	"strconv"
	"time"

	"gno.land/p/demo/avl"
	"gno.land/p/demo/json"
	"gno.land/p/demo/users"
)

func UnionNode(variant string, value *json.Node) *json.Node {
	return json.ObjectNode("", map[string]*json.Node{
		variant: value,
	})
}

func MustUnion(value *json.Node) (string, *json.Node) {
	obj := value.MustObject()
	for key, value := range obj {
		return key, value
	}

	panic("no variant in union")
}

func TimeNode(value time.Time) *json.Node {
	j, err := value.MarshalJSON()
	if err != nil {
		panic(err)
	}

	return json.StringNode("", string(j[1:len(j)-1]))
}

func MustTime(value *json.Node) time.Time {
	t := time.Time{}
	err := t.UnmarshalJSON([]byte(value.String()))
	if err != nil {
		panic(err)
	}

	return t
}

func DurationNode(value time.Duration) *json.Node {
	return Int64Node(value.Nanoseconds())
}

func EmptyObjectNode() *json.Node {
	return json.ObjectNode("", nil)
}

// int is always 64 bits in gno so we need a string to represent it without loss of precision in a lot of javascript environment, I wish bigint in json was more widely supported
func IntNode(value int) *json.Node {
	return json.StringNode("", strconv.Itoa(value))
}

func MustInt(value *json.Node) int {
	i, err := strconv.Atoi(value.MustString())
	if err != nil {
		panic(err)
	}

	return i
}

func Uint32Node(value uint32) *json.Node {
	return json.StringNode("", strconv.FormatUint(uint64(value), 10))
}

func MustUint32(value *json.Node) uint32 {
	return uint32(MustInt(value))
}

func Int64Node(value int64) *json.Node {
	return json.StringNode("", strconv.FormatInt(value, 10))
}

func MustInt64(value *json.Node) int64 {
	return int64(MustInt(value))
}

func Uint64Node(value uint64) *json.Node {
	return json.StringNode("", strconv.FormatUint(value, 10))
}

func MustUint64(value *json.Node) uint64 {
	return uint64(MustInt(value)) // FIXME: full uint64 range support (currently limited to [-2^63, 2^63-1])
}

func AVLTreeNode(root *avl.Tree, transform func(elem interface{}) *json.Node) *json.Node {
	if root == nil {
		return EmptyObjectNode()
	}

	fields := make(map[string]*json.Node)
	root.Iterate("", "", func(key string, val interface{}) bool {
		fields[key] = transform(val)
		return false
	})

	return json.ObjectNode("", fields)
}

func AddressNode(addr std.Address) *json.Node {
	return json.StringNode("", addr.String())
}

func MustAddress(value *json.Node) std.Address {
	addr := std.Address(value.MustString())
	if !addr.IsValid() {
		panic("invalid address")
	}

	return addr
}

func AddressOrNameNode(aon users.AddressOrName) *json.Node {
	return json.StringNode("", string(aon))
}

func MustAddressOrName(value *json.Node) users.AddressOrName {
	aon := users.AddressOrName(value.MustString())
	if !aon.IsValid() {
		panic("invalid address or name")
	}

	return aon
}
