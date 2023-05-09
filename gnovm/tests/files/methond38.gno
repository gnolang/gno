package main

 import (
 	"fmt"
 )

 type word uint
 type nat []word

 func (n nat) add() bool {
 	return true
 }

 func main() {
 	var abs nat
 	abs = []word{0}
 	println(abs.add())
 }
