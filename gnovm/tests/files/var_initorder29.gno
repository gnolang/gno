package main

// Regression: codaInitOrderDeps must track method expressions reached
// through a type alias, e.g. `U.Method` where `type U = T`.

type T int

type U = T

func (r T) Method() int { return a }

var dummy = U.Method(0)

var a = identity(1)

func identity(a int) int { return a }

func main() {
	if dummy != 1 {
		panic("dummy != 1")
	}
	println(dummy)
}

// Output:
// 1
