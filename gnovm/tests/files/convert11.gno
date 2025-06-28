package main

import "fmt"

// slice -> *array
func main() {
	s0 := "hello"
	s1 := []byte(s0)          // len(s1) == 5
	arrPtr1 := (*[5]byte)(s1) // Convert to *[5]byte
	println("len of s1: ", len(s1))
	println("len of arrPtr1: ", len(*arrPtr1))
	for i := 0; i < len(s1); i++ {
		fmt.Println(string((*arrPtr1)[i]))
	}

	(*arrPtr1)[0] = 'H'

	fmt.Println("modify underlying array...")

	for i := 0; i < len(s1); i++ {
		fmt.Println(string(s1[i]))
	}

	fmt.Println("convert back to slice")
	s11 := (*arrPtr1)[:]
	println("len of s11: ", len(s11))
	for i := 0; i < len(s1); i++ {
		fmt.Println(string(s11[i]))
	}

	fmt.Println("s0: ", s0)

	fmt.Println("=============================")

	s2 := []byte("hellohello") // len(s) == 10
	arrPtr2 := (*[10]byte)(s2) // Convert to *[10]byte
	println("len of s2: ", len(s2))
	println("len of arrPtr2: ", len(*arrPtr2))
	for i := 0; i < len(s2); i++ {
		fmt.Println(string((*arrPtr2)[i]))
	}
}

// Output:
// len of s1:  5
// len of arrPtr1:  8
// h
// e
// l
// l
// o
// modify underlying array...
// H
// e
// l
// l
// o
// convert back to slice
// len of s11:  8
// H
// e
// l
// l
// o
// s0:  hello
// =============================
// len of s2:  10
// len of arrPtr2:  16
// h
// e
// l
// l
// o
// h
// e
// l
// l
// o
