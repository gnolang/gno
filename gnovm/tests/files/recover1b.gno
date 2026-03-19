package main

func main() {
	defer func() {
		recover()
		panic("other panic")
	}()
	panic("test panic")
}

// Stacktrace:
// panic: other panic
// defer func(){ ... }()
//     main/recover1b.gno:6
// main<VPBlock(1,0)>()
//     main/recover1b.gno:4

// Error:
// other panic

