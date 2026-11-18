package main

// Companion to maptype2.gno: == on S must see a comparable struct, not a
// memo poisoned by an isComparable call made while nested DeclaredTypes
// were still incomplete.

type D X
type X [2]*Z
type Z map[S]int
type S struct{ d D }

func main() {
	var a, b S
	println(a == b)
}

// Output:
// true
