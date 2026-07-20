// Positive companion to method45: the ambiguity of the promoted o.Foo
// (two Foos at depth 2) does not poison other lookups on the same type.
// A name that is unique at its shallowest depth still resolves however
// deep it sits (o.Inner3, depth 2 via DeepBox), and anchoring even one
// hop past the fork removes the ambiguity entirely (o.DeepBox.Foo,
// o.AmbBox.M1.Foo). Verified against Go: all three are legal while
// o.Foo is an ambiguous selector.
package main

type M1 struct{ Foo string }
type M2 struct{ Foo string }

type AmbBox struct {
	M1
	M2
}

type Inner3 struct{ Foo string }
type Inner2 struct{ Inner3 }

type DeepBox struct {
	Inner2
}

type Outer struct {
	AmbBox
	DeepBox
}

func main() {
	o := Outer{
		AmbBox:  AmbBox{M1{Foo: "m1"}, M2{Foo: "m2"}},
		DeepBox: DeepBox{Inner2{Inner3{Foo: "deep"}}},
	}
	println(o.Inner3.Foo)
	println(o.DeepBox.Foo)
	println(o.AmbBox.M1.Foo)
}

// Output:
// deep
// deep
// m1
