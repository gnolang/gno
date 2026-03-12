package main

type Getter interface {
	GetB() int
}

type T struct{}

func (T) GetB() int { return B }

var (
	A = Getter(T{}).GetB()
	B = 42
)

func main() { println(A) }

// Output:
// 0
