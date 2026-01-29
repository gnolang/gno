package main

// Tests memory allocation patterns with large data structures.
// Expected: Memory allocation may not be reflected in gas.

func main() {
	LargeSlice()
	DeepNesting()
	LargeMap()
}

func LargeSlice() {
	s := make([]int, 100)
	for i := range s {
		s[i] = i * 2
	}
	_ = s[50]
}

func DeepNesting() {
	type Node struct {
		Value int
		Next  *Node
	}
	var head *Node
	for i := 0; i < 20; i++ {
		head = &Node{Value: i, Next: head}
	}
	// Traverse
	count := 0
	for n := head; n != nil; n = n.Next {
		count++
	}
	_ = count
}

func LargeMap() {
	m := make(map[int]int)
	for i := 0; i < 50; i++ {
		m[i] = i * 2
	}
	_ = m[25]
}

// Gasprofile:
// Opcodes:
//   OpAssign: count=193 gas=15247
//   OpBody: count=19 gas=817
//   OpCall: count=6 gas=1536
//   OpCallNativeBody: count=2 gas=848
//   OpCompositeLit: count=20 gas=1000
//   OpDefine: count=6 gas=666
//   OpEval: count=1243 gas=36047
//   OpExec: count=10 gas=250
//   OpFieldType: count=2 gas=118
//   OpForLoop: count=183 gas=4941
//   OpFuncType: count=4 gas=324
//   OpHalt: count=83 gas=83
//   OpInc: count=90 gas=6840
//   OpIndex1: count=2 gas=28
//   OpLss: count=72 gas=936
//   OpMapType: count=2 gas=118
//   OpMul: count=150 gas=2850
//   OpNeq: count=21 gas=1995
//   OpPopBlock: count=76 gas=228
//   OpPopResults: count=4 gas=4
//   OpPrecall: count=6 gas=1242
//   OpRangeIter: count=200 gas=21000
//   OpRef: count=20 gas=2500
//   OpReturn: count=2 gas=76
//   OpReturnFromBlock: count=4 gas=144
//   OpSelector: count=20 gas=640
//   OpSliceType: count=2 gas=110
//   OpStar: count=3 gas=120
//   OpStructLit: count=20 gas=3580
//   OpStructType: count=1 gas=174
//   OpTypeDecl: count=1 gas=100
//   OpValueDecl: count=1 gas=113
// Store:
//   StoreGetObject: count=1 size=0
// HotSpots:
//   large_data.gno:13 LargeSlice: count=9 gas=774
//   large_data.gno:15 -: count=700 gas=24300
//   large_data.gno:17 LargeSlice: count=5 gas=180
//   large_data.gno:21 DeepNesting: count=3 gas=154
//   large_data.gno:25 DeepNesting: count=5 gas=236
//   large_data.gno:26 -: count=23 gas=1685
//   large_data.gno:27 -: count=180 gas=11560
//   large_data.gno:30 DeepNesting: count=2 gas=140
//   large_data.gno:31 -: count=83 gas=3545
//   large_data.gno:32 -: count=20 gas=1520
//   large_data.gno:34 DeepNesting: count=2 gas=108
//   large_data.gno:38 LargeMap: count=9 gas=778
//   large_data.gno:39 -: count=53 gas=3965
//   large_data.gno:40 -: count=350 gas=12150
//   large_data.gno:42 LargeMap: count=5 gas=180
// SubOps:
//   AssignIndex: count=150
//   AssignVar: count=43
//   DefineVar: count=6
//   ParamAssign: count=4
//   RangeKey: count=100
//   StructField: count=40
// Variables:
//   large_data.gno:13 s: count=1
//   large_data.gno:15 #0: count=100
//   large_data.gno:17 _: count=1
//   large_data.gno:26 i: count=1
//   large_data.gno:27 Next: count=20
//   large_data.gno:27 Value: count=20
//   large_data.gno:27 head: count=20
//   large_data.gno:30 count: count=1
//   large_data.gno:31 n: count=21
//   large_data.gno:34 _: count=1
//   large_data.gno:38 m: count=1
//   large_data.gno:39 i: count=1
//   large_data.gno:40 #0: count=50
//   large_data.gno:42 _: count=1
