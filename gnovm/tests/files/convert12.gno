package main

import "fmt"

// slice -> array
func main() {
	s := []byte("hello") // len(s) == 5
	arr := ([5]byte)(s)  // Convert to [5]byte
	println("len of slice: ", len(s))
	println("len of array: ", len(arr))
	for i := 0; i < len(s); i++ {
		fmt.Println(string(arr[i]))
	}

	arr[0] = 'H'

	fmt.Println("modify underlying array, print slice...")

	for i := 0; i < len(s); i++ {
		fmt.Println(string(s[i]))
	}

	fmt.Println("printing array...")
	for i := 0; i < len(s); i++ {
		fmt.Println(string(arr[i]))
	}
}

// Output:
// len of slice:  5
// len of array:  8
// h
// e
// l
// l
// o
// modify underlying array, print slice...
// h
// e
// l
// l
// o
// printing array...
// H
// e
// l
// l
// o
