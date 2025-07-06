package main

import (
	"fmt"
	"math"
)

var NaN = math.NaN()

func main() {
	fmt.Println(map[float64]int{NaN: 1, NaN: 2})
}

// Output:
// map[NaN:1 NaN:2]
