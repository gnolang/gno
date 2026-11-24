package main

// A name declared later in a block by a type or const declaration does not
// shadow an outer variable for the references that come before it.
func main() {
	x := 1
	f := func() {
		a := "secret"
		g := func() { println("g sees", x) }
		g()
		println("f sees", x)
		type x int
		println(a)
	}
	f()
	h := func() {
		println("before const", x)
		const x = 5
		println("after const", x)
	}
	h()
}

// Output:
// g sees 1
// f sees 1
// secret
// before const 1
// after const 5
