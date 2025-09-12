package main

import (
	"fmt"
	"math"
)

func main() {
	n0 := math.Float64frombits(1 << 63)
	fmt.Println(n0)
	m := map[float64]string{}
	m[0] = "hey"
	m[n0] = "world"
	fmt.Println(m)
}

// Output:
// -0
// map[-0:world]
