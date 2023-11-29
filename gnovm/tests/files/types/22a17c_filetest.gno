package main

type word uint
type nat []word

func Add(n1, n2 nat) bool {
	return true
}

// mapLit
func main() {
	n0 := []word{0}
	n1 := []word{1}

	r := Add(n0, n1)
	println(r)
}

// Output:
// true
