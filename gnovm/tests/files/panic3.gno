package main

func main() {
	var p *int
	*p = 2
}

// Stacktrace:
// panic: runtime error: nil pointer dereference
// main<VPBlock(1,0)>()
//     main/panic3.gno:5

// Error:
// runtime error: nil pointer dereference
