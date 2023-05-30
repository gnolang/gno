package main

 type word uint
 type nat []word

 func (n nat) add() bool {
 	return true
 }

 // parameter
 func main() {
 	des := make([]nat, 2)
 	src := make([][]word, 2)

 	src[0] = []word{0}

 	copy(des, src)
 }

// Error:
// main/files/declaredType7.gno:17: arguments to copy have different element types, want main.nat, got []main.word