package params

import (
	"reflect"

	"github.com/gnolang/gno/tm2/pkg/amino"
	"github.com/gnolang/gno/tm2/pkg/errors"
	"github.com/gnolang/gno/tm2/pkg/std"
	sm "github.com/gnolang/gno/tm2/pkg/store"
)

// Returns list of kvpairs from param struct.
func encodeStructFields(prm any) (res []std.KVPair) {
	rvPrm := reflect.ValueOf(prm)
	tinfo, err := amino.GetTypeInfo(rvPrm.Type())
	if err != nil {
		panic(errors.Wrap(err, "Error reflecting on module param struct"))
	}
	fields := tinfo.Fields
	for i, field := range fields {
		rv := rvPrm.Field(i)
		name := field.JSONName
		value := amino.MustMarshalJSON(rv.Interface())
		res = append(res, std.KVPair{Key: []byte(name), Value: value})
	}
	return res
}

func findKV(kvz []std.KVPair, key string) (std.KVPair, bool) {
	for _, kv := range kvz {
		if string(kv.Key) == key {
			return kv, true
		}
	}
	return std.KVPair{}, false
}

// Reads list of kvpairs into param struct.
func decodeStructFields(prmPtr any, kvz []std.KVPair) {
	if reflect.TypeOf(prmPtr).Kind() != reflect.Pointer {
		panic("setStructFields expects module param struct pointer")
	}
	rvPrm := reflect.ValueOf(prmPtr).Elem()
	tinfo, err := amino.GetTypeInfo(rvPrm.Type())
	if err != nil {
		panic(errors.Wrap(err, "Error reflecting on module param struct"))
	}
	fields := tinfo.Fields
	for i, field := range fields {
		rv := rvPrm.Field(i)
		name := field.JSONName
		kv, ok := findKV(kvz, name)
		if !ok {
			continue
		}
		amino.MustUnmarshalJSON(kv.Value, rv.Addr().Interface())
	}
}

// IsParamReprEmpty reports whether the amino repr-string form of a Param
// is empty. Used by generated pbbindings.go in packages that reference
// params.Param (see tm2/pkg/amino/genproto). Mirrors the shape of
// std.IsCoinReprEmpty.
func IsParamReprEmpty(goor string) (empty bool) {
	empty = true
	if goor != "" {
		return false
	}
	return
}

// Gets list of kvpairs associated with param struct from store.
func getStructFieldsFromStore(prmPtr any, store sm.Store, key []byte) (res []std.KVPair) {
	if reflect.TypeOf(prmPtr).Kind() != reflect.Pointer {
		panic("setStructFields expects module param struct pointer")
	}
	rvPrm := reflect.ValueOf(prmPtr).Elem()
	tinfo, err := amino.GetTypeInfo(rvPrm.Type())
	if err != nil {
		panic(errors.Wrap(err, "Error reflecting on module param struct"))
	}
	fields := tinfo.Fields
	for _, field := range fields {
		name := field.JSONName
		value := store.Get(nil, []byte(string(key)+":"+name))
		if value == nil {
			continue
		}
		res = append(res, std.KVPair{Key: []byte(name), Value: value})
	}
	return res
}
