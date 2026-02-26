package main

import "fmt"

func main() {
    s := []int{1, 2, 3}
    for i, _ := range s {
        func() {
            println(i)
        }()
    }
    fmt.Println("Hello, 世界")
    println("ok")
}

// Output:
// 0
// 1
// 2
// Hello, 世界
// ok
