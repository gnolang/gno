package main

import "fmt"

func main() {
    a := make([]byte, 1.0, 2.2)
    fmt.Println(a)
}

// Error:
// main/files/make7.gno:6:10: invalid capacity argument for built-in make: 2.2 (untyped float constant) truncated to int