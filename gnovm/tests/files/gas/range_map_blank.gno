// MAXALLOC: 500000000
package main

// Ranges over a map of three [1000]int values with "_" for the value. Nothing
// reads "_", so no value is loaded or copied, as with `for k := range m`.

func main() {
	m := map[int][1000]int{1: {}, 2: {}, 3: {}}
	n := 0
	for k, _ := range m {
		n += k
	}
	println(n)
}

// Output:
// 6

// Gas:
// 293649
