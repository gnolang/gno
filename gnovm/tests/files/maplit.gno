package main

import "fmt"

func main() {
    a := map[int]int{0: 10, 0: 20}
    fmt.Println(a)
}

// Error:
// main/maplit.gno:6:10-35: duplicate key (0 int) in map literal

// TypeCheckError:
// main/maplit.gno:6:29: duplicate key 0 in map literal
