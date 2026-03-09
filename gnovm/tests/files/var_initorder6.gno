package main

// Tests that a multi-name spec initialized by a multi-return function is
// treated as one atomic step.  X and Y are ONE ValueDecl; Z must wait for
// both even though it could theoretically proceed after X alone.

func pair() (int, int) { return 10 + incr(), 20 + incr() }

var counter int

func incr() int {
	counter++
	return counter
}

var (
	X, Y = pair()
	Z    = X + incr()
)

func main() { println(X, Y, Z) }

// Output:
// 11 22 14
