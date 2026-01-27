package main

func add(a, b int) int {
	return a + b
}

func multiply(a, b int) int {
	return a * b
}

func main() {
	x := add(1, 2)
	y := multiply(x, 4)
	println(y)
}

// Output:
// 12

// Gasprofile:
// Opcodes:
//   OpAdd: count=1 gas=18
//   OpBody: count=6 gas=258
//   OpCall: count=4 gas=1024
//   OpCallNativeBody: count=1 gas=424
//   OpDefine: count=2 gas=222
//   OpEval: count=44 gas=1276
//   OpExec: count=2 gas=50
//   OpFieldType: count=6 gas=354
//   OpFuncType: count=3 gas=243
//   OpHalt: count=25 gas=25
//   OpMul: count=1 gas=19
//   OpPopBlock: count=24 gas=72
//   OpPopResults: count=2 gas=2
//   OpPrecall: count=4 gas=828
//   OpReturn: count=3 gas=114
//   OpReturnFromBlock: count=1 gas=36
// Store:
//   StoreGetObject: count=1 size=0
// Native:
//   NativePrint: count=1
// HotSpots:
//   gasprofile_func.gno:12 main: count=7 gas=690
//   gasprofile_func.gno:13 main: count=7 gas=690
// SubOps:
//   DefineVar: count=2
//   ParamAssign: count=5
