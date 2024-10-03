package main

type (
	nat  []Word
	Word uint
)

func main() {
	var abs []Word
	abs = nat{0} // abs is a not named array
	println(abs)
	println()
	a := []Word{0}
	b := nat{0}
	c := a
	d := b
	e := abs

	println(a)
	println(b)
	println(c)
	println(d)
	println(e)
}

// Output:
// slice[(0 main.Word)]
//
// slice[(0 main.Word)]
// (slice[(0 main.Word)] main.nat)
// slice[(0 main.Word)]
// (slice[(0 main.Word)] main.nat)
// slice[(0 main.Word)]
