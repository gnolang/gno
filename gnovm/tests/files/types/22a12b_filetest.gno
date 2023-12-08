package main

type word uint
type nat []word

func (n nat) add() bool {
	return true
}

// TODO: improve err, like :s.add undefined (type []word has no field or method add)
func main() {
	var abs nat
	abs = []word{0} // unname-Composit to named DeclaredTypes
	println(abs.add())

	var s []word
	s = abs
	println(s.add())
}

// Error:
// main/files/types/22a12b_filetest.gno:18: unexpected selector expression type *gnolang.SliceType
