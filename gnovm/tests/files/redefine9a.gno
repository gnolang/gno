package main

import "fmt"

func main() {
    s := []int{1, 2, 3}
    for _, v := range s {
        func() {
            println(v)
        }()
    }
    fmt.Println("Hello, 世界")
    println("ok")
}

// Output:
// 1
// 2
// 3
// Hello, 世界
// ok
