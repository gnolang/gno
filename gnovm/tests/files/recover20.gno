package main

func main() {
	var p *int
	if i := *p; i == 0 {
		println("hey")
	}
}

// Error:
// nil pointer dereference

// Stacktrace:
// panic: nil pointer dereference
// main<VPBlock(1,0)>()
//     main/recover20.gno:5
