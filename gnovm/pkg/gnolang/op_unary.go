package gnolang

import (
	"fmt"
	"math/big"

	"github.com/cockroachdb/apd/v3"
	"github.com/gnolang/gno/gnovm/pkg/gnolang/internal/softfloat"
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
		xv.SetFloat32(softfloat.Fneg32(xv.GetFloat32()))
	case Float64Type:
		xv.SetFloat64(softfloat.Fneg64(xv.GetFloat64()))
	case UntypedBigintType:
		biv := xv.V.(BigintValue)
		xv.V = BigintValue{V: new(big.Int).Neg(biv.V)}
	case UntypedBigdecType:
		bdv := xv.V.(BigdecValue)
		xv.V = BigdecValue{V: apd.New(0, 0).Neg(bdv.V)}
	case nil:
		// NOTE: for now only BigintValue is possible.
		biv := xv.V.(BigintValue)
		xv.V = BigintValue{V: new(big.Int).Neg(biv.V)}
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
		panic(fmt.Sprintf("unexpected type %v in operation",
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
	case UntypedBigintType:
		bv := xv.V.(BigintValue)
		xv.V = BigintValue{V: new(big.Int).Not(bv.V)}
	default:
		panic(fmt.Sprintf("unexpected type %s in operation",
			baseOf(xv.T)))
	}
}

func (m *Machine) doOpUrecv() {
	panic("not yet implemented")
}
