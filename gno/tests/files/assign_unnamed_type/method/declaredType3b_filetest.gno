package main

type (
	word uint
	nat  []word
)

func (n nat) add() bool {
	return true
}

// sliceLit2
func main() {
	items := []nat{1: []word{0}, 2: []word{1}}

	r := items[1]

	println(r.add())
}

// Output:
// true
