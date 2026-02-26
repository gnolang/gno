package main

import "fmt"

func main() {
    s := []int{1, 2, 3}
    for i, v := range s {
        i := i // shadow loop variable
        v := v // shadow loop variable
        func() {
            i += 1
            println(i)
            println(v)
        }()
    }
    fmt.Println("Hello, 世界")
    println("ok")
}

// Output:
// 1
// 1
// 2
// 2
// 3
// 3
// Hello, 世界
// ok
