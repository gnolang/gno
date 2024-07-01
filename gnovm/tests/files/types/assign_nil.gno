package main

func main() {
	var i interface{}
	i = 4
	var j int
	j, nil = i.(int)
}

// Error:
// main/files/types/assign_nil.gno:7: cannot assign to uverse nil
