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
//   OpAddAssign: count=10 gas=850
//   OpBody: count=4 gas=172
//   OpCall: count=2 gas=512
//   OpCallNativeBody: count=1 gas=424
//   OpDefine: count=2 gas=222
//   OpEval: count=55 gas=1595
//   OpExec: count=3 gas=75
//   OpForLoop: count=21 gas=567
//   OpFuncType: count=1 gas=81
//   OpHalt: count=13 gas=13
//   OpInc: count=10 gas=760
//   OpLss: count=11 gas=143
//   OpPopBlock: count=12 gas=36
//   OpPopResults: count=2 gas=2
//   OpPrecall: count=2 gas=414
//   OpReturn: count=1 gas=38
//   OpReturnFromBlock: count=1 gas=36
// Store:
//   StoreGetObject: count=1 size=0
// Native:
//   NativePrint: count=1
// HotSpots:
//   gasprofile_loop.gno:4 main: count=2 gas=140
//   gasprofile_loop.gno:5 -: count=13 gas=925
//   gasprofile_loop.gno:6 -: count=20 gas=1140
// SubOps:
//   DefineVar: count=2
//   ParamAssign: count=1
