package main

type Getter interface {
	Get() int
}

type T struct{}

func (T) Get() int { return B }

var (
	A = Getter(T{}).Get()
	B = Getter(T{}).Get()
)

func main() { println(A) }

// Output:
// 0
