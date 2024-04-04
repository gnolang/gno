package main

type (
	word uint
	nat  []word
)

func (n nat) add() bool {
	return true
}

// slice append
func main() {
	var items []nat

	n := []word{0}

	items = append(items, n)

	r := items[0]

	println(r.add())
}

// Output:
// true
