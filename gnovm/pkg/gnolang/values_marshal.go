package gnolang

import (
	"reflect"
)

// XXX: PoC, API will most likely to be change

type AminoTypedValue struct {
	TypedValue TypedValue

	Store     Store
	Allocator *Allocator
}

// XXX: tv.TypedValue.T need to be filled in order to be able guess the type
func (tv AminoTypedValue) TypeDesc() reflect.Type {
	typ := baseOf(tv.TypedValue.T)
	return gno2GoType(typ)
}

func (tv AminoTypedValue) MarshalAmino() (interface{}, error) {
	ret := Gno2GoValue(&tv.TypedValue, reflect.Value{})
	return ret.Interface(), nil
}

func (tv *AminoTypedValue) UnmarshalAmino(i interface{}) error {
	rv := reflect.ValueOf(i)
	tv.TypedValue = Go2GnoValue(tv.Allocator, tv.Store, rv)
	return nil
}
