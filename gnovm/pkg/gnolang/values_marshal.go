package gnolang

import (
	"reflect"
)

type AminoTypedValue struct {
	TypedValue TypedValue

	Store     Store
	Allocator *Allocator
}

func (tv AminoTypedValue) Type() Type {
	return tv.TypedValue.T
}

func (tv AminoTypedValue) Value() Value {
	return tv.TypedValue.V
}

func (tv *AminoTypedValue) SetType(t Type) {
	tv.TypedValue.T = t
}

func (tv AminoTypedValue) TypeAmino() (reflect.Type, error) {
	typ := baseOf(tv.TypedValue.T)
	return gno2GoType(typ), nil
}

func (tv AminoTypedValue) MarshalAmino() (interface{}, error) {
	ret := Gno2GoValue(&tv.TypedValue, reflect.Value{})
	return ret.Interface(), nil
}

func (tv *AminoTypedValue) UnmarshalAmino(i interface{}) error {
	rv := reflect.ValueOf(i)
	tv.TypedValue = Go2GnoValue(tv.Allocator, tv.Store, rv)
	// fmt.Printf("the end: %v\n", tv2.String())
	return nil
}
