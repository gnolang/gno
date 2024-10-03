package main

func main() {
	c := []byte{'A'}
	a := c
	println(&c[0])
	println(&c[0] == &a[0])
}

// Output:
// &(65 uint8)
// true
