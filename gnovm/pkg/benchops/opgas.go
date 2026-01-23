package benchops

// OpGas maps opcode bytes to their gas costs.
//
// These values are derived from the OpCPU* constants in gnolang/machine.go.
// They represent the relative CPU cost of each VM opcode, used for gas metering.
//
// IMPORTANT: When updating these values, ensure they stay in sync with the
// OpCPU* constants in gnolang/machine.go. The benchops/cmd tool can be used
// to re-calibrate these values through benchmarking.
var OpGas = [256]int64{
	// ---- Control Operators (0x00-0x1F)
	0x00: 1,   // OpInvalid
	0x01: 1,   // OpHalt
	0x02: 1,   // OpNoop
	0x03: 25,  // OpExec
	0x04: 207, // OpPrecall
	0x05: 100, // OpEnterCrossing
	0x06: 256, // OpCall
	0x07: 424, // OpCallNativeBody
	0x0A: 64,  // OpDefer
	0x0B: 33,  // OpCallDeferNativeBody
	0x0C: 1,   // OpGo (not implemented)
	0x0D: 1,   // OpSelect (not implemented)
	0x0E: 38,  // OpSwitchClause
	0x0F: 143, // OpSwitchClauseCase
	0x10: 171, // OpTypeSwitch
	0x11: 38,  // OpIfCond
	0x12: 1,   // OpPopValue
	0x13: 1,   // OpPopResults
	0x14: 3,   // OpPopBlock
	0x15: 15,  // OpPopFrameAndReset
	0x1A: 38,  // OpReturn
	0x1B: 38,  // OpReturnAfterCopy
	0x1C: 36,  // OpReturnFromBlock
	0x1D: 23,  // OpReturnToBlock

	// ---- Unary Operators (0x20-0x25)
	0x20: 7,  // OpUpos
	0x21: 25, // OpUneg
	0x22: 6,  // OpUnot
	0x23: 14, // OpUxor
	0x25: 1,  // OpUrecv (not implemented)

	// ---- Binary Operators (0x26-0x38)
	0x26: 26,  // OpLor
	0x27: 24,  // OpLand
	0x28: 160, // OpEql
	0x29: 95,  // OpNeq
	0x2A: 13,  // OpLss
	0x2B: 19,  // OpLeq
	0x2C: 20,  // OpGtr
	0x2D: 26,  // OpGeq
	0x2E: 18,  // OpAdd
	0x2F: 6,   // OpSub
	0x30: 23,  // OpBor
	0x31: 13,  // OpXor
	0x32: 19,  // OpMul
	0x33: 16,  // OpQuo
	0x34: 18,  // OpRem
	0x35: 22,  // OpShl
	0x36: 20,  // OpShr
	0x37: 9,   // OpBand
	0x38: 15,  // OpBandn

	// ---- Expression Operators (0x40-0x52)
	0x40: 29,  // OpEval
	0x41: 18,  // OpBinary1
	0x42: 14,  // OpIndex1
	0x43: 20,  // OpIndex2
	0x44: 32,  // OpSelector
	0x45: 103, // OpSlice
	0x46: 40,  // OpStar
	0x47: 125, // OpRef
	0x48: 22,  // OpTypeAssert1
	0x49: 34,  // OpTypeAssert2
	0x4A: 100, // OpStaticTypeOf
	0x4B: 50,  // OpCompositeLit
	0x4C: 137, // OpArrayLit
	0x4D: 183, // OpSliceLit
	0x4E: 92,  // OpSliceLit2
	0x4F: 475, // OpMapLit
	0x50: 179, // OpStructLit
	0x51: 61,  // OpFuncLit
	0x52: 16,  // OpConvert

	// ---- Type Operators (0x70-0x78)
	0x70: 59,  // OpFieldType
	0x71: 57,  // OpArrayType
	0x72: 55,  // OpSliceType
	0x73: 1,   // OpPointerType (not implemented)
	0x74: 75,  // OpInterfaceType
	0x75: 57,  // OpChanType
	0x76: 81,  // OpFuncType
	0x77: 59,  // OpMapType
	0x78: 174, // OpStructType

	// ---- Assignment Operators (0x80-0x8E)
	0x80: 79,  // OpAssign
	0x81: 85,  // OpAddAssign
	0x82: 57,  // OpSubAssign
	0x83: 55,  // OpMulAssign
	0x84: 50,  // OpQuoAssign
	0x85: 46,  // OpRemAssign
	0x86: 54,  // OpBandAssign
	0x87: 44,  // OpBandnAssign
	0x88: 55,  // OpBorAssign
	0x89: 48,  // OpXorAssign
	0x8A: 68,  // OpShlAssign
	0x8B: 76,  // OpShrAssign
	0x8C: 111, // OpDefine
	0x8D: 76,  // OpInc
	0x8E: 46,  // OpDec

	// ---- Declaration Operators (0x90-0x91)
	0x90: 113, // OpValueDecl
	0x91: 100, // OpTypeDecl

	// ---- Sticky/Loop Operators (0xD0+)
	0xD0: 1,   // OpSticky (not a real op)
	0xD1: 43,  // OpBody
	0xD2: 27,  // OpForLoop
	0xD3: 105, // OpRangeIter
	0xD4: 55,  // OpRangeIterString
	0xD5: 48,  // OpRangeIterMap
	0xD6: 46,  // OpRangeIterArrayPtr
	0xD7: 78,  // OpReturnCallDefers

	0xFF: 1, // OpVoid (profiling)
}

// GetOpGas returns the gas cost for an opcode.
func GetOpGas(op Op) int64 {
	gas := OpGas[op]
	if gas == 0 {
		return 1 // default minimum gas
	}
	return gas
}
