package main

import (
	"fmt"
	"math"
)

var NaN = math.NaN()

type X struct {
	S string
	F float64
}

func main() {
	fmt.Println(map[X]int{
		{S: "hey", F: NaN}: 1,
		{S: "hey", F: NaN}: 2,
	})
}

// Output:
// map[{hey NaN}:1 {hey NaN}:2]
