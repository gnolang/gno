package main

// Valid Go: S is comparable because the cycle reaches a pointer through
// X = [2]*Z. Nothing may call isComparable on S while a nested
// DeclaredType in its graph still has a nil Base: that would both
// false-reject the map and poison StructType.comparable on S.

type D X
type X [2]*Z
type Z map[S]int
type S struct{ d D }

func main() {
	m := Z{}
	println(len(m))
}

// Output:
// 0
