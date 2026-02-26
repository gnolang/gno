package main

func main() {

	var p *[]int
	for i, v := range p {
		println(i, v)
	}
}

// Error:
// main/ptr_array5a.gno:6:2-8:3: range iteration over pointer requires array elem type

// TypeCheckError:
// main/ptr_array5a.gno:6:20: cannot range over p (variable of type *[]int)
