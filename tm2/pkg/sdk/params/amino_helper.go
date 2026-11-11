package params

import (
	"encoding/json"
	"fmt"
	"reflect"
	"sync"

	"github.com/gnolang/gno/tm2/pkg/amino"
	"github.com/gnolang/gno/tm2/pkg/errors"
	"github.com/gnolang/gno/tm2/pkg/std"
	sm "github.com/gnolang/gno/tm2/pkg/store"
)

// Returns list of kvpairs from param struct.
func encodeStructFields(prm any) (res []std.KVPair) {
	rvPrm := reflect.ValueOf(prm)
	tinfo, err := structTypeInfo(rvPrm.Type())
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
	tinfo, err := structTypeInfo(rvPrm.Type())
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
		if err := decodeField(rv, kv.Value); err != nil {
			panic(fmt.Errorf("failed to decode param field %q: %w", name, err))
		}
	}
}

func decodeField(rv reflect.Value, value []byte) error {
	tmp := reflect.New(rv.Type())
	err := amino.UnmarshalJSON(value, tmp.Interface())
	if err == nil {
		rv.Set(tmp.Elem())
		return nil
	}
	if _, ok := rv.Addr().Interface().(json.Unmarshaler); !ok {
		return err
	}

	tmp = reflect.New(rv.Type())
	if err := json.Unmarshal(value, tmp.Interface()); err != nil {
		return err
	}
	rv.Set(tmp.Elem())
	return nil
}

// Gets list of kvpairs associated with param struct from store.
func getStructFieldsFromStore(gctx *sm.GasContext, prmPtr any, store sm.Store, key []byte) (res []std.KVPair) {
	if reflect.TypeOf(prmPtr).Kind() != reflect.Pointer {
		panic("setStructFields expects module param struct pointer")
	}
	rvPrm := reflect.ValueOf(prmPtr).Elem()
	tinfo, err := structTypeInfo(rvPrm.Type())
	if err != nil {
		panic(errors.Wrap(err, "Error reflecting on module param struct"))
	}
	fields := tinfo.Fields
	for _, field := range fields {
		name := field.JSONName
		value := store.Get(gctx, []byte(string(key)+":"+name))
		if value == nil {
			continue
		}
		res = append(res, std.KVPair{Key: []byte(name), Value: value})
	}
	return res
}

// typeInfos memoizes structTypeInfo, keyed by reflect.Type.
var typeInfos sync.Map // reflect.Type -> *amino.TypeInfo

// structTypeInfo returns the amino type info of a param struct type.
//
// amino.GetTypeInfo builds a fresh codec, well-known types included, on every
// call, and every GetStruct/SetStruct needs the type info, so ante handlers
// pay for it on every transaction. The result depends only on rt (the
// throwaway codec has nothing registered), so it is safe to compute once per
// type. The key set is the set of param struct types the binary passes to
// GetStruct/SetStruct, which is fixed at compile time. Errors are not cached.
// Callers must treat the returned *amino.TypeInfo as read-only: it is shared.
func structTypeInfo(rt reflect.Type) (*amino.TypeInfo, error) {
	if ti, ok := typeInfos.Load(rt); ok {
		return ti.(*amino.TypeInfo), nil
	}
	ti, err := amino.GetTypeInfo(rt)
	if err != nil {
		return nil, err
	}
	actual, _ := typeInfos.LoadOrStore(rt, ti)
	return actual.(*amino.TypeInfo), nil
}
