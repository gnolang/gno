// MAXALLOC: 5000000
package main

type Boom struct {
	A int
	B int
	C int
	D int
}


func main() {
	keyMap := Boom{}
	m := map[any]int{}
	m[keyMap] = 1
}


// Gas:
// 6025
