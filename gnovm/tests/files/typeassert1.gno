package main

type A interface {
	Do(s string)
}

func main() {
	var a A
	_ = a.(A)
}

// Stacktrace:
// panic interface conversion: interface is nil, not main.A
// _<VPBlock(0,0)> = a<VPBlock(1,0)>.(A<VPBlock(3,0)>)
//     main/files/typeassert1.gno:9
// main()
//     main/files/typeassert1.gno:7

// Error:
// interface conversion: interface is nil, not main.A
