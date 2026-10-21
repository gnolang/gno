package main

// Cyclic type declarations that mix aliases with a defined type are valid
// Go as long as the cycle also passes through an indirection: the
// indirection keeps the type finite, and the defined type gives the alias
// chain something concrete to resolve to. See golang/go#25838.

type (
	e = f
	f = g
	g = []h
	h i
	i = j
	j = e
)

type (
	e1 = []h1
	h1 e1
)

type (
	P = *T
	T P
)

func main() {
	var x e
	x = append(x, nil, nil)
	println(len(x))

	var y e1
	y = append(y, nil)
	println(len(y))

	var t T
	println(t == nil)
}

// Output:
// 2
// 1
// true
