package main

// Test expression operations: arrays, maps, slices, structs, indexing, refs, type assertions

type foo struct {
	i int
}

func (f foo) bark() {}

type dog interface {
	bark()
}

type foofighter struct {
	f foo
}

func main() {
	ExprOps()
}

func ExprOps() {
	a := [2]int{0, 1} // OpArrayLit

	a2 := [...]int{0, 1}                     // same as [2]int
	m := make(map[int]int)                   // OpMapLit
	s := []int{0, 1, 2, 3, 4, 5, 6, 7, 8, 9} // OpSliceLit

	s2 := []int{9: 90}     // OpSliceLit2
	f := foo{i: 1}         // OpStructLit
	ff := foofighter{f: f} // OpCompositeLit
	_ = ff
	_ = s2

	b := a[0]   // OpIndex1
	b, _ = m[0] // OpIndex2
	b = f.i     // OpSelector
	_ = b

	subs := s[1:5:10] // OpSlice
	_ = subs

	ptr := &a2[0] // OpRef
	_ = *ptr      // OpStar

	var d dog
	d = f
	f = d.(foo)      // OpTypeAssert1 concrete type
	f, ok := d.(foo) // OpTypeAssert2
	_ = f
	_ = ok
}

// Gasprofile:
// Opcodes:
//   OpArrayLit: count=2 gas=274
//   OpArrayType: count=2 gas=114
//   OpAssign: count=11 gas=869
//   OpBody: count=26 gas=1118
//   OpCall: count=3 gas=768
//   OpCallNativeBody: count=1 gas=424
//   OpCompositeLit: count=6 gas=300
//   OpDefine: count=11 gas=1221
//   OpEval: count=152 gas=4408
//   OpExec: count=7 gas=175
//   OpFieldType: count=4 gas=236
//   OpFuncType: count=5 gas=405
//   OpHalt: count=112 gas=112
//   OpIndex1: count=1 gas=14
//   OpIndex2: count=1 gas=20
//   OpInterfaceType: count=1 gas=75
//   OpMapType: count=2 gas=118
//   OpPopBlock: count=101 gas=303
//   OpPopResults: count=2 gas=2
//   OpPrecall: count=3 gas=621
//   OpRef: count=1 gas=125
//   OpReturn: count=1 gas=38
//   OpReturnFromBlock: count=2 gas=72
//   OpSelector: count=1 gas=32
//   OpSlice: count=1 gas=103
//   OpSliceLit: count=1 gas=183
//   OpSliceLit2: count=1 gas=92
//   OpSliceType: count=2 gas=110
//   OpStar: count=1 gas=40
//   OpStructLit: count=2 gas=358
//   OpStructType: count=2 gas=348
//   OpTypeAssert1: count=1 gas=22
//   OpTypeAssert2: count=1 gas=34
//   OpTypeDecl: count=3 gas=300
//   OpValueDecl: count=1 gas=113
// Store:
//   StoreGetObject: count=1 size=0
// HotSpots:
//   opcodes_expr.gno:5 -: count=3 gas=154
//   opcodes_expr.gno:11 -: count=3 gas=154
//   opcodes_expr.gno:15 -: count=3 gas=154
//   opcodes_expr.gno:24 ExprOps: count=7 gas=414
//   opcodes_expr.gno:26 ExprOps: count=7 gas=414
//   opcodes_expr.gno:27 ExprOps: count=9 gas=778
//   opcodes_expr.gno:28 ExprOps: count=15 gas=692
//   opcodes_expr.gno:30 ExprOps: count=7 gas=369
//   opcodes_expr.gno:31 ExprOps: count=6 gas=427
//   opcodes_expr.gno:32 ExprOps: count=6 gas=427
//   opcodes_expr.gno:33 ExprOps: count=2 gas=108
//   opcodes_expr.gno:34 ExprOps: count=2 gas=108
//   opcodes_expr.gno:36 ExprOps: count=5 gas=212
//   opcodes_expr.gno:37 ExprOps: count=5 gas=186
//   opcodes_expr.gno:38 ExprOps: count=4 gas=169
//   opcodes_expr.gno:39 ExprOps: count=2 gas=108
//   opcodes_expr.gno:41 ExprOps: count=7 gas=359
//   opcodes_expr.gno:42 ExprOps: count=2 gas=108
//   opcodes_expr.gno:44 ExprOps: count=5 gas=323
//   opcodes_expr.gno:45 ExprOps: count=4 gas=177
//   opcodes_expr.gno:47 ExprOps: count=3 gas=167
//   opcodes_expr.gno:48 ExprOps: count=2 gas=108
//   opcodes_expr.gno:49 ExprOps: count=5 gas=188
//   opcodes_expr.gno:50 ExprOps: count=5 gas=232
//   opcodes_expr.gno:51 ExprOps: count=2 gas=108
//   opcodes_expr.gno:52 ExprOps: count=2 gas=108
// SubOps:
//   ArrayElem: count=4
//   AssignVar: count=12
//   DefineVar: count=12
//   ParamAssign: count=2
//   StructField: count=2
// Variables:
//   opcodes_expr.gno:24 #0: count=1
//   opcodes_expr.gno:24 #1: count=1
//   opcodes_expr.gno:24 a: count=1
//   opcodes_expr.gno:26 #0: count=1
//   opcodes_expr.gno:26 #1: count=1
//   opcodes_expr.gno:26 a2: count=1
//   opcodes_expr.gno:27 m: count=1
//   opcodes_expr.gno:28 s: count=1
//   opcodes_expr.gno:30 s2: count=1
//   opcodes_expr.gno:31 f: count=1
//   opcodes_expr.gno:31 i: count=1
//   opcodes_expr.gno:32 f: count=1
//   opcodes_expr.gno:32 ff: count=1
//   opcodes_expr.gno:33 _: count=1
//   opcodes_expr.gno:34 _: count=1
//   opcodes_expr.gno:36 b: count=1
//   opcodes_expr.gno:37 _: count=1
//   opcodes_expr.gno:37 b: count=1
//   opcodes_expr.gno:38 b: count=1
//   opcodes_expr.gno:39 _: count=1
//   opcodes_expr.gno:41 subs: count=1
//   opcodes_expr.gno:42 _: count=1
//   opcodes_expr.gno:44 ptr: count=1
//   opcodes_expr.gno:45 _: count=1
//   opcodes_expr.gno:48 d: count=1
//   opcodes_expr.gno:49 f: count=1
//   opcodes_expr.gno:50 f: count=1
//   opcodes_expr.gno:50 ok: count=1
//   opcodes_expr.gno:51 _: count=1
//   opcodes_expr.gno:52 _: count=1
