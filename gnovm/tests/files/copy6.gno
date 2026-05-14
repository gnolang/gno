package main

import "fmt"

func main() {
	s1 := []int{1, 2, 3}
	copy(s1[1:], s1[0:2])
	fmt.Println("copy overlap:", s1)
}

// Output:
// copy overlap: [1 1 2]
