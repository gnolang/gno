package main

func main() {
	var p int
	_ = p[0:1]
}

// TypeCheckError:
// main/ptr_array4d.gno:5:6: cannot slice p (variable of type int)

// Error:
// main/ptr_array4d.gno:5:6-12: cannot slice variable of type int
