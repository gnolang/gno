package main

type (
	nat  []Word
	Word uint
)

func (n nat) add() nat {
	return []Word{0}
}

func main() {
	var abs nat
	abs = []Word{0}
	println(abs)
	println(abs.add().add())
}

// Output:
// (slice[(0 main.Word)] main.nat)
// (slice[(0 main.Word)] main.nat)

