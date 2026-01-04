package main

func main() {
	var i any
	i = 4
	var j int
	j, nil = i.(int)
}

// Error:
// main/assign_nil.gno:7:2-18: cannot assign to uverse nil

// TypeCheckError:
// main/assign_nil.gno:7:5: cannot assign to nil (neither addressable nor a map index expression); main/assign_nil.gno:6:6: declared and not used: j
