package main

// Valid Go: S is comparable because the cycle reaches a pointer through
// X = [2]*Z. assertValidMapKeys must wait until nested DeclaredTypes in
// the key graph are filled, otherwise isComparable sees a nil Base and
// false-rejects the map (and poisons StructType.comparable on S).

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
