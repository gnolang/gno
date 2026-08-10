package main

import "math"

// NaN keys are never findable, and must not disturb neighbouring entries.
// Exercises the isNaN path in both the lookup and the lazy index build.

func main() {
	m := map[float64]string{}
	nan := math.NaN()
	m[nan] = "nan"
	m[1.5] = "one-point-five"
	println(len(m))
	_, ok := m[nan]
	println(ok)
	println(m[1.5])

	type wrap struct{ f float64 }
	w := map[wrap]string{}
	w[wrap{nan}] = "nan"
	w[wrap{2.5}] = "two-point-five"
	println(len(w))
	_, ok2 := w[wrap{nan}]
	println(ok2)
	println(w[wrap{2.5}])
}

// Output:
// 2
// false
// one-point-five
// 2
// false
// two-point-five
