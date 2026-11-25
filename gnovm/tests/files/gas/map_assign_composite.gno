// MAXALLOC: 500000000
package main

// Assigns a [1000]int to ten new map keys. Only the assignment's copy pays
// per element: the zero value a new key used to get first was filled, charged
// and overwritten at once.

func main() {
	m := map[int][1000]int{}
	var v [1000]int
	for i := 0; i < 10; i++ {
		m[i] = v
	}
	println(len(m))
}

// Output:
// 10

// Gas:
// 585064
