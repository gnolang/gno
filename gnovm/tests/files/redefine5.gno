package main

func main() {
	a := 1
	println(a) // prints 1

	if true {
		a := 2     // valid: new scope inside the if statement
		println(a) // prints 2
	}

	println(a) // prints 1: outer variable is unchanged
}

// Output:
// 1
// 2
// 1
