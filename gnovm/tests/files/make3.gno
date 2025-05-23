package main

func main() {
    a := make(1)
    println(a)
}

// Error:
// main/files/make3.gno:4:10: (const (1 <untyped> bigint)) is not a type

// TypeCheckError:
// main/files/make3.gno:4:15: 1 is not a type