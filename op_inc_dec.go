package gno

func (m *Machine) doOpInc() {
	s := m.PopStmt().(*IncDecStmt)

	// Get reference to lhs.
	lv := m.PopAsPointer(s.X).TV

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
	default:
		panic("unexpected type in in operation")
	}
}

func (m *Machine) doOpDec() {
	s := m.PopStmt().(*IncDecStmt)

	// Get result ptr depending on lhs.
	lv := m.PopAsPointer(s.X).TV

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
	default:
		panic("unexpected type in in operation")
	}
}
