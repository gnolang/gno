package main

func Foo() {
	var other = func() {
		if false {
		} else {
			anything := something
			println(anything)
		}
	}
	other()
}

var something = "a string"

func main() {
	Foo()
}

// Output:
// a string
