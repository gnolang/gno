package main

import "io"

type (
	s [2]int
	m s
)

func main() {
	const v = len(m{1, 2})
	println(v)
}

// Output:
// 2
