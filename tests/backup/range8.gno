package main

import (
	"fmt"
	"time"
)

func main() {
	for range time.Tick(time.Millisecond) {
		fmt.Println("success")
		return
	}
}

// Output:
// success
