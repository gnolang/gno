package main

// Tests string operations at different scales to quantify gas vs time variability.
// Expected: Gas should scale with string size, but currently doesn't (2427x issue).

func main() {
	StringConcat10()
	StringConcat100()
	StringSlice()
	StringIteration()
}

// Test string concat with 10 iterations
func StringConcat10() {
	s := ""
	for i := 0; i < 10; i++ {
		s += "a"
	}
	_ = s
}

// Test string concat with 100 iterations (should show ~10x more gas)
func StringConcat100() {
	s := ""
	for i := 0; i < 100; i++ {
		s += "a"
	}
	_ = s
}

// Test string slicing
func StringSlice() {
	s := "hello world this is a test string"
	_ = s[5:10]
	_ = s[:5]
	_ = s[5:]
}

// Test string iteration
func StringIteration() {
	s := "hello"
	for _, c := range s {
		_ = c
	}
}

// Gasprofile:
// Opcodes:
//   OpAddAssign: count=110 gas=9350
//   OpAssign: count=10 gas=790
//   OpBody: count=21 gas=903
//   OpCall: count=5 gas=1280
//   OpDefine: count=6 gas=666
//   OpEval: count=499 gas=14471
//   OpExec: count=8 gas=200
//   OpForLoop: count=222 gas=5994
//   OpFuncType: count=5 gas=405
//   OpHalt: count=53 gas=53
//   OpInc: count=110 gas=8360
//   OpLss: count=112 gas=1456
//   OpPopBlock: count=49 gas=147
//   OpPopResults: count=5 gas=5
//   OpPrecall: count=5 gas=1035
//   OpRangeIterString: count=10 gas=550
//   OpReturnFromBlock: count=5 gas=180
//   OpSlice: count=3 gas=309
// Store:
//   StoreGetObject: count=1 size=0
// HotSpots:
//   string_scaling.gno:15 StringConcat10: count=2 gas=140
//   string_scaling.gno:16 -: count=13 gas=925
//   string_scaling.gno:17 -: count=20 gas=1140
//   string_scaling.gno:19 StringConcat10: count=2 gas=108
//   string_scaling.gno:24 StringConcat100: count=2 gas=140
//   string_scaling.gno:25 -: count=103 gas=7765
//   string_scaling.gno:26 -: count=200 gas=11400
//   string_scaling.gno:28 StringConcat100: count=2 gas=108
//   string_scaling.gno:33 StringSlice: count=2 gas=140
//   string_scaling.gno:34 StringSlice: count=6 gas=298
//   string_scaling.gno:35 StringSlice: count=5 gas=269
//   string_scaling.gno:36 StringSlice: count=5 gas=269
//   string_scaling.gno:41 StringIteration: count=2 gas=140
//   string_scaling.gno:43 -: count=10 gas=540
// SubOps:
//   AssignVar: count=10
//   DefineVar: count=6
//   RangeKey: count=5
//   RangeValue: count=5
// Variables:
//   string_scaling.gno:15 s: count=1
//   string_scaling.gno:16 i: count=1
//   string_scaling.gno:19 _: count=1
//   string_scaling.gno:24 s: count=1
//   string_scaling.gno:25 i: count=1
//   string_scaling.gno:28 _: count=1
//   string_scaling.gno:33 s: count=1
//   string_scaling.gno:34 _: count=1
//   string_scaling.gno:35 _: count=1
//   string_scaling.gno:36 _: count=1
//   string_scaling.gno:41 s: count=1
//   string_scaling.gno:43 _: count=5
