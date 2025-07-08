package main

type nat []int

func x() any {
	a := nat{0}
	return a
}

func main() {
	var a nat

	a = x()

	println(a)
}

// Error:
// main/return_interface.gno:13:2-9: cannot use interface {} as []int

// TypeCheckError:
// main/return_interface.gno:13:6: cannot use x() (value of interface type any) as nat value in assignment: need type assertion
