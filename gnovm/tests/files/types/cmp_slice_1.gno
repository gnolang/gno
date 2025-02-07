package main

func main() {
	expected := `hello`[:]
	a := "hello"
	println(a == expected)
}

// Output:
// true
