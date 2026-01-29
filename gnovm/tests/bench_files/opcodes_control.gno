package main

// Test control flow operations: if/else, switch, defer, logical or/and, panic/recover

func main() {
	ControlOps()
	_ = OpDefer()
	OpLor()
	OpLand()
	OpPanic()
}

func ControlOps() {
	var a int
	var b int
	if a > 0 {
		b++
	} else {
		b--
	}
	switch b {
	case 0:
		a++
	case 1:
		a--
	default:
		a = 2
	}
	_ = a
	_ = b
}

func OpDefer() int {
	a := [1]int{0}

	defer func() {
		// a[0]++
	}()

	defer func() {
		_ = len(a)
	}()

	return a[0]
}

func OpLor() {
	a := true
	b := false
	_ = b || a
	_ = a || b
}

func OpLand() {
	a := true
	b := false
	_ = a && b
}

func OpPanic() {
	defer func() {
		recover()
	}()
	panic("panic")
}

// Gasprofile:
// Opcodes:
//   OpArrayLit: count=1 gas=137
//   OpArrayType: count=1 gas=57
//   OpAssign: count=8 gas=632
//   OpBinary1: count=3 gas=54
//   OpBody: count=37 gas=1591
//   OpCall: count=9 gas=2304
//   OpCallNativeBody: count=3 gas=1272
//   OpCompositeLit: count=1 gas=50
//   OpDec: count=1 gas=46
//   OpDefer: count=3 gas=192
//   OpDefine: count=5 gas=555
//   OpEval: count=93 gas=2697
//   OpExec: count=10 gas=250
//   OpFieldType: count=1 gas=59
//   OpFuncLit: count=3 gas=183
//   OpFuncType: count=15 gas=1215
//   OpGtr: count=1 gas=20
//   OpHalt: count=71 gas=71
//   OpIfCond: count=1 gas=38
//   OpIndex1: count=1 gas=14
//   OpLand: count=1 gas=24
//   OpLor: count=1 gas=26
//   OpPanic2: count=1 gas=1
//   OpPopBlock: count=71 gas=213
//   OpPopFrameAndReset: count=1 gas=15
//   OpPopResults: count=6 gas=6
//   OpPrecall: count=9 gas=1863
//   OpReturn: count=2 gas=76
//   OpReturnCallDefers: count=6 gas=468
//   OpReturnFromBlock: count=9 gas=324
//   OpReturnToBlock: count=1 gas=23
//   OpSwitchClause: count=3 gas=114
//   OpSwitchClauseCase: count=2 gas=286
//   OpValueDecl: count=2 gas=226
// Store:
//   StoreGetObject: count=1 size=0
// HotSpots:
//   opcodes_control.gno:7 main: count=5 gas=600
//   opcodes_control.gno:14 ControlOps: count=3 gas=167
//   opcodes_control.gno:15 ControlOps: count=3 gas=167
//   opcodes_control.gno:16 ControlOps: count=5 gas=145
//   opcodes_control.gno:19 ControlOps: count=1 gas=46
//   opcodes_control.gno:21 -: count=8 gas=487
//   opcodes_control.gno:27 -: count=2 gas=108
//   opcodes_control.gno:29 ControlOps: count=2 gas=108
//   opcodes_control.gno:30 ControlOps: count=2 gas=108
//   opcodes_control.gno:34 OpDefer: count=6 gas=385
//   opcodes_control.gno:36 OpDefer: count=5 gas=264
//   opcodes_control.gno:40 OpDefer: count=5 gas=264
//   opcodes_control.gno:41 -: count=6 gas=629
//   opcodes_control.gno:48 OpLor: count=2 gas=140
//   opcodes_control.gno:49 OpLor: count=2 gas=140
//   opcodes_control.gno:50 OpLor: count=6 gas=210
//   opcodes_control.gno:51 OpLor: count=4 gas=155
//   opcodes_control.gno:55 OpLand: count=2 gas=140
//   opcodes_control.gno:56 OpLand: count=2 gas=140
//   opcodes_control.gno:57 OpLand: count=6 gas=208
//   opcodes_control.gno:61 OpPanic: count=5 gas=264
// SubOps:
//   ArrayElem: count=1
//   AssignVar: count=8
//   DefineVar: count=5
//   ParamAssign: count=2
// Variables:
//   opcodes_control.gno:7 _: count=1
//   opcodes_control.gno:27 a: count=1
//   opcodes_control.gno:29 _: count=1
//   opcodes_control.gno:30 _: count=1
//   opcodes_control.gno:34 #0: count=1
//   opcodes_control.gno:34 a: count=1
//   opcodes_control.gno:41 _: count=1
//   opcodes_control.gno:48 a: count=1
//   opcodes_control.gno:49 b: count=1
//   opcodes_control.gno:50 _: count=1
//   opcodes_control.gno:51 _: count=1
//   opcodes_control.gno:55 a: count=1
//   opcodes_control.gno:56 b: count=1
//   opcodes_control.gno:57 _: count=1
