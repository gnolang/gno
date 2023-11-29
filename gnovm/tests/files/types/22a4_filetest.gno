package main

type word uint
type nat []word

// receiver
func (n nat) add() bool {
	return true
}

// map
func main() {
	items := map[string]nat{}

	n := []word{0}

	items["test"] = n

	r := items["test"]

	println(r.add())
}

// Output:
// true
