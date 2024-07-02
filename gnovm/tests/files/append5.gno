package main

func main() {
	var x string
	y := append([]byte{'X'}, x...)
	println(string(y))
}

// Output:
// X
