package main

type (
	word uint
	nat  []word
)

func (n nat) add() bool {
	return true
}

// sliceLit
func main() {
	items := []nat{[]word{0}, []word{1}}

	r := items[0]

	println(r.add())
}

// Output:
// true
