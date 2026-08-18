package main

func main() {
	var p *int
	_ = p[0:1]
}

// TypeCheckError:
// main/slice_2.gno:5:6: cannot slice p (variable of type *int)

// Error:
// main/slice_2.gno:5:6-12: cannot slice variable of type *int
