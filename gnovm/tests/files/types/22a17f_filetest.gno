package main

type word uint
type nat []word

// receiver
func (n nat) add() bool {
	return true
}

func Gen2() (nat, nat) {
	n0 := []word{0}
	n1 := []word{1}
	return n0, n1
}

// mapLit
func main() {
	r1, r2 := Gen2()
	println(r1.add())
	println(r2.add())
}

// Output:
// true
// true
