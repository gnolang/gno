package evm

import (
	"gno.land/p/onbloc/uint256"
)

type OpCode byte

type EvmOpData struct {
	opCode    OpCode
	name      string
	operation func(*state) error
	gas       uint
}

const (
	STOP           OpCode = 0x00
	ADD            OpCode = 0x01
	MUL            OpCode = 0x02
	SUB            OpCode = 0x03
	DIV            OpCode = 0x04
	SDIV           OpCode = 0x05
	MOD            OpCode = 0x06
	SMOD           OpCode = 0x07
	ADDMOD         OpCode = 0x08
	MULMOD         OpCode = 0x09
	EXP            OpCode = 0x0A
	SIGNEXTEND     OpCode = 0x0B
	LT             OpCode = 0x10
	GT             OpCode = 0x11
	SLT            OpCode = 0x12
	SGT            OpCode = 0x13
	EQ             OpCode = 0x14
	ISZERO         OpCode = 0x15
	AND            OpCode = 0x16
	OR             OpCode = 0x17
	XOR            OpCode = 0x18
	NOT            OpCode = 0x19
	BYTE           OpCode = 0x1A
	SHL            OpCode = 0x1B
	SHR            OpCode = 0x1C
	SAR            OpCode = 0x1D
	KECCAK256      OpCode = 0x20
	ADDRESS        OpCode = 0x30
	BALANCE        OpCode = 0x31
	ORIGIN         OpCode = 0x32
	CALLER         OpCode = 0x33
	CALLVALUE      OpCode = 0x34
	CALLDATALOAD   OpCode = 0x35
	CALLDATASIZE   OpCode = 0x36
	CALLDATACOPY   OpCode = 0x37
	CODESIZE       OpCode = 0x38
	CODECOPY       OpCode = 0x39
	GASPRICE       OpCode = 0x3A
	EXTCODESIZE    OpCode = 0x3B
	EXTCODECOPY    OpCode = 0x3C
	RETURNDATASIZE OpCode = 0x3D
	RETURNDATACOPY OpCode = 0x3E
	EXTCODEHASH    OpCode = 0x3F
	BLOCKHASH      OpCode = 0x40
	COINBASE       OpCode = 0x41
	TIMESTAMP      OpCode = 0x42
	NUMBER         OpCode = 0x43
	PREVRANDAO     OpCode = 0x44
	GASLIMIT       OpCode = 0x45
	CHAINID        OpCode = 0x46
	SELFBALANCE    OpCode = 0x47
	BASEFEE        OpCode = 0x48
	BLOBHASH       OpCode = 0x49
	BLOBBASEFEE    OpCode = 0x4A
	POP            OpCode = 0x50
	MLOAD          OpCode = 0x51
	MSTORE         OpCode = 0x52
	MSTORE8        OpCode = 0x53
	SLOAD          OpCode = 0x54
	SSTORE         OpCode = 0x55
	JUMP           OpCode = 0x56
	JUMPI          OpCode = 0x57
	PC             OpCode = 0x58
	MSIZE          OpCode = 0x59
	GAS            OpCode = 0x5A
	JUMPDEST       OpCode = 0x5B
	TLOAD          OpCode = 0x5C
	TSTORE         OpCode = 0x5D
	MCOPY          OpCode = 0x5E
	PUSH0          OpCode = 0x5F
	PUSH1          OpCode = 0x60
	PUSH2          OpCode = 0x61
	PUSH3          OpCode = 0x62
	PUSH4          OpCode = 0x63
	PUSH5          OpCode = 0x64
	PUSH6          OpCode = 0x65
	PUSH7          OpCode = 0x66
	PUSH8          OpCode = 0x67
	PUSH9          OpCode = 0x68
	PUSH10         OpCode = 0x69
	PUSH11         OpCode = 0x6A
	PUSH12         OpCode = 0x6B
	PUSH13         OpCode = 0x6C
	PUSH14         OpCode = 0x6D
	PUSH15         OpCode = 0x6E
	PUSH16         OpCode = 0x6F
	PUSH17         OpCode = 0x70
	PUSH18         OpCode = 0x71
	PUSH19         OpCode = 0x72
	PUSH20         OpCode = 0x73
	PUSH21         OpCode = 0x74
	PUSH22         OpCode = 0x75
	PUSH23         OpCode = 0x76
	PUSH24         OpCode = 0x77
	PUSH25         OpCode = 0x78
	PUSH26         OpCode = 0x79
	PUSH27         OpCode = 0x7A
	PUSH28         OpCode = 0x7B
	PUSH29         OpCode = 0x7C
	PUSH30         OpCode = 0x7D
	PUSH31         OpCode = 0x7E
	PUSH32         OpCode = 0x7F
	DUP1           OpCode = 0x80
	DUP2           OpCode = 0x81
	DUP3           OpCode = 0x82
	DUP4           OpCode = 0x83
	DUP5           OpCode = 0x84
	DUP6           OpCode = 0x85
	DUP7           OpCode = 0x86
	DUP8           OpCode = 0x87
	DUP9           OpCode = 0x88
	DUP10          OpCode = 0x89
	DUP11          OpCode = 0x8A
	DUP12          OpCode = 0x8B
	DUP13          OpCode = 0x8C
	DUP14          OpCode = 0x8D
	DUP15          OpCode = 0x8E
	DUP16          OpCode = 0x8F
	SWAP1          OpCode = 0x90
	SWAP2          OpCode = 0x91
	SWAP3          OpCode = 0x92
	SWAP4          OpCode = 0x93
	SWAP5          OpCode = 0x94
	SWAP6          OpCode = 0x95
	SWAP7          OpCode = 0x96
	SWAP8          OpCode = 0x97
	SWAP9          OpCode = 0x98
	SWAP10         OpCode = 0x99
	SWAP11         OpCode = 0x9A
	SWAP12         OpCode = 0x9B
	SWAP13         OpCode = 0x9C
	SWAP14         OpCode = 0x9D
	SWAP15         OpCode = 0x9E
	SWAP16         OpCode = 0x9F
	LOG0           OpCode = 0xA0
	LOG1           OpCode = 0xA1
	LOG2           OpCode = 0xA2
	LOG3           OpCode = 0xA3
	LOG4           OpCode = 0xA4
	CREATE         OpCode = 0xF0
	CALL           OpCode = 0xF1
	CALLCODE       OpCode = 0xF2
	RETURN         OpCode = 0xF3
	DELEGATECALL   OpCode = 0xF4
	CREATE2        OpCode = 0xF5
	STATICCALL     OpCode = 0xFA
	REVERT         OpCode = 0xFD
	INVALID        OpCode = 0xFE
	SELFDESTRUCT   OpCode = 0xFF
)

var (
	opCodes map[OpCode]EvmOpData
)

func init() {
	opCodes = map[OpCode]EvmOpData{
		STOP:       {STOP, "STOP", opNop, 0},
		ADD:        {ADD, "ADD", opOperation(func(a, b *uint256.Uint) *uint256.Uint { return a.Add(a, b) }), 3},
		MUL:        {MUL, "MUL", opOperation(func(a, b *uint256.Uint) *uint256.Uint { return a.Mul(a, b) }), 5},
		SUB:        {SUB, "SUB", opOperation(func(a, b *uint256.Uint) *uint256.Uint { return a.Sub(a, b) }), 3},
		DIV:        {DIV, "DIV", opOperation(func(a, b *uint256.Uint) *uint256.Uint { return a.Div(a, b) }), 5},
		SDIV:       {SDIV, "SDIV", opOperation(func(a, b *uint256.Uint) *uint256.Uint { return a.SDiv(a, b) }), 5},
		MOD:        {MOD, "MOD", opOperation(func(a, b *uint256.Uint) *uint256.Uint { return a.Mod(a, b) }), 5},
		SMOD:       {SMOD, "SMOD", opOperation(func(a, b *uint256.Uint) *uint256.Uint { return a.SMod(a, b) }), 5},
		ADDMOD:     {ADDMOD, "ADDMOD", opAddmod, 8},
		MULMOD:     {MULMOD, "MULMOD", opMulmod, 8},
		EXP:        {EXP, "EXP", opExp, 10},
		SIGNEXTEND: {SIGNEXTEND, "SIGNEXTEND", opOperation(func(a, b *uint256.Uint) *uint256.Uint { return a.ExtendSign(b, a) }), 5},
		LT: {LT, "LT", opOperation(func(a, b *uint256.Uint) *uint256.Uint {
			if a.Lt(b) {
				return uint256.One()
			}
			return uint256.Zero()
		}), 3},
		GT: {GT, "GT", opOperation(func(a, b *uint256.Uint) *uint256.Uint {
			if a.Gt(b) {
				return uint256.One()
			}
			return uint256.Zero()
		}), 3},
		SLT: {SLT, "SLT", opSlt, 3},
		SGT: {SGT, "SGT", opSgt, 3},
		EQ: {EQ, "EQ", opOperation(func(a, b *uint256.Uint) *uint256.Uint {
			if a.Eq(b) {
				return uint256.One()
			}
			return uint256.Zero()
		}), 3},
		ISZERO:         {ISZERO, "ISZERO", opIszero, 3},
		AND:            {AND, "AND", opOperation(func(a, b *uint256.Uint) *uint256.Uint { return a.And(a, b) }), 3},
		OR:             {OR, "OR", opOperation(func(a, b *uint256.Uint) *uint256.Uint { return a.Or(a, b) }), 3},
		XOR:            {XOR, "XOR", opOperation(func(a, b *uint256.Uint) *uint256.Uint { return a.Xor(a, b) }), 3},
		NOT:            {NOT, "NOT", opNot, 3},
		BYTE:           {BYTE, "BYTE", opByte, 3},
		SHL:            {SHL, "SHL", opOperation(func(a, b *uint256.Uint) *uint256.Uint { return b.Lsh(b, uint(a.Uint64())) }), 3},
		SHR:            {SHR, "SHR", opOperation(func(a, b *uint256.Uint) *uint256.Uint { return b.Rsh(b, uint(a.Uint64())) }), 3},
		SAR:            {SAR, "SAR", opSar, 3},
		KECCAK256:      {KECCAK256, "KECCAK256", opKeccak256, 30}, // can change
		ADDRESS:        {ADDRESS, "ADDRESS", opAddress, 2},
		BALANCE:        {BALANCE, "BALANCE", opBalance, 100}, // can change
		ORIGIN:         {ORIGIN, "ORIGIN", opOrigin, 2},
		CALLER:         {CALLER, "CALLER", opCaller, 2},
		CALLVALUE:      {CALLVALUE, "CALLVALUE", opCallvalue, 2},
		CALLDATALOAD:   {CALLDATALOAD, "CALLDATALOAD", opCalldataload, 3},
		CALLDATASIZE:   {CALLDATASIZE, "CALLDATASIZE", opCalldatasize, 2},
		CALLDATACOPY:   {CALLDATACOPY, "CALLDATACOPY", opCalldatacopy, 3}, // can change
		CODESIZE:       {CODESIZE, "CODESIZE", opCodeSize, 2},
		CODECOPY:       {CODECOPY, "CODECOPY", opCodeCopy, 3}, // can change
		GASPRICE:       {GASPRICE, "GASPRICE", opGasprice, 2},
		EXTCODESIZE:    {EXTCODESIZE, "EXTCODESIZE", opExtCodeSize, 100}, // can change
		EXTCODECOPY:    {EXTCODECOPY, "EXTCODECOPY", opExtCodeCopy, 100}, // can change
		RETURNDATASIZE: {RETURNDATASIZE, "RETURNDATASIZE", opReturnDataSize, 2},
		RETURNDATACOPY: {RETURNDATACOPY, "RETURNDATACOPY", opReturnDataCopy, 3}, // can change
		EXTCODEHASH:    {EXTCODEHASH, "EXTCODEHASH", opExtCodeHash, 100},        // can change
		BLOCKHASH:      {BLOCKHASH, "BLOCKHASH", opBlockHash, 20},
		COINBASE:       {COINBASE, "COINBASE", opCoinbase, 2},
		TIMESTAMP:      {TIMESTAMP, "TIMESTAMP", opTimestamp, 2},
		NUMBER:         {NUMBER, "NUMBER", opNumber, 2},
		PREVRANDAO:     {PREVRANDAO, "PREVRANDAO", opPrevRandao, 2},
		GASLIMIT:       {GASLIMIT, "GASLIMIT", opGaslimit, 2},
		CHAINID:        {CHAINID, "CHAINID", opChainId, 2},
		SELFBALANCE:    {SELFBALANCE, "SELFBALANCE", opSelfBalance, 5},
		BASEFEE:        {BASEFEE, "BASEFEE", opBaseFee, 2},
		BLOBHASH:       {BLOBHASH, "BLOBHASH", opBlobHash, 3},
		BLOBBASEFEE:    {BLOBBASEFEE, "BLOBBASEFEE", opBlobBaseFee, 2},
		POP:            {POP, "POP", opPop, 2},
		MLOAD:          {MLOAD, "MLOAD", opMload, 3},       // can change
		MSTORE:         {MSTORE, "MSTORE", opMstore, 3},    // can change
		MSTORE8:        {MSTORE8, "MSTORE8", opMstore8, 3}, // can change
		SLOAD:          {SLOAD, "SLOAD", opSload, 100},     // can change
		SSTORE:         {SSTORE, "SSTORE", opSstore, 100},  // can change
		JUMP:           {JUMP, "JUMP", opJump, 8},
		JUMPI:          {JUMPI, "JUMPI", opJumpi, 10},
		PC:             {PC, "PC", opPc, 2},
		MSIZE:          {MSIZE, "MSIZE", opMsize, 2},
		GAS:            {GAS, "GAS", opGas, 2},
		JUMPDEST:       {JUMPDEST, "JUMPDEST", opNop, 1},
		TLOAD:          {TLOAD, "TLOAD", opTload, 100},
		TSTORE:         {TSTORE, "TSTORE", opTstore, 100},
		MCOPY:          {MCOPY, "MCOPY", opMcopy, 3}, // can change
		PUSH0:          {PUSH0, "PUSH0", opPush0, 3},
		PUSH1:          {PUSH1, "PUSH1", opPush(1), 3},
		PUSH2:          {PUSH2, "PUSH2", opPush(2), 3},
		PUSH3:          {PUSH3, "PUSH3", opPush(3), 3},
		PUSH4:          {PUSH4, "PUSH4", opPush(4), 3},
		PUSH5:          {PUSH5, "PUSH5", opPush(5), 3},
		PUSH6:          {PUSH6, "PUSH6", opPush(6), 3},
		PUSH7:          {PUSH7, "PUSH7", opPush(7), 3},
		PUSH8:          {PUSH8, "PUSH8", opPush(8), 3},
		PUSH9:          {PUSH9, "PUSH9", opPush(9), 3},
		PUSH10:         {PUSH10, "PUSH10", opPush(10), 3},
		PUSH11:         {PUSH11, "PUSH11", opPush(11), 3},
		PUSH12:         {PUSH12, "PUSH12", opPush(12), 3},
		PUSH13:         {PUSH13, "PUSH13", opPush(13), 3},
		PUSH14:         {PUSH14, "PUSH14", opPush(14), 3},
		PUSH15:         {PUSH15, "PUSH15", opPush(15), 3},
		PUSH16:         {PUSH16, "PUSH16", opPush(16), 3},
		PUSH17:         {PUSH17, "PUSH17", opPush(17), 3},
		PUSH18:         {PUSH18, "PUSH18", opPush(18), 3},
		PUSH19:         {PUSH19, "PUSH19", opPush(19), 3},
		PUSH20:         {PUSH20, "PUSH20", opPush(20), 3},
		PUSH21:         {PUSH21, "PUSH21", opPush(21), 3},
		PUSH22:         {PUSH22, "PUSH22", opPush(22), 3},
		PUSH23:         {PUSH23, "PUSH23", opPush(23), 3},
		PUSH24:         {PUSH24, "PUSH24", opPush(24), 3},
		PUSH25:         {PUSH25, "PUSH25", opPush(25), 3},
		PUSH26:         {PUSH26, "PUSH26", opPush(26), 3},
		PUSH27:         {PUSH27, "PUSH27", opPush(27), 3},
		PUSH28:         {PUSH28, "PUSH28", opPush(28), 3},
		PUSH29:         {PUSH29, "PUSH29", opPush(29), 3},
		PUSH30:         {PUSH30, "PUSH30", opPush(30), 3},
		PUSH31:         {PUSH31, "PUSH31", opPush(31), 3},
		PUSH32:         {PUSH32, "PUSH32", opPush(32), 3},
		DUP1:           {DUP1, "DUP1", opDup(1), 3},
		DUP2:           {DUP2, "DUP2", opDup(2), 3},
		DUP3:           {DUP3, "DUP3", opDup(3), 3},
		DUP4:           {DUP4, "DUP4", opDup(4), 3},
		DUP5:           {DUP5, "DUP5", opDup(5), 3},
		DUP6:           {DUP6, "DUP6", opDup(6), 3},
		DUP7:           {DUP7, "DUP7", opDup(7), 3},
		DUP8:           {DUP8, "DUP8", opDup(8), 3},
		DUP9:           {DUP9, "DUP9", opDup(9), 3},
		DUP10:          {DUP10, "DUP10", opDup(10), 3},
		DUP11:          {DUP11, "DUP11", opDup(11), 3},
		DUP12:          {DUP12, "DUP12", opDup(12), 3},
		DUP13:          {DUP13, "DUP13", opDup(13), 3},
		DUP14:          {DUP14, "DUP14", opDup(14), 3},
		DUP15:          {DUP15, "DUP15", opDup(15), 3},
		DUP16:          {DUP16, "DUP16", opDup(16), 3},
		SWAP1:          {SWAP1, "SWAP1", opSwap(1), 3},
		SWAP2:          {SWAP2, "SWAP2", opSwap(2), 3},
		SWAP3:          {SWAP3, "SWAP3", opSwap(3), 3},
		SWAP4:          {SWAP4, "SWAP4", opSwap(4), 3},
		SWAP5:          {SWAP5, "SWAP5", opSwap(5), 3},
		SWAP6:          {SWAP6, "SWAP6", opSwap(6), 3},
		SWAP7:          {SWAP7, "SWAP7", opSwap(7), 3},
		SWAP8:          {SWAP8, "SWAP8", opSwap(8), 3},
		SWAP9:          {SWAP9, "SWAP9", opSwap(9), 3},
		SWAP10:         {SWAP10, "SWAP10", opSwap(10), 3},
		SWAP11:         {SWAP11, "SWAP11", opSwap(11), 3},
		SWAP12:         {SWAP12, "SWAP12", opSwap(12), 3},
		SWAP13:         {SWAP13, "SWAP13", opSwap(13), 3},
		SWAP14:         {SWAP14, "SWAP14", opSwap(14), 3},
		SWAP15:         {SWAP15, "SWAP15", opSwap(15), 3},
		SWAP16:         {SWAP16, "SWAP16", opSwap(16), 3},
		LOG0:           {LOG0, "LOG0", opLog(0), 375}, // all bellow this one can change
		LOG1:           {LOG1, "LOG1", opLog(1), 750},
		LOG2:           {LOG2, "LOG2", opLog(2), 1125},
		LOG3:           {LOG3, "LOG3", opLog(3), 1500},
		LOG4:           {LOG4, "LOG4", opLog(4), 1875},
		CREATE:         {CREATE, "CREATE", opCreate, 32000},
		CALL:           {CALL, "CALL", opCall, 100},
		CALLCODE:       {CALLCODE, "CALLCODE", opCallCode, 700},
		RETURN:         {RETURN, "RETURN", opReturn, 0},
		DELEGATECALL:   {DELEGATECALL, "DELEGATECALL", opDelegateCall, 100},
		CREATE2:        {CREATE2, "CREATE2", opCreate2, 32000},
		STATICCALL:     {STATICCALL, "STATICCALL", opStaticCall, 700},
		REVERT:         {REVERT, "REVERT", opRevert, 0},
		INVALID:        {INVALID, "INVALID", opInvalid, 0},
		SELFDESTRUCT:   {SELFDESTRUCT, "SELFDESTRUCT", opSelfDestruct, 5000},
	}
}
