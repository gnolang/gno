package main

func main() {
    a := len(42)
    println(a)
}

// Error:
// main/files/len10.gno:4:10: unexpected type for len(): <untyped> bigint

// TypeCheckError:
// main/files/len10.gno:4:14: invalid argument: 42 (untyped int constant) for built-in len