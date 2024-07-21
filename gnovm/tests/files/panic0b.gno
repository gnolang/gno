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
// panic first
// f<VPBlock(3,1)>()
//     main/files/panic0b.gno:14
// main<VPBlock(1,0)>()
//     main/files/panic0b.gno:4
// panic second
// f<VPBlock(3,1)>()
//     main/files/panic0b.gno:12
// main<VPBlock(1,0)>()
//     main/files/panic0b.gno:4
// panic third
// f<VPBlock(3,1)>()
//     main/files/panic0b.gno:9
// main<VPBlock(1,0)>()
//     main/files/panic0b.gno:4

// Error:
// first
// 	second
// 	third
