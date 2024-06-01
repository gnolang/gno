package main

type (
	word uint
	nat  []word
)

func (n nat) add() bool {
	return true
}

// mapLit
func main() {
	n := []word{0}

	items := map[string]nat{
		"test": n,
	}

	r := items["test"]

	println(r.add())
}

// Output:
// true
