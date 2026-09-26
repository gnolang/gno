package main

// The same alias chain as recursive12.gno, but closed with an array rather
// than a slice. An array holds its elements inline, so it is not an
// indirection and the cycle stays infinitely sized.

type (
	e = f
	f = g
	g = [3]h
	h i
	i = j
	j = e
)

func main() {}

// Error:
// main/recursive9j.gno:10:2-10: invalid recursive type: h refers to itself

// TypeCheckError:
// main/recursive9j.gno:11:2: invalid recursive type: h refers to itself
