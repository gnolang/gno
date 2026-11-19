package main

// K is recovered inside a defer of D1 while E is live; the next panic, P,
// links onto E and must find no stale link left from F.
func G() {
	defer func() { panic("P") }()
	defer func() { // D1
		defer func() { println("D1b recovered:", recover()) }()
		defer func() { panic("K") }()
		panic("F")
	}()
	panic("E")
}

func main() {
	defer func() { println("main recovered:", recover()) }()
	G()
}

// Output:
// D1b recovered: K
// main recovered: P
