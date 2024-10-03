package main

type (
	word int
	nat  []word
)

func (n nat) add() bool {
	return true
}

func main() {
	var abs nat
	var b []word
	b = []word{0}
	abs = b

	println(abs)
}

// Output:
// (slice[(0 main.word)] main.nat)
