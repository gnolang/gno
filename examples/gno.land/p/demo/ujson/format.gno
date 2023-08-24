package ujson

// This package strives to have the same behavior as json.Marshal but does not support all types and returns strings

import (
	"std"
	"strconv"
	"strings"

	"gno.land/p/demo/avl"
)

type JSONAble interface {
	ToJSON() string
}

type FormatKV struct {
	Key   string
	Value interface{}
	Raw   bool
}

// does not work for slices, use FormatSlice instead
func FormatAny(p interface{}) string {
	switch p.(type) {
	case std.Address:
		return FormatString(string(p.(std.Address)))
	case *avl.Tree:
		return FormatAVLTree(p.(*avl.Tree))
	case avl.Tree:
		return FormatAVLTree(&p.(avl.Tree))
	case JSONAble:
		return p.(JSONAble).ToJSON()
	case string:
		return FormatString(p.(string))
	case uint64:
		return FormatUint64(p.(uint64))
	case uint32:
		return FormatUint64(uint64(p.(uint32)))
	case uint:
		return FormatUint64(uint64(p.(uint)))
	case int64:
		return FormatInt64(p.(int64))
	case int32:
		return FormatInt64(int64(p.(int32)))
	case int:
		return FormatInt64(int64(p.(int)))
	case float32:
		panic("float32 not implemented")
	case float64:
		panic("float64 not implemented")
	case bool:
		return FormatBool(p.(bool))
	default:
		return "null"
	}
}

func FormatUint64(i uint64) string {
	return strconv.FormatUint(i, 10)
}

func FormatInt64(i int64) string {
	return strconv.FormatInt(i, 10)
}

func FormatSlice(s []interface{}) string {
	elems := make([]string, len(s))
	for i, elem := range s {
		elems[i] = FormatAny(elem)
	}
	return "[" + strings.Join(elems, ",") + "]"
}

func FormatObject(kv []FormatKV) string {
	elems := make([]string, len(kv))
	i := 0
	for _, elem := range kv {
		var val string
		if elem.Raw {
			val = elem.Value.(string)
		} else {
			val = FormatAny(elem.Value)
		}
		elems[i] = FormatString(elem.Key) + ":" + val
		i++
	}
	return "{" + strings.Join(elems, ",") + "}"
}

func FormatBool(b bool) string {
	if b {
		return "true"
	}
	return "false"
}

func FormatAVLTree(t *avl.Tree) string {
	if t == nil {
		return "{}"
	}
	kv := make([]FormatKV, 0, t.Size())
	t.Iterate("", "", func(key string, value interface{}) bool {
		kv = append(kv, FormatKV{key, value, false})
		return false
	})
	return FormatObject(kv)
}
