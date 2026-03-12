package main

type T struct{}

func (*T) GetB() int { return B }

var t T

var (
	A = t.GetB()
	B = 42
)

func main() { println(A) }

// Output:
// 42
