// Verifies that a tie at depth N is resolved (not reported as ambiguous)
// when a later sibling has a strictly shallower match at depth < N.
// Exercises the "don't return on tie" branch of the lookup.
package main

type Inner struct {
	Foo string
}

type Mid1 struct {
	Inner
}

type Mid2 struct {
	Inner
}

type Direct struct {
	Foo string
}

type Outer struct {
	Mid1
	Mid2
	Direct
}

func main() {
	o := Outer{
		Mid1:   Mid1{Inner: Inner{Foo: "mid1"}},
		Mid2:   Mid2{Inner: Inner{Foo: "mid2"}},
		Direct: Direct{Foo: "direct"},
	}
	if o.Foo != "direct" {
		panic(o.Foo)
	}
}

// Output:
