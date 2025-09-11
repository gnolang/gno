package main

type (
	nat  []Word
	Word uint
)

func (n nat) add(ws []Word) any {
	println(ws)

	return ws
}

func (n nat) add2(ws nat) any {
	println(ws)

	return ws
}

func main() {
	var abs nat
	abs = []Word{0}

	println(abs.add(abs))
	println()
	println(abs.add2(abs))
}

// Output:
// slice[(0 main.Word)]
// slice[(0 main.Word)]
//
// (slice[(0 main.Word)] main.nat)
// (slice[(0 main.Word)] main.nat)
