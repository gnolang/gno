package main

import (
	"fmt"
	"time"
)

func main() {
	var a int64 = 2
	fmt.Println(time.Second * a)
}

// Error:
// main/files/time16_native.gno:10:14: incompatible types in binary expression: go:time.Duration MUL int64
