package main

import "math"

type T struct {
	_ int
	X int
}

type F struct {
	_ float64
	X int
}

func main() {
	m := map[T]string{T{1, 2}: "a"}
	println(m[T{3, 2}])

	m[T{3, 2}] = "b"
	println(len(m))
	println(m[T{1, 2}])

	delete(m, T{4, 2})
	println(len(m))

	f := map[F]string{F{math.NaN(), 2}: "nan"}
	println(f[F{math.NaN(), 2}])
	println(len(f))
}

// Output:
// a
// 1
// b
// 0
// nan
// 1
