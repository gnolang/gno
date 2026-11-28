package main

// Companion to maptype2.gno: after the map-key check has deferred past
// incomplete nested DeclaredTypes, == on S must still see a comparable
// struct (not a memo poisoned by an earlier nil-Base walk).

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
