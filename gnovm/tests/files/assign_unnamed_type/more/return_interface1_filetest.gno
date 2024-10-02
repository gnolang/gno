package main

func x1() interface{} {
	a := "1"
	return a
}

func main() {
	var a uint
	a = x1() // should fail
	println("1")
}

// Error:
// main/files/assign_unnamed_type/more/return_interface1_filetest.gno:10:2: cannot use interface{} as uint
