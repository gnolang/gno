package main

func main() {
	sum := 0
	for i := 0; i < 10; i++ {
		sum += i
	}
	println(sum)
}

// Output:
// 45

// Gasprofile:
// Opcodes:
//   OpAddAssign: count=10 gas=85
//   OpBody: count=4 gas=43
//   OpCall: count=2 gas=256
//   OpCallNativeBody: count=1 gas=424
//   OpDefine: count=2 gas=111
//   OpEval: count=55 gas=29
//   OpExec: count=3 gas=25
//   OpForLoop: count=21 gas=27
//   OpFuncType: count=1 gas=81
//   OpHalt: count=13 gas=1
//   OpInc: count=10 gas=76
//   OpLss: count=11 gas=13
//   OpPopBlock: count=12 gas=3
//   OpPopResults: count=2 gas=1
//   OpPrecall: count=2 gas=207
//   OpReturn: count=1 gas=38
//   OpReturnFromBlock: count=1 gas=36
//   OpVoid: count=158 gas=1
// Store:
//   StoreGetObject: count=1 size=0
// Native:
//   NativePrint: count=1
