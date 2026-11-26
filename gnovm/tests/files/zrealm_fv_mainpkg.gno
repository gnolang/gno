package main

var count int
var saved func()

type box struct{ n int }

var b = &box{}

func init() { saved = func() { count++; b.n++ } }

// A non-realm program's state is not chain state: a stored function value
// called by value writes it.
func main() {
	g := saved
	g()
	saved()
	println("m01 main package writes", count, b.n)
}

// Output:
// m01 main package writes 2 2
