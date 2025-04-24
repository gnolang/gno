package evm

type OpCode byte

type EvmOpData struct {
	opCode    OpCode
	name      string
	operation func(*state) error
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
	RETOUR         OpCode = 0xF3
	DELEGATECALL   OpCode = 0xF4
	CREATE2        OpCode = 0xF5
	STATICCALL     OpCode = 0xFA
	REVERT         OpCode = 0xFD
	INVALID        OpCode = 0xFE
	SELFDESTRUCT   OpCode = 0xFF
)

var (
	opCodes map[OpCode]EvmOpData = map[OpCode]EvmOpData{
		STOP:           {STOP, "STOP", opNotImplemented /*opStop*/},
		ADD:            {ADD, "ADD", opAdd},
		MUL:            {MUL, "MUL", opMul},
		SUB:            {SUB, "SUB", opSub},
		DIV:            {DIV, "DIV", opDiv},
		SDIV:           {SDIV, "SDIV", opSdiv},
		MOD:            {MOD, "MOD", opMod},
		SMOD:           {SMOD, "SMOD", opSmod},
		ADDMOD:         {ADDMOD, "ADDMOD", opAddmod},
		MULMOD:         {MULMOD, "MULMOD", opMulmod},
		EXP:            {EXP, "EXP", opExp},
		SIGNEXTEND:     {SIGNEXTEND, "SIGNEXTEND", opNotImplemented /*opSignextend*/},
		LT:             {LT, "LT", opNotImplemented /*opLt*/},
		GT:             {GT, "GT", opNotImplemented /*opGt*/},
		SLT:            {SLT, "SLT", opNotImplemented /*opSlt*/},
		SGT:            {SGT, "SGT", opNotImplemented /*opSgt*/},
		EQ:             {EQ, "EQ", opNotImplemented /*opEq*/},
		ISZERO:         {ISZERO, "ISZERO", opIszero},
		AND:            {AND, "AND", opNotImplemented /*opAnd*/},
		OR:             {OR, "OR", opNotImplemented /*opOr*/},
		XOR:            {XOR, "XOR", opNotImplemented /*opXor*/},
		NOT:            {NOT, "NOT", opNotImplemented /*opNot*/},
		BYTE:           {BYTE, "BYTE", opNotImplemented /*opByte*/},
		SHL:            {SHL, "SHL", opNotImplemented /*opShl*/},
		SHR:            {SHR, "SHR", opNotImplemented /*opShr*/},
		SAR:            {SAR, "SAR", opNotImplemented /*opSar*/},
		KECCAK256:      {KECCAK256, "KECCAK256", opNotImplemented /*opKeccak256*/},
		ADDRESS:        {ADDRESS, "ADDRESS", opNotImplemented /*opAddress*/},
		BALANCE:        {BALANCE, "BALANCE", opNotImplemented /*opBalance*/},
		ORIGIN:         {ORIGIN, "ORIGIN", opNotImplemented /*opOrigin*/},
		CALLER:         {CALLER, "CALLER", opNotImplemented /*opCaller*/},
		CALLVALUE:      {CALLVALUE, "CALLVALUE", opNotImplemented /*opCallvalue*/},
		CALLDATALOAD:   {CALLDATALOAD, "CALLDATALOAD", opNotImplemented /*opCalldataload*/},
		CALLDATASIZE:   {CALLDATASIZE, "CALLDATASIZE", opNotImplemented /*opCalldatasize*/},
		CALLDATACOPY:   {CALLDATACOPY, "CALLDATACOPY", opNotImplemented /*opCalldatacopy*/},
		CODESIZE:       {CODESIZE, "CODESIZE", opNotImplemented /*opCodesize*/},
		CODECOPY:       {CODECOPY, "CODECOPY", opNotImplemented /*opCodecopy*/},
		GASPRICE:       {GASPRICE, "GASPRICE", opNotImplemented /*opGasprice*/},
		EXTCODESIZE:    {EXTCODESIZE, "EXTCODESIZE", opNotImplemented /*opExtcodesize*/},
		EXTCODECOPY:    {EXTCODECOPY, "EXTCODECOPY", opNotImplemented /*opExtcodecopy*/},
		RETURNDATASIZE: {RETURNDATASIZE, "RETURNDATASIZE", opNotImplemented /*opReturndatasize*/},
		RETURNDATACOPY: {RETURNDATACOPY, "RETURNDATACOPY", opNotImplemented /*opReturndatacopy*/},
		EXTCODEHASH:    {EXTCODEHASH, "EXTCODEHASH", opNotImplemented /*opExtcodehash*/},
		BLOCKHASH:      {BLOCKHASH, "BLOCKHASH", opNotImplemented /*opBlockhash*/},
		COINBASE:       {COINBASE, "COINBASE", opNotImplemented /*opCoinbase*/},
		TIMESTAMP:      {TIMESTAMP, "TIMESTAMP", opNotImplemented /*opTimestamp*/},
		NUMBER:         {NUMBER, "NUMBER", opNotImplemented /*opNumber*/},
		PREVRANDAO:     {PREVRANDAO, "PREVRANDAO", opNotImplemented /*opPrevrandao*/},
		GASLIMIT:       {GASLIMIT, "GASLIMIT", opNotImplemented /*opGaslimit*/},
		CHAINID:        {CHAINID, "CHAINID", opNotImplemented /*opChainid*/},
		SELFBALANCE:    {SELFBALANCE, "SELFBALANCE", opNotImplemented /*opSelfbalance*/},
		BASEFEE:        {BASEFEE, "BASEFEE", opNotImplemented /*opBasefee*/},
		BLOBHASH:       {BLOBHASH, "BLOBHASH", opNotImplemented /*opBlobhash*/},
		BLOBBASEFEE:    {BLOBBASEFEE, "BLOBBASEFEE", opNotImplemented /*opBlobbasefee*/},
		POP:            {POP, "POP", opPop},
		MLOAD:          {MLOAD, "MLOAD", opMload},
		MSTORE:         {MSTORE, "MSTORE", opMstore},
		MSTORE8:        {MSTORE8, "MSTORE8", opMstore8},
		SLOAD:          {SLOAD, "SLOAD", opNotImplemented /*opSload*/},
		SSTORE:         {SSTORE, "SSTORE", opNotImplemented /*opSstore*/},
		JUMP:           {JUMP, "JUMP", opJump},
		JUMPI:          {JUMPI, "JUMPI", opJumpi},
		PC:             {PC, "PC", opNotImplemented /*opPc*/},
		MSIZE:          {MSIZE, "MSIZE", opNotImplemented /*opMsize*/},
		GAS:            {GAS, "GAS", opNotImplemented /*opGas*/},
		JUMPDEST:       {JUMPDEST, "JUMPDEST", opNop},
		TLOAD:          {TLOAD, "TLOAD", opNotImplemented /*opTload*/},
		TSTORE:         {TSTORE, "TSTORE", opNotImplemented /*opTstore*/},
		MCOPY:          {MCOPY, "MCOPY", opNotImplemented /*opMcopy*/},
		PUSH0:          {PUSH0, "PUSH0", opPush0},
		PUSH1:          {PUSH1, "PUSH1", opPush(1)},
		PUSH2:          {PUSH2, "PUSH2", opPush(2)},
		PUSH3:          {PUSH3, "PUSH3", opPush(3)},
		PUSH4:          {PUSH4, "PUSH4", opPush(4)},
		PUSH5:          {PUSH5, "PUSH5", opPush(5)},
		PUSH6:          {PUSH6, "PUSH6", opPush(6)},
		PUSH7:          {PUSH7, "PUSH7", opPush(7)},
		PUSH8:          {PUSH8, "PUSH8", opPush(8)},
		PUSH9:          {PUSH9, "PUSH9", opPush(9)},
		PUSH10:         {PUSH10, "PUSH10", opPush(10)},
		PUSH11:         {PUSH11, "PUSH11", opPush(11)},
		PUSH12:         {PUSH12, "PUSH12", opPush(12)},
		PUSH13:         {PUSH13, "PUSH13", opPush(13)},
		PUSH14:         {PUSH14, "PUSH14", opPush(14)},
		PUSH15:         {PUSH15, "PUSH15", opPush(15)},
		PUSH16:         {PUSH16, "PUSH16", opPush(16)},
		PUSH17:         {PUSH17, "PUSH17", opPush(17)},
		PUSH18:         {PUSH18, "PUSH18", opPush(18)},
		PUSH19:         {PUSH19, "PUSH19", opPush(19)},
		PUSH20:         {PUSH20, "PUSH20", opPush(20)},
		PUSH21:         {PUSH21, "PUSH21", opPush(21)},
		PUSH22:         {PUSH22, "PUSH22", opPush(22)},
		PUSH23:         {PUSH23, "PUSH23", opPush(23)},
		PUSH24:         {PUSH24, "PUSH24", opPush(24)},
		PUSH25:         {PUSH25, "PUSH25", opPush(25)},
		PUSH26:         {PUSH26, "PUSH26", opPush(26)},
		PUSH27:         {PUSH27, "PUSH27", opPush(27)},
		PUSH28:         {PUSH28, "PUSH28", opPush(28)},
		PUSH29:         {PUSH29, "PUSH29", opPush(29)},
		PUSH30:         {PUSH30, "PUSH30", opPush(30)},
		PUSH31:         {PUSH31, "PUSH31", opPush(31)},
		PUSH32:         {PUSH32, "PUSH32", opPush(32)},
		DUP1:           {DUP1, "DUP1", opDup(1)},
		DUP2:           {DUP2, "DUP2", opDup(2)},
		DUP3:           {DUP3, "DUP3", opDup(3)},
		DUP4:           {DUP4, "DUP4", opDup(4)},
		DUP5:           {DUP5, "DUP5", opDup(5)},
		DUP6:           {DUP6, "DUP6", opDup(6)},
		DUP7:           {DUP7, "DUP7", opDup(7)},
		DUP8:           {DUP8, "DUP8", opDup(8)},
		DUP9:           {DUP9, "DUP9", opDup(9)},
		DUP10:          {DUP10, "DUP10", opDup(10)},
		DUP11:          {DUP11, "DUP11", opDup(11)},
		DUP12:          {DUP12, "DUP12", opDup(12)},
		DUP13:          {DUP13, "DUP13", opDup(13)},
		DUP14:          {DUP14, "DUP14", opDup(14)},
		DUP15:          {DUP15, "DUP15", opDup(15)},
		DUP16:          {DUP16, "DUP16", opDup(16)},
		SWAP1:          {SWAP1, "SWAP1", opSwap(1)},
		SWAP2:          {SWAP2, "SWAP2", opSwap(2)},
		SWAP3:          {SWAP3, "SWAP3", opSwap(3)},
		SWAP4:          {SWAP4, "SWAP4", opSwap(4)},
		SWAP5:          {SWAP5, "SWAP5", opSwap(5)},
		SWAP6:          {SWAP6, "SWAP6", opSwap(6)},
		SWAP7:          {SWAP7, "SWAP7", opSwap(7)},
		SWAP8:          {SWAP8, "SWAP8", opSwap(8)},
		SWAP9:          {SWAP9, "SWAP9", opSwap(9)},
		SWAP10:         {SWAP10, "SWAP10", opSwap(10)},
		SWAP11:         {SWAP11, "SWAP11", opSwap(11)},
		SWAP12:         {SWAP12, "SWAP12", opSwap(12)},
		SWAP13:         {SWAP13, "SWAP13", opSwap(13)},
		SWAP14:         {SWAP14, "SWAP14", opSwap(14)},
		SWAP15:         {SWAP15, "SWAP15", opSwap(15)},
		SWAP16:         {SWAP16, "SWAP16", opSwap(16)},
		LOG0:           {LOG0, "LOG0", opNotImplemented /*opLog0*/},
		LOG1:           {LOG1, "LOG1", opNotImplemented /*opLog1*/},
		LOG2:           {LOG2, "LOG2", opNotImplemented /*opLog2*/},
		LOG3:           {LOG3, "LOG3", opNotImplemented /*opLog3*/},
		LOG4:           {LOG4, "LOG4", opNotImplemented /*opLog4*/},
		CREATE:         {CREATE, "CREATE", opNotImplemented /*opCreate*/},
		CALL:           {CALL, "CALL", opNotImplemented /*opCall*/},
		CALLCODE:       {CALLCODE, "CALLCODE", opNotImplemented /*opCallcode*/},
		RETOUR:         {RETOUR, "RETOUR", opNotImplemented /*opRetour*/},
		DELEGATECALL:   {DELEGATECALL, "DELEGATECALL", opNotImplemented /*opDelegatecall*/},
		CREATE2:        {CREATE2, "CREATE2", opNotImplemented /*opCreate2*/},
		STATICCALL:     {STATICCALL, "STATICCALL", opNotImplemented /*opStaticcall*/},
		REVERT:         {REVERT, "REVERT", opNotImplemented /*opRevert*/},
		INVALID:        {INVALID, "INVALID", opNotImplemented /*opInvalid*/},
		SELFDESTRUCT:   {SELFDESTRUCT, "SELFDESTRUCT", opNotImplemented /*opSelfdestruct*/},
	}
)
