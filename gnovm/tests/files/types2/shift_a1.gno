package main

// both typed(different) const
func main() {
	println(int(1) << int8(1))
	println(int(1) >> int8(1))
}

// Output:
// 2
// 0
