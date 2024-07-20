package main

type A interface {
	Do(s string)
}

func main() {
	defer func() {
		e := recover()
		_ = e.(A)
	}()
}

// Stacktrace:
// panic interface conversion: interface is nil, not main.A
// _<VPBlock(0,0)> = e<VPBlock(1,0)>.(A<VPBlock(4,0)>)
//     main/files/typeassert2a.gno:10
// main()
//     main/files/typeassert2a.gno:7

// Error:
// interface conversion: interface is nil, not main.A
