package gno

import (
	"fmt"
	"reflect"
)

//----------------------------------------
// Machine ops

func (m *Machine) doOpBinary1() {
	bx := m.PopExpr().(*BinaryExpr)
	switch bx.Op {
	case LAND:
		res := m.PeekValue(1) // re-use
		if res.GetBool() {
			// continuation
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
			// continuation
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
		assertTypes(lv.T, rv.T)
	}

	// set result in lv.
	lv.SetBool(rv.GetBool())
}

func (m *Machine) doOpLand() {
	// get right and left operands.
	rv := m.PopValue()
	lv := m.PeekValue(1) // also the result
	if debug {
		assertTypes(lv.T, rv.T)
	}

	// set result in lv.
	lv.SetBool(rv.GetBool())
}

func (m *Machine) doOpEql() {
	m.PopExpr()

	// get right and left operands.
	rv := m.PopValue()
	lv := m.PeekValue(1) // also the result
	if debug {
		assertTypes(lv.T, rv.T)
	}

	// set result in lv.
	res := isEql(lv, rv)
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
		assertTypes(lv.T, rv.T)
	}

	// set result in lv.
	res := !isEql(lv, rv)
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
		assertTypes(lv.T, rv.T)
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
		assertTypes(lv.T, rv.T)
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
		assertTypes(lv.T, rv.T)
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
		assertTypes(lv.T, rv.T)
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
		assertTypes(lv.T, rv.T)
	}

	// add rv to lv.
	addAssign(lv, rv)
}

func (m *Machine) doOpSub() {
	m.PopExpr()

	// get right and left operands.
	rv := m.PopValue()
	lv := m.PeekValue(1) // also result
	if debug {
		assertTypes(lv.T, rv.T)
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
		assertTypes(lv.T, rv.T)
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
		assertTypes(lv.T, rv.T)
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
		assertTypes(lv.T, rv.T)
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
		assertTypes(lv.T, rv.T)
	}

	// lv / rv
	quoAssign(lv, rv)
}

func (m *Machine) doOpRem() {
	m.PopExpr()

	// get right and left operands.
	rv := m.PopValue()
	lv := m.PeekValue(1) // also result
	if debug {
		assertTypes(lv.T, rv.T)
	}

	// lv % rv
	remAssign(lv, rv)
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
	shlAssign(lv, rv)
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
	shrAssign(lv, rv)
}

func (m *Machine) doOpBand() {
	m.PopExpr()

	// get right and left operands.
	rv := m.PopValue()
	lv := m.PeekValue(1) // also result
	if debug {
		assertTypes(lv.T, rv.T)
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
		assertTypes(lv.T, rv.T)
	}

	// lv &^ rv
	bandnAssign(lv, rv)
}

//----------------------------------------
// logic functions

// TODO: document what class of problems its for.
// One of them can be nil, and this lets uninitialized primitives and
// others serve as empty values.  See doOpAdd()
func assertTypes(lt, rt Type) {
	if lt == nil && rt == nil {
		panic("assertTypes() requires at least one type")
	} else if lt == nil || rt == nil {
		// one is nil.
	} else if lt.Kind() == rt.Kind() &&
		isUntyped(lt) || isUntyped(rt) {
		// one is untyped of same kind.
	} else if lt.TypeID() != rt.TypeID() {
		panic(fmt.Sprintf(
			"incompatible operands in binary expression: %s and %s",
			lt.String(),
			rt.String(),
		))
	} else {
		// non-nil types are identical.
	}
}

// TODO: can be much faster.
func isEql(lv, rv *TypedValue) bool {
	switch lv.T.Kind() {
	case BoolKind:
		return (lv.GetBool() == rv.GetBool())
	case StringKind:
		return (lv.V.(StringValue) == rv.V.(StringValue))
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
	case BigintKind:
		lb := lv.V.(BigintValue).V
		rb := rv.V.(BigintValue).V
		return lb.Cmp(rb) == 0
	case ArrayKind:
		la := lv.V.(*ArrayValue).List
		ra := rv.V.(*ArrayValue).List
		if debug {
			if len(la) != len(ra) {
				panic("comparison on arrays of unequal length")
			}
			if lv.T.(*ArrayType).Elt.TypeID() !=
				rv.T.(*ArrayType).Elt.TypeID() {
				panic("comparison on arrays of unequal type")
			}
		}
		for i := 0; i < len(la); i++ {
			if !isEql(&la[i], &ra[i]) {
				return false
			}
		}
		return true
	default:
		panic(fmt.Sprintf(
			"comparison operator == not defined for %s kind",
			KindOf(lv.T),
		))
	}
}

// TODO: can be much faster.
func isLss(lv, rv *TypedValue) bool {
	switch baseOf(lv.T) {
	case StringType:
		return (lv.V.(StringValue) < rv.V.(StringValue))
	case IntType:
		return (lv.GetInt() < rv.GetInt())
	case Int8Type:
		return (lv.GetInt8() < rv.GetInt8())
	case Int16Type:
		return (lv.GetInt16() < rv.GetInt16())
	case Int32Type:
		return (lv.GetInt32() < rv.GetInt32())
	case Int64Type:
		return (lv.GetInt64() < rv.GetInt64())
	case UintType:
		return (lv.GetUint() < rv.GetUint())
	case Uint8Type:
		return (lv.GetUint8() < rv.GetUint8())
	case Uint16Type:
		return (lv.GetUint16() < rv.GetUint16())
	case Uint32Type:
		return (lv.GetUint32() < rv.GetUint32())
	case Uint64Type:
		return (lv.GetUint64() < rv.GetUint64())
	case BigintType:
		lb := lv.V.(BigintValue).V
		rb := rv.V.(BigintValue).V
		return lb.Cmp(rb) < 0
	default:
		panic(fmt.Sprintf(
			"comparison operator < not defined for %s kind",
			KindOf(lv.T),
		))
	}
}

func isLeq(lv, rv *TypedValue) bool {
	switch baseOf(lv.T) {
	case StringType:
		return (lv.V.(StringValue) <= rv.V.(StringValue))
	case IntType:
		return (lv.GetInt() <= rv.GetInt())
	case Int8Type:
		return (lv.GetInt8() <= rv.GetInt8())
	case Int16Type:
		return (lv.GetInt16() <= rv.GetInt16())
	case Int32Type:
		return (lv.GetInt32() <= rv.GetInt32())
	case Int64Type:
		return (lv.GetInt64() <= rv.GetInt64())
	case UintType:
		return (lv.GetUint() <= rv.GetUint())
	case Uint8Type:
		return (lv.GetUint8() <= rv.GetUint8())
	case Uint16Type:
		return (lv.GetUint16() <= rv.GetUint16())
	case Uint32Type:
		return (lv.GetUint32() <= rv.GetUint32())
	case Uint64Type:
		return (lv.GetUint64() <= rv.GetUint64())
	case BigintType:
		lb := lv.V.(BigintValue).V
		rb := rv.V.(BigintValue).V
		return lb.Cmp(rb) <= 0
	default:
		panic(fmt.Sprintf(
			"comparison operator <= not defined for %s kind",
			KindOf(lv.T),
		))
	}
}

func isGtr(lv, rv *TypedValue) bool {
	switch baseOf(lv.T) {
	case StringType:
		return (lv.V.(StringValue) > rv.V.(StringValue))
	case IntType:
		return (lv.GetInt() > rv.GetInt())
	case Int8Type:
		return (lv.GetInt8() > rv.GetInt8())
	case Int16Type:
		return (lv.GetInt16() > rv.GetInt16())
	case Int32Type:
		return (lv.GetInt32() > rv.GetInt32())
	case Int64Type:
		return (lv.GetInt64() > rv.GetInt64())
	case UintType:
		return (lv.GetUint() > rv.GetUint())
	case Uint8Type:
		return (lv.GetUint8() > rv.GetUint8())
	case Uint16Type:
		return (lv.GetUint16() > rv.GetUint16())
	case Uint32Type:
		return (lv.GetUint32() > rv.GetUint32())
	case Uint64Type:
		return (lv.GetUint64() > rv.GetUint64())
	case BigintType:
		lb := lv.V.(BigintValue).V
		rb := rv.V.(BigintValue).V
		return lb.Cmp(rb) > 0
	default:
		panic(fmt.Sprintf(
			"comparison operator > not defined for %s kind",
			KindOf(lv.T),
		))
	}
}

func isGeq(lv, rv *TypedValue) bool {
	switch baseOf(lv.T) {
	case StringType:
		return (lv.V.(StringValue) >= rv.V.(StringValue))
	case IntType:
		return (lv.GetInt() >= rv.GetInt())
	case Int8Type:
		return (lv.GetInt8() >= rv.GetInt8())
	case Int16Type:
		return (lv.GetInt16() >= rv.GetInt16())
	case Int32Type:
		return (lv.GetInt32() >= rv.GetInt32())
	case Int64Type:
		return (lv.GetInt64() >= rv.GetInt64())
	case UintType:
		return (lv.GetUint() >= rv.GetUint())
	case Uint8Type:
		return (lv.GetUint8() >= rv.GetUint8())
	case Uint16Type:
		return (lv.GetUint16() >= rv.GetUint16())
	case Uint32Type:
		return (lv.GetUint32() >= rv.GetUint32())
	case Uint64Type:
		return (lv.GetUint64() >= rv.GetUint64())
	case BigintType:
		lb := lv.V.(BigintValue).V
		rb := rv.V.(BigintValue).V
		return lb.Cmp(rb) >= 0
	default:
		panic(fmt.Sprintf(
			"comparison operator >= not defined for %s kind",
			KindOf(lv.T),
		))
	}
}

// for doOpAdd and doOpAddAssign.
func addAssign(lv, rv *TypedValue) {
	// set the result in lv.
	// NOTE this block is replicated in op_assign.go
	switch baseOf(lv.T) {
	case UntypedStringType, StringType:
		lv.V = lv.GetString() + rv.GetString()
	case IntType:
		lv.SetInt(lv.GetInt() + rv.GetInt())
	case Int8Type:
		lv.SetInt8(lv.GetInt8() + rv.GetInt8())
	case Int16Type:
		lv.SetInt16(lv.GetInt16() + rv.GetInt16())
	case UntypedRuneType, Int32Type:
		lv.SetInt32(lv.GetInt32() + rv.GetInt32())
	case Int64Type:
		lv.SetInt64(lv.GetInt64() + rv.GetInt64())
	case UintType:
		lv.SetUint(lv.GetUint() + rv.GetUint())
	case Uint8Type:
		lv.SetUint8(lv.GetUint8() + rv.GetUint8())
	case Uint16Type:
		lv.SetUint16(lv.GetUint16() + rv.GetUint16())
	case Uint32Type:
		lv.SetUint32(lv.GetUint32() + rv.GetUint32())
	case Uint64Type:
		lv.SetUint64(lv.GetUint64() + rv.GetUint64())
	case UntypedBigintType, BigintType:
		lb := lv.GetBig()
		lb.Add(lb, rv.GetBig())
	default:
		panic(fmt.Sprintf(
			"operator + not defined for %s kind %s",
			KindOf(lv.T),
			reflect.TypeOf(baseOf(lv.T)),
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
	case UntypedRuneType, Int32Type:
		lv.SetInt32(lv.GetInt32() - rv.GetInt32())
	case Int64Type:
		lv.SetInt64(lv.GetInt64() - rv.GetInt64())
	case UintType:
		lv.SetUint(lv.GetUint() - rv.GetUint())
	case Uint8Type:
		lv.SetUint8(lv.GetUint8() - rv.GetUint8())
	case Uint16Type:
		lv.SetUint16(lv.GetUint16() - rv.GetUint16())
	case Uint32Type:
		lv.SetUint32(lv.GetUint32() - rv.GetUint32())
	case Uint64Type:
		lv.SetUint64(lv.GetUint64() - rv.GetUint64())
	case UntypedBigintType, BigintType:
		panic("not yet implemented")
	default:
		panic(fmt.Sprintf(
			"operators - and -= not defined for %s kind %s",
			KindOf(lv.T),
			reflect.TypeOf(baseOf(lv.T)),
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
	case UntypedRuneType, Int32Type:
		lv.SetInt32(lv.GetInt32() * rv.GetInt32())
	case Int64Type:
		lv.SetInt64(lv.GetInt64() * rv.GetInt64())
	case UintType:
		lv.SetUint(lv.GetUint() * rv.GetUint())
	case Uint8Type:
		lv.SetUint8(lv.GetUint8() * rv.GetUint8())
	case Uint16Type:
		lv.SetUint16(lv.GetUint16() * rv.GetUint16())
	case Uint32Type:
		lv.SetUint32(lv.GetUint32() * rv.GetUint32())
	case Uint64Type:
		lv.SetUint64(lv.GetUint64() * rv.GetUint64())
	case UntypedBigintType, BigintType:
		panic("not yet implemented")
	default:
		panic(fmt.Sprintf(
			"operators * and *= not defined for %s kind %s",
			KindOf(lv.T),
			reflect.TypeOf(baseOf(lv.T)),
		))
	}
}

// for doOpQuo and doOpQuoAssign.
func quoAssign(lv, rv *TypedValue) {
	// set the result in lv.
	// NOTE this block is replicated in op_assign.go
	switch baseOf(lv.T) {
	case IntType:
		lv.SetInt(lv.GetInt() / rv.GetInt())
	case Int8Type:
		lv.SetInt8(lv.GetInt8() / rv.GetInt8())
	case Int16Type:
		lv.SetInt16(lv.GetInt16() / rv.GetInt16())
	case UntypedRuneType, Int32Type:
		lv.SetInt32(lv.GetInt32() / rv.GetInt32())
	case Int64Type:
		lv.SetInt64(lv.GetInt64() / rv.GetInt64())
	case UintType:
		lv.SetUint(lv.GetUint() / rv.GetUint())
	case Uint8Type:
		lv.SetUint8(lv.GetUint8() / rv.GetUint8())
	case Uint16Type:
		lv.SetUint16(lv.GetUint16() / rv.GetUint16())
	case Uint32Type:
		lv.SetUint32(lv.GetUint32() / rv.GetUint32())
	case Uint64Type:
		lv.SetUint64(lv.GetUint64() / rv.GetUint64())
	case UntypedBigintType, BigintType:
		panic("not yet implemented")
	default:
		panic(fmt.Sprintf(
			"operators / and /= not defined for %s kind %s",
			KindOf(lv.T),
			reflect.TypeOf(baseOf(lv.T)),
		))
	}
}

// for doOpRem and doOpRemAssign.
func remAssign(lv, rv *TypedValue) {
	// set the result in lv.
	// NOTE this block is replicated in op_assign.go
	switch baseOf(lv.T) {
	case IntType:
		lv.SetInt(lv.GetInt() % rv.GetInt())
	case Int8Type:
		lv.SetInt8(lv.GetInt8() % rv.GetInt8())
	case Int16Type:
		lv.SetInt16(lv.GetInt16() % rv.GetInt16())
	case UntypedRuneType, Int32Type:
		lv.SetInt32(lv.GetInt32() % rv.GetInt32())
	case Int64Type:
		lv.SetInt64(lv.GetInt64() % rv.GetInt64())
	case UintType:
		lv.SetUint(lv.GetUint() % rv.GetUint())
	case Uint8Type:
		lv.SetUint8(lv.GetUint8() % rv.GetUint8())
	case Uint16Type:
		lv.SetUint16(lv.GetUint16() % rv.GetUint16())
	case Uint32Type:
		lv.SetUint32(lv.GetUint32() % rv.GetUint32())
	case Uint64Type:
		lv.SetUint64(lv.GetUint64() % rv.GetUint64())
	case UntypedBigintType, BigintType:
		panic("not yet implemented")
	default:
		panic(fmt.Sprintf(
			"operators %% and %%= not defined for %s kind %s",
			KindOf(lv.T),
			reflect.TypeOf(baseOf(lv.T)),
		))
	}
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
	case UntypedRuneType, Int32Type:
		lv.SetInt32(lv.GetInt32() & rv.GetInt32())
	case Int64Type:
		lv.SetInt64(lv.GetInt64() & rv.GetInt64())
	case UintType:
		lv.SetUint(lv.GetUint() & rv.GetUint())
	case Uint8Type:
		lv.SetUint8(lv.GetUint8() & rv.GetUint8())
	case Uint16Type:
		lv.SetUint16(lv.GetUint16() & rv.GetUint16())
	case Uint32Type:
		lv.SetUint32(lv.GetUint32() & rv.GetUint32())
	case Uint64Type:
		lv.SetUint64(lv.GetUint64() & rv.GetUint64())
	case UntypedBigintType, BigintType:
		panic("not yet implemented")
	default:
		panic(fmt.Sprintf(
			"operators & and &= not defined for %s kind %s",
			KindOf(lv.T),
			reflect.TypeOf(baseOf(lv.T)),
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
	case UntypedRuneType, Int32Type:
		lv.SetInt32(lv.GetInt32() &^ rv.GetInt32())
	case Int64Type:
		lv.SetInt64(lv.GetInt64() &^ rv.GetInt64())
	case UintType:
		lv.SetUint(lv.GetUint() &^ rv.GetUint())
	case Uint8Type:
		lv.SetUint8(lv.GetUint8() &^ rv.GetUint8())
	case Uint16Type:
		lv.SetUint16(lv.GetUint16() &^ rv.GetUint16())
	case Uint32Type:
		lv.SetUint32(lv.GetUint32() &^ rv.GetUint32())
	case Uint64Type:
		lv.SetUint64(lv.GetUint64() &^ rv.GetUint64())
	case UntypedBigintType, BigintType:
		panic("not yet implemented")
	default:
		panic(fmt.Sprintf(
			"operators &^ and &^= not defined for %s kind %s",
			KindOf(lv.T),
			reflect.TypeOf(baseOf(lv.T)),
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
	case UntypedRuneType, Int32Type:
		lv.SetInt32(lv.GetInt32() | rv.GetInt32())
	case Int64Type:
		lv.SetInt64(lv.GetInt64() | rv.GetInt64())
	case UintType:
		lv.SetUint(lv.GetUint() | rv.GetUint())
	case Uint8Type:
		lv.SetUint8(lv.GetUint8() | rv.GetUint8())
	case Uint16Type:
		lv.SetUint16(lv.GetUint16() | rv.GetUint16())
	case Uint32Type:
		lv.SetUint32(lv.GetUint32() | rv.GetUint32())
	case Uint64Type:
		lv.SetUint64(lv.GetUint64() | rv.GetUint64())
	case UntypedBigintType, BigintType:
		panic("not yet implemented")
	default:
		panic(fmt.Sprintf(
			"operators | and |= not defined for %s kind %s",
			KindOf(lv.T),
			reflect.TypeOf(baseOf(lv.T)),
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
	case UntypedRuneType, Int32Type:
		lv.SetInt32(lv.GetInt32() ^ rv.GetInt32())
	case Int64Type:
		lv.SetInt64(lv.GetInt64() ^ rv.GetInt64())
	case UintType:
		lv.SetUint(lv.GetUint() ^ rv.GetUint())
	case Uint8Type:
		lv.SetUint8(lv.GetUint8() ^ rv.GetUint8())
	case Uint16Type:
		lv.SetUint16(lv.GetUint16() ^ rv.GetUint16())
	case Uint32Type:
		lv.SetUint32(lv.GetUint32() ^ rv.GetUint32())
	case Uint64Type:
		lv.SetUint64(lv.GetUint64() ^ rv.GetUint64())
	case UntypedBigintType, BigintType:
		panic("not yet implemented")
	default:
		panic(fmt.Sprintf(
			"operators ^ and ^= not defined for %s kind %s",
			KindOf(lv.T),
			reflect.TypeOf(baseOf(lv.T)),
		))
	}
}

// for doOpShl and doOpShlAssign.
func shlAssign(lv, rv *TypedValue) {
	// set the result in lv.
	// NOTE: baseOf(rv.T) is always UintType.
	switch baseOf(lv.T) {
	case IntType:
		lv.SetInt(lv.GetInt() << rv.GetUint())
	case Int8Type:
		lv.SetInt8(lv.GetInt8() << rv.GetUint())
	case Int16Type:
		lv.SetInt16(lv.GetInt16() << rv.GetUint())
	case UntypedRuneType, Int32Type:
		lv.SetInt32(lv.GetInt32() << rv.GetUint())
	case Int64Type:
		lv.SetInt64(lv.GetInt64() << rv.GetUint())
	case UintType:
		lv.SetUint(lv.GetUint() << rv.GetUint())
	case Uint8Type:
		lv.SetUint8(lv.GetUint8() << rv.GetUint())
	case Uint16Type:
		lv.SetUint16(lv.GetUint16() << rv.GetUint())
	case Uint32Type:
		lv.SetUint32(lv.GetUint32() << rv.GetUint())
	case Uint64Type:
		lv.SetUint64(lv.GetUint64() << rv.GetUint())
	case UntypedBigintType, BigintType:
		lbi := lv.V.(BigintValue).V
		lbi.Lsh(lbi, rv.GetUint())
	default:
		panic(fmt.Sprintf(
			"operators << and <<= not defined for %s kind %s",
			KindOf(lv.T),
			reflect.TypeOf(baseOf(lv.T)),
		))
	}
}

// for doOpShr and doOpShrAssign.
func shrAssign(lv, rv *TypedValue) {
	// set the result in lv.
	// NOTE: baseOf(rv.T) is always UintType.
	switch baseOf(lv.T) {
	case IntType:
		lv.SetInt(lv.GetInt() >> rv.GetUint())
	case Int8Type:
		lv.SetInt8(lv.GetInt8() >> rv.GetUint())
	case Int16Type:
		lv.SetInt16(lv.GetInt16() >> rv.GetUint())
	case UntypedRuneType, Int32Type:
		lv.SetInt32(lv.GetInt32() >> rv.GetUint())
	case Int64Type:
		lv.SetInt64(lv.GetInt64() >> rv.GetUint())
	case UintType:
		lv.SetUint(lv.GetUint() >> rv.GetUint())
	case Uint8Type:
		lv.SetUint8(lv.GetUint8() >> rv.GetUint())
	case Uint16Type:
		lv.SetUint16(lv.GetUint16() >> rv.GetUint())
	case Uint32Type:
		lv.SetUint32(lv.GetUint32() >> rv.GetUint())
	case Uint64Type:
		lv.SetUint64(lv.GetUint64() >> rv.GetUint())
	case UntypedBigintType, BigintType:
		lbi := lv.V.(BigintValue).V
		lbi.Rsh(lbi, rv.GetUint())
	default:
		panic(fmt.Sprintf(
			"operators >> and >>= not defined for %s kind %s",
			KindOf(lv.T),
			reflect.TypeOf(baseOf(lv.T)),
		))
	}
}
