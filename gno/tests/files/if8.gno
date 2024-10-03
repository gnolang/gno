package main

func main() {
	v := T
	if v {
		println("true")
	}
}

type C bool

const (
	F C = false
	T C = true
)

// Output:
// true
