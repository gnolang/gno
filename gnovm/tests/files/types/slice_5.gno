package main

func main() {
	var p **[3]int
	_ = p[0:1]
}

// Error:
// main/slice_5.gno:5:6-12: cannot slice variable of type **[3]int

// TypeCheckError:
// main/slice_5.gno:5:6: cannot slice p (variable of type **[3]int)
