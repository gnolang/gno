// run

package main

// Regression: codaInitOrderDeps must descend into closure bodies so that
// method expressions hidden inside a func literal still contribute to the
// init order graph.

type T int

func (r T) Method() int { return a }

var fn = func() int { return T.Method(0) }

var dummy = fn()

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
