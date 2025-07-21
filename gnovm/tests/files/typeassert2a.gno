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
// panic: interface conversion: interface is nil, not main.A
// defer func(){ ... }()
//     main/typeassert2a.gno:10
// main<VPBlock(1,1)>()
//     main/typeassert2a.gno:8

// Error:
// interface conversion: interface is nil, not main.A
