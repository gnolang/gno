package main

func main() {
	expected := `hello`[:] // this kind of untyped non const will be converted

	a := "hello"

	println(a == expected)

}

// Output:
// true
