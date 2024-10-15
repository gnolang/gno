package main

type (
	word uint
	nat  []word
)

func (n nat) add() bool {
	return true
}

// map
func main() {
	items := map[string]nat{}

	n := []word{0}

	// this is assignment
	items["test"] = n

	r := items["test"]

	println(r.add())
}

// Output:
// true
