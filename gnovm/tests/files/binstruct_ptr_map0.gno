package main

import (
	"fmt"
)

type Point struct{ X, Y int }

func main() {
	v := map[string]*Point{
		"foo": {X: 3, Y: 2},
		"bar": {X: 4, Y: 5},
	}
	fmt.Println(v["foo"], v["bar"])
}

// Output:
// &{3 2} &{4 5}
