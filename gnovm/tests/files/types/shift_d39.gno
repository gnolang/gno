package main

import "fmt"

func main() {
	x := 1
	s := []int{1 << x}

	v := float32(s[0]) // type determined when indexed
	println(v)
	fmt.Printf("%T \n", v)
}

// Output:
// 2
// float32
