package gas

import "math"

// Operation is a typed identifier for gas operations.
type Operation uint8

// Bounds for Operation type.
const (
	OperationListMaxSize = math.MaxUint8 + 1
	OperationMaxValue    = math.MaxUint8
)

// Operation enumeration.
const (
	// Store operations
	OpStoreReadFlat Operation = iota
	OpStoreReadPerByte
	OpStoreWriteFlat
	OpStoreWritePerByte
	OpStoreHas
	OpStoreDelete
	OpStoreIterNextFlat
	OpStoreValuePerByte

	// KVStore operations
	OpKVStoreGetObjectPerByte
	OpKVStoreSetObjectPerByte
	OpKVStoreGetTypePerByte
	OpKVStoreSetTypePerByte
	OpKVStoreGetPackageRealmPerByte
	OpKVStoreSetPackageRealmPerByte
	OpKVStoreAddMemPackagePerByte
	OpKVStoreGetMemPackagePerByte
	OpKVStoreDeleteObject

	// VM Memory operations
	OpMemoryAllocPerByte
	OpMemoryGarbageCollect

	// VM CPU operations - Control operators
	OpCPUInvalid
	OpCPUHalt
	OpCPUNoop
	OpCPUExec
	OpCPUPrecall
	OpCPUEnterCrossing
	OpCPUCall
	OpCPUCallNativeBody
	OpCPUDefer
	OpCPUCallDeferNativeBody
	OpCPUGo
	OpCPUSelect
	OpCPUSwitchClause
	OpCPUSwitchClauseCase
	OpCPUTypeSwitch
	OpCPUIfCond
	OpCPUPopValue
	OpCPUPopResults
	OpCPUPopBlock
	OpCPUPopFrameAndReset
	OpCPUPanic1
	OpCPUPanic2
	OpCPUReturn
	OpCPUReturnAfterCopy
	OpCPUReturnFromBlock
	OpCPUReturnToBlock

	// VM CPU operations - Unary & binary operators
	OpCPUUpos
	OpCPUUneg
	OpCPUUnot
	OpCPUUxor
	OpCPUUrecv
	OpCPULor
	OpCPULand
	OpCPUEql
	OpCPUNeq
	OpCPULss
	OpCPULeq
	OpCPUGtr
	OpCPUGeq
	OpCPUAdd
	OpCPUSub
	OpCPUBor
	OpCPUXor
	OpCPUMul
	OpCPUQuo
	OpCPURem
	OpCPUShl
	OpCPUShr
	OpCPUBand
	OpCPUBandn

	// VM CPU operations - Other expression operators
	OpCPUEval
	OpCPUBinary1
	OpCPUIndex1
	OpCPUIndex2
	OpCPUSelector
	OpCPUSlice
	OpCPUStar
	OpCPURef
	OpCPUTypeAssert1
	OpCPUTypeAssert2
	OpCPUStaticTypeOf
	OpCPUCompositeLit
	OpCPUArrayLit
	OpCPUSliceLit
	OpCPUSliceLit2
	OpCPUMapLit
	OpCPUStructLit
	OpCPUFuncLit
	OpCPUConvert

	// VM CPU operations - Type operators
	OpCPUFieldType
	OpCPUArrayType
	OpCPUSliceType
	OpCPUPointerType
	OpCPUInterfaceType
	OpCPUChanType
	OpCPUFuncType
	OpCPUMapType
	OpCPUStructType

	// VM CPU operations - Statement operators
	OpCPUAssign
	OpCPUAddAssign
	OpCPUSubAssign
	OpCPUMulAssign
	OpCPUQuoAssign
	OpCPURemAssign
	OpCPUBandAssign
	OpCPUBandnAssign
	OpCPUBorAssign
	OpCPUXorAssign
	OpCPUShlAssign
	OpCPUShrAssign
	OpCPUDefine
	OpCPUInc
	OpCPUDec

	// VM CPU operations - Decl operators
	OpCPUValueDecl
	OpCPUTypeDecl

	// VM CPU operations - Loop operators
	OpCPUSticky
	OpCPUBody
	OpCPUForLoop
	OpCPURangeIter
	OpCPURangeIterString
	OpCPURangeIterMap
	OpCPURangeIterArrayPtr
	OpCPUReturnCallDefers

	// Parsing operations
	OpParsingToken
	OpParsingNesting

	// Transaction operations
	OpTransactionPerByte
	OpTransactionSigVerifyEd25519
	OpTransactionSigVerifySecp256k1

	// Native operations
	OpNativePrintFlat
	OpNativePrintPerByte

	// Block summation operations
	OpBlockGasSum

	// Test operation
	OpTesting = OperationMaxValue // Last value reserved for testing purposes
)

// operationNames maps Operation values to their string names.
var operationNames = [OperationListMaxSize]string{
	// Store operations
	OpStoreReadFlat:     "StoreReadFlat",
	OpStoreReadPerByte:  "StoreReadPerByte",
	OpStoreWriteFlat:    "StoreWriteFlat",
	OpStoreWritePerByte: "StoreWritePerByte",
	OpStoreHas:          "StoreHas",
	OpStoreDelete:       "StoreDelete",
	OpStoreIterNextFlat: "StoreIterNextFlat",
	OpStoreValuePerByte: "StoreValuePerByte",

	// KVStore operations
	OpKVStoreGetObjectPerByte:       "KVStoreGetObjectPerByte",
	OpKVStoreSetObjectPerByte:       "KVStoreSetObjectPerByte",
	OpKVStoreGetTypePerByte:         "KVStoreGetTypePerByte",
	OpKVStoreSetTypePerByte:         "KVStoreSetTypePerByte",
	OpKVStoreGetPackageRealmPerByte: "KVStoreGetPackageRealmPerByte",
	OpKVStoreSetPackageRealmPerByte: "KVStoreSetPackageRealmPerByte",
	OpKVStoreAddMemPackagePerByte:   "KVStoreAddMemPackagePerByte",
	OpKVStoreGetMemPackagePerByte:   "KVStoreGetMemPackagePerByte",
	OpKVStoreDeleteObject:           "KVStoreDeleteObject",

	// VM Memory operations
	OpMemoryAllocPerByte:   "MemoryAllocPerByte",
	OpMemoryGarbageCollect: "MemoryGarbageCollect",

	// VM CPU operations - Control operators
	OpCPUInvalid:             "CPUInvalid",
	OpCPUHalt:                "CPUHalt",
	OpCPUNoop:                "CPUNoop",
	OpCPUExec:                "CPUExec",
	OpCPUPrecall:             "CPUPrecall",
	OpCPUEnterCrossing:       "CPUEnterCrossing",
	OpCPUCall:                "CPUCall",
	OpCPUCallNativeBody:      "CPUCallNativeBody",
	OpCPUDefer:               "CPUDefer",
	OpCPUCallDeferNativeBody: "CPUCallDeferNativeBody",
	OpCPUGo:                  "CPUGo",
	OpCPUSelect:              "CPUSelect",
	OpCPUSwitchClause:        "CPUSwitchClause",
	OpCPUSwitchClauseCase:    "CPUSwitchClauseCase",
	OpCPUTypeSwitch:          "CPUTypeSwitch",
	OpCPUIfCond:              "CPUIfCond",
	OpCPUPopValue:            "CPUPopValue",
	OpCPUPopResults:          "CPUPopResults",
	OpCPUPopBlock:            "CPUPopBlock",
	OpCPUPopFrameAndReset:    "CPUPopFrameAndReset",
	OpCPUPanic1:              "CPUPanic1",
	OpCPUPanic2:              "CPUPanic2",
	OpCPUReturn:              "CPUReturn",
	OpCPUReturnAfterCopy:     "CPUReturnAfterCopy",
	OpCPUReturnFromBlock:     "CPUReturnFromBlock",
	OpCPUReturnToBlock:       "CPUReturnToBlock",

	// VM CPU operations - Unary & binary operators
	OpCPUUpos:  "CPUUpos",
	OpCPUUneg:  "CPUUneg",
	OpCPUUnot:  "CPUUnot",
	OpCPUUxor:  "CPUUxor",
	OpCPUUrecv: "CPUUrecv",
	OpCPULor:   "CPULor",
	OpCPULand:  "CPULand",
	OpCPUEql:   "CPUEql",
	OpCPUNeq:   "CPUNeq",
	OpCPULss:   "CPULss",
	OpCPULeq:   "CPULeq",
	OpCPUGtr:   "CPUGtr",
	OpCPUGeq:   "CPUGeq",
	OpCPUAdd:   "CPUAdd",
	OpCPUSub:   "CPUSub",
	OpCPUBor:   "CPUBor",
	OpCPUXor:   "CPUXor",
	OpCPUMul:   "CPUMul",
	OpCPUQuo:   "CPUQuo",
	OpCPURem:   "CPURem",
	OpCPUShl:   "CPUShl",
	OpCPUShr:   "CPUShr",
	OpCPUBand:  "CPUBand",
	OpCPUBandn: "CPUBandn",

	// VM CPU operations - Other expression operators
	OpCPUEval:         "CPUEval",
	OpCPUBinary1:      "CPUBinary1",
	OpCPUIndex1:       "CPUIndex1",
	OpCPUIndex2:       "CPUIndex2",
	OpCPUSelector:     "CPUSelector",
	OpCPUSlice:        "CPUSlice",
	OpCPUStar:         "CPUStar",
	OpCPURef:          "CPURef",
	OpCPUTypeAssert1:  "CPUTypeAssert1",
	OpCPUTypeAssert2:  "CPUTypeAssert2",
	OpCPUStaticTypeOf: "CPUStaticTypeOf",
	OpCPUCompositeLit: "CPUCompositeLit",
	OpCPUArrayLit:     "CPUArrayLit",
	OpCPUSliceLit:     "CPUSliceLit",
	OpCPUSliceLit2:    "CPUSliceLit2",
	OpCPUMapLit:       "CPUMapLit",
	OpCPUStructLit:    "CPUStructLit",
	OpCPUFuncLit:      "CPUFuncLit",
	OpCPUConvert:      "CPUConvert",

	// VM CPU operations - Type operators
	OpCPUFieldType:     "CPUFieldType",
	OpCPUArrayType:     "CPUArrayType",
	OpCPUSliceType:     "CPUSliceType",
	OpCPUPointerType:   "CPUPointerType",
	OpCPUInterfaceType: "CPUInterfaceType",
	OpCPUChanType:      "CPUChanType",
	OpCPUFuncType:      "CPUFuncType",
	OpCPUMapType:       "CPUMapType",
	OpCPUStructType:    "CPUStructType",

	// VM CPU operations - Statement operators
	OpCPUAssign:      "CPUAssign",
	OpCPUAddAssign:   "CPUAddAssign",
	OpCPUSubAssign:   "CPUSubAssign",
	OpCPUMulAssign:   "CPUMulAssign",
	OpCPUQuoAssign:   "CPUQuoAssign",
	OpCPURemAssign:   "CPURemAssign",
	OpCPUBandAssign:  "CPUBandAssign",
	OpCPUBandnAssign: "CPUBandnAssign",
	OpCPUBorAssign:   "CPUBorAssign",
	OpCPUXorAssign:   "CPUXorAssign",
	OpCPUShlAssign:   "CPUShlAssign",
	OpCPUShrAssign:   "CPUShrAssign",
	OpCPUDefine:      "CPUDefine",
	OpCPUInc:         "CPUInc",
	OpCPUDec:         "CPUDec",

	// VM CPU operations - Decl operators
	OpCPUValueDecl: "CPUValueDecl",
	OpCPUTypeDecl:  "CPUTypeDecl",

	// VM CPU operations - Loop operators
	OpCPUSticky:            "CPUSticky",
	OpCPUBody:              "CPUBody",
	OpCPUForLoop:           "CPUForLoop",
	OpCPURangeIter:         "CPURangeIter",
	OpCPURangeIterString:   "CPURangeIterString",
	OpCPURangeIterMap:      "CPURangeIterMap",
	OpCPURangeIterArrayPtr: "CPURangeIterArrayPtr",
	OpCPUReturnCallDefers:  "CPUReturnCallDefers",

	// Parsing operations
	OpParsingToken:   "ParsingToken",
	OpParsingNesting: "ParsingNesting",

	// Transaction operations
	OpTransactionPerByte:            "TransactionPerByte",
	OpTransactionSigVerifyEd25519:   "TransactionSigVerifyEd25519",
	OpTransactionSigVerifySecp256k1: "TxansactionSigVerifySecp256k1",

	// Native operations
	OpNativePrintFlat:    "NativePrintFlat",
	OpNativePrintPerByte: "NativePrintPerByte",

	// Block summation operations
	OpBlockGasSum: "BlockGasSum",

	// Test operation
	OpTesting: "Testing",
}

// String returns the name of the operation for logging purposes.
func (o Operation) String() string {
	if name := operationNames[o]; name != "" {
		return name
	}
	return "UnknownOperation"
}
