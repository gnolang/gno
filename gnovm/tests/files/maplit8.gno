package main

import "fmt"

func main() {
    a := map[[]int]int{[]int{1, 2}: 10, []int{2, 3}: 11}
    fmt.Println(a)
}

// Error:
// main/maplit8.gno:6:10-57: invalid map key type []int

// TypeCheckError:
// main/maplit8.gno:6:14: invalid map key type []int
