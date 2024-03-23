package ujson

// This package strives to have the same behavior as json.Marshal but does not support all types and returns strings

import (
	"errors"
	"std"
	"strconv"
	"strings"
	"time"

	"gno.land/p/demo/avl"
	"gno.land/p/demo/users"
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
		t := p.(avl.Tree)
		return FormatAVLTree(&t)
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
	case time.Time:
		return FormatTime(p.(time.Time))
	case time.Duration:
		return FormatInt64(int64(p.(time.Duration)))
	case users.AddressOrName:
		return FormatString(string(p.(users.AddressOrName)))
	default:
		return "null"
	}
}

// loosely ported from https://cs.opensource.google/go/go/+/master:src/time/time.go;l=1357?q=appendStrictRFC3339&ss=go%2Fgo
func FormatTime(t time.Time) string {
	s := t.Format(time.RFC3339Nano)
	b := []byte(s)

	// Not all valid Go timestamps can be serialized as valid RFC 3339.
	// Explicitly check for these edge cases.
	// See https://go.dev/issue/4556 and https://go.dev/issue/54580.
	n0 := 0
	num2 := func(b []byte) byte { return 10*(b[0]-'0') + (b[1] - '0') }
	switch {
	case b[n0+len("9999")] != '-': // year must be exactly 4 digits wide
		panic(errors.New("year outside of range [0,9999]"))
	case b[len(b)-1] != 'Z':
		c := b[len(b)-len("Z07:00")]
		if ('0' <= c && c <= '9') || num2(b[len(b)-len("07:00"):]) >= 24 {
			panic(errors.New("timezone hour outside of range [0,23]"))
		}
	}
	return FormatString(string(b))
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

func FormatUnionMember(name string, val interface{}, raw bool) string {
	return FormatObject([]FormatKV{
		{Key: name, Value: val, Raw: raw},
	})
}
