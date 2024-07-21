package main

func main() {
	panic("wtf")
}

// Stacktrace:
// panic wtf
// main<VPBlock(1,0)>()
//     main/files/panic0.gno:4

// Error:
// wtf
