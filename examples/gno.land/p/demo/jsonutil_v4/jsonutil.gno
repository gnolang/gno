package jsonutil

// This package strives to have the same behavior as json.Marshal but has no support for nested slices and returns strings

import (
	"std"
	"strconv"
	"strings"
	"unicode/utf8"

	"gno.land/p/demo/avl"
	"gno.land/p/demo/ufmt"
)

type JSONAble interface {
	ToJSON() string
}

type KeyValue struct {
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
	case bool:
		return FormatBool(p.(bool))
	default:
		return "null"
	}
}

// Ported from https://cs.opensource.google/go/go/+/refs/tags/go1.20.6:src/encoding/json/encode.go
func FormatString(s string) string {
	const escapeHTML = true
	e := `"` // e.WriteByte('"')
	start := 0
	for i := 0; i < len(s); {
		if b := s[i]; b < utf8.RuneSelf {
			if htmlSafeSet[b] || (!escapeHTML && safeSet[b]) {
				i++
				continue
			}
			if start < i {
				e += s[start:i] // e.WriteString(s[start:i])
			}
			e += "\\" // e.WriteByte('\\')
			switch b {
			case '\\', '"':
				e += string(b) // e.WriteByte(b)
			case '\n':
				e += "n" // e.WriteByte('n')
			case '\r':
				e += "r" // e.WriteByte('r')
			case '\t':
				e += "t" // e.WriteByte('t')
			default:
				// This encodes bytes < 0x20 except for \t, \n and \r.
				// If escapeHTML is set, it also escapes <, >, and &
				// because they can lead to security holes when
				// user-controlled strings are rendered into JSON
				// and served to some browsers.
				e += `u00`              // e.WriteString(`u00`)
				e += string(hex[b>>4])  // e.WriteByte(hex[b>>4])
				e += string(hex[b&0xF]) // e.WriteByte(hex[b&0xF])
			}
			i++
			start = i
			continue
		}
		c, size := utf8.DecodeRuneInString(s[i:])
		if c == utf8.RuneError && size == 1 {
			if start < i {
				e += s[start:i] // e.WriteString(s[start:i])
			}
			e += `\ufffd` // e.WriteString(`\ufffd`)
			i += size
			start = i
			continue
		}
		// U+2028 is LINE SEPARATOR.
		// U+2029 is PARAGRAPH SEPARATOR.
		// They are both technically valid characters in JSON strings,
		// but don't work in JSONP, which has to be evaluated as JavaScript,
		// and can lead to security holes there. It is valid JSON to
		// escape them, so we do so unconditionally.
		// See http://timelessrepo.com/json-isnt-a-javascript-subset for discussion.
		if c == '\u2028' || c == '\u2029' {
			if start < i {
				e += s[start:i] // e.WriteString(s[start:i])
			}
			e += `\u202`            // e.WriteString(`\u202`)
			e += string(hex[c&0xF]) // e.WriteByte(hex[c&0xF])
			i += size
			start = i
			continue
		}
		i += size
	}
	if start < len(s) {
		e += s[start:] // e.WriteString(s[start:])
	}
	e += `"` // e.WriteByte('"')
	return e
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

func FormatObject(kv []KeyValue) string {
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
	kv := make([]KeyValue, 0, t.Size())
	t.Iterate("", "", func(key string, value interface{}) bool {
		kv = append(kv, KeyValue{key, value, false})
		return false
	})
	return FormatObject(kv)
}
