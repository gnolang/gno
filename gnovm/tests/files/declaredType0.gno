package main

 type word uint
 type nat []word

 func (n nat) add() bool {
 	return true
 }

 func main() {
 	var abs nat
 	abs = []word{0} // unname-Composit to named DeclaredTypes
 	println(abs.add())
 }
