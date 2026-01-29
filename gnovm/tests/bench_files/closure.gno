package main

// Tests closure variable capture and nested closures.
// Expected: Variable capture may have hidden allocation costs not reflected in gas.

func main() {
	SimpleClosure()
	NestedClosure()
	ClosureInLoop()
}

func SimpleClosure() {
	x := 10
	f := func() int {
		return x * 2 // Captures x
	}
	_ = f()
}

func NestedClosure() {
	x := 1
	f := func() func() int {
		y := 2
		return func() int {
			return x + y // Captures both x and y
		}
	}
	_ = f()()
}

func ClosureInLoop() {
	sum := 0
	for i := 0; i < 5; i++ {
		f := func() int {
			return i // Captures loop variable
		}
		sum += f()
	}
	_ = sum
}

// Gasprofile:
// Opcodes:
//   OpAdd: count=1 gas=18
//   OpAddAssign: count=5 gas=425
//   OpAssign: count=3 gas=237
//   OpBody: count=25 gas=1075
//   OpCall: count=12 gas=3072
//   OpDefine: count=12 gas=1332
//   OpEval: count=138 gas=4002
//   OpExec: count=6 gas=150
//   OpFieldType: count=20 gas=1180
//   OpForLoop: count=16 gas=432
//   OpFuncLit: count=8 gas=488
//   OpFuncType: count=24 gas=1944
//   OpHalt: count=46 gas=46
//   OpInc: count=5 gas=380
//   OpLss: count=6 gas=78
//   OpMul: count=1 gas=19
//   OpPopBlock: count=45 gas=135
//   OpPopResults: count=4 gas=4
//   OpPrecall: count=12 gas=2484
//   OpReturn: count=8 gas=304
//   OpReturnFromBlock: count=4 gas=144
// Store:
//   StoreGetObject: count=1 size=0
// HotSpots:
//   closure.gno:13 SimpleClosure: count=2 gas=140
//   closure.gno:14 SimpleClosure: count=8 gas=428
//   closure.gno:17 SimpleClosure: count=5 gas=600
//   closure.gno:21 NestedClosure: count=2 gas=140
//   closure.gno:22 NestedClosure: count=12 gas=626
//   closure.gno:23 -: count=2 gas=140
//   closure.gno:28 NestedClosure: count=8 gas=1092
//   closure.gno:32 ClosureInLoop: count=2 gas=140
//   closure.gno:33 -: count=8 gas=545
//   closure.gno:34 -: count=40 gas=2140
//   closure.gno:37 -: count=25 gas=3030
//   closure.gno:39 ClosureInLoop: count=2 gas=108
// SubOps:
//   AssignVar: count=3
//   DefineVar: count=12
// Variables:
//   closure.gno:13 x: count=1
//   closure.gno:14 f: count=1
//   closure.gno:17 _: count=1
//   closure.gno:21 x: count=1
//   closure.gno:22 f: count=1
//   closure.gno:23 y: count=1
//   closure.gno:28 _: count=1
//   closure.gno:32 sum: count=1
//   closure.gno:33 i: count=1
//   closure.gno:34 f: count=5
//   closure.gno:39 _: count=1
