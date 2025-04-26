package main

func main() {
	f()
}

func f() {
	defer func() {
		panic("third")
	}()
	defer func() {
		panic("second")
	}()
	panic("first")
}

// Stacktrace:
// panic: first
// f<VPBlock(3,1)>()
//     main/files/panic0b.gno:14
// main<VPBlock(1,0)>()
//     main/files/panic0b.gno:4
// ... 1 panic(s) elided ...
// panic: third
// defer func(){ ... }()
//     main/files/panic0b.gno:9
// f<VPBlock(3,1)>()
//     main/files/panic0b.gno:8
// main<VPBlock(1,0)>()
//     main/files/panic0b.gno:4

// Error:
// first
// 	second
// 	third
