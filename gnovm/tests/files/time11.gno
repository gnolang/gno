package main

import (
	"fmt"
	"time"
)

const df = time.Minute * 30

func main() {
	println(df)
	fmt.Printf("df: %v %T\n", df, df)
}

// Output:
// 30m0s
// df: 30m0s time.Duration
