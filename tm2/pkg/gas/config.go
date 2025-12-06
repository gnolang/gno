package gas

// Cost represents the gas cost for an operation.
type Cost float64

// Config defines gas costs for all operations.
type Config struct {
	// Global multiplier applied to all gas consumption (default 1.0).
	// Allows fractional adjustments (e.g., 0.5 to halve gas, 2.0 to double).
	GlobalMultiplier float64

	// Fixed-size array of gas costs indexed by Operation.
	// Each Operation constant maps to its cost in this array.
	Costs [OperationListMaxSize]Cost
}

// GetCostForOperation returns the base cost for an operation.
func (c *Config) GetCostForOperation(op Operation) Cost {
	return c.Costs[op]
}

// defaultConfig is the default gas configuration with predefined costs.
var defaultConfig = Config{
	GlobalMultiplier: 1,
	Costs: [OperationListMaxSize]Cost{
		// Store operations
		OpStoreReadFlat:     1000,
		OpStoreReadPerByte:  3,
		OpStoreWriteFlat:    2000,
		OpStoreWritePerByte: 30,
		OpStoreHas:          1000,
		OpStoreDelete:       1000,
		OpStoreIterNextFlat: 30,
		OpStoreValuePerByte: 3,

		// KVStore operations
		OpKVStoreGetObjectPerByte:       16,
		OpKVStoreSetObjectPerByte:       16,
		OpKVStoreGetTypePerByte:         52,
		OpKVStoreSetTypePerByte:         52,
		OpKVStoreGetPackageRealmPerByte: 524,
		OpKVStoreSetPackageRealmPerByte: 524,
		OpKVStoreAddMemPackagePerByte:   8,
		OpKVStoreGetMemPackagePerByte:   8,
		OpKVStoreDeleteObject:           3715,

		// VM Memory operations
		OpMemoryAllocPerByte: 1,
		// Represents the "time unit" cost for a single garbage collection visit.
		// It's similar to "CPU cycles" and is calculated based on a rough
		// benchmarking results.
		// TODO: more accurate benchmark.
		OpMemoryGarbageCollect: 8,

		// VM CPU operations - Control operators
		OpCPUInvalid:             1,
		OpCPUHalt:                1,
		OpCPUNoop:                1,
		OpCPUExec:                25,
		OpCPUPrecall:             207,
		OpCPUEnterCrossing:       100,
		OpCPUCall:                256,
		OpCPUCallNativeBody:      424,
		OpCPUDefer:               64,
		OpCPUCallDeferNativeBody: 33,
		OpCPUGo:                  1,
		OpCPUSelect:              1,
		OpCPUSwitchClause:        38,
		OpCPUSwitchClauseCase:    143,
		OpCPUTypeSwitch:          171,
		OpCPUIfCond:              38,
		OpCPUPopValue:            1,
		OpCPUPopResults:          1,
		OpCPUPopBlock:            3,
		OpCPUPopFrameAndReset:    15,
		OpCPUPanic1:              121,
		OpCPUPanic2:              21,
		OpCPUReturn:              38,
		OpCPUReturnAfterCopy:     38,
		OpCPUReturnFromBlock:     36,
		OpCPUReturnToBlock:       23,

		// VM CPU operations - Unary & binary operators
		OpCPUUpos:  7,
		OpCPUUneg:  25,
		OpCPUUnot:  6,
		OpCPUUxor:  14,
		OpCPUUrecv: 1,
		OpCPULor:   26,
		OpCPULand:  24,
		OpCPUEql:   160,
		OpCPUNeq:   95,
		OpCPULss:   13,
		OpCPULeq:   19,
		OpCPUGtr:   20,
		OpCPUGeq:   26,
		OpCPUAdd:   18,
		OpCPUSub:   6,
		OpCPUBor:   23,
		OpCPUXor:   13,
		OpCPUMul:   19,
		OpCPUQuo:   16,
		OpCPURem:   18,
		OpCPUShl:   22,
		OpCPUShr:   20,
		OpCPUBand:  9,
		OpCPUBandn: 15,

		// VM CPU operations - Other expression operators
		OpCPUEval:         29,
		OpCPUBinary1:      19,
		OpCPUIndex1:       77,
		OpCPUIndex2:       195,
		OpCPUSelector:     32,
		OpCPUSlice:        103,
		OpCPUStar:         40,
		OpCPURef:          125,
		OpCPUTypeAssert1:  30,
		OpCPUTypeAssert2:  25,
		OpCPUStaticTypeOf: 100,
		OpCPUCompositeLit: 50,
		OpCPUArrayLit:     137,
		OpCPUSliceLit:     183,
		OpCPUSliceLit2:    467,
		OpCPUMapLit:       475,
		OpCPUStructLit:    179,
		OpCPUFuncLit:      61,
		OpCPUConvert:      16,

		// VM CPU operations - Type operators
		OpCPUFieldType:     59,
		OpCPUArrayType:     57,
		OpCPUSliceType:     55,
		OpCPUPointerType:   1,
		OpCPUInterfaceType: 75,
		OpCPUChanType:      57,
		OpCPUFuncType:      81,
		OpCPUMapType:       59,
		OpCPUStructType:    174,

		// VM CPU operations - Statement operators
		OpCPUAssign:      79,
		OpCPUAddAssign:   85,
		OpCPUSubAssign:   57,
		OpCPUMulAssign:   55,
		OpCPUQuoAssign:   50,
		OpCPURemAssign:   46,
		OpCPUBandAssign:  54,
		OpCPUBandnAssign: 44,
		OpCPUBorAssign:   55,
		OpCPUXorAssign:   48,
		OpCPUShlAssign:   68,
		OpCPUShrAssign:   76,
		OpCPUDefine:      111,
		OpCPUInc:         76,
		OpCPUDec:         46,

		// VM CPU operations - Decl operators
		OpCPUValueDecl: 113,
		OpCPUTypeDecl:  100,

		// VM CPU operations - Loop operators
		OpCPUSticky:            1,
		OpCPUBody:              43,
		OpCPUForLoop:           27,
		OpCPURangeIter:         105,
		OpCPURangeIterString:   55,
		OpCPURangeIterMap:      48,
		OpCPURangeIterArrayPtr: 46,
		OpCPUReturnCallDefers:  78,

		// Parsing operations
		OpParsingToken:   1, // TODO: adjust this with benchmarks
		OpParsingNesting: 1, // TODO: adjust this with benchmarks

		// Transaction operations
		OpTransactionPerByte:            10,
		OpTransactionSigVerifyEd25519:   590,
		OpTransactionSigVerifySecp256k1: 1000,

		// Print operations
		// OpNativePrintFlat is the base gas cost for the Print function.
		// The actual cost is 1800, but we subtract OpCPUCallNativeBody (424), resulting in 1376.
		OpNativePrintFlat:    1376,
		OpNativePrintPerByte: 0.1, // 10 bytes per gas

		// Block summation operations
		OpBlockGasSum: 1,

		// Test operation
		OpTesting: 1,

		// All other operations default to 0
	},
}

// DefaultConfig returns a copy of the default configuration.
func DefaultConfig() Config {
	return defaultConfig
}
