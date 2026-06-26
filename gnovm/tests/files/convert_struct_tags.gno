package main

type A struct{ X int `a` }
type B struct{ X int `b` }

func main() {
	var b B
	a := A(b)
	println(a.X)
}

// Output:
// 0
