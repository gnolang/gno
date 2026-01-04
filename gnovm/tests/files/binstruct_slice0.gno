package main

import (
	"fmt"
)

type Point struct{ X, Y int }

func main() {
	v := []Point{
		{X: 3, Y: 2},
	}
	fmt.Println(v)
}

// Output:
// [{3 2}]
