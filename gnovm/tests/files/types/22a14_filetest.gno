package main

type word uint
type nat []word

func (n nat) add() bool {
	return true
}

// parameter
// TODO: should fix in another place, this should not work
func main() {
	des := make([]nat, 2)
	src := make([][]word, 2)

	src[0] = []word{0}

	copy(des, src)
}

// Error:
