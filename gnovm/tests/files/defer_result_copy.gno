package main

type Inner struct{ X int }

type Outer struct{ I Inner }

var g = Outer{I: Inner{X: 1}}

// r holds a copy of g, so the deferred write must not reach g.
func f() (r Outer) {
	defer func() { r.I.X = 2 }()
	return g
}

func main() {
	println(f().I.X, g.I.X)
}

// Output:
// 2 1
