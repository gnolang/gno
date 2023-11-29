package main

type word uint
type nat []word

// receiver
func (n nat) add() bool {
	return true
}

func Gen() nat {
	n := []word{0}
	return n
}

// mapLit
func main() {
	r := Gen()
	println(r.add())
}

// Output:
// true
