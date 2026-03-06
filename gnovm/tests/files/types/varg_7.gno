package main

func main() {
	var a uint = 1
	var s []int
	s = append(s, 1.0<<a)

	println(s)
}


// Output:
// slice[(2 int)]