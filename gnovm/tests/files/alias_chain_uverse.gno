package main

// An alias chain that leaves the declaration group: nothing in it has a
// slot to reserve, so each alias is resolved once its target is defined.

type A = B
type B = C
type C = int

type D struct{ a A }

func main() {
	var d D
	d.a = 3
	println(d.a + 1)
}

// Output:
// 4
