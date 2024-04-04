package main

type (
	word uint
	nat  []word
)

func (n nat) add() bool {
	return true
}

func main() {
	var abs nat
	abs = []word{0}
	println(abs.add())
}

// Output:
// true
