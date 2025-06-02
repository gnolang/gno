package main

import (
	"fmt"
)

type myInt struct {
	num int
}

type stringMap map[myInt]int

func main() {
	var (
		accounts = stringMap{}
		tmpInt   = myInt{1}
	)
	accounts[tmpInt] = 1

	tmpInt.num = 2
	accounts[tmpInt] = 1
	for k, v := range accounts {
		fmt.Printf("k: %v, v: %v\n", k, v)
	}
}

// Output:
// k: {1}, v: 1
// k: {2}, v: 1
