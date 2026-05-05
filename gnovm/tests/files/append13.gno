package main

import "fmt"

func main() {
	s1 := []int{0, 1, 2}
	r1 := append(s1[:1], s1[:2]...)
	fmt.Println("append overlap:", r1) // [0 0 1]

	s2 := []int{0, 1, 2, 3, 4}
	r2 := append(s2[:1], s2[2:4]...)
	fmt.Println("append overlap forward:", r2) // [0 2 3]

	s3 := []int{0, 1, 2, 3, 4, 5}
	r3 := append(s3[:3], s3[1:4]...)
	fmt.Println("append overlap backward:", r3) // [0 1 2 1 2 3]

	s4 := []byte{0, 1, 2, 3, 4}
	r4 := append(s4[:1], s4[2:4]...)
	fmt.Println("append overlap bytes:", r4) // [0 2 3]
}

// Output:
// append overlap: [0 0 1]
// append overlap forward: [0 2 3]
// append overlap backward: [0 1 2 1 2 3]
// append overlap bytes: [0 2 3]
