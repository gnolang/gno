// MAXALLOC: 50000000
package main

type Boom struct {
	Data [4000]byte
}

type OuterBoom struct {
	A    [2000]Boom
	B    [2000]Boom
	C    [2000]Boom
	D    [2000]Boom
	E    [2000]Boom
}

func main() {
	keyMap := OuterBoom{}
	m := map[any]int{}
	m[keyMap] = 1
}


// Gas:
// 128371131
