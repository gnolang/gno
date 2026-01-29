package main

// Test native function calls with different string sizes

func main() {
	Print_1()
	Print_100()
	Print_1000()
}

func Print_1() {
	b := ""
	for i := 0; i < 1; i++ {
		b += "a"
	}
	print(b)
}

func Print_100() {
	b := ""
	for i := 0; i < 100; i++ {
		b += "a"
	}
	print(b)
}

func Print_1000() {
	b := ""
	for i := 0; i < 1000; i++ {
		b += "a"
	}
	print(b)
}

// Output:
// aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa

// Gasprofile:
// Opcodes:
//   OpAddAssign: count=1101 gas=93585
//   OpBody: count=16 gas=688
//   OpCall: count=7 gas=1792
//   OpCallNativeBody: count=3 gas=1272
//   OpDefine: count=6 gas=666
//   OpEval: count=4455 gas=129195
//   OpExec: count=8 gas=200
//   OpForLoop: count=2205 gas=59535
//   OpFuncType: count=4 gas=324
//   OpHalt: count=39 gas=39
//   OpInc: count=1101 gas=83676
//   OpLss: count=1104 gas=14352
//   OpPopBlock: count=38 gas=114
//   OpPopResults: count=7 gas=7
//   OpPrecall: count=7 gas=1449
//   OpReturn: count=3 gas=114
//   OpReturnFromBlock: count=4 gas=144
// Store:
//   StoreGetObject: count=1 size=0
// Native:
//   NativePrint: count=3
// HotSpots:
//   native.gno:12 Print_1: count=2 gas=140
//   native.gno:13 -: count=4 gas=241
//   native.gno:14 -: count=2 gas=114
//   native.gno:20 Print_100: count=2 gas=140
//   native.gno:21 -: count=103 gas=7765
//   native.gno:22 -: count=200 gas=11400
//   native.gno:28 Print_1000: count=2 gas=140
//   native.gno:29 -: count=1003 gas=76165
//   native.gno:30 -: count=2000 gas=114000
// SubOps:
//   DefineVar: count=6
//   ParamAssign: count=3
// Variables:
//   native.gno:12 b: count=1
//   native.gno:13 i: count=1
//   native.gno:20 b: count=1
//   native.gno:21 i: count=1
//   native.gno:28 b: count=1
//   native.gno:29 i: count=1
