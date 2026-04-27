// Go semantics: [N]T(s) shallow-copies elements.
// Pointer fields inside elements alias between slice and array.
package main

type Box struct {
	P *int
}

func main() {
	x := 1
	s := []Box{{P: &x}}
	a := [1]Box(s)
	*a[0].P = 99
	println(*s[0].P)
}

// Output:
// 99
