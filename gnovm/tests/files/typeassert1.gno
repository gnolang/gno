package main

type A interface {
	Do(s string)
}

func main() {
	var a A
	_ = a.(A)
}

// Stacktrace:
// panic: interface conversion: interface is nil, not main.A
// main<VPBlock(1,1)>()
//     main/files/typeassert1.gno:9

// Error:
// interface conversion: interface is nil, not main.A
