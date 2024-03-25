package vm

import (
	"encoding/json"
	"reflect"

	gno "github.com/gnolang/gno/gnovm/pkg/gnolang"
)

// XXX: Gas Consumption - Measure and add gas consumption for JSON marshal/unmarshal operations.

func UnmarshalTypedValueJSON(alloc *gno.Allocator, store gno.Store, b []byte, t gno.Type) (gno.TypedValue, error) {
	tv := gno.TypedValue{T: t}
	gvalue := gno.Gno2GoValue(&tv, reflect.Value{})
	v := reflect.New(gvalue.Type())
	if err := json.Unmarshal(b, v.Interface()); err != nil {
		return gno.TypedValue{}, err
	}

	return gno.Go2GnoValue(alloc, store, v.Elem()), nil
}

func UnmarshalNativeValueJSON(alloc *gno.Allocator, b []byte, t gno.Type) (gno.TypedValue, error) {
	tv := gno.TypedValue{T: t}
	gvalue := gno.Gno2GoValue(&tv, reflect.Value{})
	v := reflect.New(gvalue.Type())
	if err := json.Unmarshal(b, v.Interface()); err != nil {
		return gno.TypedValue{}, err
	}

	return gno.Go2GnoNativeValue(alloc, v.Elem()), nil
}

func MarshalTypedValueJSON(tv *gno.TypedValue) ([]byte, error) {
	rv := gno.Gno2GoValue(tv, reflect.Value{})
	return json.Marshal(rv.Interface())
}
