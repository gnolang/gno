package amino

import (
	"fmt"
	"reflect"
	"unicode"
)

// ----------------------------------------
// Constants

var errorType = reflect.TypeOf(new(error)).Elem()

// ----------------------------------------
// encode: see binary-encode.go and json-encode.go
// decode: see binary-decode.go and json-decode.go

// ----------------------------------------
// Misc.

// CONTRACT: by the time this is called, len(bz) >= _n
// Returns true so you can write one-liners.
func slide(bz *[]byte, n *int, _n int) bool {
	if bz != nil {
		if _n < 0 || _n > len(*bz) {
			panic(fmt.Sprintf("impossible slide: len:%v _n:%v", len(*bz), _n))
		}
		*bz = (*bz)[_n:]
	}
	if n != nil {
		*n += _n
	}
	return true
}

// maybe dereference if pointer.
// drv: the final non-pointer value (which may be invalid).
// isPtr: whether rv.Kind() == reflect.Ptr.
// isNilPtr: whether a nil pointer at any level.
func maybeDerefValue(rv reflect.Value) (drv reflect.Value, rvIsPtr bool, rvIsNilPtr bool) {
	if rv.Kind() == reflect.Ptr {
		rvIsPtr = true
		if rv.IsNil() {
			rvIsNilPtr = true
			return
		}
		rv = rv.Elem()
	}
	drv = rv
	return
}

// Dereference-and-construct pointers.
func maybeDerefAndConstruct(rv reflect.Value) reflect.Value {
	if rv.Kind() == reflect.Ptr {
		if rv.IsNil() {
			newPtr := reflect.New(rv.Type().Elem())
			rv.Set(newPtr)
		}
		rv = rv.Elem()
	}
	if rv.Kind() == reflect.Ptr {
		panic("unexpected pointer pointer")
	}
	return rv
}

// Returns isDefaultValue=true iff is zero and isn't a struct.
// NOTE: Also works for Maps, Chans, and Funcs, though they are not
// otherwise supported by Amino.  For future?
func isNonstructDefaultValue(rv reflect.Value) (isDefault bool) {
	// time.Duration is a special case,
	// it is considered a struct for encoding purposes.
	switch rv.Type() {
	case durationType:
		return false
	}
	// general cae
	switch rv.Kind() {
	case reflect.Ptr:
		if rv.IsNil() {
			return true
		} else {
			erv := rv.Elem()
			return isNonstructDefaultValue(erv)
		}
	case reflect.Bool:
		return rv.Bool() == false //nolint:staticcheck
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return rv.Int() == 0
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		return rv.Uint() == 0
	case reflect.String:
		return rv.Len() == 0
	case reflect.Chan, reflect.Map, reflect.Slice:
		return rv.IsNil() || rv.Len() == 0
	case reflect.Func, reflect.Interface:
		return rv.IsNil()
	case reflect.Struct:
		return false
	default:
		return false
	}
}

// Returns the default value of a type.  For a time type or a
// pointer(s) to time, the default value is not zero (or nil), but the
// time value of 1970.
//
// The default value of a struct pointer is nil, while the default value of
// other pointers is not nil.  This is due to a proto3 wart, e.g. while there
// is a way to distinguish between a nil struct/message vs an empty one (via
// its presence or absence in an outer struct), there is no such way to
// distinguish between nil bytes/lists and empty bytes/lists, are they are all
// absent in binary encoding.
func defaultValue(rt reflect.Type) (rv reflect.Value) {
	switch rt.Kind() {
	case reflect.Ptr:
		// Dereference all the way and see if it's a time type.
		ert := rt.Elem()
		if ert.Kind() == reflect.Ptr {
			panic("nested pointers not allowed")
		}
		if ert == timeType {
			// Start from the top and construct pointers as needed.
			rv = reflect.New(rt.Elem())
			// Set to 1970, the whole point of this function.
			rv.Elem().Set(reflect.ValueOf(emptyTime))
			return rv
		} else if ert.Kind() == reflect.Struct {
			rv = reflect.Zero(rt)
			return rv
		} else {
			rv = reflect.New(rt.Elem())
			return rv
		}
	case reflect.Struct:
		if rt == timeType {
			// Set to 1970, the whole point of this function.
			rv = reflect.New(rt).Elem()
			rv.Set(reflect.ValueOf(emptyTime))
			return rv
		} else {
			return reflect.Zero(rt)
		}
	}

	// Just return the default Go zero object.
	// Return an empty struct.
	return reflect.Zero(rt)
}

// NOTE: Also works for Maps and Chans, though they are not
// otherwise supported by Amino.  For future?
func isNil(rv reflect.Value) bool {
	switch rv.Kind() {
	case reflect.Interface, reflect.Chan, reflect.Func, reflect.Map, reflect.Ptr, reflect.Slice:
		return rv.IsNil()
	default:
		return false
	}
}

// constructConcreteType creates the concrete value as
// well as the corresponding settable value for it.
// Return irvSet which should be set on caller's interface rv.
func constructConcreteType(cinfo *TypeInfo) (crv, irvSet reflect.Value) {
	// Construct new concrete type.
	if cinfo.PointerPreferred {
		cPtrRv := reflect.New(cinfo.Type)
		crv = cPtrRv.Elem()
		irvSet = cPtrRv
	} else {
		crv = reflect.New(cinfo.Type).Elem()
		irvSet = crv
	}
	return
}

func toReprObject(rv reflect.Value) (rrv reflect.Value, err error) {
	var mwrm reflect.Value
	if rv.CanAddr() {
		mwrm = rv.Addr().MethodByName("MarshalAmino")
	} else {
		mwrm = rv.MethodByName("MarshalAmino")
	}
	mwouts := mwrm.Call(nil)
	if !mwouts[1].IsNil() {
		erri := mwouts[1].Interface()
		if erri != nil {
			err = erri.(error)
			return rrv, err
		}
	}
	rrv = mwouts[0]
	return
}

func isExported(field reflect.StructField) bool {
	// Test 1:
	if field.PkgPath != "" {
		return false
	}
	// Test 2:
	var first rune
	for _, c := range field.Name {
		first = c
		break
	}
	// TODO: JAE: I'm not sure that the unicode spec
	// is the correct spec to use, so this might be wrong.

	return unicode.IsUpper(first)
}

func marshalAminoReprType(rm reflect.Method) (rrt reflect.Type) {
	// Verify form of this method.
	if rm.Type.NumIn() != 1 {
		panic(fmt.Sprintf("MarshalAmino should have 1 input parameters (including receiver); got %v", rm.Type))
	}
	if rm.Type.NumOut() != 2 {
		panic(fmt.Sprintf("MarshalAmino should have 2 output parameters; got %v", rm.Type))
	}
	if out := rm.Type.Out(1); out != errorType {
		panic(fmt.Sprintf("MarshalAmino should have second output parameter of error type, got %v", out))
	}
	rrt = rm.Type.Out(0)
	if rrt.Kind() == reflect.Ptr {
		panic(fmt.Sprintf("Representative objects cannot be pointers; got %v", rrt))
	}
	return
}

func unmarshalAminoReprType(rm reflect.Method) (rrt reflect.Type) {
	// Verify form of this method.
	if rm.Type.NumIn() != 2 {
		panic(fmt.Sprintf("UnmarshalAmino should have 2 input parameters (including receiver); got %v", rm.Type))
	}
	if in1 := rm.Type.In(0); in1.Kind() != reflect.Ptr {
		panic(fmt.Sprintf("UnmarshalAmino first input parameter should be pointer type but got %v", in1))
	}
	if rm.Type.NumOut() != 1 {
		panic(fmt.Sprintf("UnmarshalAmino should have 1 output parameters; got %v", rm.Type))
	}
	if out := rm.Type.Out(0); out != errorType {
		panic(fmt.Sprintf("UnmarshalAmino should have first output parameter of error type, got %v", out))
	}
	rrt = rm.Type.In(1)
	if rrt.Kind() == reflect.Ptr {
		panic(fmt.Sprintf("Representative objects cannot be pointers; got %v", rrt))
	}
	return
}

// NOTE: do not change this definition.
// It is also defined for genproto.
func isListType(rt reflect.Type) bool {
	return rt.Kind() == reflect.Slice ||
		rt.Kind() == reflect.Array
}
