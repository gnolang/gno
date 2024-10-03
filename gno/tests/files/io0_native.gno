package main

import (
	"crypto/rand"
	"fmt"
	"io"
)

func main() {
	var buf [16]byte
	fmt.Println(buf)
	io.ReadFull(rand.Reader, buf[:])
	fmt.Println(buf)
}

// Output:
// [0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0]
// [100 101 102 103 104 105 106 107 108 109 110 111 112 113 114 115]
