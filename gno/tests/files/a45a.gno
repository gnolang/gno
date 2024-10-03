package main

import "fmt"

func main() {
	// NOTE: this becomes a .List array
	x := [11]uint16{3, 3, 3, 3, 3, 3, 3, 3, 3, 3, 3}
	fmt.Println(x)
	x[0] += 2
	x[1] -= 2
	x[2] *= 2
	x[3] /= 2
	x[4] %= 2
	x[5] &= 2
	x[6] &^= 2
	x[7] |= 2
	x[8] &= 2
	x[9] <<= 2
	x[10] >>= 2
	fmt.Println(x)
}

// Output:
// [3 3 3 3 3 3 3 3 3 3 3]
// [5 1 6 1 1 2 1 3 2 12 0]
