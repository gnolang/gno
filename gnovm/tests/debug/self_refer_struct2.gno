package main

type A struct {
	X B
}

type B struct {
	X C
}

type C struct {
	X A
}

func main() {
	var p, q A
	println(p == q)
}

// Error:
// main/debug/self_refer_struct2.gno:1: Cyclic dependency detected: [C A B C]
// :
// --- preprocess stack ---
// stack 0: file{ package main; type A (const-type main.A); type B (const-type main.B); type C (const-type main.C); func main() { var p<VPUverse(0)>, q<VPUverse(0)> A<VPUverse(0)>; println<VPUverse(0)>(p<VPUverse(0)> == q<VPUverse(0)>) } }
// ------------------------
