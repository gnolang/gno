package main

type A *int

func main() {
	b := new(int)
	var a A
	a = b

	var c *int
	c = b

	println(areEqual(a, b))
	println(areEqual(b, c))
}

func areEqual(v, w interface{}) bool {
	return v == w
}

// Output:
// false
// true
