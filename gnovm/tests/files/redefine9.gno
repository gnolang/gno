package main

import "fmt"

func main() {
    for i := 0; i < 3; i++ {
        i := i // shadow loop variable
        func() {
            i += 1
            println(i)
        }()
    }
    fmt.Println("Hello, 世界")
}

// Output:
// 1
// 2
// 3
// Hello, 世界
