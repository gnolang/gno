package main

import (
	"fmt"
	"math"
)

var NaN = float32(math.NaN())

func main() {
	fmt.Println(map[float32]int{NaN: 1, NaN: 2})
}

// Output:
// map[NaN:1 NaN:2]
