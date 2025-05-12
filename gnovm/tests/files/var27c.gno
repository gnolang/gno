package main

var other = func() {
	something := 2
	if false {
	} else {
		anything := something
		println(anything)
	}
}

var something = "a string"

func main() {
	other()
}

// Output:
// 2
