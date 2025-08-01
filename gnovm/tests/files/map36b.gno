package main

import (
	"fmt"
	"math"
)

type S struct {
	F, G float64
}

func main() {
	n0 := math.Float64frombits(1 << 63)
	sn := S{n0, 0}
	sp := S{0, n0}
	fmt.Println(sn == sp)
	m := map[S]string{}
	m[sp] = "hey"
	fmt.Println(m)
	m[sn] = "world"
	fmt.Println(m)
	m[sp] = "sup"
	fmt.Println(m)
}

// Output:
// true
// map[{0 -0}:hey]
// map[{-0 0}:world]
// map[{0 -0}:sup]
