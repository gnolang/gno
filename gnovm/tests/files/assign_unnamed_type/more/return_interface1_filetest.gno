package main

func x1() any {
	a := "1"
	return a
}

func main() {
	var a uint
	a = x1() // should fail
	println("1")
}

// Error:
// main/return_interface1.gno:10:2-10: cannot use interface {} as uint

// TypeCheckError:
// main/return_interface1.gno:10:6: cannot use x1() (value of interface type any) as uint value in assignment: need type assertion; main/return_interface1.gno:9:6: declared and not used: a
