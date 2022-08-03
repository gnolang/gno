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
// df: 1800000000000 int64
