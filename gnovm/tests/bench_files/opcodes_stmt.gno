package main

// Test statement operations: increment, decrement, assignment operators, unary/binary operators

func main() {
	StmtOps()
	OpUnary()
	OpBinary()
}

func StmtOps() {
	i := 1_000_000
	i++            // 1,000,001
	i--            // 1,000,000
	i += 1_000_000 // 2_000_000
	i -= 1_000_000 // 1_000_000
	i *= 1         // 1,000,000
	i /= 1_000_000 // 1
	i %= 3         // 1
	i &= 2         // 01 & 10 = 0
	i |= 12        // 12( 1100 )
	i &^= 6        // 12&^6=8 (1100 &^ 0110 = 1000)
	i ^= 14        // 8^14 = 6 (1000 ^ 1110 = 0110)
	i <<= 2        // 24
	i >>= 2        // 6
	_ = i
}

func OpUnary() {
	a := 1
	b := -a
	b = +a
	b = ^a

	c := true
	d := !c
	_ = b
	_ = d
}

func OpBinary() {
	a := 1_000_000
	b := 1_000_001

	var c bool
	var d int
	// boolean
	c = true || false
	c = true && false
	c = a == b
	c = a != b
	c = a < b
	c = a <= b
	c = a > b
	c = a >= b
	c = true && a < b
	c = true || a < b
	//
	d = a + b // 1
	d = a - b // 0
	d = a | b // 1
	d = a ^ b // 1
	d = a * b // 1,000,001,000,000
	d = a / b
	d = a % b
	d = 63 << a
	d = a >> 63
	d = a & b
	d = a &^ b
	_ = c
	_ = d
}

// Gasprofile:
// Opcodes:
//   OpAdd: count=1 gas=18
//   OpAddAssign: count=1 gas=85
//   OpAssign: count=28 gas=2212
//   OpBand: count=1 gas=9
//   OpBandAssign: count=1 gas=54
//   OpBandn: count=1 gas=15
//   OpBandnAssign: count=1 gas=44
//   OpBinary1: count=4 gas=72
//   OpBody: count=57 gas=2451
//   OpBor: count=1 gas=23
//   OpBorAssign: count=1 gas=55
//   OpCall: count=4 gas=1024
//   OpConvert: count=9 gas=144
//   OpDec: count=1 gas=46
//   OpDefine: count=7 gas=777
//   OpEql: count=1 gas=160
//   OpEval: count=159 gas=4611
//   OpExec: count=7 gas=175
//   OpFuncType: count=4 gas=324
//   OpGeq: count=1 gas=26
//   OpGtr: count=1 gas=20
//   OpHalt: count=155 gas=155
//   OpInc: count=1 gas=76
//   OpLand: count=2 gas=48
//   OpLeq: count=1 gas=19
//   OpLss: count=2 gas=26
//   OpMul: count=1 gas=19
//   OpMulAssign: count=1 gas=55
//   OpNeq: count=1 gas=95
//   OpPopBlock: count=154 gas=462
//   OpPopResults: count=4 gas=4
//   OpPrecall: count=13 gas=2691
//   OpQuo: count=1 gas=16
//   OpQuoAssign: count=1 gas=50
//   OpRem: count=1 gas=18
//   OpRemAssign: count=1 gas=46
//   OpReturnFromBlock: count=4 gas=144
//   OpShl: count=1 gas=22
//   OpShlAssign: count=1 gas=68
//   OpShr: count=1 gas=20
//   OpShrAssign: count=1 gas=76
//   OpSub: count=1 gas=6
//   OpSubAssign: count=1 gas=57
//   OpUneg: count=1 gas=25
//   OpUnot: count=1 gas=6
//   OpUpos: count=1 gas=7
//   OpUxor: count=1 gas=14
//   OpValueDecl: count=2 gas=226
//   OpXor: count=1 gas=13
//   OpXorAssign: count=1 gas=48
// Store:
//   StoreGetObject: count=1 size=0
// HotSpots:
//   opcodes_stmt.gno:12 StmtOps: count=2 gas=140
//   opcodes_stmt.gno:13 StmtOps: count=1 gas=76
//   opcodes_stmt.gno:14 StmtOps: count=1 gas=46
//   opcodes_stmt.gno:15 StmtOps: count=2 gas=114
//   opcodes_stmt.gno:16 StmtOps: count=2 gas=86
//   opcodes_stmt.gno:17 StmtOps: count=2 gas=84
//   opcodes_stmt.gno:18 StmtOps: count=2 gas=79
//   opcodes_stmt.gno:19 StmtOps: count=2 gas=75
//   opcodes_stmt.gno:20 StmtOps: count=2 gas=83
//   opcodes_stmt.gno:21 StmtOps: count=2 gas=84
//   opcodes_stmt.gno:22 StmtOps: count=2 gas=73
//   opcodes_stmt.gno:23 StmtOps: count=2 gas=77
//   opcodes_stmt.gno:24 StmtOps: count=2 gas=97
//   opcodes_stmt.gno:25 StmtOps: count=2 gas=105
//   opcodes_stmt.gno:26 StmtOps: count=2 gas=108
//   opcodes_stmt.gno:30 OpUnary: count=2 gas=140
//   opcodes_stmt.gno:31 OpUnary: count=4 gas=194
//   opcodes_stmt.gno:32 OpUnary: count=4 gas=144
//   opcodes_stmt.gno:33 OpUnary: count=4 gas=151
//   opcodes_stmt.gno:35 OpUnary: count=2 gas=140
//   opcodes_stmt.gno:36 OpUnary: count=4 gas=175
//   opcodes_stmt.gno:37 OpUnary: count=2 gas=108
//   opcodes_stmt.gno:38 OpUnary: count=2 gas=108
//   opcodes_stmt.gno:42 OpBinary: count=2 gas=140
//   opcodes_stmt.gno:43 OpBinary: count=2 gas=140
//   opcodes_stmt.gno:45 OpBinary: count=3 gas=167
//   opcodes_stmt.gno:46 OpBinary: count=3 gas=167
//   opcodes_stmt.gno:48 OpBinary: count=2 gas=108
//   opcodes_stmt.gno:49 OpBinary: count=2 gas=108
//   opcodes_stmt.gno:50 OpBinary: count=9 gas=607
//   opcodes_stmt.gno:51 OpBinary: count=9 gas=542
//   opcodes_stmt.gno:52 OpBinary: count=9 gas=460
//   opcodes_stmt.gno:53 OpBinary: count=9 gas=466
//   opcodes_stmt.gno:54 OpBinary: count=9 gas=467
//   opcodes_stmt.gno:55 OpBinary: count=9 gas=473
//   opcodes_stmt.gno:56 OpBinary: count=13 gas=560
//   opcodes_stmt.gno:57 OpBinary: count=4 gas=155
//   opcodes_stmt.gno:59 OpBinary: count=5 gas=184
//   opcodes_stmt.gno:60 OpBinary: count=5 gas=172
//   opcodes_stmt.gno:61 OpBinary: count=5 gas=189
//   opcodes_stmt.gno:62 OpBinary: count=5 gas=179
//   opcodes_stmt.gno:63 OpBinary: count=5 gas=185
//   opcodes_stmt.gno:64 OpBinary: count=5 gas=182
//   opcodes_stmt.gno:65 OpBinary: count=5 gas=184
//   opcodes_stmt.gno:66 OpBinary: count=9 gas=469
//   opcodes_stmt.gno:67 OpBinary: count=5 gas=186
//   opcodes_stmt.gno:68 OpBinary: count=5 gas=175
//   opcodes_stmt.gno:69 OpBinary: count=5 gas=181
//   opcodes_stmt.gno:70 OpBinary: count=2 gas=108
//   opcodes_stmt.gno:71 OpBinary: count=2 gas=108
// SubOps:
//   AssignVar: count=28
//   DefineVar: count=7
// Variables:
//   opcodes_stmt.gno:12 i: count=1
//   opcodes_stmt.gno:26 _: count=1
//   opcodes_stmt.gno:30 a: count=1
//   opcodes_stmt.gno:31 b: count=1
//   opcodes_stmt.gno:32 b: count=1
//   opcodes_stmt.gno:33 b: count=1
//   opcodes_stmt.gno:35 c: count=1
//   opcodes_stmt.gno:36 d: count=1
//   opcodes_stmt.gno:37 _: count=1
//   opcodes_stmt.gno:38 _: count=1
//   opcodes_stmt.gno:42 a: count=1
//   opcodes_stmt.gno:43 b: count=1
//   opcodes_stmt.gno:48 c: count=1
//   opcodes_stmt.gno:49 c: count=1
//   opcodes_stmt.gno:50 c: count=1
//   opcodes_stmt.gno:51 c: count=1
//   opcodes_stmt.gno:52 c: count=1
//   opcodes_stmt.gno:53 c: count=1
//   opcodes_stmt.gno:54 c: count=1
//   opcodes_stmt.gno:55 c: count=1
//   opcodes_stmt.gno:56 c: count=1
//   opcodes_stmt.gno:57 c: count=1
//   opcodes_stmt.gno:59 d: count=1
//   opcodes_stmt.gno:60 d: count=1
//   opcodes_stmt.gno:61 d: count=1
//   opcodes_stmt.gno:62 d: count=1
//   opcodes_stmt.gno:63 d: count=1
//   opcodes_stmt.gno:64 d: count=1
//   opcodes_stmt.gno:65 d: count=1
//   opcodes_stmt.gno:66 d: count=1
//   opcodes_stmt.gno:67 d: count=1
//   opcodes_stmt.gno:68 d: count=1
//   opcodes_stmt.gno:69 d: count=1
//   opcodes_stmt.gno:70 _: count=1
//   opcodes_stmt.gno:71 _: count=1
