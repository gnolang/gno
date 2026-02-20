package main

func main() {
	var p *[3]int
	_ = p[0:1]
}

// Error:
// nil pointer dereference

// Stacktrace:
// panic: nil pointer dereference
// main<VPBlock(1,0)>()
//     main/ptr_array4a.gno:5
