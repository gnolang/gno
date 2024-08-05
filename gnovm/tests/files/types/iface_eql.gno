package main

func main() {
	var l interface{} = 1
	println(int8(1) == l)
}

// Output:
// false
