// https://github.com/gnolang/gno/issues/5957 (review follow-up)
// An alias of a pointer type resolves to the *T spelling and is a
// legal embedded field, unlike a defined pointer type.
package main

type D2 struct{ A int }

type P = *D2 // alias of a pointer type: legal as an embedded field

type S struct{ P }

func main() {
	var s S
	s.P = &D2{A: 3}
	println(s.A)
}

// Output:
// 3
