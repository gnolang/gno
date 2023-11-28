package main

func main() {
	a := append([]byte(nil), "foobar"...)
	println(string(a))
}

// Output:
// foobar
