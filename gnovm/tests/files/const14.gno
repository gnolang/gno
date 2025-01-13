package main

import "math"

func f1(i float64) { println("i:", i) }

func main() {
	i := math.Pi
	f1(i)
}

// Output:
// i: 3.141592653589793
