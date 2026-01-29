package main

// Tests recursive calls to measure stack growth and OpCall accumulation.
// Expected: Gas should scale linearly with recursion depth.

func main() {
	_ = Fibonacci(15) // 15 to keep test fast
	_ = Factorial(10)
}

func Fibonacci(n int) int {
	if n <= 1 {
		return n
	}
	return Fibonacci(n-1) + Fibonacci(n-2)
}

func Factorial(n int) int {
	if n <= 1 {
		return 1
	}
	return n * Factorial(n-1)
}

// Gasprofile:
// Opcodes:
//   OpAdd: count=986 gas=17748
//   OpAssign: count=2 gas=158
//   OpBody: count=3969 gas=170667
//   OpCall: count=1984 gas=507904
//   OpEval: count=17877 gas=518433
//   OpExec: count=2 gas=50
//   OpFieldType: count=4 gas=236
//   OpFuncType: count=3 gas=243
//   OpHalt: count=38 gas=38
//   OpIfCond: count=1983 gas=75354
//   OpLeq: count=1983 gas=37677
//   OpMul: count=9 gas=171
//   OpPopBlock: count=1032 gas=3096
//   OpPopResults: count=1 gas=1
//   OpPrecall: count=1984 gas=410688
//   OpReturn: count=1983 gas=75354
//   OpReturnFromBlock: count=1 gas=36
//   OpSub: count=1981 gas=11886
// Store:
//   StoreGetObject: count=1 size=0
// HotSpots:
//   recursion.gno:7 main: count=6 gas=629
//   recursion.gno:8 main: count=6 gas=629
//   recursion.gno:12 Fibonacci: count=9865 gas=284112
//   recursion.gno:19 Factorial: count=50 gas=1440
// SubOps:
//   AssignVar: count=2
//   ParamAssign: count=1983
// Variables:
//   recursion.gno:7 _: count=1
//   recursion.gno:8 _: count=1
//   recursion.gno:11 #0: count=1973
//   recursion.gno:18 #0: count=10
