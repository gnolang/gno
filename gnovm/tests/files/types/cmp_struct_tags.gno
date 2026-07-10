package main

func main() {
	var x struct{ A int `a` }
	var y struct{ A int `b` }
	println(x == y)
}

// Error:
// main/cmp_struct_tags.gno:6:10-16: invalid operation: (mismatched types struct{A int "a"} and struct{A int "b"})

// TypeCheckError:
// main/cmp_struct_tags.gno:6:15: invalid operation: x == y (mismatched types struct{A int "a"} and struct{A int "b"})
