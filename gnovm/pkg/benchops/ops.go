package benchops

import "fmt"

// Op represents a GnoVM opcode.
type Op byte

// GnoVM opcode constants.
const (
	// ---- Control Operators
	OpInvalid             Op = 0x00
	OpHalt                Op = 0x01
	OpNoop                Op = 0x02
	OpExec                Op = 0x03
	OpPrecall             Op = 0x04
	OpEnterCrossing       Op = 0x05
	OpCall                Op = 0x06
	OpCallNativeBody      Op = 0x07
	OpDefer               Op = 0x0A
	OpCallDeferNativeBody Op = 0x0B
	OpGo                  Op = 0x0C
	OpSelect              Op = 0x0D
	OpSwitchClause        Op = 0x0E
	OpSwitchClauseCase    Op = 0x0F
	OpTypeSwitch          Op = 0x10
	OpIfCond              Op = 0x11
	OpPopValue            Op = 0x12
	OpPopResults          Op = 0x13
	OpPopBlock            Op = 0x14
	OpPopFrameAndReset    Op = 0x15
	OpPanic1              Op = 0x16
	OpPanic2              Op = 0x17
	OpReturn              Op = 0x1A
	OpReturnAfterCopy     Op = 0x1B
	OpReturnFromBlock     Op = 0x1C
	OpReturnToBlock       Op = 0x1D

	// ---- Unary Operators
	OpUpos  Op = 0x20
	OpUneg  Op = 0x21
	OpUnot  Op = 0x22
	OpUxor  Op = 0x23
	OpUrecv Op = 0x25

	// ---- Binary Operators
	OpLor   Op = 0x26
	OpLand  Op = 0x27
	OpEql   Op = 0x28
	OpNeq   Op = 0x29
	OpLss   Op = 0x2A
	OpLeq   Op = 0x2B
	OpGtr   Op = 0x2C
	OpGeq   Op = 0x2D
	OpAdd   Op = 0x2E
	OpSub   Op = 0x2F
	OpBor   Op = 0x30
	OpXor   Op = 0x31
	OpMul   Op = 0x32
	OpQuo   Op = 0x33
	OpRem   Op = 0x34
	OpShl   Op = 0x35
	OpShr   Op = 0x36
	OpBand  Op = 0x37
	OpBandn Op = 0x38

	// ---- Expression Operators
	OpEval         Op = 0x40
	OpBinary1      Op = 0x41
	OpIndex1       Op = 0x42
	OpIndex2       Op = 0x43
	OpSelector     Op = 0x44
	OpSlice        Op = 0x45
	OpStar         Op = 0x46
	OpRef          Op = 0x47
	OpTypeAssert1  Op = 0x48
	OpTypeAssert2  Op = 0x49
	OpStaticTypeOf Op = 0x4A
	OpCompositeLit Op = 0x4B
	OpArrayLit     Op = 0x4C
	OpSliceLit     Op = 0x4D
	OpSliceLit2    Op = 0x4E
	OpMapLit       Op = 0x4F
	OpStructLit    Op = 0x50
	OpFuncLit      Op = 0x51
	OpConvert      Op = 0x52

	// ---- Type Operators
	OpFieldType     Op = 0x70
	OpArrayType     Op = 0x71
	OpSliceType     Op = 0x72
	OpPointerType   Op = 0x73
	OpInterfaceType Op = 0x74
	OpChanType      Op = 0x75
	OpFuncType      Op = 0x76
	OpMapType       Op = 0x77
	OpStructType    Op = 0x78

	// ---- Assignment Operators
	OpAssign      Op = 0x80
	OpAddAssign   Op = 0x81
	OpSubAssign   Op = 0x82
	OpMulAssign   Op = 0x83
	OpQuoAssign   Op = 0x84
	OpRemAssign   Op = 0x85
	OpBandAssign  Op = 0x86
	OpBandnAssign Op = 0x87
	OpBorAssign   Op = 0x88
	OpXorAssign   Op = 0x89
	OpShlAssign   Op = 0x8A
	OpShrAssign   Op = 0x8B
	OpDefine      Op = 0x8C
	OpInc         Op = 0x8D
	OpDec         Op = 0x8E

	// ---- Declaration Operators
	OpValueDecl Op = 0x90
	OpTypeDecl  Op = 0x91

	// ---- Sticky/Loop Operators
	OpSticky            Op = 0xD0
	OpBody              Op = 0xD1
	OpForLoop           Op = 0xD2
	OpRangeIter         Op = 0xD3
	OpRangeIterString   Op = 0xD4
	OpRangeIterMap      Op = 0xD5
	OpRangeIterArrayPtr Op = 0xD6
	OpReturnCallDefers  Op = 0xD7

	// ---- Special
	OpVoid Op = 0xFF
)

var opNames = [256]string{
	0x00: "OpInvalid",
	0x01: "OpHalt",
	0x02: "OpNoop",
	0x03: "OpExec",
	0x04: "OpPrecall",
	0x05: "OpEnterCrossing",
	0x06: "OpCall",
	0x07: "OpCallNativeBody",
	0x0A: "OpDefer",
	0x0B: "OpCallDeferNativeBody",
	0x0C: "OpGo",
	0x0D: "OpSelect",
	0x0E: "OpSwitchClause",
	0x0F: "OpSwitchClauseCase",
	0x10: "OpTypeSwitch",
	0x11: "OpIfCond",
	0x12: "OpPopValue",
	0x13: "OpPopResults",
	0x14: "OpPopBlock",
	0x15: "OpPopFrameAndReset",
	0x16: "OpPanic1",
	0x17: "OpPanic2",
	0x1A: "OpReturn",
	0x1B: "OpReturnAfterCopy",
	0x1C: "OpReturnFromBlock",
	0x1D: "OpReturnToBlock",
	0x20: "OpUpos",
	0x21: "OpUneg",
	0x22: "OpUnot",
	0x23: "OpUxor",
	0x25: "OpUrecv",
	0x26: "OpLor",
	0x27: "OpLand",
	0x28: "OpEql",
	0x29: "OpNeq",
	0x2A: "OpLss",
	0x2B: "OpLeq",
	0x2C: "OpGtr",
	0x2D: "OpGeq",
	0x2E: "OpAdd",
	0x2F: "OpSub",
	0x30: "OpBor",
	0x31: "OpXor",
	0x32: "OpMul",
	0x33: "OpQuo",
	0x34: "OpRem",
	0x35: "OpShl",
	0x36: "OpShr",
	0x37: "OpBand",
	0x38: "OpBandn",
	0x40: "OpEval",
	0x41: "OpBinary1",
	0x42: "OpIndex1",
	0x43: "OpIndex2",
	0x44: "OpSelector",
	0x45: "OpSlice",
	0x46: "OpStar",
	0x47: "OpRef",
	0x48: "OpTypeAssert1",
	0x49: "OpTypeAssert2",
	0x4A: "OpStaticTypeOf",
	0x4B: "OpCompositeLit",
	0x4C: "OpArrayLit",
	0x4D: "OpSliceLit",
	0x4E: "OpSliceLit2",
	0x4F: "OpMapLit",
	0x50: "OpStructLit",
	0x51: "OpFuncLit",
	0x52: "OpConvert",
	0x70: "OpFieldType",
	0x71: "OpArrayType",
	0x72: "OpSliceType",
	0x73: "OpPointerType",
	0x74: "OpInterfaceType",
	0x75: "OpChanType",
	0x76: "OpFuncType",
	0x77: "OpMapType",
	0x78: "OpStructType",
	0x80: "OpAssign",
	0x81: "OpAddAssign",
	0x82: "OpSubAssign",
	0x83: "OpMulAssign",
	0x84: "OpQuoAssign",
	0x85: "OpRemAssign",
	0x86: "OpBandAssign",
	0x87: "OpBandnAssign",
	0x88: "OpBorAssign",
	0x89: "OpXorAssign",
	0x8A: "OpShlAssign",
	0x8B: "OpShrAssign",
	0x8C: "OpDefine",
	0x8D: "OpInc",
	0x8E: "OpDec",
	0x90: "OpValueDecl",
	0x91: "OpTypeDecl",
	0xD0: "OpSticky",
	0xD1: "OpBody",
	0xD2: "OpForLoop",
	0xD3: "OpRangeIter",
	0xD4: "OpRangeIterString",
	0xD5: "OpRangeIterMap",
	0xD6: "OpRangeIterArrayPtr",
	0xD7: "OpReturnCallDefers",
	0xFF: "OpVoid",
}

func (o Op) String() string {
	if name := opNames[o]; name != "" {
		return name
	}
	return "OpUnknown"
}

// StoreOp represents a store operation.
type StoreOp byte

// Store operation constants.
const (
	StoreOpInvalid StoreOp = 0x00

	// ---- Gno Store Operations
	StoreGetObject       StoreOp = 0x01
	StoreSetObject       StoreOp = 0x02
	StoreDeleteObject    StoreOp = 0x03
	StoreGetPackage      StoreOp = 0x04
	StoreSetPackage      StoreOp = 0x05
	StoreGetType         StoreOp = 0x06
	StoreSetType         StoreOp = 0x07
	StoreGetBlockNode    StoreOp = 0x08
	StoreSetBlockNode    StoreOp = 0x09
	StoreAddMemPackage   StoreOp = 0x0A
	StoreGetMemPackage   StoreOp = 0x0B
	StoreGetPackageRealm StoreOp = 0x0C
	StoreSetPackageRealm StoreOp = 0x0D

	// ---- Amino Serialization
	AminoMarshal    StoreOp = 0x0E
	AminoMarshalAny StoreOp = 0x0F
	AminoUnmarshal  StoreOp = 0x10

	// ---- Underlying Store
	StoreGet StoreOp = 0x11
	StoreSet StoreOp = 0x12

	// ---- Transaction
	FinalizeTx StoreOp = 0x13
)

var storeOpNames = [256]string{
	0x00: "StoreOpInvalid",
	0x01: "StoreGetObject",
	0x02: "StoreSetObject",
	0x03: "StoreDeleteObject",
	0x04: "StoreGetPackage",
	0x05: "StoreSetPackage",
	0x06: "StoreGetType",
	0x07: "StoreSetType",
	0x08: "StoreGetBlockNode",
	0x09: "StoreSetBlockNode",
	0x0A: "StoreAddMemPackage",
	0x0B: "StoreGetMemPackage",
	0x0C: "StoreGetPackageRealm",
	0x0D: "StoreSetPackageRealm",
	0x0E: "AminoMarshal",
	0x0F: "AminoMarshalAny",
	0x10: "AminoUnmarshal",
	0x11: "StoreGet",
	0x12: "StoreSet",
	0x13: "FinalizeTx",
}

func (o StoreOp) String() string {
	if name := storeOpNames[o]; name != "" {
		return name
	}
	return "StoreOpUnknown"
}

// NativeOp represents a native function operation.
type NativeOp byte

// Native operation constants.
const (
	NativeOpInvalid NativeOp = 0x00
	NativePrint     NativeOp = 0x01
	NativePrint1    NativeOp = 0x02
	NativePrint1000 NativeOp = 0x03
	NativePrint1e4  NativeOp = 0x04
)

var nativeOpNames = [256]string{
	0x00: "NativeOpInvalid",
	0x01: "NativePrint",
	0x02: "NativePrint1",
	0x03: "NativePrint1000",
	0x04: "NativePrint1e4",
}

func (o NativeOp) String() string {
	if name := nativeOpNames[o]; name != "" {
		return name
	}
	return "NativeOpUnknown"
}

// GetNativePrintCode returns the appropriate NativeOp for the given print size.
func GetNativePrintCode(size int) NativeOp {
	switch size {
	case 1:
		return NativePrint1
	case 1000:
		return NativePrint1000
	case 10000:
		return NativePrint1e4
	default:
		panic(fmt.Sprintf("invalid print size: %d", size))
	}
}
