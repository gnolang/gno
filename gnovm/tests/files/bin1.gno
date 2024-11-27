package main

import (
	"crypto/sha256"
	"fmt"
)

func main() {
	a := sha256.Sum256([]byte("password"))
	fmt.Println(a)
}

// Output:
// [94 136 72 152 218 40 4 113 81 208 229 111 141 198 41 39 115 96 61 13 106 171 189 214 42 17 239 114 29 21 66 216]
