package main

type (
	word uint
	nat  []word
)

func (n nat) add(x, y nat) bool {
	return true
}

// parameter
func main() {
	var abs nat
	abs = []word{0}
	x := []word{1}
	y := []word{2}

	println(abs.add(x, y))
}

// Output:
// true
