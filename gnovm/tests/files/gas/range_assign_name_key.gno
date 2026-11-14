// MAXALLOC: 500000000
package main

// Ranges into existing variables, `for k, v = range x`, over a slice and a
// map of [100]int elements. A key target that is a name cannot reach the
// element, so the element is copied once, by the value's write, as it was
// before range assignments converted their values.
type S struct {
	N int
	A [100]int
}

func main() {
	s := make([]S, 10)
	var i int
	var v S
	for i, v = range s {
	}
	m := map[int]S{}
	for j := 0; j < 10; j++ {
		m[j] = S{}
	}
	var k int
	for k, v = range m {
	}
	for i = range s {
	}
	println(i, k, v.N)
}

// Output:
// 9 9 0

// Gas:
// 303895
