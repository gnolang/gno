package main

// Tests promoted pointer-receiver method through embedding.
// Outer embeds *Inner (pointer embed), Inner has a pointer receiver method.

type Inner struct{}

func (*Inner) GetB() int { return B }

type Outer struct{ *Inner }

var (
	A = Outer{Inner: &Inner{}}.GetB()
	B = 42
)

func main() { println(A) }

// Output:
// 42
