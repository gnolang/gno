package main

import "fmt"

func main() {
    for i := 0; i < 3; i++ {
        i := i // shadow loop variable
        func() {
            i += 1
        }()
    }
    fmt.Println("Hello, 世界")
}

// Error:
// unexpected NameExpr type for GetPointerToMaybeHeapDefine
