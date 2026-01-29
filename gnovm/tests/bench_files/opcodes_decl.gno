package main

// Test declarations and literals: type declarations, variable declarations, const declarations, numeric literals

func main() {
	OpDecl()
	OpEvalInt()
	OpEvalFloat()
}

func OpDecl() {
	type T int
	var i int
	const c int = 1
	_ = i
	_ = c
}

func OpEvalInt() {
	i := 1234567891
	i2 := 123_456_7892
	_ = i
	_ = i2
}

func OpEvalFloat() {
	// decimal
	f := 1.23456789
	f2 := 1.234_56789
	// exp
	f3 := 1.23456789e2
	f4 := 1.23456789e-2
	// hex
	f5 := 0x2.p10  // == 2048.0
	f6 := 0x1.Fp+0 // == 1.9375
	_ = f
	_ = f2
	_ = f3
	_ = f4
	_ = f5
	_ = f6
}

// Gasprofile:
// Opcodes:
//   OpAssign: count=10 gas=790
//   OpBody: count=28 gas=1204
//   OpCall: count=4 gas=1024
//   OpDefine: count=8 gas=888
//   OpEval: count=48 gas=1392
//   OpExec: count=8 gas=200
//   OpFuncType: count=4 gas=324
//   OpHalt: count=42 gas=42
//   OpPopBlock: count=41 gas=123
//   OpPopResults: count=4 gas=4
//   OpPrecall: count=4 gas=828
//   OpReturnFromBlock: count=4 gas=144
//   OpTypeDecl: count=1 gas=100
//   OpValueDecl: count=2 gas=226
// Store:
//   StoreGetObject: count=1 size=0
// HotSpots:
//   opcodes_decl.gno:12 OpDecl: count=3 gas=154
//   opcodes_decl.gno:13 OpDecl: count=3 gas=167
//   opcodes_decl.gno:14 OpDecl: count=4 gas=196
//   opcodes_decl.gno:15 OpDecl: count=2 gas=108
//   opcodes_decl.gno:16 OpDecl: count=2 gas=108
//   opcodes_decl.gno:20 OpEvalInt: count=2 gas=140
//   opcodes_decl.gno:21 OpEvalInt: count=2 gas=140
//   opcodes_decl.gno:22 OpEvalInt: count=2 gas=108
//   opcodes_decl.gno:23 OpEvalInt: count=2 gas=108
//   opcodes_decl.gno:28 OpEvalFloat: count=2 gas=140
//   opcodes_decl.gno:29 OpEvalFloat: count=2 gas=140
//   opcodes_decl.gno:31 OpEvalFloat: count=2 gas=140
//   opcodes_decl.gno:32 OpEvalFloat: count=2 gas=140
//   opcodes_decl.gno:34 OpEvalFloat: count=2 gas=140
//   opcodes_decl.gno:35 OpEvalFloat: count=2 gas=140
//   opcodes_decl.gno:36 OpEvalFloat: count=2 gas=108
//   opcodes_decl.gno:37 OpEvalFloat: count=2 gas=108
//   opcodes_decl.gno:38 OpEvalFloat: count=2 gas=108
//   opcodes_decl.gno:39 OpEvalFloat: count=2 gas=108
//   opcodes_decl.gno:40 OpEvalFloat: count=2 gas=108
//   opcodes_decl.gno:41 OpEvalFloat: count=2 gas=108
// SubOps:
//   AssignVar: count=10
//   DefineVar: count=8
// Variables:
//   opcodes_decl.gno:15 _: count=1
//   opcodes_decl.gno:16 _: count=1
//   opcodes_decl.gno:20 i: count=1
//   opcodes_decl.gno:21 i2: count=1
//   opcodes_decl.gno:22 _: count=1
//   opcodes_decl.gno:23 _: count=1
//   opcodes_decl.gno:28 f: count=1
//   opcodes_decl.gno:29 f2: count=1
//   opcodes_decl.gno:31 f3: count=1
//   opcodes_decl.gno:32 f4: count=1
//   opcodes_decl.gno:34 f5: count=1
//   opcodes_decl.gno:35 f6: count=1
//   opcodes_decl.gno:36 _: count=1
//   opcodes_decl.gno:37 _: count=1
//   opcodes_decl.gno:38 _: count=1
//   opcodes_decl.gno:39 _: count=1
//   opcodes_decl.gno:40 _: count=1
//   opcodes_decl.gno:41 _: count=1
