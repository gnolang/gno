package main

// Test loop operations: range (array, string, map, pointer), for loop, OpValues

func main() {
	OpRange()
	OpForLoop()
	OpOpValues()
}

func OpRange() {
	a := [1]int{1}
	s := "h"
	m := map[int]string{
		1: "one",
	}
	one := 1
	p := &[1]int{one}

	for i := range a {
		println(a[i])
	}

	for i := range s {
		println(s[i])
	}

	for i := range m {
		println(m[i])
	}

	for i := range p {
		println(p[i])
	}
}

func OpForLoop() {
	a := [1]int{0}
	for i := 0; i < 1; i++ {
		_ = len(a)
	}
}

func OpOpValues() {
	x := 1
	for i := 0; i < 1; i++ {
		_ = x + 1
	}

	if x == 1 {
		_ = x + 1
	}
}

// Output:
// 1
// 104
// one
// 1

// Gasprofile:
// Opcodes:
//   OpAdd: count=2 gas=36
//   OpArrayLit: count=3 gas=411
//   OpArrayType: count=3 gas=171
//   OpAssign: count=3 gas=237
//   OpBody: count=23 gas=989
//   OpCall: count=9 gas=2304
//   OpCallNativeBody: count=5 gas=2120
//   OpCompositeLit: count=4 gas=200
//   OpDefine: count=9 gas=999
//   OpEql: count=1 gas=160
//   OpEval: count=126 gas=3654
//   OpExec: count=7 gas=175
//   OpForLoop: count=6 gas=162
//   OpFuncType: count=4 gas=324
//   OpHalt: count=90 gas=90
//   OpIfCond: count=1 gas=38
//   OpInc: count=2 gas=152
//   OpIndex1: count=4 gas=56
//   OpLss: count=4 gas=52
//   OpMapLit: count=1 gas=475
//   OpMapType: count=1 gas=59
//   OpPopBlock: count=84 gas=252
//   OpPopResults: count=8 gas=8
//   OpPrecall: count=9 gas=1863
//   OpRangeIter: count=2 gas=210
//   OpRangeIterArrayPtr: count=2 gas=92
//   OpRangeIterMap: count=2 gas=96
//   OpRangeIterString: count=2 gas=110
//   OpRef: count=1 gas=125
//   OpReturn: count=5 gas=190
//   OpReturnFromBlock: count=4 gas=144
//   OpStar: count=1 gas=40
// Store:
//   StoreGetObject: count=1 size=0
// Native:
//   NativePrint: count=4
// HotSpots:
//   opcodes_loop.gno:12 OpRange: count=6 gas=385
//   opcodes_loop.gno:13 OpRange: count=2 gas=140
//   opcodes_loop.gno:14 OpRange: count=7 gas=752
//   opcodes_loop.gno:17 OpRange: count=2 gas=140
//   opcodes_loop.gno:18 OpRange: count=8 gas=539
//   opcodes_loop.gno:38 OpForLoop: count=6 gas=385
//   opcodes_loop.gno:39 -: count=4 gas=241
//   opcodes_loop.gno:40 -: count=6 gas=629
//   opcodes_loop.gno:45 OpOpValues: count=2 gas=140
//   opcodes_loop.gno:46 -: count=4 gas=241
//   opcodes_loop.gno:47 -: count=5 gas=184
//   opcodes_loop.gno:50 OpOpValues: count=5 gas=285
//   opcodes_loop.gno:51 OpOpValues: count=5 gas=184
// SubOps:
//   ArrayElem: count=3
//   AssignVar: count=3
//   DefineVar: count=9
//   MapEntry: count=1
//   ParamAssign: count=5
//   RangeKey: count=4
// Variables:
//   opcodes_loop.gno:12 #0: count=1
//   opcodes_loop.gno:12 a: count=1
//   opcodes_loop.gno:13 s: count=1
//   opcodes_loop.gno:14 #0: count=1
//   opcodes_loop.gno:14 m: count=1
//   opcodes_loop.gno:17 one: count=1
//   opcodes_loop.gno:18 #0: count=1
//   opcodes_loop.gno:18 p: count=1
//   opcodes_loop.gno:38 #0: count=1
//   opcodes_loop.gno:38 a: count=1
//   opcodes_loop.gno:39 i: count=1
//   opcodes_loop.gno:40 _: count=1
//   opcodes_loop.gno:45 x: count=1
//   opcodes_loop.gno:46 i: count=1
//   opcodes_loop.gno:47 _: count=1
//   opcodes_loop.gno:51 _: count=1
