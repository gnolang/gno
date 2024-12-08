package main

func main() {
	a := 1
	i := a + nil
}

// Error:
// main/files/add3.gno:5:7: invalid operation: a<VPBlock(1,0)> + (const (undefined)) (mismatched types int and untyped nil)
