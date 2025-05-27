package main

func main() {
	a := 1
	i := a + nil
}

// Error:
// main/add3.gno:5:7-14: invalid operation: a<VPBlock(1,0)> + (const (undefined)) (mismatched types int and untyped nil)

// TypeCheckError:
// main/add3.gno:5:7: invalid operation: a + nil (mismatched types int and untyped nil); main/add3.gno:5:2: declared and not used: i
