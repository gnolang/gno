package main

type (
	nat  []Word
	Word uint
)

func (n nat) add() []Word {
	return nat{0}
}

func main() {
	var abs nat
	abs = []Word{0}
	println(abs)
	println(abs.add())
	// println(abs.add().add()) // it should build failed.
}

// Output:
// (slice[(0 main.Word)] main.nat)
// slice[(0 main.Word)]
