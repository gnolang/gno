package main

type S struct {
	expected string
}

func main() {
	expected := `hello`[:] // this kind of untyped non const will be converted
	a := "hello"

	a == expected

	//println(r)

}

// Output:
