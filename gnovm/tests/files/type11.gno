package main

import (
	"fmt"
	"time"
)

const i1 = int(time.Nanosecond)

var weirdArray [i1 + len("123456789")]time.Duration

func main() {
	fmt.Printf("%T\n", weirdArray)
}

// Output:
// [10]time.Duration
