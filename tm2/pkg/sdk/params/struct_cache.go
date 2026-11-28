package params

import (
	"bytes"
	"reflect"
	"sync"

	"github.com/gnolang/gno/tm2/pkg/amino"
	"github.com/gnolang/gno/tm2/pkg/std"
)

// structCache remembers, for each struct key, the field bytes last read from
// the store and the values they decoded to, so that GetStruct skips the amino
// JSON decode while the store still holds the same bytes. The store is read on
// every call and only the decode is saved: the cache trusts nothing but the
// bytes just read, so it is right for whatever store the context carries — a
// transaction's cache wrap, a write reverted with its transaction, a query at
// an old height — with no invalidation to get wrong. (auth's GetParams decoded
// its ten fields once a block from EndBlocker: 8–13% of a block sync's CPU on
// a mainnet node.)
type structCache struct {
	mu      sync.Mutex
	entries map[string]*structEntry // struct key, such as "auth:p" → its last decode
}

// structEntry is one decode: the bytes it was made from and the values it
// made. The values are never handed out, copies of them are.
type structEntry struct {
	typ    reflect.Type
	kvz    []std.KVPair
	fields []int // for each kv, the struct field it decoded into
	vals   []reflect.Value
}

func newStructCache() *structCache {
	return &structCache{entries: map[string]*structEntry{}}
}

// fill sets the fields of *prmPtr from the decode remembered for key, when
// kvz are the bytes that decode was made from, and reports whether it did.
// Fields without a kv are left as they are, as decodeStructFields leaves them.
func (c *structCache) fill(key string, prmPtr any, kvz []std.KVPair) bool {
	if c == nil {
		return false
	}
	rvPrm := reflect.ValueOf(prmPtr).Elem()
	c.mu.Lock()
	defer c.mu.Unlock()
	e := c.entries[key]
	if e == nil || e.typ != rvPrm.Type() || !sameKVs(e.kvz, kvz) {
		return false
	}
	copies := make([]reflect.Value, len(e.vals))
	for i, v := range e.vals {
		cp, ok := copyValue(v)
		if !ok {
			return false // remember keeps only values copyValue accepts
		}
		copies[i] = cp
	}
	for i, cp := range copies {
		rvPrm.Field(e.fields[i]).Set(cp)
	}
	return true
}

// remember keeps the decode of *prmPtr from kvz under key: for each kv, a
// copy of the field it was decoded into. A field of a kind copyValue cannot
// copy keeps the whole struct out of the cache, so it is decoded every time.
func (c *structCache) remember(key string, prmPtr any, kvz []std.KVPair) {
	if c == nil {
		return
	}
	rvPrm := reflect.ValueOf(prmPtr).Elem()
	tinfo, err := amino.GetTypeInfo(rvPrm.Type())
	if err != nil {
		return
	}
	e := &structEntry{typ: rvPrm.Type()}
	for i, field := range tinfo.Fields {
		kv, ok := findKV(kvz, field.JSONName)
		if !ok {
			continue
		}
		// field i of the struct: the one decodeStructFields writes for this kv
		rv := rvPrm.Field(i)
		v, ok := reflect.Value{}, rv.CanInterface() // an unexported field is never served from the cache
		if ok {
			v, ok = copyValue(rv)
		}
		if !ok {
			c.mu.Lock()
			delete(c.entries, key)
			c.mu.Unlock()
			return
		}
		e.kvz = append(e.kvz, std.KVPair{Key: bytes.Clone(kv.Key), Value: bytes.Clone(kv.Value)})
		e.fields = append(e.fields, i)
		e.vals = append(e.vals, v)
	}
	c.mu.Lock()
	c.entries[key] = e
	c.mu.Unlock()
}

// sameKVs reports whether a and b hold the same keys and values in the same
// order.
func sameKVs(a, b []std.KVPair) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if !bytes.Equal(a[i].Key, b[i].Key) || !bytes.Equal(a[i].Value, b[i].Value) {
			return false
		}
	}
	return true
}

// copyValue is a deep copy of v, for the kinds a param struct is made of:
// numbers, strings, bools, and arrays, slices and structs of those. It reports
// false for a kind it cannot copy safely — a pointer, map, interface, function
// or channel, or a struct with an unexported field.
func copyValue(v reflect.Value) (reflect.Value, bool) {
	if plainKind(v.Kind()) {
		return v, true // Set copies these by value
	}
	switch v.Kind() {
	case reflect.Array:
		if plainKind(v.Type().Elem().Kind()) {
			return v, true // Set copies an array by value
		}
		return copyItems(reflect.New(v.Type()).Elem(), v)
	case reflect.Slice:
		if v.IsNil() {
			return v, true
		}
		out := reflect.MakeSlice(v.Type(), v.Len(), v.Len())
		if plainKind(v.Type().Elem().Kind()) {
			reflect.Copy(out, v)
			return out, true
		}
		return copyItems(out, v)
	case reflect.Struct:
		out := reflect.New(v.Type()).Elem()
		for i := range v.NumField() {
			if !v.Type().Field(i).IsExported() {
				return reflect.Value{}, false
			}
			cp, ok := copyValue(v.Field(i))
			if !ok {
				return reflect.Value{}, false
			}
			out.Field(i).Set(cp)
		}
		return out, true
	default:
		return reflect.Value{}, false
	}
}

// copyItems copies the elements of v into out, an array or slice of the same
// type and length.
func copyItems(out, v reflect.Value) (reflect.Value, bool) {
	for i := range v.Len() {
		cp, ok := copyValue(v.Index(i))
		if !ok {
			return reflect.Value{}, false
		}
		out.Index(i).Set(cp)
	}
	return out, true
}

// plainKind reports whether values of kind k are copied whole by Set, with
// nothing shared.
func plainKind(k reflect.Kind) bool {
	switch k {
	case reflect.Bool, reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64,
		reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64, reflect.Uintptr,
		reflect.Float32, reflect.Float64, reflect.Complex64, reflect.Complex128, reflect.String:
		return true
	default:
		return false
	}
}
