package main

import "fmt"

func main() {
	bs := []byte("hello")
	bs[0] = 'y'
	fmt.Println(bs[0])
	fmt.Printf("%c\n", bs[0])
}

// Output:
// 121
// y
