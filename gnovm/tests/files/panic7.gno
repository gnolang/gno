package main

func main() {
	var p *int
	println(*p)	
}

// Stacktrace:
// panic: nil pointer dereference
// main<VPBlock(1,0)>()
//     main/panic7.gno:5

// Error:
// nil pointer dereference
