package main

func bar(s ...int) {
	var s1 []int
	s1 = append(s1, s...)
	println(s1)
}
func main() {
	var a uint = 2
	bar(1.0<<a, 2)
}

// Output:
// slice[(4 int),(2 int)]