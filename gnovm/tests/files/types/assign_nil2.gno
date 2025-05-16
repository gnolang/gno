package main

func main() {
	var i any
	i = 4
	var ok bool
	nil, nil = i.(int)
}

// Error:
// main/files/types/assign_nil2.gno:7:2: cannot assign to uverse nil

// TypeCheckError:
// main/files/types/assign_nil2.gno:7:2: cannot assign to nil (neither addressable nor a map index expression); main/files/types/assign_nil2.gno:7:7: cannot assign to nil (neither addressable nor a map index expression); main/files/types/assign_nil2.gno:6:6: declared and not used: ok
