package main

func main() {
	s := []int{1, 2, 3}
	s[0] = 0
	println(s)
}

// Output:
// slice[(0 int),(2 int),(3 int)]
