package main

import (
	"crypto/chacha20/rand"
	"fmt"
	"io"
)

func main() {
	var buf [16]byte
	fmt.Println(buf)
	randr := rand.NewFromSeed([]byte("someseedsomeseedsomeseedsomeseed"))
	io.ReadFull(randr, buf[:])
	fmt.Println(buf)
}

// Output:
// [0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0]
// [97 144 241 136 231 32 178 3 204 43 206 42 168 170 181 215]
