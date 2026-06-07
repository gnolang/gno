//go:build ignore

package main

import (
	"crypto/rand"
	"encoding/binary"
	"flag"
	"fmt"
)

func main() {
	n := flag.Int("n", 781, "how many random numbers to generate")
	flag.Parse()

	buf := make([]byte, 8)
	for i := 0; i < *n; i++ {
		_, err := rand.Read(buf)
		if err != nil {
			panic(err)
		}
		n := binary.LittleEndian.Uint64(buf)
		fmt.Printf("0x%016x,", n)
		if i%4 == 3 {
			fmt.Printf("\n")
		} else {
			fmt.Printf(" ")
		}
	}
	if *n%4 != 0 {
		fmt.Println()
	}
}
