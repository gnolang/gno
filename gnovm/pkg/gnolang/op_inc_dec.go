package gnolang

import (
	"fmt"
	"github.com/cockroachdb/apd/v3"
	"math/big"
)

func (m *Machine) doOpInc() {
	s := m.PopStmt().(*IncDecStmt)

	// Get reference to lhs.
	pv := m.PopAsPointer(s.X)
	lv := pv.TV

	// Switch on the base type.  NOTE: this is faster
	// than computing the kind of kv.T.  TODO: consider
	// optimizing away this switch by implementing a
	// general SetAnyInt(n int64) function that handles
	// bounds checking.  NOTE: no need to set .V to nil,
	// as the type should be the same, and thus .V is
	// expected to be nil.
	if debug {
		if lv.V != nil {
			panic("expected lv.V to be nil for primitive type for OpInc")
		}
	}

	// here we can't just switch on the value type
	// because it could be a type alias
	// type num int
	switch baseOf(lv.T) {
	case IntType:
		lv.SetInt(lv.GetInt() + 1)
	case Int8Type:
		lv.SetInt8(lv.GetInt8() + 1)
	case Int16Type:
		lv.SetInt16(lv.GetInt16() + 1)
	case Int32Type:
		lv.SetInt32(lv.GetInt32() + 1)
	case Int64Type:
		lv.SetInt64(lv.GetInt64() + 1)
	case UintType:
		lv.SetUint(lv.GetUint() + 1)
	case Uint8Type:
		lv.SetUint8(lv.GetUint8() + 1)
	case DataByteType:
		lv.SetDataByte(lv.GetDataByte() + 1)
	case Uint16Type:
		lv.SetUint16(lv.GetUint16() + 1)
	case Uint32Type:
		lv.SetUint32(lv.GetUint32() + 1)
	case Uint64Type:
		lv.SetUint64(lv.GetUint64() + 1)
	case Float32Type:
		lv.SetFloat32(lv.GetFloat32() + 1)
	case Float64Type:
		lv.SetFloat64(lv.GetFloat64() + 1)
	case BigintType, UntypedBigintType:
		lb := lv.GetBigInt()
		lb = big.NewInt(0).Add(lb, big.NewInt(1))
		lv.V = BigintValue{V: lb}
	case BigdecType, UntypedBigdecType:
		lb := lv.GetBigDec()
		sum := apd.New(0, 0)
		cond, err := apd.BaseContext.WithPrecision(0).Add(sum, lb, apd.New(1, 0))
		if err != nil {
			panic(fmt.Sprintf("bigdec addition error: %v", err))
		} else if cond.Inexact() {
			panic(fmt.Sprintf("bigdec addition inexact: %v + 1", lb))
		}
		lv.V = BigdecValue{V: sum}
	default:
		panic(fmt.Sprintf("unexpected type %s in inc/dec operation", lv.T))
	}

	// Mark dirty in realm.
	if m.Realm != nil && pv.Base != nil {
		m.Realm.DidUpdate(pv.Base.(Object), nil, nil)
	}
}

func (m *Machine) doOpDec() {
	s := m.PopStmt().(*IncDecStmt)

	// Get result ptr depending on lhs.
	pv := m.PopAsPointer(s.X)
	lv := pv.TV

	// Switch on the base type.  NOTE: this is faster
	// than computing the kind of kv.T.  TODO: consider
	// optimizing away this switch by implementing a
	// general SetAnyInt(n int64) function that handles
	// bounds checking.  NOTE: no need to set .V to nil,
	// as the type should be the same, and thus .V is
	// expected to be nil.
	if debug {
		if lv.V != nil {
			panic("expected lv.V to be nil for primitive type for OpDec")
		}
	}
	switch baseOf(lv.T) {
	case IntType:
		lv.SetInt(lv.GetInt() - 1)
	case Int8Type:
		lv.SetInt8(lv.GetInt8() - 1)
	case Int16Type:
		lv.SetInt16(lv.GetInt16() - 1)
	case Int32Type:
		lv.SetInt32(lv.GetInt32() - 1)
	case Int64Type:
		lv.SetInt64(lv.GetInt64() - 1)
	case UintType:
		lv.SetUint(lv.GetUint() - 1)
	case Uint8Type:
		lv.SetUint8(lv.GetUint8() - 1)
	case DataByteType:
		lv.SetDataByte(lv.GetDataByte() - 1)
	case Uint16Type:
		lv.SetUint16(lv.GetUint16() - 1)
	case Uint32Type:
		lv.SetUint32(lv.GetUint32() - 1)
	case Uint64Type:
		lv.SetUint64(lv.GetUint64() - 1)
	case Float32Type:
		lv.SetFloat32(lv.GetFloat32() - 1)
	case Float64Type:
		lv.SetFloat64(lv.GetFloat64() - 1)
	case BigintType, UntypedBigintType:
		lb := lv.GetBigInt()
		lb = big.NewInt(0).Sub(lb, big.NewInt(1))
		lv.V = BigintValue{V: lb}
	case BigdecType, UntypedBigdecType:
		lb := lv.GetBigDec()
		sum := apd.New(0, 0)
		cond, err := apd.BaseContext.WithPrecision(0).Sub(sum, lb, apd.New(1, 0))
		if err != nil {
			panic(fmt.Sprintf("bigdec addition error: %v", err))
		} else if cond.Inexact() {
			panic(fmt.Sprintf("bigdec addition inexact: %v + 1", lb))
		}
		lv.V = BigdecValue{V: sum}
	default:
		panic(fmt.Sprintf("unexpected type %s in inc/dec operation", lv.T))
	}

	// Mark dirty in realm.
	if m.Realm != nil && pv.Base != nil {
		m.Realm.DidUpdate(pv.Base.(Object), nil, nil)
	}
}
