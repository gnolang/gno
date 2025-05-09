package main

import "fmt"

func main() {
    a := make([]byte, 1.0, 2.0)
    fmt.Println(a)
}

// Output:
// [0]