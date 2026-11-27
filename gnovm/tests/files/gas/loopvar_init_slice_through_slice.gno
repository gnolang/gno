// MAXALLOC: 500000000
package main

// Slicing an array that sits behind a slice of the loop's init variable
// aliases the slice's backing array, not the variable, so the variable stays
// in its block slot and its [1000]int is not copied per iteration.

type S struct {
	big [1000]int
	sl  [][2]int
}

func main() {
	n := 0
	for s := (S{sl: make([][2]int, 1)}); s.big[0] < 10; s.big[0]++ {
		n += len(s.sl[0][:])
	}
	println(n)
}

// Output:
// 20

// Gas:
// 162523
