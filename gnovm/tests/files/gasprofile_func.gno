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
//   OpBody: count=6 gas=43
//   OpCall: count=4 gas=256
//   OpCallNativeBody: count=1 gas=424
//   OpDefine: count=2 gas=111
//   OpEval: count=44 gas=29
//   OpExec: count=2 gas=25
//   OpFieldType: count=6 gas=59
//   OpFuncType: count=3 gas=81
//   OpHalt: count=25 gas=1
//   OpMul: count=1 gas=19
//   OpPopBlock: count=24 gas=3
//   OpPopResults: count=2 gas=1
//   OpPrecall: count=4 gas=207
//   OpReturn: count=3 gas=38
//   OpReturnFromBlock: count=1 gas=36
//   OpVoid: count=142 gas=1
// Store:
//   StoreGetObject: count=1 size=0
// Native:
//   NativePrint: count=1
// HotSpots:
//   gasprofile_func.gno:12 main: count=7 gas=690
//   gasprofile_func.gno:13 main: count=7 gas=690
