package main

func Search(n int, f func(int) bool) int {
	i, j := 0, n
	for i := 0; i < j; i++ {
		h := int(uint(i+j) >> 1) // avoid overflow when computing h
		f(h)
	}
	return i
}

func main() {
	for size := 0; size < 2; size++ {
		for targ := 0; targ <= size; targ++ {
			Search(size, func(i int) bool { return i >= targ })
		}
	}
}

// Output:
