package main

func main() {
	var p *int
	println(*p)
}

// Stacktrace:
// panic: runtime error: nil pointer dereference
// main<VPBlock(1,0)>()
//     main/panic3a.gno:5

// Error:
// runtime error: nil pointer dereference
