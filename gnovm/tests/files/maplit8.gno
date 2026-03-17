package main

import "fmt"

func main() {
    a := map[[]int]int{[]int{1, 2}: 10, []int{2, 3}: 11}
    fmt.Println(a)
}

// Error:
// runtime error: slice type cannot be used as map key

// TypeCheckError:
// main/maplit8.gno:6:14: invalid map key type []int
