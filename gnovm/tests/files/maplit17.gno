// A typed nil pointer key is not a Go constant either: two of them are
// not a duplicate-key error, they are the same key at run time.
package main

func main() {
	m := map[any]int{(*int)(nil): 1, (*int)(nil): 2}
	n := map[*int]int{nil: 1, nil: 2}
	println(len(m), len(n))
}

// Output:
// 1 1
