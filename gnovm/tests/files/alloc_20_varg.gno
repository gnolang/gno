// MAXALLOC: 1000000
package main

// Regression: an allocator buffer under construction must stay counted.
//
// popCopyArgs' variadic branch has the same hole twice: the varg `list`
// buffer, and the `args` buffer it is then stored into.
type Big [300000]byte

var a Big

func f(xs ...Big) int { return len(xs) }

func main() {
	println(f(a, a, a, a, a, a, a, a, a, a))
}

// Error:
// allocation limit exceeded
