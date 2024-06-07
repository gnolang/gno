package main

func foo() (int, bool) {
	return 1, true
}

func main() {
	var a, b = 2, foo()

	println(a, b)
}

// Error:
// main/files/var21.gno:8: multiple-value foo<VPBlock(3,0)> (value of type [int bool]) in single-value context:
// --- preprocess stack ---
// stack 0: func main() { var a<VPBlock(1,0)>, b<VPBlock(1,1)> = (const (2 <untyped> bigint)), foo<VPBlock(3,0)>(); println<VPUverse(0)>(a<VPUverse(0)>, b<VPUverse(0)>) }
// ------------------------
