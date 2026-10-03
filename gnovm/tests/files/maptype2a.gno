package main

// Same as maptype2.gno, but X is a type alias.

type D X
type X = [2]*Z
type Z map[S]int
type S struct{ d D }

func main() {
	m := Z{}
	println(len(m))
}

// Output:
// 0
