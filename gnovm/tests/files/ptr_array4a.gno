package main

func main() {
	var p *[]int
	_ = p[0:1]
}

// Error:
// main/ptr_array4a.gno:5:6-12: slice pointer requires array elem type

// TypeCheckError:
// main/ptr_array4a.gno:5:6: cannot slice p (variable of type *[]int)
