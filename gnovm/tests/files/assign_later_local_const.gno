package main

// A const or type declared later in a block does not make an outer variable
// of the same name unassignable for the statements that come before it.
func main() {
	x := 1
	f := func() {
		x = 2
		const x = 5
		println("f", x)
	}
	f()
	println("after f", x)
	{
		x++
		for x = range []int{0, 1, 2, 3} {
		}
		const x = 6
		println("block", x)
	}
	println("after block", x)
	g := func() {
		x = 4
		type x int
		var y x = 7
		println("g", int(y))
	}
	g()
	println("after g", x)
}

// Output:
// f 5
// after f 2
// block 6
// after block 3
// g 7
// after g 4
