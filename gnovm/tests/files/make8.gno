package main

import "fmt"

func main() {
    a := make([]byte, 1.5, 2.0)
    fmt.Println(a)
}

// Error:
// main/files/make8.gno:6:10: invalid length argument for built-in make: 1.5 (untyped float constant) truncated to int

// TypeCheckError:
// main/files/make8.gno:6:23: 1.5 (untyped float constant) truncated to int