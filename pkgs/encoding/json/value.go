package json

import (
	"reflect"
	"unicode"

	gno "github.com/gnolang/gno/pkgs/gnolang"
)

type Value interface {
	Kind() gno.Kind

	// Array & Slice & Map
	Len() int
	Index(i int) Value

	// Struct
	StructFields() []StructField

	// Elem returns the value that the interface v contains or that the pointer v points to.
	Elem() Value

	String() string
	Bool() bool
	Int() int64
	Uint() uint64

	IsNil() bool
	IsZero() bool
}

type gnoValue struct {
	v gno.TypedValue
	s gno.Store
}

type nativeValue struct {
	v reflect.Value
}

func newValue(tv gno.TypedValue, s gno.Store) Value {
	if nv, ok := tv.V.(*gno.NativeValue); ok {
		return nativeValue{nv.Value}
	}
	return gnoValue{tv, s}
}

func (gv gnoValue) Kind() gno.Kind {
	if gv.v.T == nil {
		return gno.InvalidKind
	}
	return gv.v.T.Kind()
}

func (gv gnoValue) Len() int {
	return gv.v.GetLength()
}

func (gv gnoValue) Index(i int) Value {
	switch v := gv.v.V.(type) {
	case *gno.ArrayValue:
		return gnoValue{v.List[i], gv.s}
	case *gno.SliceValue:
		sv := v.GetBase(gv.s)
		return gnoValue{sv.List[i], gv.s}
	}
	// TODO panic
	panic("should not happen")
}

func (gv gnoValue) String() string {
	return gv.v.GetString()
}

func (gv gnoValue) Int() int64 {
	return gv.v.GetInt64()
}

func (gv gnoValue) Uint() uint64 {
	return gv.v.GetUint64()
}

func (gv gnoValue) Bool() bool {
	return gv.v.GetBool()
}

func (gv gnoValue) IsNil() bool {
	return gv.v.IsUndefined()
}

func (gv gnoValue) IsZero() bool {
	// todo
	return false
}

func (gv gnoValue) StructFields() []StructField {
	if gv.v.T.Kind() != gno.StructKind {
		panic("not a struct")
	}

	var stt *gno.StructType
	if dt, ok := gv.v.T.(*gno.DeclaredType); ok {
		stt = dt.Base.(*gno.StructType)
	} else {
		stt = gv.v.T.(*gno.StructType)
	}

	stv := gv.v.V.(*gno.StructValue)

	var fields []StructField

	for i, ft := range stt.Fields {
		fname := string(ft.Name)
		if fname == "" || !isFirstCharUpper(fname) {
			continue
		}
		fields = append(fields, gnoStructField{
			fieldType: ft,
			value:     newValue(stv.Fields[i], gv.s),
		})
	}
	return fields
}

func (gv gnoValue) Elem() Value {
	//todo: check kind is interface or pointer
	if gv.v.T.Kind() == gno.PointerKind {
		pv := gv.v.V.(gno.PointerValue)
		return newValue(pv.Deref(), gv.s)
	}
	return nil
}

func (nv nativeValue) Kind() gno.Kind {
	switch nv.v.Kind() {
	case reflect.Bool:
		return gno.BoolKind
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return gno.IntKind
	case reflect.String:
		return gno.StringKind
	case reflect.Array:
		return gno.ArrayKind
	case reflect.Slice:
		return gno.SliceKind
	case reflect.Struct:
		return gno.StructKind
	case reflect.Pointer:
		return gno.PointerKind
	case reflect.Interface:
		return gno.InterfaceKind
	}
	return gno.InvalidKind
}

func (nv nativeValue) Len() int {
	return nv.v.Len()
}

func (nv nativeValue) Index(i int) Value {
	return nativeValue{nv.v.Index(i)}
}

func (nv nativeValue) String() string {
	return nv.v.String()
}

func (nv nativeValue) Int() int64 {
	return nv.v.Int()
}

func (nv nativeValue) Uint() uint64 {
	return nv.v.Uint()
}

func (nv nativeValue) Bool() bool {
	return nv.v.Bool()
}

func (nv nativeValue) IsNil() bool {
	return nv.v.IsNil()
}

func (nv nativeValue) IsZero() bool {
	return nv.v.IsZero()
}

//TODO
func (nv nativeValue) StructFields() []StructField {
	var fields []StructField
	for i := 0; i < nv.v.NumField(); i++ {
		fname := nv.v.Type().Field(i).Name
		if fname == "" || !isFirstCharUpper(fname) {
			continue
		}
		fields = append(fields, nativeStructField{
			field: nv.v.Type().Field(i),
			value: nativeValue{nv.v.Field(i)},
		})
	}
	return fields
}

func (nv nativeValue) Elem() Value {
	return nativeValue{nv.v.Elem()}
}

func isFirstCharUpper(s string) bool {
	firstChar := []rune(s)[0]
	return unicode.IsUpper(firstChar)
}
