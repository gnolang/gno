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
//   OpBody: count=5 gas=215
//   OpCall: count=2 gas=512
//   OpCallNativeBody: count=1 gas=424
//   OpDefine: count=3 gas=333
//   OpEval: count=14 gas=406
//   OpExec: count=2 gas=50
//   OpFuncType: count=1 gas=81
//   OpHalt: count=10 gas=10
//   OpPopBlock: count=9 gas=27
//   OpPopResults: count=2 gas=2
//   OpPrecall: count=2 gas=414
//   OpReturn: count=1 gas=38
//   OpReturnFromBlock: count=1 gas=36
// Store:
//   StoreGetObject: count=1 size=0
// Native:
//   NativePrint: count=1
// HotSpots:
//   gasprofile_basic.gno:4 main: count=2 gas=140
//   gasprofile_basic.gno:5 main: count=2 gas=140
//   gasprofile_basic.gno:6 main: count=5 gas=216
// SubOps:
//   DefineVar: count=3
//   ParamAssign: count=1
