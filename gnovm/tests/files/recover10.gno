package main

func main() {
	defer recover()

	panic("ahhhhh")
}

// Stacktrace:
// panic: ahhhhh
// main<VPBlock(1,0)>()
//     main/files/recover10.gno:6

// Error:
// panic: ahhhhh
