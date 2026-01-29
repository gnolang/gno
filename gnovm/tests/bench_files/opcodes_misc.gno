package main

// Test miscellaneous operations: deferred native function calls

func main() {
	OpCallDeferNativeBody()
}

func OpCallDeferNativeBody() {
	a := [1]int{0}
	defer func() {
		_ = len(a)
	}()
}

// Gasprofile:
// Opcodes:
//   OpArrayLit: count=1 gas=137
//   OpArrayType: count=1 gas=57
//   OpAssign: count=1 gas=79
//   OpBody: count=7 gas=301
//   OpCall: count=3 gas=768
//   OpCallNativeBody: count=1 gas=424
//   OpCompositeLit: count=1 gas=50
//   OpDefer: count=1 gas=64
//   OpDefine: count=1 gas=111
//   OpEval: count=25 gas=725
//   OpExec: count=4 gas=100
//   OpFuncLit: count=1 gas=61
//   OpFuncType: count=5 gas=405
//   OpHalt: count=16 gas=16
//   OpPopBlock: count=15 gas=45
//   OpPopResults: count=2 gas=2
//   OpPrecall: count=3 gas=621
//   OpReturn: count=1 gas=38
//   OpReturnCallDefers: count=2 gas=156
//   OpReturnFromBlock: count=3 gas=108
// Store:
//   StoreGetObject: count=1 size=0
// HotSpots:
//   opcodes_misc.gno:10 OpCallDeferNativeBody: count=6 gas=385
//   opcodes_misc.gno:11 OpCallDeferNativeBody: count=5 gas=264
//   opcodes_misc.gno:12 -: count=6 gas=629
// SubOps:
//   ArrayElem: count=1
//   AssignVar: count=1
//   DefineVar: count=1
//   ParamAssign: count=1
// Variables:
//   opcodes_misc.gno:10 #0: count=1
//   opcodes_misc.gno:10 a: count=1
//   opcodes_misc.gno:12 _: count=1
