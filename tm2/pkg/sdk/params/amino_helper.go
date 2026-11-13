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
	for i, name := range fieldNames(rvPrm.Type()) {
		rv := rvPrm.Field(i)
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
	for i, name := range fieldNames(rvPrm.Type()) {
		rv := rvPrm.Field(i)
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
	for _, name := range fieldNames(rvPrm.Type()) {
		value := store.Get(gctx, []byte(string(key)+":"+name))
		if value == nil {
			continue
		}
		res = append(res, std.KVPair{Key: []byte(name), Value: value})
	}
	return res
}

// fieldNamesCache maps a param struct type to its fields' JSON names.
var fieldNamesCache sync.Map // reflect.Type -> []string

// fieldNames returns the JSON names of the fields amino encodes for rt, in
// amino.TypeInfo.Fields order, cached per type. The slice is shared.
func fieldNames(rt reflect.Type) []string {
	if names, ok := fieldNamesCache.Load(rt); ok {
		return names.([]string)
	}
	tinfo, err := amino.GetTypeInfo(rt)
	if err != nil {
		panic(errors.Wrap(err, "Error reflecting on module param struct"))
	}
	names := make([]string, len(tinfo.Fields))
	for i, field := range tinfo.Fields {
		names[i] = field.JSONName
	}
	cached, _ := fieldNamesCache.LoadOrStore(rt, names)
	return cached.([]string)
}
