package gnolang

import (
	"fmt"
	"math/big"

	"github.com/cockroachdb/apd"
)

func (m *Machine) doOpUpos() {
	ux := m.PopExpr().(*UnaryExpr)
	if debug {
		debug.Printf("doOpUpos(%v)\n", ux)
	}
	// nothing to do, +x is just x?
}

func (m *Machine) doOpUneg() {
	ux := m.PopExpr().(*UnaryExpr)
	if debug {
		debug.Printf("doOpUneg(%v)\n", ux)
	}
	xv := m.PeekValue(1)

	// Switch on the base type.
	// NOTE: this is faster than computing the kind of kv.T.
	switch baseOf(xv.T) {
	case IntType:
		xv.SetInt(-xv.GetInt())
	case Int8Type:
		xv.SetInt8(-xv.GetInt8())
	case Int16Type:
		xv.SetInt16(-xv.GetInt16())
	case Int32Type:
		xv.SetInt32(-xv.GetInt32())
	case Int64Type:
		xv.SetInt64(-xv.GetInt64())
	case UintType:
		xv.SetUint(-xv.GetUint())
	case Uint8Type:
		xv.SetUint8(-xv.GetUint8())
	case Uint16Type:
		xv.SetUint16(-xv.GetUint16())
	case Uint32Type:
		xv.SetUint32(-xv.GetUint32())
	case Uint64Type:
		xv.SetUint64(-xv.GetUint64())
	case Float32Type:
		xv.SetFloat32(-xv.GetFloat32())
	case Float64Type:
		xv.SetFloat64(-xv.GetFloat64())
	case UntypedBigintType, BigintType:
		bv := xv.V.(BigintValue)
		xv.V = BigintValue{V: new(big.Int).Neg(bv.V)}
	case UntypedBigdecType, BigdecType:
		bv := xv.V.(BigdecValue)
		xv.V = BigdecValue{V: apd.New(0, 0).Neg(bv.V)}
	case nil:
		// NOTE: for now only BigintValue is possible.
		bv := xv.V.(BigintValue)
		xv.V = BigintValue{V: new(big.Int).Neg(bv.V)}
	default:
		panic(fmt.Sprintf("unexpected type %s in operation",
			baseOf(xv.T)))
	}
}

func (m *Machine) doOpUnot() {
	ux := m.PopExpr().(*UnaryExpr)
	if debug {
		debug.Printf("doOpUnot(%v)\n", ux)
	}
	xv := m.PeekValue(1)

	// Switch on the base type.
	switch baseOf(xv.T) {
	case BoolType, UntypedBoolType:
		xv.SetBool(!xv.GetBool())
	default:
		panic(fmt.Sprintf("unexpected type %s in operation",
			baseOf(xv.T)))
	}
}

func (m *Machine) doOpUxor() {
	ux := m.PopExpr().(*UnaryExpr)
	if debug {
		debug.Printf("doOpUxor(%v)\n", ux)
	}
	xv := m.PeekValue(1)

	// Switch on the base type.
	switch baseOf(xv.T) {
	case IntType:
		xv.SetInt(^xv.GetInt())
	case Int8Type:
		xv.SetInt8(^xv.GetInt8())
	case Int16Type:
		xv.SetInt16(^xv.GetInt16())
	case Int32Type:
		xv.SetInt32(^xv.GetInt32())
	case Int64Type:
		xv.SetInt64(^xv.GetInt64())
	case UintType:
		xv.SetUint(^xv.GetUint())
	case Uint8Type:
		xv.SetUint8(^xv.GetUint8())
	case Uint16Type:
		xv.SetUint16(^xv.GetUint16())
	case Uint32Type:
		xv.SetUint32(^xv.GetUint32())
	case Uint64Type:
		xv.SetUint64(^xv.GetUint64())
	case UntypedBigintType, BigintType:
		// XXX can it even be implemented?
		panic("not yet implemented")
	default:
		panic(fmt.Sprintf("unexpected type %s in operation",
			baseOf(xv.T)))
	}
}

func (m *Machine) doOpUrecv() {
	panic("not yet implemented")
}
