package gnolang

import (
	"fmt"
	"math"
	"reflect"

	"github.com/gnolang/gno/gnovm/pkg/gnolang/internal/softfloat"
)

// This file contains functions to convert values and types from Go to Gno, and
// vice versa. It supports a small subset of types, which are used in native
// functions, linked using misc/genstd.
//
// Much of this code is eventually meant to be converted into code generation
// performed by misc/genstd.

// ----------------------------------------
// Go to Gno conversion

// for native type unary/binary expression conversion.
// also used for untyped gno -> native conversion intermediary.
// XXX support unary conversions as we did for binary.
func go2GnoType(rt reflect.Type) Type {
	switch rk := rt.Kind(); rk {
	case reflect.Bool:
		return BoolType
	case reflect.String:
		return StringType
	case reflect.Int:
		return IntType
	case reflect.Int8:
		return Int8Type
	case reflect.Int16:
		return Int16Type
	case reflect.Int32:
		return Int32Type
	case reflect.Int64:
		return Int64Type
	case reflect.Uint:
		return UintType
	case reflect.Uint8:
		return Uint8Type
	case reflect.Uint16:
		return Uint16Type
	case reflect.Uint32:
		return Uint32Type
	case reflect.Uint64:
		return Uint64Type
	case reflect.Float32:
		return Float32Type
	case reflect.Float64:
		return Float64Type
	case reflect.Array:
		return &ArrayType{
			Elt: go2GnoType(rt.Elem()),
			Len: rt.Len(),
			Vrd: false,
		}
	case reflect.Slice:
		return &SliceType{
			Elt: go2GnoType(rt.Elem()),
			Vrd: false,
		}
	case reflect.Ptr:
		return &PointerType{
			Elt: go2GnoType(rt.Elem()), // recursive
		}
	default:
		panic(fmt.Sprintf(
			"unexpected type %v", rt))
	}
}

// NOTE: used by vm module.  Recursively converts.
// If recursive is false, this function is like go2GnoValue() but less lazy
// (but still not recursive/eager). When recursive is false, it is for
// converting Go types to Gno types upon an explicit conversion (via
// ConvertTo).  Panics on unexported/private fields. Some types that cannot be
// converted remain native. Unlike go2GnoValue(), rv must be valid.
func Go2GnoValue(alloc *Allocator, store Store, rv reflect.Value) (tv TypedValue) {
	if debug {
		if !rv.IsValid() {
			panic("go2GnoValue2() requires valid rv")
		}
	}
	tv.T = go2GnoType(rv.Type())
	switch rk := rv.Kind(); rk {
	case reflect.Bool:
		tv.SetBool(rv.Bool())
	case reflect.String:
		tv.V = alloc.NewString(rv.String())
	case reflect.Int:
		tv.SetInt(rv.Int())
	case reflect.Int8:
		tv.SetInt8(int8(rv.Int()))
	case reflect.Int16:
		tv.SetInt16(int16(rv.Int()))
	case reflect.Int32:
		tv.SetInt32(int32(rv.Int()))
	case reflect.Int64:
		tv.SetInt64(rv.Int())
	case reflect.Uint:
		tv.SetUint(rv.Uint())
	case reflect.Uint8:
		tv.SetUint8(uint8(rv.Uint()))
	case reflect.Uint16:
		tv.SetUint16(uint16(rv.Uint()))
	case reflect.Uint32:
		tv.SetUint32(uint32(rv.Uint()))
	case reflect.Uint64:
		tv.SetUint64(rv.Uint())
	case reflect.Float32:
		tv.SetFloat32(softfloat.F64to32(math.Float64bits(rv.Float())))
	case reflect.Float64:
		tv.SetFloat64(math.Float64bits(rv.Float()))
	case reflect.Array:
		rvl := rv.Len()
		if rv.Type().Elem().Kind() == reflect.Uint8 {
			av := alloc.NewDataArray(rvl)
			data := av.Data
			reflect.Copy(reflect.ValueOf(data), rv)
			tv.V = av
		} else {
			av := alloc.NewListArray(rvl)
			list := av.List
			for i := range rvl {
				list[i] = Go2GnoValue(alloc, store, rv.Index(i))
			}
			tv.V = av
		}
	case reflect.Slice:
		rvl := rv.Len()
		rvc := rv.Cap()

		baseArray := alloc.NewListArray2(rvl, rvc)
		list := baseArray.List
		for i := range rvl {
			list[i] = Go2GnoValue(alloc, store, rv.Index(i))
		}
		tv.V = alloc.NewSlice(baseArray, 0, rvl, rvc)
	case reflect.Ptr:
		val := Go2GnoValue(alloc, store, rv.Elem())
		tv.V = PointerValue{TV: &val} // heap alloc
	default:
		panic("not yet implemented")
	}
	return
}

// ----------------------------------------
// Gno to Go conversion

// NOTE: Recursive types are not supported, as named types are not
// supported.  See https://github.com/golang/go/issues/20013 and
// https://github.com/golang/go/issues/39717.
func gno2GoType(t Type) reflect.Type {
	switch ct := baseOf(t).(type) {
	case PrimitiveType:
		switch ct {
		case BoolType, UntypedBoolType:
			return reflect.TypeOf(false)
		case StringType, UntypedStringType:
			return reflect.TypeOf("")
		case IntType:
			return reflect.TypeOf(int(0))
		case Int8Type:
			return reflect.TypeOf(int8(0))
		case Int16Type:
			return reflect.TypeOf(int16(0))
		case Int32Type, UntypedRuneType:
			return reflect.TypeOf(int32(0))
		case Int64Type:
			return reflect.TypeOf(int64(0))
		case UintType:
			return reflect.TypeOf(uint(0))
		case Uint8Type:
			return reflect.TypeOf(uint8(0))
		case Uint16Type:
			return reflect.TypeOf(uint16(0))
		case Uint32Type:
			return reflect.TypeOf(uint32(0))
		case Uint64Type:
			return reflect.TypeOf(uint64(0))
		case Float32Type:
			return reflect.TypeOf(float32(0))
		case Float64Type:
			return reflect.TypeOf(float64(0))
		case UntypedBigintType:
			panic("not yet implemented")
		case UntypedBigdecType:
			panic("not yet implemented")
		default:
			panic("should not happen")
		}
	case *PointerType:
		et := gno2GoType(ct.Elem())
		return reflect.PointerTo(et)
	case *ArrayType:
		ne := ct.Len
		et := gno2GoType(ct.Elem())
		return reflect.ArrayOf(ne, et)
	case *SliceType:
		et := gno2GoType(ct.Elem())
		return reflect.SliceOf(et)
	default:
		panic(fmt.Sprintf("unexpected type %v with base %v", t, baseOf(t)))
	}
}

// rv must be addressable, or zero (invalid) (say if tv is referred to from a
// gno.PointerValue). In the latter case, an addressable one will be
// constructed and returned, otherwise returns rv.  if tv is undefined, rv must
// be valid.
//
// NOTE It doesn't make sense to add a 'store' argument here to support lazy
// loading (e.g. from native function bindings for SDKParams) because it
// doesn't really make sense in the general case, and there is a FillValue()
// function available for eager fetching of ref values.
func Gno2GoValue(tv *TypedValue, rv reflect.Value) (ret reflect.Value) {
	if tv.IsUndefined() {
		if debug {
			if !rv.IsValid() {
				panic("unexpected undefined gno value")
			}
		}
		return rv
	}
	var rt reflect.Type
	bt := baseOf(tv.T)
	if !rv.IsValid() {
		rt = gno2GoType(bt)
		rv = reflect.New(rt).Elem()
		ret = rv
	} else if rv.Kind() == reflect.Interface {
		if debug {
			if !rv.IsZero() {
				panic("should not happen")
			}
		}
		rt = gno2GoType(bt)
		rv1 := rv
		rv2 := reflect.New(rt).Elem()
		rv = rv2       // swaparoo
		defer func() { // TODO: improve?
			rv1.Set(rv2)
			ret = rv
		}()
	} else {
		ret = rv
		rt = rv.Type()
	}

	// Only need to support the types supported by native bindings:
	// see misc/genstd/mapping.go
	switch ct := bt.(type) {
	case PrimitiveType:
		switch ct {
		case BoolType, UntypedBoolType:
			rv.SetBool(tv.GetBool())
		case StringType, UntypedStringType:
			rv.SetString(tv.GetString())
		case IntType:
			rv.SetInt(tv.GetInt())
		case Int8Type:
			rv.SetInt(int64(tv.GetInt8()))
		case Int16Type:
			rv.SetInt(int64(tv.GetInt16()))
		case Int32Type, UntypedRuneType:
			rv.SetInt(int64(tv.GetInt32()))
		case Int64Type:
			rv.SetInt(tv.GetInt64())
		case UintType:
			rv.SetUint(tv.GetUint())
		case Uint8Type:
			rv.SetUint(uint64(tv.GetUint8()))
		case Uint16Type:
			rv.SetUint(uint64(tv.GetUint16()))
		case Uint32Type:
			rv.SetUint(uint64(tv.GetUint32()))
		case Uint64Type:
			rv.SetUint(tv.GetUint64())
		case Float32Type:
			rv.SetFloat(math.Float64frombits(softfloat.F32to64(tv.GetFloat32())))
		case Float64Type:
			rv.SetFloat(math.Float64frombits(tv.GetFloat64()))
		default:
			panic(fmt.Sprintf(
				"unexpected type %s",
				tv.T.String()))
		}
	case *PointerType:
		// This doesn't take into account pointer relativity, or even
		// identical pointers -- every non-nil gno pointer type results in a
		// new addressable value in go.
		if tv.V == nil {
			// do nothing
		} else {
			rve := reflect.New(rv.Type().Elem()).Elem()
			rv2 := Gno2GoValue(tv.V.(PointerValue).TV, rve)
			rv.Set(rv2.Addr())
		}
	case *ArrayType:
		if debug {
			if tv.V == nil {
				// all arguments and recursively fetched arrays
				// should have been initialized if not already so.
				panic("unexpected uninitialized array")
			}
		}
		// General case.
		av := tv.V.(*ArrayValue)
		if av.Data == nil {
			for i := range ct.Len {
				etv := &av.List[i]
				if etv.IsUndefined() {
					continue
				}
				Gno2GoValue(etv, rv.Index(i))
			}
		} else {
			for i := range ct.Len {
				val := av.Data[i]
				erv := rv.Index(i)
				erv.SetUint(uint64(val))
			}
		}
	case *SliceType:
		st := rt
		// If uninitialized slice, return zero value.
		if tv.V == nil {
			return
		}
		// General case.
		sv := tv.V.(*SliceValue)
		svo := sv.Offset
		svl := sv.Length
		svc := sv.Maxcap
		svb := sv.GetBase(nil)
		if svb.Data == nil {
			rv.Set(reflect.MakeSlice(st, svl, svc))
			for i := range svl {
				etv := &(svb.List[svo+i])
				if etv.IsUndefined() {
					continue
				}
				Gno2GoValue(etv, rv.Index(i))
			}
		} else {
			data := make([]byte, svl, svc)
			copy(data[:svc], svb.Data[svo:svo+svc])
			rv.Set(reflect.ValueOf(data))
		}
	default:
		panic(fmt.Sprintf(
			"unexpected type %s",
			tv.T.String()))
	}
	return
}
