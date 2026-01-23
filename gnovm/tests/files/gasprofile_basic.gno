package main

func main() {
	x := 1
	y := 2
	z := x + y
	println(z)
}

// Output:
// 3

// Gasprofile:
// Opcodes:
//   OpAdd: count=1 gas=18
//   OpBody: count=5 gas=43
//   OpCall: count=2 gas=256
//   OpCallNativeBody: count=1 gas=424
//   OpDefine: count=3 gas=111
//   OpEval: count=14 gas=29
//   OpExec: count=2 gas=25
//   OpFuncType: count=1 gas=81
//   OpHalt: count=10 gas=1
//   OpPopBlock: count=9 gas=3
//   OpPopResults: count=2 gas=1
//   OpPrecall: count=2 gas=207
//   OpReturn: count=1 gas=38
//   OpReturnFromBlock: count=1 gas=36
//   OpVoid: count=60 gas=1
// Store:
//   StoreGetObject: count=1 size=0
// Native:
//   NativePrint: count=1
// HotSpots:
//   gasprofile_basic.gno:4 main: count=2 gas=140
//   gasprofile_basic.gno:5 main: count=2 gas=140
//   gasprofile_basic.gno:6 main: count=5 gas=216
