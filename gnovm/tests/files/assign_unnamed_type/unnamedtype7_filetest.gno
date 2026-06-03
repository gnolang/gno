package main

type myptr *int

// *int is unnamed
func main() {
	var n myptr = nil
	var u *int = nil
	n = u

	println(n)
}

// Output:
// (nil main.myptr)
