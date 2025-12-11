package gnolang

import (
	"fmt"
	"math"
	"math/big"

	"github.com/cockroachdb/apd/v3"
	"github.com/gnolang/gno/gnovm/pkg/gnolang/internal/softfloat"
)

// ----------------------------------------
// Machine ops

func (m *Machine) doOpBinary1() {
	bx := m.PopExpr().(*BinaryExpr)
	switch bx.Op {
	case LAND:
		res := m.PeekValue(1) // re-use
		if res.GetBool() {
			m.PushOp(OpLand)
			// evaluate right
			m.PushExpr(bx.Right)
			m.PushOp(OpEval)
		} else {
			return // done, already false.
		}
	case LOR:
		res := m.PeekValue(1) // re-use
		if res.GetBool() {
			return // done, already true.
		} else {
			m.PushOp(OpLor)
			// evaluate right
			m.PushExpr(bx.Right)
			m.PushOp(OpEval)
		}
	default:
		panic(fmt.Sprintf(
			"unexpected binary(1) expr %s",
			bx.String()))
	}
}

func (m *Machine) doOpLor() {
	// get right and left operands.
	rv := m.PopValue()
	lv := m.PeekValue(1) // also the result
	if debug {
		debugAssertSameTypes(lv.T, rv.T)
	}

	// set result in lv.
	if isUntyped(lv.T) {
		lv.T = rv.T
	}
	lv.SetBool(lv.GetBool() || rv.GetBool())
}

func (m *Machine) doOpLand() {
	// get right and left operands.
	rv := m.PopValue()
	lv := m.PeekValue(1) // also the result
	if debug {
		debugAssertSameTypes(lv.T, rv.T)
	}

	// set result in lv.
	if isUntyped(lv.T) {
		lv.T = rv.T
	}
	lv.SetBool(lv.GetBool() && rv.GetBool())
}

func (m *Machine) doOpEql() {
	m.PopExpr()

	// get right and left operands.
	rv := m.PopValue()
	lv := m.PeekValue(1) // also the result
	if debug {
		debugAssertEqualityTypes(lv.T, rv.T)
	}
	// set result in lv.
	res := isEql(m.Store, lv, rv)
	lv.T = UntypedBoolType
	lv.V = nil
	lv.SetBool(res)
}

func (m *Machine) doOpNeq() {
	m.PopExpr()

	// get right and left operands.
	rv := m.PopValue()
	lv := m.PeekValue(1) // also the result
	if debug {
		debugAssertEqualityTypes(lv.T, rv.T)
	}

	// set result in lv.
	res := !isEql(m.Store, lv, rv)
	lv.T = UntypedBoolType
	lv.V = nil
	lv.SetBool(res)
}

func (m *Machine) doOpLss() {
	m.PopExpr()

	// get right and left operands.
	rv := m.PopValue()
	lv := m.PeekValue(1) // also the result
	if debug {
		debugAssertSameTypes(lv.T, rv.T)
	}

	// set the result in lv.
	res := isLss(lv, rv)
	lv.T = UntypedBoolType
	lv.V = nil
	lv.SetBool(res)
}

func (m *Machine) doOpLeq() {
	m.PopExpr()

	// get right and left operands.
	rv := m.PopValue()
	lv := m.PeekValue(1) // also the result
	if debug {
		debugAssertSameTypes(lv.T, rv.T)
	}

	// set the result in lv.
	res := isLeq(lv, rv)
	lv.T = UntypedBoolType
	lv.V = nil
	lv.SetBool(res)
}

func (m *Machine) doOpGtr() {
	m.PopExpr()

	// get right and left operands.
	rv := m.PopValue()
	lv := m.PeekValue(1) // also the result
	if debug {
		debugAssertSameTypes(lv.T, rv.T)
	}

	// set the result in lv.
	res := isGtr(lv, rv)
	lv.T = UntypedBoolType
	lv.V = nil
	lv.SetBool(res)
}

func (m *Machine) doOpGeq() {
	m.PopExpr()

	// get right and left operands.
	rv := m.PopValue()
	lv := m.PeekValue(1) // also the result
	if debug {
		debugAssertSameTypes(lv.T, rv.T)
	}

	// set the result in lv.
	res := isGeq(lv, rv)
	lv.T = UntypedBoolType
	lv.V = nil
	lv.SetBool(res)
}

func (m *Machine) doOpAdd() {
	m.PopExpr()

	// get right and left operands.
	rv := m.PopValue()
	lv := m.PeekValue(1) // also result
	if debug {
		debugAssertSameTypes(lv.T, rv.T)
	}

	// add rv to lv.
	addAssign(m.Alloc, lv, rv)
}

func (m *Machine) doOpSub() {
	m.PopExpr()

	// get right and left operands.
	rv := m.PopValue()
	lv := m.PeekValue(1) // also result
	if debug {
		debugAssertSameTypes(lv.T, rv.T)
	}

	// sub rv from lv.
	subAssign(lv, rv)
}

func (m *Machine) doOpBor() {
	m.PopExpr()

	// get right and left operands.
	rv := m.PopValue()
	lv := m.PeekValue(1) // also result
	if debug {
		debugAssertSameTypes(lv.T, rv.T)
	}

	// lv | rv
	borAssign(lv, rv)
}

func (m *Machine) doOpXor() {
	m.PopExpr()

	// get right and left operands.
	rv := m.PopValue()
	lv := m.PeekValue(1) // also result
	if debug {
		debugAssertSameTypes(lv.T, rv.T)
	}

	// lv ^ rv
	xorAssign(lv, rv)
}

func (m *Machine) doOpMul() {
	m.PopExpr()

	// get right and left operands.
	rv := m.PopValue()
	lv := m.PeekValue(1) // also result
	if debug {
		debugAssertSameTypes(lv.T, rv.T)
	}

	// lv * rv
	mulAssign(lv, rv)
}

func (m *Machine) doOpQuo() {
	m.PopExpr()

	// get right and left operands.
	rv := m.PopValue()
	lv := m.PeekValue(1) // also result
	if debug {
		debugAssertSameTypes(lv.T, rv.T)
	}

	// lv / rv
	err := quoAssign(lv, rv)
	if err != nil {
		panic(err)
	}
}

func (m *Machine) doOpRem() {
	m.PopExpr()

	// get right and left operands.
	rv := m.PopValue()
	lv := m.PeekValue(1) // also result
	if debug {
		debugAssertSameTypes(lv.T, rv.T)
	}

	// lv % rv
	err := remAssign(lv, rv)
	if err != nil {
		panic(err)
	}
}

func (m *Machine) doOpShl() {
	m.PopExpr()

	// get right and left operands.
	rv := m.PopValue()
	lv := m.PeekValue(1) // also result
	if debug {
		if rv.T.Kind() != UintKind {
			panic("should not happen")
		}
	}

	// lv << rv
	shlAssign(m, lv, rv)
}

func (m *Machine) doOpShr() {
	m.PopExpr()

	// get right and left operands.
	rv := m.PopValue()
	lv := m.PeekValue(1) // also result
	if debug {
		if rv.T.Kind() != UintKind {
			panic("should not happen")
		}
	}

	// lv >> rv
	shrAssign(m, lv, rv)
}

func (m *Machine) doOpBand() {
	m.PopExpr()

	// get right and left operands.
	rv := m.PopValue()
	lv := m.PeekValue(1) // also result
	if debug {
		debugAssertSameTypes(lv.T, rv.T)
	}

	// lv & rv
	bandAssign(lv, rv)
}

func (m *Machine) doOpBandn() {
	m.PopExpr()

	// get right and left operands.
	rv := m.PopValue()
	lv := m.PeekValue(1) // also result
	if debug {
		debugAssertSameTypes(lv.T, rv.T)
	}

	// lv &^ rv
	bandnAssign(lv, rv)
}

// ----------------------------------------
// logic functions

// TODO: can be much faster.
func isEql(store Store, lv, rv *TypedValue) bool {
	// If one is undefined, the other must be as well.
	// Fields/items are set to defaultTypedValue along the way.
	lvu := lv.IsUndefined()
	rvu := rv.IsUndefined()
	if lvu {
		return rvu
	} else if rvu {
		return false
	}
	if err := checkSame(lv.T, rv.T, ""); err != nil {
		return false
	}
	switch lv.T.Kind() {
	case BoolKind:
		return (lv.GetBool() == rv.GetBool())
	case StringKind:
		return (lv.GetString() == rv.GetString())
	case IntKind:
		return (lv.GetInt() == rv.GetInt())
	case Int8Kind:
		return (lv.GetInt8() == rv.GetInt8())
	case Int16Kind:
		return (lv.GetInt16() == rv.GetInt16())
	case Int32Kind:
		return (lv.GetInt32() == rv.GetInt32())
	case Int64Kind:
		return (lv.GetInt64() == rv.GetInt64())
	case UintKind:
		return (lv.GetUint() == rv.GetUint())
	case Uint8Kind:
		return (lv.GetUint8() == rv.GetUint8())
	case Uint16Kind:
		return (lv.GetUint16() == rv.GetUint16())
	case Uint32Kind:
		return (lv.GetUint32() == rv.GetUint32())
	case Uint64Kind:
		return (lv.GetUint64() == rv.GetUint64())
	case Float32Kind:
		return softfloat.Feq32(lv.GetFloat32(), rv.GetFloat32())
	case Float64Kind:
		return softfloat.Feq64(lv.GetFloat64(), rv.GetFloat64())
	case BigintKind:
		lb := lv.V.(BigintValue).V
		rb := rv.V.(BigintValue).V
		return lb.Cmp(rb) == 0
	case BigdecKind:
		lb := lv.V.(BigdecValue).V
		rb := rv.V.(BigdecValue).V
		return lb.Cmp(rb) == 0
	case ArrayKind:
		la := lv.V.(*ArrayValue)
		ra := rv.V.(*ArrayValue)
		at := baseOf(lv.T).(*ArrayType)
		et := at.Elt
		if debug {
			if la.GetLength() != ra.GetLength() {
				panic("comparison on arrays of unequal length")
			}
			rat := baseOf(lv.T).(*ArrayType)
			if at.TypeID() != rat.TypeID() {
				panic("comparison on arrays of unequal type")
			}
		}
		for i := range la.GetLength() {
			li := la.GetPointerAtIndexInt2(store, i, et).Deref()
			ri := ra.GetPointerAtIndexInt2(store, i, et).Deref()
			if !isEql(store, &li, &ri) {
				return false
			}
		}
		return true
	case StructKind:
		ls := lv.V.(*StructValue)
		rs := rv.V.(*StructValue)
		if debug {
			lt := baseOf(lv.T).(*StructType)
			rt := baseOf(rv.T).(*StructType)
			if lt.TypeID() != rt.TypeID() {
				panic("comparison on structs of unequal types")
			}
			if len(ls.Fields) != len(rs.Fields) {
				panic("comparison on structs of unequal size")
			}
		}
		for i := range ls.Fields {
			lf := ls.GetPointerToInt(store, i).Deref()
			rf := rs.GetPointerToInt(store, i).Deref()
			if !isEql(store, &lf, &rf) {
				return false
			}
		}
		return true
	case MapKind:
		if debug {
			if lv.V != nil && rv.V != nil {
				panic("map can only be compared with `nil`")
			}
		}
		return lv.V == rv.V
	case SliceKind:
		if debug {
			if lv.V != nil && rv.V != nil {
				panic("slice can only be compared with `nil`")
			}
		}
		return lv.V == rv.V
	case FuncKind:
		if debug {
			if lv.V != nil && rv.V != nil {
				panic("function can only be compared with `nil`")
			}
		}
		return lv.V == rv.V
	case PointerKind:
		if lv.T != rv.T &&
			lv.T.Elem() != DataByteType &&
			lv.T.TypeID() != rv.T.TypeID() {
			return false
		}

		if lv.V != nil && rv.V != nil {
			lpv := lv.V.(PointerValue)
			rpv := rv.V.(PointerValue)
			if lpv.TV.T == DataByteType && rpv.TV.T == DataByteType {
				return *(lpv.TV) == *(rpv.TV) && lpv.Base == rpv.Base && lpv.Index == rpv.Index
			}
		}
		return lv.V == rv.V
	default:
		panic(fmt.Sprintf(
			"comparison operator == not defined for %s",
			lv.T.Kind(),
		))
	}
}

// TODO: can be much faster.
func isLss(lv, rv *TypedValue) bool {
	switch lv.T.Kind() {
	case StringKind:
		return (lv.GetString() < rv.GetString())
	case IntKind:
		return (lv.GetInt() < rv.GetInt())
	case Int8Kind:
		return (lv.GetInt8() < rv.GetInt8())
	case Int16Kind:
		return (lv.GetInt16() < rv.GetInt16())
	case Int32Kind:
		return (lv.GetInt32() < rv.GetInt32())
	case Int64Kind:
		return (lv.GetInt64() < rv.GetInt64())
	case UintKind:
		return (lv.GetUint() < rv.GetUint())
	case Uint8Kind:
		return (lv.GetUint8() < rv.GetUint8())
	case Uint16Kind:
		return (lv.GetUint16() < rv.GetUint16())
	case Uint32Kind:
		return (lv.GetUint32() < rv.GetUint32())
	case Uint64Kind:
		return (lv.GetUint64() < rv.GetUint64())
	case Float32Kind:
		return softfloat.Flt32(lv.GetFloat32(), rv.GetFloat32())
	case Float64Kind:
		return softfloat.Flt64(lv.GetFloat64(), rv.GetFloat64())
	case BigintKind:
		lb := lv.V.(BigintValue).V
		rb := rv.V.(BigintValue).V
		return lb.Cmp(rb) < 0
	case BigdecKind:
		lb := lv.V.(BigdecValue).V
		rb := rv.V.(BigdecValue).V
		return lb.Cmp(rb) < 0
	default:
		panic(fmt.Sprintf(
			"comparison operator < not defined for %s",
			lv.T.Kind(),
		))
	}
}

func isLeq(lv, rv *TypedValue) bool {
	switch lv.T.Kind() {
	case StringKind:
		return (lv.GetString() <= rv.GetString())
	case IntKind:
		return (lv.GetInt() <= rv.GetInt())
	case Int8Kind:
		return (lv.GetInt8() <= rv.GetInt8())
	case Int16Kind:
		return (lv.GetInt16() <= rv.GetInt16())
	case Int32Kind:
		return (lv.GetInt32() <= rv.GetInt32())
	case Int64Kind:
		return (lv.GetInt64() <= rv.GetInt64())
	case UintKind:
		return (lv.GetUint() <= rv.GetUint())
	case Uint8Kind:
		return (lv.GetUint8() <= rv.GetUint8())
	case Uint16Kind:
		return (lv.GetUint16() <= rv.GetUint16())
	case Uint32Kind:
		return (lv.GetUint32() <= rv.GetUint32())
	case Uint64Kind:
		return (lv.GetUint64() <= rv.GetUint64())
	case Float32Kind:
		return softfloat.Fle32(lv.GetFloat32(), rv.GetFloat32())
	case Float64Kind:
		return softfloat.Fle64(lv.GetFloat64(), rv.GetFloat64())
	case BigintKind:
		lb := lv.V.(BigintValue).V
		rb := rv.V.(BigintValue).V
		return lb.Cmp(rb) <= 0
	case BigdecKind:
		lb := lv.V.(BigdecValue).V
		rb := rv.V.(BigdecValue).V
		return lb.Cmp(rb) <= 0
	default:
		panic(fmt.Sprintf(
			"comparison operator <= not defined for %s",
			lv.T.Kind(),
		))
	}
}

func isGtr(lv, rv *TypedValue) bool {
	switch lv.T.Kind() {
	case StringKind:
		return (lv.GetString() > rv.GetString())
	case IntKind:
		return (lv.GetInt() > rv.GetInt())
	case Int8Kind:
		return (lv.GetInt8() > rv.GetInt8())
	case Int16Kind:
		return (lv.GetInt16() > rv.GetInt16())
	case Int32Kind:
		return (lv.GetInt32() > rv.GetInt32())
	case Int64Kind:
		return (lv.GetInt64() > rv.GetInt64())
	case UintKind:
		return (lv.GetUint() > rv.GetUint())
	case Uint8Kind:
		return (lv.GetUint8() > rv.GetUint8())
	case Uint16Kind:
		return (lv.GetUint16() > rv.GetUint16())
	case Uint32Kind:
		return (lv.GetUint32() > rv.GetUint32())
	case Uint64Kind:
		return (lv.GetUint64() > rv.GetUint64())
	case Float32Kind:
		return softfloat.Fgt32(lv.GetFloat32(), rv.GetFloat32())
	case Float64Kind:
		return softfloat.Fgt64(lv.GetFloat64(), rv.GetFloat64())
	case BigintKind:
		lb := lv.V.(BigintValue).V
		rb := rv.V.(BigintValue).V
		return lb.Cmp(rb) > 0
	case BigdecKind:
		lb := lv.V.(BigdecValue).V
		rb := rv.V.(BigdecValue).V
		return lb.Cmp(rb) > 0
	default:
		panic(fmt.Sprintf(
			"comparison operator > not defined for %s",
			lv.T.Kind(),
		))
	}
}

func isGeq(lv, rv *TypedValue) bool {
	switch lv.T.Kind() {
	case StringKind:
		return (lv.GetString() >= rv.GetString())
	case IntKind:
		return (lv.GetInt() >= rv.GetInt())
	case Int8Kind:
		return (lv.GetInt8() >= rv.GetInt8())
	case Int16Kind:
		return (lv.GetInt16() >= rv.GetInt16())
	case Int32Kind:
		return (lv.GetInt32() >= rv.GetInt32())
	case Int64Kind:
		return (lv.GetInt64() >= rv.GetInt64())
	case UintKind:
		return (lv.GetUint() >= rv.GetUint())
	case Uint8Kind:
		return (lv.GetUint8() >= rv.GetUint8())
	case Uint16Kind:
		return (lv.GetUint16() >= rv.GetUint16())
	case Uint32Kind:
		return (lv.GetUint32() >= rv.GetUint32())
	case Uint64Kind:
		return (lv.GetUint64() >= rv.GetUint64())
	case Float32Kind:
		return softfloat.Fge32(lv.GetFloat32(), rv.GetFloat32())
	case Float64Kind:
		return softfloat.Fge64(lv.GetFloat64(), rv.GetFloat64())
	case BigintKind:
		lb := lv.V.(BigintValue).V
		rb := rv.V.(BigintValue).V
		return lb.Cmp(rb) >= 0
	case BigdecKind:
		lb := lv.V.(BigdecValue).V
		rb := rv.V.(BigdecValue).V
		return lb.Cmp(rb) >= 0
	default:
		panic(fmt.Sprintf(
			"comparison operator >= not defined for %s",
			lv.T.Kind(),
		))
	}
}

// for doOpAdd and doOpAddAssign.
func addAssign(alloc *Allocator, lv, rv *TypedValue) {
	// set the result in lv.
	// NOTE this block is replicated in op_assign.go
	switch baseOf(lv.T) {
	case StringType, UntypedStringType:
		lv.V = alloc.NewString(lv.GetString() + rv.GetString())
	case IntType:
		lv.SetInt(lv.GetInt() + rv.GetInt())
	case Int8Type:
		lv.SetInt8(lv.GetInt8() + rv.GetInt8())
	case Int16Type:
		lv.SetInt16(lv.GetInt16() + rv.GetInt16())
	case Int32Type, UntypedRuneType:
		lv.SetInt32(lv.GetInt32() + rv.GetInt32())
	case Int64Type:
		lv.SetInt64(lv.GetInt64() + rv.GetInt64())
	case UintType:
		lv.SetUint(lv.GetUint() + rv.GetUint())
	case Uint8Type:
		lv.SetUint8(lv.GetUint8() + rv.GetUint8())
	case DataByteType:
		lv.SetDataByte(lv.GetDataByte() + rv.GetUint8())
	case Uint16Type:
		lv.SetUint16(lv.GetUint16() + rv.GetUint16())
	case Uint32Type:
		lv.SetUint32(lv.GetUint32() + rv.GetUint32())
	case Uint64Type:
		lv.SetUint64(lv.GetUint64() + rv.GetUint64())
	case Float32Type:
		// NOTE: gno doesn't fuse *+.
		lv.SetFloat32(softfloat.Fadd32(lv.GetFloat32(), rv.GetFloat32()))
	case Float64Type:
		// NOTE: gno doesn't fuse *+.
		lv.SetFloat64(softfloat.Fadd64(lv.GetFloat64(), rv.GetFloat64()))
	case UntypedBigintType:
		lb := lv.GetBigInt()
		lb = big.NewInt(0).Add(lb, rv.GetBigInt())
		lv.V = BigintValue{V: lb}
	case UntypedBigdecType:
		lb := lv.GetBigDec()
		rb := rv.GetBigDec()
		sum := apd.New(0, 0)
		cond, err := apd.BaseContext.WithPrecision(0).Add(sum, lb, rb)
		if err != nil {
			panic(fmt.Sprintf("bigdec addition error: %v", err))
		} else if cond.Inexact() {
			panic(fmt.Sprintf("bigdec addition inexact: %v + %v", lb, rb))
		}
		lv.V = BigdecValue{V: sum}
	default:
		panic(fmt.Sprintf(
			"operators + and += not defined for %s",
			lv.T,
		))
	}
}

// for doOpSub and doOpSubAssign.
func subAssign(lv, rv *TypedValue) {
	// set the result in lv.
	// NOTE this block is replicated in op_assign.go
	switch baseOf(lv.T) {
	case IntType:
		lv.SetInt(lv.GetInt() - rv.GetInt())
	case Int8Type:
		lv.SetInt8(lv.GetInt8() - rv.GetInt8())
	case Int16Type:
		lv.SetInt16(lv.GetInt16() - rv.GetInt16())
	case Int32Type, UntypedRuneType:
		lv.SetInt32(lv.GetInt32() - rv.GetInt32())
	case Int64Type:
		lv.SetInt64(lv.GetInt64() - rv.GetInt64())
	case UintType:
		lv.SetUint(lv.GetUint() - rv.GetUint())
	case Uint8Type:
		lv.SetUint8(lv.GetUint8() - rv.GetUint8())
	case DataByteType:
		lv.SetDataByte(lv.GetDataByte() - rv.GetUint8())
	case Uint16Type:
		lv.SetUint16(lv.GetUint16() - rv.GetUint16())
	case Uint32Type:
		lv.SetUint32(lv.GetUint32() - rv.GetUint32())
	case Uint64Type:
		lv.SetUint64(lv.GetUint64() - rv.GetUint64())
	case Float32Type:
		// NOTE: gno doesn't fuse *+.
		lv.SetFloat32(softfloat.Fsub32(lv.GetFloat32(), rv.GetFloat32()))
	case Float64Type:
		// NOTE: gno doesn't fuse *+.
		lv.SetFloat64(softfloat.Fsub64(lv.GetFloat64(), rv.GetFloat64()))
	case UntypedBigintType:
		lb := lv.GetBigInt()
		lb = big.NewInt(0).Sub(lb, rv.GetBigInt())
		lv.V = BigintValue{V: lb}
	case UntypedBigdecType:
		lb := lv.GetBigDec()
		rb := rv.GetBigDec()
		diff := apd.New(0, 0)
		cond, err := apd.BaseContext.WithPrecision(0).Sub(diff, lb, rb)
		if err != nil {
			panic(fmt.Sprintf("bigdec subtraction error: %v", err))
		} else if cond.Inexact() {
			panic(fmt.Sprintf("bigdec subtraction inexact: %v + %v", lb, rb))
		}
		lv.V = BigdecValue{V: diff}
	default:
		panic(fmt.Sprintf(
			"operators - and -= not defined for %s",
			lv.T,
		))
	}
}

// for doOpMul and doOpMulAssign.
func mulAssign(lv, rv *TypedValue) {
	// set the result in lv.
	// NOTE this block is replicated in op_assign.go
	switch baseOf(lv.T) {
	case IntType:
		lv.SetInt(lv.GetInt() * rv.GetInt())
	case Int8Type:
		lv.SetInt8(lv.GetInt8() * rv.GetInt8())
	case Int16Type:
		lv.SetInt16(lv.GetInt16() * rv.GetInt16())
	case Int32Type, UntypedRuneType:
		lv.SetInt32(lv.GetInt32() * rv.GetInt32())
	case Int64Type:
		lv.SetInt64(lv.GetInt64() * rv.GetInt64())
	case UintType:
		lv.SetUint(lv.GetUint() * rv.GetUint())
	case Uint8Type:
		lv.SetUint8(lv.GetUint8() * rv.GetUint8())
	case DataByteType:
		lv.SetDataByte(lv.GetDataByte() * rv.GetUint8())
	case Uint16Type:
		lv.SetUint16(lv.GetUint16() * rv.GetUint16())
	case Uint32Type:
		lv.SetUint32(lv.GetUint32() * rv.GetUint32())
	case Uint64Type:
		lv.SetUint64(lv.GetUint64() * rv.GetUint64())
	case Float32Type:
		// NOTE: gno doesn't fuse *+.
		lv.SetFloat32(softfloat.Fmul32(lv.GetFloat32(), rv.GetFloat32()))
	case Float64Type:
		// NOTE: gno doesn't fuse *+.
		lv.SetFloat64(softfloat.Fmul64(lv.GetFloat64(), rv.GetFloat64()))
	case UntypedBigintType:
		lb := lv.GetBigInt()
		lb = big.NewInt(0).Mul(lb, rv.GetBigInt())
		lv.V = BigintValue{V: lb}
	case UntypedBigdecType:
		lb := lv.GetBigDec()
		rb := rv.GetBigDec()
		prod := apd.New(0, 0)
		_, err := apd.BaseContext.WithPrecision(1024).Mul(prod, lb, rb)
		if err != nil {
			panic(fmt.Sprintf("bigdec multiplication error: %v", err))
		}
		lv.V = BigdecValue{V: prod}
	default:
		panic(fmt.Sprintf(
			"operators * and *= not defined for %s",
			lv.T,
		))
	}
}

// for doOpQuo and doOpQuoAssign.
func quoAssign(lv, rv *TypedValue) *Exception {
	expt := &Exception{
		Value: typedString("division by zero"),
	}

	// set the result in lv.
	// NOTE this block is replicated in op_assign.go
	switch baseOf(lv.T) {
	case IntType:
		if rv.GetInt() == 0 {
			return expt
		}
		lv.SetInt(lv.GetInt() / rv.GetInt())
	case Int8Type:
		if rv.GetInt8() == 0 {
			return expt
		}
		lv.SetInt8(lv.GetInt8() / rv.GetInt8())
	case Int16Type:
		if rv.GetInt16() == 0 {
			return expt
		}
		lv.SetInt16(lv.GetInt16() / rv.GetInt16())
	case Int32Type, UntypedRuneType:
		if rv.GetInt32() == 0 {
			return expt
		}
		lv.SetInt32(lv.GetInt32() / rv.GetInt32())
	case Int64Type:
		if rv.GetInt64() == 0 {
			return expt
		}
		lv.SetInt64(lv.GetInt64() / rv.GetInt64())
	case UintType:
		if rv.GetUint() == 0 {
			return expt
		}
		lv.SetUint(lv.GetUint() / rv.GetUint())
	case Uint8Type:
		if rv.GetUint8() == 0 {
			return expt
		}
		lv.SetUint8(lv.GetUint8() / rv.GetUint8())
	case DataByteType:
		if rv.GetUint8() == 0 {
			return expt
		}
		lv.SetDataByte(lv.GetDataByte() / rv.GetUint8())
	case Uint16Type:
		if rv.GetUint16() == 0 {
			return expt
		}
		lv.SetUint16(lv.GetUint16() / rv.GetUint16())
	case Uint32Type:
		if rv.GetUint32() == 0 {
			return expt
		}
		lv.SetUint32(lv.GetUint32() / rv.GetUint32())
	case Uint64Type:
		if rv.GetUint64() == 0 {
			return expt
		}
		lv.SetUint64(lv.GetUint64() / rv.GetUint64())
	case Float32Type:
		// NOTE: gno doesn't fuse *+.
		lv.SetFloat32(softfloat.Fdiv32(lv.GetFloat32(), rv.GetFloat32()))
	case Float64Type:
		// NOTE: gno doesn't fuse *+.
		lv.SetFloat64(softfloat.Fdiv64(lv.GetFloat64(), rv.GetFloat64()))
	case UntypedBigintType:
		if rv.GetBigInt().Sign() == 0 {
			return expt
		}
		lb := lv.GetBigInt()
		lb = big.NewInt(0).Quo(lb, rv.GetBigInt())
		lv.V = BigintValue{V: lb}
	case UntypedBigdecType:
		if rv.GetBigDec().Cmp(apd.New(0, 0)) == 0 {
			return expt
		}
		lb := lv.GetBigDec()
		rb := rv.GetBigDec()
		quo := apd.New(0, 0)
		_, err := apd.BaseContext.WithPrecision(1024).Quo(quo, lb, rb)
		if err != nil {
			panic(fmt.Sprintf("bigdec division error: %v", err))
		}
		lv.V = BigdecValue{V: quo}
	default:
		panic(fmt.Sprintf(
			"operators / and /= not defined for %s",
			lv.T,
		))
	}

	return nil
}

// for doOpRem and doOpRemAssign.
func remAssign(lv, rv *TypedValue) *Exception {
	expt := &Exception{
		Value: typedString("division by zero"),
	}

	// set the result in lv.
	// NOTE this block is replicated in op_assign.go
	switch baseOf(lv.T) {
	case IntType:
		if rv.GetInt() == 0 {
			return expt
		}
		lv.SetInt(lv.GetInt() % rv.GetInt())
	case Int8Type:
		if rv.GetInt8() == 0 {
			return expt
		}
		lv.SetInt8(lv.GetInt8() % rv.GetInt8())
	case Int16Type:
		if rv.GetInt16() == 0 {
			return expt
		}
		lv.SetInt16(lv.GetInt16() % rv.GetInt16())
	case Int32Type, UntypedRuneType:
		if rv.GetInt32() == 0 {
			return expt
		}
		lv.SetInt32(lv.GetInt32() % rv.GetInt32())
	case Int64Type:
		if rv.GetInt64() == 0 {
			return expt
		}
		lv.SetInt64(lv.GetInt64() % rv.GetInt64())
	case UintType:
		if rv.GetUint() == 0 {
			return expt
		}
		lv.SetUint(lv.GetUint() % rv.GetUint())
	case Uint8Type:
		if rv.GetUint8() == 0 {
			return expt
		}
		lv.SetUint8(lv.GetUint8() % rv.GetUint8())
	case DataByteType:
		if rv.GetUint8() == 0 {
			return expt
		}
		lv.SetDataByte(lv.GetDataByte() % rv.GetUint8())
	case Uint16Type:
		if rv.GetUint16() == 0 {
			return expt
		}
		lv.SetUint16(lv.GetUint16() % rv.GetUint16())
	case Uint32Type:
		if rv.GetUint32() == 0 {
			return expt
		}
		lv.SetUint32(lv.GetUint32() % rv.GetUint32())
	case Uint64Type:
		if rv.GetUint64() == 0 {
			return expt
		}
		lv.SetUint64(lv.GetUint64() % rv.GetUint64())
	case UntypedBigintType:
		if rv.GetBigInt().Sign() == 0 {
			return expt
		}

		lb := lv.GetBigInt()
		lb = big.NewInt(0).Rem(lb, rv.GetBigInt())
		lv.V = BigintValue{V: lb}
	default:
		panic(fmt.Sprintf(
			"operators %% and %%= not defined for %s",
			lv.T,
		))
	}

	return nil
}

// for doOpBand and doOpBandAssign.
func bandAssign(lv, rv *TypedValue) {
	// set the result in lv.
	// NOTE this block is replicated in op_assign.go
	switch baseOf(lv.T) {
	case IntType:
		lv.SetInt(lv.GetInt() & rv.GetInt())
	case Int8Type:
		lv.SetInt8(lv.GetInt8() & rv.GetInt8())
	case Int16Type:
		lv.SetInt16(lv.GetInt16() & rv.GetInt16())
	case Int32Type, UntypedRuneType:
		lv.SetInt32(lv.GetInt32() & rv.GetInt32())
	case Int64Type:
		lv.SetInt64(lv.GetInt64() & rv.GetInt64())
	case UintType:
		lv.SetUint(lv.GetUint() & rv.GetUint())
	case Uint8Type:
		lv.SetUint8(lv.GetUint8() & rv.GetUint8())
	case DataByteType:
		lv.SetDataByte(lv.GetDataByte() & rv.GetUint8())
	case Uint16Type:
		lv.SetUint16(lv.GetUint16() & rv.GetUint16())
	case Uint32Type:
		lv.SetUint32(lv.GetUint32() & rv.GetUint32())
	case Uint64Type:
		lv.SetUint64(lv.GetUint64() & rv.GetUint64())
	case UntypedBigintType:
		lb := lv.GetBigInt()
		lb = big.NewInt(0).And(lb, rv.GetBigInt())
		lv.V = BigintValue{V: lb}
	default:
		panic(fmt.Sprintf(
			"operators & and &= not defined for %s",
			lv.T,
		))
	}
}

// for doOpBandn and doOpBandnAssign.
func bandnAssign(lv, rv *TypedValue) {
	// set the result in lv.
	// NOTE this block is replicated in op_assign.go
	switch baseOf(lv.T) {
	case IntType:
		lv.SetInt(lv.GetInt() &^ rv.GetInt())
	case Int8Type:
		lv.SetInt8(lv.GetInt8() &^ rv.GetInt8())
	case Int16Type:
		lv.SetInt16(lv.GetInt16() &^ rv.GetInt16())
	case Int32Type, UntypedRuneType:
		lv.SetInt32(lv.GetInt32() &^ rv.GetInt32())
	case Int64Type:
		lv.SetInt64(lv.GetInt64() &^ rv.GetInt64())
	case UintType:
		lv.SetUint(lv.GetUint() &^ rv.GetUint())
	case Uint8Type:
		lv.SetUint8(lv.GetUint8() &^ rv.GetUint8())
	case DataByteType:
		lv.SetDataByte(lv.GetDataByte() &^ rv.GetUint8())
	case Uint16Type:
		lv.SetUint16(lv.GetUint16() &^ rv.GetUint16())
	case Uint32Type:
		lv.SetUint32(lv.GetUint32() &^ rv.GetUint32())
	case Uint64Type:
		lv.SetUint64(lv.GetUint64() &^ rv.GetUint64())
	case UntypedBigintType:
		lb := lv.GetBigInt()
		lb = big.NewInt(0).AndNot(lb, rv.GetBigInt())
		lv.V = BigintValue{V: lb}
	default:
		panic(fmt.Sprintf(
			"operators &^ and &^= not defined for %s",
			lv.T,
		))
	}
}

// for doOpBor and doOpBorAssign.
func borAssign(lv, rv *TypedValue) {
	// set the result in lv.
	// NOTE this block is replicated in op_assign.go
	switch baseOf(lv.T) {
	case IntType:
		lv.SetInt(lv.GetInt() | rv.GetInt())
	case Int8Type:
		lv.SetInt8(lv.GetInt8() | rv.GetInt8())
	case Int16Type:
		lv.SetInt16(lv.GetInt16() | rv.GetInt16())
	case Int32Type, UntypedRuneType:
		lv.SetInt32(lv.GetInt32() | rv.GetInt32())
	case Int64Type:
		lv.SetInt64(lv.GetInt64() | rv.GetInt64())
	case UintType:
		lv.SetUint(lv.GetUint() | rv.GetUint())
	case Uint8Type:
		lv.SetUint8(lv.GetUint8() | rv.GetUint8())
	case DataByteType:
		lv.SetDataByte(lv.GetDataByte() | rv.GetUint8())
	case Uint16Type:
		lv.SetUint16(lv.GetUint16() | rv.GetUint16())
	case Uint32Type:
		lv.SetUint32(lv.GetUint32() | rv.GetUint32())
	case Uint64Type:
		lv.SetUint64(lv.GetUint64() | rv.GetUint64())
	case UntypedBigintType:
		lb := lv.GetBigInt()
		lb = big.NewInt(0).Or(lb, rv.GetBigInt())
		lv.V = BigintValue{V: lb}
	default:
		panic(fmt.Sprintf(
			"operators | and |= not defined for %s",
			lv.T,
		))
	}
}

// for doOpXor and doOpXorAssign.
func xorAssign(lv, rv *TypedValue) {
	// set the result in lv.
	// NOTE this block is replicated in op_assign.go
	switch baseOf(lv.T) {
	case IntType:
		lv.SetInt(lv.GetInt() ^ rv.GetInt())
	case Int8Type:
		lv.SetInt8(lv.GetInt8() ^ rv.GetInt8())
	case Int16Type:
		lv.SetInt16(lv.GetInt16() ^ rv.GetInt16())
	case Int32Type, UntypedRuneType:
		lv.SetInt32(lv.GetInt32() ^ rv.GetInt32())
	case Int64Type:
		lv.SetInt64(lv.GetInt64() ^ rv.GetInt64())
	case UintType:
		lv.SetUint(lv.GetUint() ^ rv.GetUint())
	case Uint8Type:
		lv.SetUint8(lv.GetUint8() ^ rv.GetUint8())
	case DataByteType:
		lv.SetDataByte(lv.GetDataByte() ^ rv.GetUint8())
	case Uint16Type:
		lv.SetUint16(lv.GetUint16() ^ rv.GetUint16())
	case Uint32Type:
		lv.SetUint32(lv.GetUint32() ^ rv.GetUint32())
	case Uint64Type:
		lv.SetUint64(lv.GetUint64() ^ rv.GetUint64())
	case UntypedBigintType:
		lb := lv.GetBigInt()
		lb = big.NewInt(0).Xor(lb, rv.GetBigInt())
		lv.V = BigintValue{V: lb}
	default:
		panic(fmt.Sprintf(
			"operators ^ and ^= not defined for %s",
			lv.T,
		))
	}
}

// for doOpShl and doOpShlAssign.
func shlAssign(m *Machine, lv, rv *TypedValue) {
	rv.AssertNonNegative("runtime error: negative shift amount")

	checkOverflow := func(v func() bool) {
		if m.Stage == StagePre && !v() {
			panic(`constant overflows`)
		}
	}

	// set the result in lv.
	// NOTE: baseOf(rv.T) is always UintType.
	switch baseOf(lv.T) {
	case IntType:
		checkOverflow(func() bool {
			l := big.NewInt(lv.GetInt())
			r := big.NewInt(0).Lsh(l, uint(rv.GetUint()))

			return r.Cmp(big.NewInt(math.MaxInt)) != 1
		})

		lv.SetInt(lv.GetInt() << rv.GetUint())
	case Int8Type:
		checkOverflow(func() bool {
			l := big.NewInt(int64(lv.GetInt8()))
			r := big.NewInt(0).Lsh(l, uint(rv.GetUint()))

			return r.Cmp(big.NewInt(math.MaxInt8)) != 1
		})

		lv.SetInt8(lv.GetInt8() << rv.GetUint())
	case Int16Type:
		checkOverflow(func() bool {
			l := big.NewInt(int64(lv.GetInt16()))
			r := big.NewInt(0).Lsh(l, uint(rv.GetUint()))

			return r.Cmp(big.NewInt(math.MaxInt16)) != 1
		})

		lv.SetInt16(lv.GetInt16() << rv.GetUint())
	case Int32Type, UntypedRuneType:
		checkOverflow(func() bool {
			l := big.NewInt(int64(lv.GetInt32()))
			r := big.NewInt(0).Lsh(l, uint(rv.GetUint()))

			return r.Cmp(big.NewInt(math.MaxInt32)) != 1
		})

		lv.SetInt32(lv.GetInt32() << rv.GetUint())
	case Int64Type:
		checkOverflow(func() bool {
			l := big.NewInt(lv.GetInt64())
			r := big.NewInt(0).Lsh(l, uint(rv.GetUint()))

			return r.Cmp(big.NewInt(math.MaxInt64)) != 1
		})

		lv.SetInt64(lv.GetInt64() << rv.GetUint())
	case UintType:
		checkOverflow(func() bool {
			l := big.NewInt(0).SetUint64(lv.GetUint())
			r := big.NewInt(0).Lsh(l, uint(rv.GetUint()))

			return r.Cmp(big.NewInt(0).SetUint64(math.MaxUint)) != 1
		})

		lv.SetUint(lv.GetUint() << rv.GetUint())
	case Uint8Type:
		checkOverflow(func() bool {
			l := big.NewInt(0).SetUint64(uint64(lv.GetUint8()))
			r := big.NewInt(0).Lsh(l, uint(rv.GetUint()))

			return r.Cmp(big.NewInt(math.MaxUint8)) != 1
		})

		lv.SetUint8(lv.GetUint8() << rv.GetUint())
	case DataByteType:
		checkOverflow(func() bool {
			l := big.NewInt(0).SetUint64(uint64(lv.GetDataByte()))
			r := big.NewInt(0).Lsh(l, uint(rv.GetUint()))

			return r.Cmp(big.NewInt(math.MaxUint8)) != 1
		})

		lv.SetDataByte(lv.GetDataByte() << rv.GetUint())
	case Uint16Type:
		checkOverflow(func() bool {
			l := big.NewInt(0).SetUint64(uint64(lv.GetUint16()))
			r := big.NewInt(0).Lsh(l, uint(rv.GetUint()))

			return r.Cmp(big.NewInt(math.MaxUint16)) != 1
		})

		lv.SetUint16(lv.GetUint16() << rv.GetUint())
	case Uint32Type:
		checkOverflow(func() bool {
			l := big.NewInt(0).SetUint64(uint64(lv.GetUint32()))
			r := big.NewInt(0).Lsh(l, uint(rv.GetUint()))

			return r.Cmp(big.NewInt(math.MaxUint32)) != 1
		})

		lv.SetUint32(lv.GetUint32() << rv.GetUint())
	case Uint64Type:
		checkOverflow(func() bool {
			l := big.NewInt(0).SetUint64(lv.GetUint64())
			r := big.NewInt(0).Lsh(l, uint(rv.GetUint()))

			return r.Cmp(big.NewInt(0).SetUint64(math.MaxUint64)) != 1
		})

		lv.SetUint64(lv.GetUint64() << rv.GetUint())
	case UntypedBigintType:
		lb := lv.GetBigInt()
		lb = big.NewInt(0).Lsh(lb, uint(rv.GetUint()))
		lv.V = BigintValue{V: lb}
	default:
		panic(fmt.Sprintf(
			"operators << and <<= not defined for %s",
			lv.T,
		))
	}
}

// for doOpShr and doOpShrAssign.
func shrAssign(m *Machine, lv, rv *TypedValue) {
	rv.AssertNonNegative("runtime error: negative shift amount")

	checkOverflow := func(v func() bool) {
		if m.Stage == StagePre && !v() {
			panic(`constant overflows`)
		}
	}

	// set the result in lv.
	// NOTE: baseOf(rv.T) is always UintType.
	switch baseOf(lv.T) {
	case IntType:
		checkOverflow(func() bool {
			l := big.NewInt(lv.GetInt())
			r := big.NewInt(0).Rsh(l, uint(rv.GetUint()))

			return r.Cmp(big.NewInt(math.MaxInt)) != 1
		})

		lv.SetInt(lv.GetInt() >> rv.GetUint())
	case Int8Type:
		checkOverflow(func() bool {
			l := big.NewInt(int64(lv.GetInt8()))
			r := big.NewInt(0).Rsh(l, uint(rv.GetUint()))

			return r.Cmp(big.NewInt(math.MaxInt8)) != 1
		})

		lv.SetInt8(lv.GetInt8() >> rv.GetUint())
	case Int16Type:
		checkOverflow(func() bool {
			l := big.NewInt(int64(lv.GetInt16()))
			r := big.NewInt(0).Rsh(l, uint(rv.GetUint()))

			return r.Cmp(big.NewInt(math.MaxInt16)) != 1
		})

		lv.SetInt16(lv.GetInt16() >> rv.GetUint())
	case Int32Type, UntypedRuneType:
		checkOverflow(func() bool {
			l := big.NewInt(int64(lv.GetInt32()))
			r := big.NewInt(0).Rsh(l, uint(rv.GetUint()))

			return r.Cmp(big.NewInt(math.MaxInt32)) != 1
		})

		lv.SetInt32(lv.GetInt32() >> rv.GetUint())
	case Int64Type:
		checkOverflow(func() bool {
			l := big.NewInt(lv.GetInt64())
			r := big.NewInt(0).Rsh(l, uint(rv.GetUint()))

			return r.Cmp(big.NewInt(math.MaxInt64)) != 1
		})

		lv.SetInt64(lv.GetInt64() >> rv.GetUint())
	case UintType:
		checkOverflow(func() bool {
			l := big.NewInt(0).SetUint64(lv.GetUint())
			r := big.NewInt(0).Rsh(l, uint(rv.GetUint()))

			return r.Cmp(big.NewInt(0).SetUint64(math.MaxUint)) != 1
		})

		lv.SetUint(lv.GetUint() >> rv.GetUint())
	case Uint8Type:
		checkOverflow(func() bool {
			l := big.NewInt(0).SetUint64(uint64(lv.GetUint8()))
			r := big.NewInt(0).Rsh(l, uint(rv.GetUint()))

			return r.Cmp(big.NewInt(math.MaxUint8)) != 1
		})

		lv.SetUint8(lv.GetUint8() >> rv.GetUint())
	case DataByteType:
		checkOverflow(func() bool {
			l := big.NewInt(0).SetUint64(uint64(lv.GetDataByte()))
			r := big.NewInt(0).Rsh(l, uint(rv.GetUint()))

			return r.Cmp(big.NewInt(math.MaxUint8)) != 1
		})

		lv.SetDataByte(lv.GetDataByte() >> rv.GetUint())
	case Uint16Type:
		checkOverflow(func() bool {
			l := big.NewInt(0).SetUint64(uint64(lv.GetUint16()))
			r := big.NewInt(0).Rsh(l, uint(rv.GetUint()))

			return r.Cmp(big.NewInt(math.MaxUint16)) != 1
		})

		lv.SetUint16(lv.GetUint16() >> rv.GetUint())
	case Uint32Type:
		checkOverflow(func() bool {
			l := big.NewInt(0).SetUint64(uint64(lv.GetUint32()))
			r := big.NewInt(0).Rsh(l, uint(rv.GetUint()))

			return r.Cmp(big.NewInt(math.MaxUint32)) != 1
		})

		lv.SetUint32(lv.GetUint32() >> rv.GetUint())
	case Uint64Type:
		checkOverflow(func() bool {
			l := big.NewInt(0).SetUint64(lv.GetUint64())
			r := big.NewInt(0).Rsh(l, uint(rv.GetUint()))

			return r.Cmp(big.NewInt(0).SetUint64(math.MaxUint64)) != 1
		})

		lv.SetUint64(lv.GetUint64() >> rv.GetUint())
	case UntypedBigintType:
		lb := lv.GetBigInt()
		lb = big.NewInt(0).Rsh(lb, uint(rv.GetUint()))
		lv.V = BigintValue{V: lb}
	default:
		panic(fmt.Sprintf(
			"operators >> and >>= not defined for %s",
			lv.T,
		))
	}
}
