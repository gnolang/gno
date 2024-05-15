package gnoamino

import (
	"reflect"

	"github.com/gnolang/gno/gnovm/pkg/gnolang"
)

// XXX: PoC - API will most likely change

type Marshaler interface {
	TypeDesc() reflect.Type
	UnmarshalAmino(interface{}) error
	MarshalAmino() (interface{}, error)
	GnoValue() *gnolang.TypedValue
}

type TypedValueMarshaler struct {
	Allocator *gnolang.Allocator
	Store     gnolang.Store
}

func NewTypedValueMarshaler(alloc *gnolang.Allocator /* , store gnolang.Store */) *TypedValueMarshaler {
	return &TypedValueMarshaler{
		Allocator: alloc,
		// Store:     store,
	}
}

func (tm *TypedValueMarshaler) Wrap(tv *gnolang.TypedValue) Marshaler {
	return &wrapperTypedValue{
		TypedValue: tv,
		Allocator:  tm.Allocator,
		// Store:      tm.Store,
	}
}

func (tm *TypedValueMarshaler) From(t gnolang.Type) Marshaler {
	tv := &gnolang.TypedValue{T: t}
	return &wrapperTypedValue{
		TypedValue: tv,
		Allocator:  tm.Allocator,
		// Store:      tm.Store,
	}
}
