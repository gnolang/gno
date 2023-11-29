package main

func printit(x ...byte) {
	println(string(x))
}

func main() {
	x := []byte{}
	x = append(x, "foobar"...)
	println(string(x))
}

// Output:
// foobar
