package main

import (
	"encoding/base32"
	"fmt"
)

func main() {
	data := []byte("any + old & data")
	str := base32.StdEncoding.EncodeToString(data)
	fmt.Println(str)
}

// Output:
// MFXHSIBLEBXWYZBAEYQGIYLUME======