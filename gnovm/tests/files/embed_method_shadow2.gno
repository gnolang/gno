package main

// Shadowing works through a pointer-embedded field too: Shallow.M at
// depth 1 shadows *Deep.M reachable at depth 2.

type I interface{ M() string }

type Deep struct{}

func (Deep) M() string { return "deep" }

type Mid struct{ *Deep }

type Shallow struct{}

func (Shallow) M() string { return "shallow" }

type T struct {
	Shallow
	Mid
}

func main() {
	var i I = T{Shallow{}, Mid{&Deep{}}}
	println(i.M())
}

// Output:
// shallow
