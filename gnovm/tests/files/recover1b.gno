package main

func main() {
	defer func() {
		recover()
		panic("other panic")
	}()
	panic("test panic")
}

// Stacktrace:
// panic: test panic
// main<VPBlock(1,0)>()
//     main/files/recover1b.gno:8
// panic: other panic
// func(){ ... }()
//     main/:6
// main<VPBlock(1,0)>()
//     main/files/recover1b.gno:4

// Error:
// panic: test panic[recovered]
// 	panic: other panic
