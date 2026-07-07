package main

// Shadowing applies across more than one level of depth: a method at
// depth 1 shadows the same-named method reachable at depth 3.

type I interface{ M() string }

type L3 struct{}

func (L3) M() string { return "l3" }

type L2 struct{ L3 }

type L1 struct{ L2 }

type S struct{}

func (S) M() string { return "s" }

type T struct {
	S  // M at depth 1
	L1 // M at depth 3 (L1.L2.L3.M)
}

func main() {
	var i I = T{}
	println(i.M())
}

// Output:
// s
