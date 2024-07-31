package main

func main() {
	var i interface{}
	i = 4
	var ok bool
	nil, nil = i.(int)
}

// Error:
// main/files/types/assign_nil2.gno:7:2: cannot assign to uverse nil
