package main

type (
	word uint
	nat  []word
)

func (n nat) add() bool {
	return true
}

// ArrayLit
func main() {
	items := [3]nat{[]word{0}, []word{1}}

	r := items[0]

	println(r.add())
}

// Output:
// true
