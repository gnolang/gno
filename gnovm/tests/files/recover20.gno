package main

func main() {
	var p *int
	if i := *p; i == 0 {
		println("hey")
	}
}

// Error:
// runtime error: invalid memory address or nil pointer dereference

// Stacktrace:
// panic: runtime error: invalid memory address or nil pointer dereference
// main<VPBlock(1,0)>()
//     main/recover20.gno:5
