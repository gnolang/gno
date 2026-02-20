package main

func main() {
	var p *[3]int
	for i, v := range p {
		println(i, v)
	}
}

// Error:
// nil pointer dereference

// Stacktrace:
// panic: nil pointer dereference
// main<VPBlock(1,0)>()
//     main/ptr_array5a.gno:5
